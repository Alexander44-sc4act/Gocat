-- TCP Proxy Example for GoCat
-- Creates a simple TCP proxy/relay

local net = require("net")
local time = require("time")

-- Configuration
local config = {
    listen_port = tonumber(arg[1]) or 8080,
    target_host = arg[2] or "127.0.0.1",
    target_port = tonumber(arg[3]) or 80,
    buffer_size = 64 * 1024,
    timeout = 30000
}

-- Statistics
local stats = {
    connections = 0,
    bytes_forwarded = 0,
    bytes_returned = 0,
    start_time = 0
}

local function format_bytes(bytes)
    local units = {"B", "KB", "MB", "GB"}
    local unit = 1
    while bytes >= 1024 and unit < #units do
        bytes = bytes / 1024
        unit = unit + 1
    end
    return string.format("%.2f %s", bytes, units[unit])
end

-- Relay data between connections
local function relay(src, dst, direction)
    local total = 0
    while true do
        local data, err = src:read(config.buffer_size)
        if not data or #data == 0 then
            break
        end
        
        local written, werr = dst:write(data)
        if werr then
            break
        end
        
        total = total + #data
        
        if direction == "forward" then
            stats.bytes_forwarded = stats.bytes_forwarded + #data
        else
            stats.bytes_returned = stats.bytes_returned + #data
        end
    end
    return total
end

-- Handle a single connection
local function handle_connection(client)
    stats.connections = stats.connections + 1
    local conn_id = stats.connections
    
    print(string.format("[%d] New connection from %s", conn_id, client:remote_addr()))
    
    -- Connect to target
    local target_addr = config.target_host .. ":" .. config.target_port
    local target, err = net.dial("tcp", target_addr, {
        timeout = config.timeout
    })
    
    if not target then
        print(string.format("[%d] Failed to connect to target: %s", conn_id, err or "unknown"))
        client:close()
        return
    end
    
    print(string.format("[%d] Connected to target %s", conn_id, target_addr))
    
    -- Start bidirectional relay
    -- Note: In real implementation, these would run concurrently
    local forward_bytes = relay(client, target, "forward")
    local return_bytes = relay(target, client, "return")
    
    -- Cleanup
    client:close()
    target:close()
    
    print(string.format("[%d] Connection closed (fwd: %s, ret: %s)", 
        conn_id, format_bytes(forward_bytes), format_bytes(return_bytes)))
end

-- Main function
function main()
    print("GoCat TCP Proxy")
    print("")
    print(string.format("Listen: 0.0.0.0:%d", config.listen_port))
    print(string.format("Target: %s:%d", config.target_host, config.target_port))
    print("")
    
    stats.start_time = time.now()
    
    -- Create listener
    local listener, err = net.listen("tcp", ":" .. config.listen_port)
    if not listener then
        print("[-] Failed to listen: " .. (err or "unknown"))
        return
    end
    
    print("[*] Proxy started, waiting for connections...")
    print("[*] Press Ctrl+C to stop")
    print("")
    
    -- Accept loop
    while true do
        local client, err = listener:accept()
        if client then
            handle_connection(client)
        end
    end
end

main()
