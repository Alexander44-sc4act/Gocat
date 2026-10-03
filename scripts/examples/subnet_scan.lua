-- GoCat Subnet Scanner
-- Usage: gocat script run subnet_scan.lua <subnet> [port]
-- Example: gocat script run subnet_scan.lua 192.168.1 22

-- Parse arguments
local subnet = arg[1]
local port = tonumber(arg[2]) or 80

-- Validate
if not subnet then
    print("GoCat Subnet Scanner")
    print("====================")
    print("")
    print("Usage: gocat script run subnet_scan.lua <subnet> [port]")
    print("")
    print("Arguments:")
    print("  subnet    First 3 octets of subnet (e.g., 192.168.1)")
    print("  port      Port to scan (default: 80)")
    print("")
    print("Examples:")
    print("  gocat script run subnet_scan.lua 192.168.1")
    print("  gocat script run subnet_scan.lua 10.0.0 22")
    print("  gocat script run subnet_scan.lua 172.16.0 443")
    return
end

-- Main function
local function main()
    print("GoCat Subnet Scanner")
    print("====================")
    print("")
    print(string.format("Subnet: %s.0/24", subnet))
    print(string.format("Port: %d", port))
    print("")
    print("Scanning 256 hosts...")
    print("")
    
    local found = {}
    local scanned = 0
    
    for i = 1, 254 do
        local ip = subnet .. "." .. i
        scanned = scanned + 1
        
        -- Progress indicator
        if scanned % 50 == 0 then
            io.write(string.format("\rScanned: %d/254 hosts, Found: %d", scanned, #found))
            io.flush()
        end
        
        local conn, err = net.connect(ip, port, "tcp")
        if conn then
            net.close(conn)
            table.insert(found, ip)
        end
    end
    
    print("")
    print("")
    
    if #found > 0 then
        print("Active hosts with port " .. port .. " open:")
        print("------------------------------------------")
        for _, ip in ipairs(found) do
            print("  " .. ip)
        end
        print("")
        print(string.format("Found %d active host(s)", #found))
    else
        print("No hosts found with port " .. port .. " open")
    end
end

main()
