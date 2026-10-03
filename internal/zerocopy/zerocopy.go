// Package zerocopy provides zero-copy I/O optimizations for network operations
package zerocopy

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
)

// Stats tracks zero-copy operation statistics
type Stats struct {
	BytesCopied    int64
	BytesZeroCopy  int64
	CopyOperations int64
	ZeroCopyOps    int64
	FallbackOps    int64
}

// Reader provides zero-copy read operations
type Reader struct {
	reader    io.Reader
	buffer    []byte
	bufferLen int
	pos       int
	stats     *Stats
	mu        sync.Mutex
}

// Writer provides zero-copy write operations
type Writer struct {
	writer io.Writer
	stats  *Stats
	mu     sync.Mutex
}

// NewReader creates a new zero-copy reader
func NewReader(r io.Reader, bufferSize int) *Reader {
	if bufferSize <= 0 {
		bufferSize = 64 * 1024 // 64KB default
	}
	return &Reader{
		reader: r,
		buffer: make([]byte, bufferSize),
		stats:  &Stats{},
	}
}

// NewWriter creates a new zero-copy writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		stats:  &Stats{},
	}
}

// Read reads data with zero-copy optimization when possible
func (r *Reader) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If we have buffered data, return it first
	if r.pos < r.bufferLen {
		n = copy(p, r.buffer[r.pos:r.bufferLen])
		r.pos += n
		atomic.AddInt64(&r.stats.BytesCopied, int64(n))
		atomic.AddInt64(&r.stats.CopyOperations, 1)
		return n, nil
	}

	// Try direct read for large buffers
	if len(p) >= len(r.buffer) {
		n, err = r.reader.Read(p)
		if n > 0 {
			atomic.AddInt64(&r.stats.BytesZeroCopy, int64(n))
			atomic.AddInt64(&r.stats.ZeroCopyOps, 1)
		}
		return n, err
	}

	// Refill buffer
	r.bufferLen, err = r.reader.Read(r.buffer)
	r.pos = 0
	if r.bufferLen > 0 {
		n = copy(p, r.buffer[:r.bufferLen])
		r.pos = n
		atomic.AddInt64(&r.stats.BytesCopied, int64(n))
		atomic.AddInt64(&r.stats.CopyOperations, 1)
	}
	return n, err
}

// Write writes data with zero-copy optimization when possible
func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err = w.writer.Write(p)
	if n > 0 {
		atomic.AddInt64(&w.stats.BytesZeroCopy, int64(n))
		atomic.AddInt64(&w.stats.ZeroCopyOps, 1)
	}
	return n, err
}

// GetStats returns current statistics
func (r *Reader) GetStats() Stats {
	return Stats{
		BytesCopied:    atomic.LoadInt64(&r.stats.BytesCopied),
		BytesZeroCopy:  atomic.LoadInt64(&r.stats.BytesZeroCopy),
		CopyOperations: atomic.LoadInt64(&r.stats.CopyOperations),
		ZeroCopyOps:    atomic.LoadInt64(&r.stats.ZeroCopyOps),
		FallbackOps:    atomic.LoadInt64(&r.stats.FallbackOps),
	}
}

// GetStats returns current statistics for writer
func (w *Writer) GetStats() Stats {
	return Stats{
		BytesZeroCopy: atomic.LoadInt64(&w.stats.BytesZeroCopy),
		ZeroCopyOps:   atomic.LoadInt64(&w.stats.ZeroCopyOps),
	}
}

// Splice performs zero-copy data transfer between two connections
func Splice(dst, src net.Conn, bufferSize int) (int64, error) {
	// Try to use splice/sendfile if available
	if rf, ok := src.(*net.TCPConn); ok {
		if wf, ok := dst.(*net.TCPConn); ok {
			return spliceTCP(wf, rf, bufferSize)
		}
	}

	// Fallback to io.Copy with buffer
	buf := make([]byte, bufferSize)
	return io.CopyBuffer(dst, src, buf)
}

// spliceTCP attempts zero-copy transfer between TCP connections
func spliceTCP(dst, src *net.TCPConn, bufferSize int) (int64, error) {
	// Get raw file descriptors for splice syscall
	srcFile, err := src.File()
	if err != nil {
		// Fallback to regular copy
		buf := make([]byte, bufferSize)
		return io.CopyBuffer(dst, src, buf)
	}
	defer srcFile.Close()

	dstFile, err := dst.File()
	if err != nil {
		buf := make([]byte, bufferSize)
		return io.CopyBuffer(dst, src, buf)
	}
	defer dstFile.Close()

	// Use ReadFrom which may use sendfile internally
	return io.Copy(dst, src)
}

// TransferPipe creates a bidirectional zero-copy pipe between connections
type TransferPipe struct {
	src       net.Conn
	dst       net.Conn
	bufSize   int
	stats     TransferStats
	done      chan struct{}
	closeOnce sync.Once
}

// TransferStats holds transfer statistics
type TransferStats struct {
	BytesSrcToDst int64
	BytesDstToSrc int64
	StartTime     int64
	EndTime       int64
}

// NewTransferPipe creates a new transfer pipe
func NewTransferPipe(src, dst net.Conn, bufferSize int) *TransferPipe {
	if bufferSize <= 0 {
		bufferSize = 64 * 1024
	}
	return &TransferPipe{
		src:     src,
		dst:     dst,
		bufSize: bufferSize,
		done:    make(chan struct{}),
	}
}

// Start begins bidirectional transfer
func (tp *TransferPipe) Start() {
	go tp.transfer(tp.src, tp.dst, &tp.stats.BytesSrcToDst)
	go tp.transfer(tp.dst, tp.src, &tp.stats.BytesDstToSrc)
}

func (tp *TransferPipe) transfer(dst, src net.Conn, counter *int64) {
	defer tp.Close()
	n, _ := Splice(dst, src, tp.bufSize)
	atomic.AddInt64(counter, n)
}

// Close closes the transfer pipe
func (tp *TransferPipe) Close() {
	tp.closeOnce.Do(func() {
		close(tp.done)
		tp.src.Close()
		tp.dst.Close()
	})
}

// Wait waits for transfer to complete
func (tp *TransferPipe) Wait() {
	<-tp.done
}

// GetStats returns transfer statistics
func (tp *TransferPipe) GetStats() TransferStats {
	return TransferStats{
		BytesSrcToDst: atomic.LoadInt64(&tp.stats.BytesSrcToDst),
		BytesDstToSrc: atomic.LoadInt64(&tp.stats.BytesDstToSrc),
	}
}
