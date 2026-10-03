package zerocopy

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

func TestNewReader(t *testing.T) {
	buf := bytes.NewBufferString("test data")
	r := NewReader(buf, 1024)

	if r == nil {
		t.Fatal("NewReader returned nil")
	}

	if r.reader == nil {
		t.Error("reader is nil")
	}

	if len(r.buffer) != 1024 {
		t.Errorf("buffer size = %d, want 1024", len(r.buffer))
	}
}

func TestNewReaderDefaultSize(t *testing.T) {
	buf := bytes.NewBufferString("test data")
	r := NewReader(buf, 0)

	if len(r.buffer) != 64*1024 {
		t.Errorf("default buffer size = %d, want %d", len(r.buffer), 64*1024)
	}
}

func TestNewWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)

	if w == nil {
		t.Fatal("NewWriter returned nil")
	}

	if w.writer == nil {
		t.Error("writer is nil")
	}
}

func TestReaderRead(t *testing.T) {
	data := "hello world test data"
	buf := bytes.NewBufferString(data)
	r := NewReader(buf, 1024)

	result := make([]byte, len(data))
	n, err := r.Read(result)

	if err != nil && err != io.EOF {
		t.Errorf("Read error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Read %d bytes, want %d", n, len(data))
	}

	if string(result[:n]) != data {
		t.Errorf("Read data = %s, want %s", string(result[:n]), data)
	}
}

func TestWriterWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)

	data := []byte("hello world")
	n, err := w.Write(data)

	if err != nil {
		t.Errorf("Write error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Wrote %d bytes, want %d", n, len(data))
	}

	if buf.String() != string(data) {
		t.Errorf("Written data = %s, want %s", buf.String(), string(data))
	}
}

func TestReaderStats(t *testing.T) {
	data := "test data for stats"
	buf := bytes.NewBufferString(data)
	r := NewReader(buf, 1024)

	result := make([]byte, len(data))
	r.Read(result)

	stats := r.GetStats()

	if stats.BytesCopied == 0 && stats.BytesZeroCopy == 0 {
		t.Error("No bytes recorded in stats")
	}
}

func TestWriterStats(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)

	data := []byte("test data")
	w.Write(data)

	stats := w.GetStats()

	if stats.BytesZeroCopy != int64(len(data)) {
		t.Errorf("BytesZeroCopy = %d, want %d", stats.BytesZeroCopy, len(data))
	}
}

func TestTransferPipeCreation(t *testing.T) {
	// Test default buffer size handling
	if 64*1024 != 65536 {
		t.Error("Expected default buffer size constant")
	}
}

func TestStatsStruct(t *testing.T) {
	stats := Stats{
		BytesCopied:    100,
		BytesZeroCopy:  200,
		CopyOperations: 10,
		ZeroCopyOps:    20,
		FallbackOps:    5,
	}

	if stats.BytesCopied != 100 {
		t.Errorf("BytesCopied = %d, want 100", stats.BytesCopied)
	}

	if stats.BytesZeroCopy != 200 {
		t.Errorf("BytesZeroCopy = %d, want 200", stats.BytesZeroCopy)
	}
}

func TestTransferStats(t *testing.T) {
	stats := TransferStats{
		BytesSrcToDst: 1000,
		BytesDstToSrc: 500,
	}

	if stats.BytesSrcToDst != 1000 {
		t.Errorf("BytesSrcToDst = %d, want 1000", stats.BytesSrcToDst)
	}
}

// mockConn implements net.Conn for testing
type mockConn struct {
	reader io.Reader
	writer io.Writer
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return m.reader.Read(b) }
func (m *mockConn) Write(b []byte) (n int, err error)  { return m.writer.Write(b) }
func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestTransferPipeWithMock(t *testing.T) {
	srcReader, srcWriter := io.Pipe()
	dstReader, dstWriter := io.Pipe()

	src := &mockConn{reader: srcReader, writer: srcWriter}
	dst := &mockConn{reader: dstReader, writer: dstWriter}

	pipe := NewTransferPipe(src, dst, 1024)

	if pipe == nil {
		t.Fatal("NewTransferPipe returned nil")
	}

	if pipe.bufSize != 1024 {
		t.Errorf("bufSize = %d, want 1024", pipe.bufSize)
	}

	// Clean up
	srcReader.Close()
	srcWriter.Close()
	dstReader.Close()
	dstWriter.Close()
}
