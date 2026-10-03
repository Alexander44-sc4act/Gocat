# GoCat Package Documentation

## internal/sniffer

Package sniffer provides advanced network packet capture and analysis capabilities.

### Types

#### Packet

```go
type Packet struct {
    Timestamp   time.Time
    Type        PacketType
    SrcIP       net.IP
    DstIP       net.IP
    SrcPort     uint16
    DstPort     uint16
    Protocol    string
    Length      int
    Payload     []byte
    Flags       TCPFlags
    Metadata    map[string]interface{}
    CaptureInfo CaptureInfo
}
```

Packet represents a captured network packet with all relevant metadata.

#### Sniffer

```go
type Sniffer struct {
    // contains filtered or unexported fields
}
```

Sniffer represents a network packet sniffer.

### Functions

#### NewSniffer

```go
func NewSniffer(config *Config) *Sniffer
```

NewSniffer creates a new packet sniffer with the given configuration.

#### AddHandler

```go
func (s *Sniffer) AddHandler(handler PacketHandler)
```

AddHandler adds a packet handler callback that will be called for each captured packet.

#### GetStatistics

```go
func (s *Sniffer) GetStatistics() Statistics
```

GetStatistics returns current capture statistics.

---

## internal/zerocopy

Package zerocopy provides zero-copy I/O optimizations for network operations.

### Types

#### Reader

```go
type Reader struct {
    // contains filtered or unexported fields
}
```

Reader provides zero-copy read operations with internal buffering.

#### Writer

```go
type Writer struct {
    // contains filtered or unexported fields
}
```

Writer provides zero-copy write operations.

#### TransferPipe

```go
type TransferPipe struct {
    // contains filtered or unexported fields
}
```

TransferPipe creates a bidirectional zero-copy pipe between connections.

### Functions

#### NewReader

```go
func NewReader(r io.Reader, bufferSize int) *Reader
```

NewReader creates a new zero-copy reader with the specified buffer size.

#### Splice

```go
func Splice(dst, src net.Conn, bufferSize int) (int64, error)
```

Splice performs zero-copy data transfer between two connections using splice/sendfile when available.

---

## internal/profiling

Package profiling provides memory profiling and optimization tools.

### Types

#### Profiler

```go
type Profiler struct {
    // contains filtered or unexported fields
}
```

Profiler provides memory profiling capabilities.

#### MemoryStats

```go
type MemoryStats struct {
    Alloc        uint64
    TotalAlloc   uint64
    Sys          uint64
    NumGC        uint32
    HeapAlloc    uint64
    HeapSys      uint64
    HeapIdle     uint64
    HeapInuse    uint64
    HeapReleased uint64
    HeapObjects  uint64
    StackInuse   uint64
    StackSys     uint64
    GCPauseNs    uint64
    Goroutines   int
    Timestamp    time.Time
}
```

MemoryStats holds memory statistics.

### Functions

#### NewProfiler

```go
func NewProfiler(maxSamples int, interval time.Duration) *Profiler
```

NewProfiler creates a new memory profiler.

#### GetMemoryReport

```go
func (p *Profiler) GetMemoryReport() string
```

GetMemoryReport generates a human-readable memory report.

---

## internal/network

Package network provides connection management and network utilities.

### Types

#### OptimizedPool

```go
type OptimizedPool struct {
    // contains filtered or unexported fields
}
```

OptimizedPool provides high-performance connection pooling.

### Functions

#### NewOptimizedPool

```go
func NewOptimizedPool(config *OptimizedPoolConfig, dialFunc DialFunc) *OptimizedPool
```

NewOptimizedPool creates a new optimized connection pool.

#### Get

```go
func (p *OptimizedPool) Get(ctx context.Context, address string) (net.Conn, error)
```

Get retrieves a connection from the pool or creates a new one.

#### Stats

```go
func (p *OptimizedPool) Stats() OptimizedPoolStats
```

Stats returns pool statistics.

---

## internal/buffer

Package buffer provides high-performance buffer pool management.

### Types

#### BufferPool

```go
type BufferPool struct {
    // contains filtered or unexported fields
}
```

BufferPool manages a pool of reusable byte buffers with size-based allocation.

#### ManagedBuffer

```go
type ManagedBuffer struct {
    Data []byte
    Info BufferInfo
    // contains filtered or unexported fields
}
```

ManagedBuffer wraps a byte slice with metadata.

### Functions

#### NewBufferPool

```go
func NewBufferPool(minSize, maxSize int, adaptive bool) *BufferPool
```

NewBufferPool creates a new buffer pool with specified configuration.

#### Get

```go
func (bp *BufferPool) Get(size int) *ManagedBuffer
```

Get retrieves a buffer of at least the specified size.

#### Release

```go
func (mb *ManagedBuffer) Release()
```

Release returns the buffer to the pool.
