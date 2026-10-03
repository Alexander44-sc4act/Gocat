package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	// Test invalid port
	err := connect("127.0.0.1", "99999", "/bin/sh")
	if err == nil {
		t.Error("Expected error for invalid port")
	}
}

func TestDialWithOptions(t *testing.T) {
	// Test with invalid address
	_, err := dialWithOptions("tcp", "invalid:99999")
	if err == nil {
		t.Error("Expected error for invalid address")
	}
}

func TestHostPortParsing(t *testing.T) {
	tests := []struct {
		args    []string
		expHost string
		expPort string
	}{
		{[]string{"8080"}, "127.0.0.1", "8080"},
		{[]string{"192.168.1.1", "9090"}, "192.168.1.1", "9090"},
	}

	for _, test := range tests {
		var host, port string
		if len(test.args) == 1 {
			host = "127.0.0.1"
			port = test.args[0]
		} else {
			host = test.args[0]
			port = test.args[1]
		}

		if host != test.expHost || port != test.expPort {
			t.Errorf("Expected %s:%s, got %s:%s", test.expHost, test.expPort, host, port)
		}
	}
}

func TestRetryLogic(t *testing.T) {
	// Save original values
	origRetryCount := retryCount
	origTimeout := timeout

	// Set test values
	retryCount = 2
	timeout = 100 * time.Millisecond

	// Restore original values
	defer func() {
		retryCount = origRetryCount
		timeout = origTimeout
	}()

	// Test connection to non-existent service
	start := time.Now()
	err := connect("127.0.0.1", "12345", "/bin/sh")
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected connection to fail")
	}

	// Should have taken at least the timeout duration * retry attempts
	expectedMinDuration := timeout * time.Duration(retryCount+1)
	if duration < expectedMinDuration {
		t.Errorf("Expected at least %v, got %v", expectedMinDuration, duration)
	}
}

func TestZeroIO(t *testing.T) {
	// Start a test server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// Track if server receives any data
	dataReceived := false
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Try to read with timeout
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		if n > 0 {
			dataReceived = true
		}
	}()

	// Enable zero-IO mode
	origZeroIO := useZeroIO
	useZeroIO = true
	defer func() { useZeroIO = origZeroIO }()

	// Test connection
	addr := listener.Addr().(*net.TCPAddr)
	err = connect("127.0.0.1", fmt.Sprintf("%d", addr.Port), "/bin/sh")

	if err != nil {
		t.Errorf("Zero-IO connection failed: %v", err)
	}

	// Wait a bit to ensure no data was sent
	time.Sleep(600 * time.Millisecond)

	if dataReceived {
		t.Error("Zero-IO mode should not transfer data")
	}
}

func TestCRLFWriter(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	writer := &crlfWriter{writer: &buf}

	// Write data with LF
	input := []byte("line1\nline2\nline3\n")
	n, err := writer.Write(input)

	if err != nil {
		t.Fatalf("CRLF writer error: %v", err)
	}

	if n != len(input) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(input), n)
	}

	// Check output has CRLF
	output := buf.Bytes()
	expected := []byte("line1\r\nline2\r\nline3\r\n")

	if !bytes.Equal(output, expected) {
		t.Errorf("CRLF conversion failed.\nExpected: %v\nGot: %v", expected, output)
	}
}

func TestTelnetConn(t *testing.T) {
	// Create a mock telnet server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// Server goroutine that sends telnet commands
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Send telnet DO command (IAC DO ECHO)
		telnetCmd := []byte{telnetIAC, telnetDO, 0x01}
		conn.Write(telnetCmd)

		// Send some normal data
		conn.Write([]byte("Hello"))

		// Send IAC IAC (escaped IAC)
		conn.Write([]byte{telnetIAC, telnetIAC})

		// Wait for response
		buf := make([]byte, 100)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		conn.Read(buf)
	}()

	// Connect to mock server
	addr := listener.Addr().(*net.TCPAddr)
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Wrap with telnet handler
	telnet := newTelnetConn(conn)

	// Read data - may need multiple reads due to packet fragmentation
	var received []byte
	buf := make([]byte, 100)
	telnet.SetReadDeadline(time.Now().Add(2 * time.Second))

	// Read multiple times to get all data
	for len(received) < 6 { // Expecting 5 bytes ("Hello") + 1 byte (IAC)
		n, err := telnet.Read(buf)
		if err != nil && err != io.EOF {
			if len(received) > 0 {
				break // Got some data, that's okay
			}
			t.Fatalf("Telnet read error: %v", err)
		}
		if n > 0 {
			received = append(received, buf[:n]...)
		}
		if n == 0 && err == io.EOF {
			break
		}
		time.Sleep(50 * time.Millisecond) // Small delay for network
	}

	// Should receive "Hello" + single IAC (0xFF)
	expected := []byte("Hello")
	expected = append(expected, telnetIAC)

	if !bytes.Equal(received, expected) {
		t.Errorf("Telnet filtering failed.\nExpected: %v\nGot: %v", expected, received)
	}
}

func TestTelnetConnWrite(t *testing.T) {
	// Create a buffer connection
	var buf bytes.Buffer
	mockConn := &mockNetConn{buffer: &buf}

	telnet := newTelnetConn(mockConn)

	// Write data containing IAC
	input := []byte{0x48, 0x65, telnetIAC, 0x6C, 0x6F} // "He" + IAC + "lo"
	n, err := telnet.Write(input)

	if err != nil {
		t.Fatalf("Telnet write error: %v", err)
	}

	if n != len(input) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(input), n)
	}

	// IAC should be escaped (doubled)
	expected := []byte{0x48, 0x65, telnetIAC, telnetIAC, 0x6C, 0x6F}
	output := buf.Bytes()

	if !bytes.Equal(output, expected) {
		t.Errorf("Telnet IAC escaping failed.\nExpected: %v\nGot: %v", expected, output)
	}
}

// Mock net.Conn for testing telnet write
type mockNetConn struct {
	buffer *bytes.Buffer
	net.Conn
}

func (m *mockNetConn) Write(b []byte) (n int, err error) {
	return m.buffer.Write(b)
}

func (m *mockNetConn) Read(b []byte) (n int, err error) {
	return m.buffer.Read(b)
}

func (m *mockNetConn) Close() error {
	return nil
}

func (m *mockNetConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockNetConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func BenchmarkConnect(b *testing.B) {

	// Start a test server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			b.Logf("Error closing listener: %v", err)
		}
	}()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			if err := conn.Close(); err != nil {
				b.Logf("Error closing connection: %v", err)
			}
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Just test the dial part, not the shell execution
		conn, err := dialWithOptions("tcp", addr.String())
		if err != nil {
			b.Error(err)
			continue
		}
		if err := conn.Close(); err != nil {
			b.Logf("Error closing connection: %v", err)
		}
	}
}
