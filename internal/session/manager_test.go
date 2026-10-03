package session

import (
	"net"
	"sync"
	"testing"

	"github.com/realibrahimsql/Gocat/internal/logger"
)

func newManagerTestSession(t *testing.T, m *Manager, name string) (*Session, func()) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	serverCh := make(chan net.Conn, 1)
	errCh := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			errCh <- err
			return
		}
		serverCh <- conn
	}()

	client, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		_ = ln.Close()
		t.Fatalf("dial: %v", err)
	}

	var server net.Conn
	select {
	case server = <-serverCh:
	case err := <-errCh:
		_ = client.Close()
		_ = ln.Close()
		t.Fatalf("accept: %v", err)
	}

	tcpAddr := ln.Addr().(*net.TCPAddr)
	sess := NewSession(server, "127.0.0.1", tcpAddr.Port, 1, SourceReverse)
	sess.Name = name
	sess.NameColored = name
	sess.User = "tester"
	m.AddSession(sess)

	cleanup := func() {
		_ = client.Close()
		_ = ln.Close()
		sess.Kill()
	}
	return sess, cleanup
}

func TestManagerConcurrentAttachDetach(t *testing.T) {
	logger.SetLevel(logger.LevelFatal)
	t.Cleanup(func() { logger.SetLevel(logger.LevelInfo) })

	m := NewManager(t.TempDir())
	m.NoLog = true

	var cleanups []func()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		_, cleanup := newManagerTestSession(t, m, name)
		cleanups = append(cleanups, cleanup)
	}
	defer func() {
		for _, cleanup := range cleanups {
			cleanup()
		}
	}()

	sessions := m.ListSessions()
	if len(sessions) != 3 {
		t.Fatalf("len(ListSessions()) = %d, want 3", len(sessions))
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		for _, sess := range sessions {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				if err := m.AttachSession(id); err != nil {
					t.Errorf("AttachSession(%d): %v", id, err)
				}
				m.DetachSession()
			}(sess.ID)
		}
	}
	wg.Wait()

	if m.AttachedSession != nil {
		t.Fatalf("AttachedSession = %v, want nil", m.AttachedSession.ID)
	}
}

func TestManagerRemoveSessionCleansIndexes(t *testing.T) {
	logger.SetLevel(logger.LevelFatal)
	t.Cleanup(func() { logger.SetLevel(logger.LevelInfo) })

	m := NewManager(t.TempDir())
	m.NoLog = true

	sess, cleanup := newManagerTestSession(t, m, "target")
	defer cleanup()

	if err := m.AttachSession(sess.ID); err != nil {
		t.Fatalf("AttachSession() error = %v", err)
	}

	m.RemoveSession(sess.ID)

	if got := m.GetSession(sess.ID); got != nil {
		t.Fatalf("GetSession(%d) = %+v, want nil", sess.ID, got)
	}
	if m.AttachedSession != nil {
		t.Fatalf("AttachedSession = %+v, want nil", m.AttachedSession)
	}
	if got := m.SessionsForHost("target"); len(got) != 0 {
		t.Fatalf("SessionsForHost(target) len = %d, want 0", len(got))
	}
}

func TestManagerMaxSessionsCap(t *testing.T) {
	logger.SetLevel(logger.LevelFatal)
	t.Cleanup(func() { logger.SetLevel(logger.LevelInfo) })

	m := NewManager(t.TempDir())
	m.NoLog = true
	m.MaxSessions = 1

	sess1, cleanup1 := newManagerTestSession(t, m, "victim")
	defer cleanup1()

	// Second session for the same host must be rejected with -1.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	serverCh := make(chan net.Conn, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		serverCh <- conn
	}()
	client, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()
	server := <-serverCh
	defer server.Close()

	sess2 := NewSession(server, "127.0.0.1", 0, 1, SourceReverse)
	sess2.Name = "victim"
	sess2.NameColored = "victim"
	if id := m.AddSession(sess2); id != -1 {
		t.Fatalf("AddSession over cap = %d, want -1", id)
	}
	if n := m.SessionCount(); n != 1 {
		t.Fatalf("SessionCount = %d, want 1", n)
	}
	_ = sess1
}
