# GoCat Lua Script Examples

This directory contains example Lua scripts for GoCat's scripting engine.

## Usage

All scripts accept command-line arguments. Run without arguments to see help:

```bash
gocat script run scripts/examples/<script>.lua
```

Run with arguments:

```bash
gocat script run scripts/examples/<script>.lua <arg1> <arg2> ...
```

## Available Scripts

### Network Scanning

| Script | Description | Usage |
|--------|-------------|-------|
| `port_scanner.lua` | TCP port scanner | `gocat script run port_scanner.lua <target> [ports]` |
| `subnet_scan.lua` | Subnet host discovery | `gocat script run subnet_scan.lua <subnet> [port]` |
| `service_detector.lua` | Service detection | `gocat script run service_detector.lua <target>` |
| `banner_grabber.lua` | Service banner grabbing | `gocat script run banner_grabber.lua <target> [ports]` |

### Network Utilities

| Script | Description | Usage |
|--------|-------------|-------|
| `tcp_ping.lua` | TCP ping utility | `gocat script run tcp_ping.lua <host> <port> [count]` |
| `network_monitor.lua` | Host availability monitor | `gocat script run network_monitor.lua <host> [interval] [count]` |
| `dns_lookup.lua` | DNS lookup tool | `gocat script run dns_lookup.lua <domain>` |

### HTTP/Web

| Script | Description | Usage |
|--------|-------------|-------|
| `http_client.lua` | HTTP client | `gocat script run http_client.lua <url> [method] [data]` |
| `ssl_client.lua` | SSL/TLS connection test | `gocat script run ssl_client.lua <host> [port]` |
| `web_headers.lua` | Security headers analyzer | `gocat script run web_headers.lua <url>` |

## Examples

### Port Scanning

```bash
# Scan common ports
gocat script run scripts/examples/port_scanner.lua scanme.nmap.org

# Scan specific ports
gocat script run scripts/examples/port_scanner.lua 192.168.1.1 22,80,443,8080

# Scan port range
gocat script run scripts/examples/port_scanner.lua example.com 1-1000
```

### Banner Grabbing

```bash
# Grab banners from common ports
gocat script run scripts/examples/banner_grabber.lua 192.168.1.1

# Grab banners from specific ports
gocat script run scripts/examples/banner_grabber.lua scanme.nmap.org 22,80,443
```

### HTTP Requests

```bash
# GET request
gocat script run scripts/examples/http_client.lua https://httpbin.org/ip

# POST request
gocat script run scripts/examples/http_client.lua https://httpbin.org/post POST '{"key":"value"}'
```

### Network Monitoring

```bash
# Monitor host every 5 seconds, 10 times
gocat script run scripts/examples/network_monitor.lua google.com 5 10

# TCP ping
gocat script run scripts/examples/tcp_ping.lua google.com 443 5
```

### Security Analysis

```bash
# Analyze web security headers
gocat script run scripts/examples/web_headers.lua https://github.com

# Test SSL/TLS connection
gocat script run scripts/examples/ssl_client.lua google.com 443
```

## Writing Your Own Scripts

Scripts have access to these modules:

- `net` - Network operations (connect, listen, scan, banner_grab)
- `http` - HTTP client (get, post, put, delete, request, download)
- `file` - File operations
- `crypto` - Cryptographic functions
- `json` - JSON encoding/decoding
- `time` - Time utilities
- `system` - System information

### Accessing Arguments

```lua
-- Arguments are available in the 'arg' table
local target = arg[1]  -- First argument
local port = arg[2]    -- Second argument

-- Show help if no arguments
if not target then
    print("Usage: gocat script run myscript.lua <target>")
    return
end
```

### Example Template

```lua
-- My Custom Script
-- Usage: gocat script run myscript.lua <target> [options]

local target = arg[1]
local option = arg[2] or "default"

if not target then
    print("Usage: gocat script run myscript.lua <target> [options]")
    return
end

-- Your code here
print("Target: " .. target)
print("Option: " .. option)
```

## License

Apache-2.0 License - See main project LICENSE file.
