-- GoCat Port Scanner
-- Usage: gocat script run port_scanner.lua <target> [ports]
-- Example: gocat script run port_scanner.lua scanme.nmap.org 1-1000
-- Example: gocat script run port_scanner.lua 192.168.1.1 22,80,443,8080

-- Parse command line arguments
local target = arg[1]
local ports_arg = arg[2] or "1-1024"

-- Validate arguments
if not target then
    print("GoCat Port Scanner")
    print("==================")
    print("")
    print("Usage: gocat script run port_scanner.lua <target> [ports]")
    print("")
    print("Arguments:")
    print("  target    Target host (IP or hostname)")
    print("  ports     Port specification (optional, default: 1-1024)")
    print("")
    print("Port formats:")
    print("  80              Single port")
    print("  80,443,8080     Multiple ports")
    print("  1-1000          Port range")
    print("  22,80,443,8000-9000  Combined")
    print("")
    print("Examples:")
    print("  gocat script run port_scanner.lua scanme.nmap.org")
    print("  gocat script run port_scanner.lua 192.168.1.1 22,80,443")
    print("  gocat script run port_scanner.lua example.com 1-65535")
    return
end

-- Common service names
local services = {
    [21] = "ftp",
    [22] = "ssh",
    [23] = "telnet",
    [25] = "smtp",
    [53] = "dns",
    [80] = "http",
    [110] = "pop3",
    [111] = "rpcbind",
    [135] = "msrpc",
    [139] = "netbios-ssn",
    [143] = "imap",
    [443] = "https",
    [445] = "microsoft-ds",
    [993] = "imaps",
    [995] = "pop3s",
    [1723] = "pptp",
    [3306] = "mysql",
    [3389] = "ms-wbt-server",
    [5432] = "postgresql",
    [5900] = "vnc",
    [6379] = "redis",
    [8080] = "http-proxy",
    [8443] = "https-alt",
    [27017] = "mongodb"
}

-- Parse port specification
local function parse_ports(spec)
    local ports = {}
    
    -- Split by comma
    for part in spec:gmatch("[^,]+") do
        part = part:match("^%s*(.-)%s*$")  -- trim
        
        -- Check if range
        local start_port, end_port = part:match("(%d+)%-(%d+)")
        if start_port and end_port then
            start_port = tonumber(start_port)
            end_port = tonumber(end_port)
            for p = start_port, end_port do
                if p > 0 and p <= 65535 then
                    table.insert(ports, p)
                end
            end
        else
            -- Single port
            local p = tonumber(part)
            if p and p > 0 and p <= 65535 then
                table.insert(ports, p)
            end
        end
    end
    
    return ports
end

-- Get service name
local function get_service(port)
    return services[port] or "unknown"
end

-- Main scan function
local function scan()
    local ports = parse_ports(ports_arg)
    
    print("GoCat Port Scanner")
    print("==================")
    print("")
    print("Target: " .. target)
    print("Ports: " .. #ports .. " ports to scan")
    print("")
    print("Scanning...")
    print("")
    
    local open_ports = net.scan(target, ports_arg)
    
    if not open_ports or #open_ports == 0 then
        print("No open ports found.")
        return
    end
    
    print("PORT      STATE   SERVICE")
    print("----      -----   -------")
    
    for _, port in ipairs(open_ports) do
        local service = get_service(port)
        print(string.format("%-9d open    %s", port, service))
    end
    
    print("")
    print(string.format("Scan complete: %d open port(s) found", #open_ports))
end

-- Run
scan()
