-- GoCat TCP Ping
-- Usage: gocat script run tcp_ping.lua <host> <port> [count]
-- Example: gocat script run tcp_ping.lua google.com 443 10

local host = arg[1]
local port = tonumber(arg[2])
local count = tonumber(arg[3]) or 5

-- Validate
if not host or not port then
    print("GoCat TCP Ping")
    print("")
    print("Usage: gocat script run tcp_ping.lua <host> <port> [count]")
    print("")
    print("Arguments:")
    print("  host    Target hostname or IP")
    print("  port    Target port number")
    print("  count   Number of pings (default: 5)")
    print("")
    print("Examples:")
    print("  gocat script run tcp_ping.lua google.com 443")
    print("  gocat script run tcp_ping.lua 192.168.1.1 22 10")
    print("  gocat script run tcp_ping.lua github.com 443 20")
    return
end

-- Statistics
local stats = {
    sent = 0,
    received = 0,
    times = {}
}

-- TCP ping function
local function tcp_ping()
    local start = os.clock()
    local conn, err = net.connect(host, port, "tcp")
    local elapsed = (os.clock() - start) * 1000
    
    if conn then
        net.close(conn)
        return true, elapsed
    end
    
    return false, elapsed
end

-- Main function
local function main()
    print("GoCat TCP Ping")
    print("")
    print(string.format("TCP ping %s:%d", host, port))
    print("")
    
    for i = 1, count do
        stats.sent = stats.sent + 1
        local success, time = tcp_ping()
        
        if success then
            stats.received = stats.received + 1
            table.insert(stats.times, time)
            print(string.format("Connected to %s:%d - time=%.2fms", host, port, time))
        else
            print(string.format("Connection to %s:%d failed", host, port))
        end
        
        if i < count then
            sleep(1000)
        end
    end
    
    -- Statistics
    print("")
    print(string.format("--- %s:%d TCP ping statistics ---", host, port))
    print(string.format("%d packets transmitted, %d received, %.1f%% packet loss",
        stats.sent, stats.received, ((stats.sent - stats.received) / stats.sent) * 100))
    
    if #stats.times > 0 then
        local min_time = stats.times[1]
        local max_time = stats.times[1]
        local sum = 0
        
        for _, t in ipairs(stats.times) do
            sum = sum + t
            if t < min_time then min_time = t end
            if t > max_time then max_time = t end
        end
        
        local avg = sum / #stats.times
        print(string.format("rtt min/avg/max = %.2f/%.2f/%.2f ms", min_time, avg, max_time))
    end
end

main()
