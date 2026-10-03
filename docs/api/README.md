# GoCat API Documentation

## Overview

GoCat is a modern, feature-rich netcat alternative written in Go. This document provides comprehensive API documentation for developers.

## Package Structure

```
github.com/realibrahimsql/Gocat/
├── cmd/                    # CLI commands
├── internal/
│   ├── buffer/            # Buffer pool management
│   ├── network/           # Network utilities
│   ├── sniffer/           # Packet capture
│   ├── zerocopy/          # Zero-copy I/O
│   ├── profiling/         # Memory profiling
│   ├── scripting/         # Lua scripting engine
│   └── ...
└── pkg/                   # Public packages
```

## Core Packages

### internal/network

Connection management and network utilities.

```go
import "github.com/realibrahimsql/Gocat/internal/network"

// Create connection pool
config := network.DefaultPoolConfig()
pool := network.NewConnectionPool(config, dialer)

// Get connection from pool
conn, err := pool.Get(ctx, "example.com:80")

// Return connection to pool
pool.Put(conn)

// Get pool statistics
stats := pool.Stats()
```

### internal/buffer

High-performance buffer pool with adaptive sizing.

```go
import "github.com/realibrahimsql/Gocat/internal/buffer"

// Create buffer pool
pool := buffer.NewBufferPool(1024, 65536, true)

// Get buffer
buf := pool.Get(4096)

// Use buffer
copy(buf.Data, data)

// Return buffer
buf.Release()
```

### internal/sniffer

Network packet capture and analysis.

```go
import "github.com/realibrahimsql/Gocat/internal/sniffer"

// Create sniffer
config := sniffer.DefaultConfig()
config.Interface = "eth0"
config.Filter = "tcp port 80"

s := sniffer.NewSniffer(config)

// Add packet handler
s.AddHandler(func(packet *sniffer.Packet) {
    fmt.Println(sniffer.FormatPacket(packet, sniffer.OutputFormatText))
})

// Get statistics
stats := s.GetStatistics()
```

### internal/zerocopy

Zero-copy I/O optimizations.

```go
import "github.com/realibrahimsql/Gocat/internal/zerocopy"

// Create zero-copy reader
reader := zerocopy.NewReader(conn, 64*1024)

// Create transfer pipe
pipe := zerocopy.NewTransferPipe(src, dst, 64*1024)
pipe.Start()
pipe.Wait()
```

### internal/profiling

Memory profiling and monitoring.

```go
import "github.com/realibrahimsql/Gocat/internal/profiling"

// Create profiler
profiler := profiling.NewProfiler(100, time.Second)
profiler.Start()

// Get current stats
stats := profiler.GetCurrentStats()

// Get memory report
report := profiler.GetMemoryReport()

// Track allocations
profiler.TrackAllocation("buffer", 4096)
```

## Lua Scripting API

### net module

```lua
local net = require("net")

-- TCP connection
local conn = net.dial("tcp", "example.com:80")
conn:write("GET / HTTP/1.1\r\n\r\n")
local data = conn:read(1024)
conn:close()

-- TLS connection
local conn = net.dial_tls("tcp", "example.com:443")

-- DNS lookup
local ips = net.lookup("example.com", {type = "A"})
```

### http module

```lua
local http = require("http")

-- GET request
local resp = http.get("https://api.example.com/data")

-- POST request
local resp = http.post("https://api.example.com/data", {
    headers = {["Content-Type"] = "application/json"},
    body = '{"key": "value"}'
})
```

### crypto module

```lua
local crypto = require("crypto")

-- Hash functions
local hash = crypto.sha256("data")
local hash = crypto.md5("data")

-- Encryption
local encrypted = crypto.aes_encrypt(data, key)
local decrypted = crypto.aes_decrypt(encrypted, key)
```

## Configuration

### YAML Configuration

```yaml
# ~/.gocat.yml
defaults:
  timeout: 30s
  retry: 3
  keep_alive: true

logging:
  level: info
  file: /var/log/gocat.log

network:
  buffer_size: 65536
  pool_size: 10

security:
  verify_cert: true
```

## Error Handling

GoCat uses structured errors with error codes:

```go
import "github.com/realibrahimsql/Gocat/internal/errors"

// Create error
err := errors.NetworkError("NET001", "Connection failed")

// Add context
err = err.WithContext("host", "example.com")

// Check error type
if errors.IsNetworkError(err) {
    // Handle network error
}
```

## Performance Tips

1. Use connection pooling for repeated connections
2. Enable buffer pooling for high-throughput scenarios
3. Use zero-copy I/O for large data transfers
4. Monitor memory with the profiling package
5. Use BPF filters to reduce packet processing overhead

## License

MIT License - see LICENSE file for details.
