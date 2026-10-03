// Package zerocopy provides zero-copy I/O optimizations for network operations.
//
// Zero-copy I/O minimizes data copying between kernel and user space,
// significantly improving performance for high-throughput network applications.
//
// # Reader
//
// The Reader provides buffered reading with zero-copy optimization:
//
//	reader := zerocopy.NewReader(conn, 64*1024)
//	data := make([]byte, 1024)
//	n, err := reader.Read(data)
//
// # Writer
//
// The Writer provides zero-copy write operations:
//
//	writer := zerocopy.NewWriter(conn)
//	n, err := writer.Write(data)
//
// # Splice
//
// Splice performs zero-copy data transfer between connections:
//
//	n, err := zerocopy.Splice(dst, src, 64*1024)
//
// This uses splice/sendfile system calls when available.
//
// # TransferPipe
//
// TransferPipe creates a bidirectional zero-copy pipe:
//
//	pipe := zerocopy.NewTransferPipe(src, dst, 64*1024)
//	pipe.Start()
//	pipe.Wait()
//	stats := pipe.GetStats()
//
// # Performance
//
// Zero-copy I/O can provide significant performance improvements:
//   - Reduced CPU usage
//   - Lower memory bandwidth consumption
//   - Higher throughput for large transfers
//   - Reduced latency
package zerocopy
