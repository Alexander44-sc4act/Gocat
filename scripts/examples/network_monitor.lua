-- GoCat Network Monitor
-- Usage: gocat script run network_monitor.lua <host> [interval]
-- Example: gocat script run network_monitor.lua google.com 5

-- Parse arguments
local host = arg[1]
local interval = tonumber(arg[2]) or 5
local count = tonumber(arg[3]) or 10

-- Validate
if not host then
    print("GoCat Network Monitor")
    print("=====================")
    print("")
    print("Usage: gocat script run network_monitor.lua <host> [interval] [count]")
    print("")
    print("Arguments:")
    print("  host      Target host to monitor")
    print("  interval  Check interval in seconds (default: 5)")
    print("  count     Number of checks (default: 10)")
    print("")
    print("Examples:")
    print("  gocat script run network_monitor.lua google.com")
    print("  gocat script run network_monitor.lua 192.168.1.1 2 20")
    print("  gocat script run network_monitor.lua api.example.com 10 100")
    return
end

-- Statistics
local stats = {
    checks = 0,
    successes = 0,
    failures = 0,
    total_time = 0,
    min_time = nil,
    max_time = 0
}

-- Check host availability
local function check_host()
    local start_time = os.clock()
    
    -- Try to connect to port 80 or 443
    local conn, err = net.connect(host, 80, "tcp")
    if not conn then
        conn, err = net.connect(host, 443, "tcp")
    end
    
    local elapsed = (os.clock() - start_time) * 1000  -- ms
    
    if conn then
        net.close(conn)
        return true, elapsed
    end
    
    return false, elapsed
end

-- Format time
local function format_time(ms)
    if ms < 1 then
        return string.format("%.2f ms", ms)
    elseif ms < 1000 then
        return string.format("%.1f ms", ms)
    else
        return string.format("%.2f s", ms / 1000)
    end
end

-- Main function
local function main()
    print("GoCat Network Monitor")
    print("=====================")
    print("")
    print("Target: " .. host)
    print("Interval: " .. interval .. " seconds")
    print("Checks: " .. count)
    print("")
    print("Starting monitoring...")
    print("")
    print("TIME                 STATUS    LATENCY")
    print("----                 ------    -------")
    
    for i = 1, count do
        local timestamp = os.date("%Y-%m-%d %H:%M:%S")
        local success, latency = check_host()
        
        stats.checks = stats.checks + 1
        stats.total_time = stats.total_time + latency
        
        if success then
            stats.successes = stats.successes + 1
            if not stats.min_time or latency < stats.min_time then
                stats.min_time = latency
            end
            if latency > stats.max_time then
                stats.max_time = latency
            end
            print(string.format("%s  UP        %s", timestamp, format_time(latency)))
        else
            stats.failures = stats.failures + 1
            print(string.format("%s  DOWN      -", timestamp))
        end
        
        if i < count then
            sleep(interval * 1000)
        end
    end
    
    -- Print summary
    print("")
    print("Summary")
    print("-------")
    print(string.format("Total checks: %d", stats.checks))
    print(string.format("Successful: %d (%.1f%%)", stats.successes, (stats.successes / stats.checks) * 100))
    print(string.format("Failed: %d (%.1f%%)", stats.failures, (stats.failures / stats.checks) * 100))
    
    if stats.successes > 0 then
        local avg = stats.total_time / stats.successes
        print(string.format("Avg latency: %s", format_time(avg)))
        print(string.format("Min latency: %s", format_time(stats.min_time or 0)))
        print(string.format("Max latency: %s", format_time(stats.max_time)))
    end
end

main()
