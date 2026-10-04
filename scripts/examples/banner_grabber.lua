-- GoCat Banner Grabber
-- Usage: gocat script run banner_grabber.lua <target> [ports]
-- Example: gocat script run banner_grabber.lua 192.168.1.1 22,80,21

local target = arg[1]
local ports_arg = arg[2] or "21,22,25,80,110,143,443"

-- Validate
if not target then
    print("GoCat Banner Grabber")
    print("")
    print("Usage: gocat script run banner_grabber.lua <target> [ports]")
    print("")
    print("Arguments:")
    print("  target    Target host (IP or hostname)")
    print("  ports     Comma-separated ports (default: 21,22,25,80,110,143,443)")
    print("")
    print("Examples:")
    print("  gocat script run banner_grabber.lua 192.168.1.1")
    print("  gocat script run banner_grabber.lua scanme.nmap.org 22,80,443")
    return
end

-- Protocol probes for different services
local probes = {
    [80] = "GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n",
    [443] = "GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n",
    [8080] = "GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n",
    [25] = "EHLO gocat\r\n",
    [587] = "EHLO gocat\r\n"
}

-- Parse ports
local function parse_ports(spec)
    local ports = {}
    for p in spec:gmatch("(%d+)") do
        local port = tonumber(p)
        if port and port > 0 and port <= 65535 then
            table.insert(ports, port)
        end
    end
    return ports
end

-- Identify service from banner
local function identify_service(banner, port)
    if not banner or banner == "" then
        return "unknown"
    end
    
    local patterns = {
        {"SSH%-", "SSH"},
        {"OpenSSH", "OpenSSH"},
        {"HTTP/", "HTTP"},
        {"Apache", "Apache"},
        {"nginx", "nginx"},
        {"Microsoft%-IIS", "IIS"},
        {"220.*FTP", "FTP"},
        {"vsftpd", "vsftpd"},
        {"ProFTPD", "ProFTPD"},
        {"220.*SMTP", "SMTP"},
        {"220.*Postfix", "Postfix"},
        {"220.*ESMTP", "ESMTP"},
        {"MySQL", "MySQL"},
        {"MariaDB", "MariaDB"},
        {"PostgreSQL", "PostgreSQL"},
        {"%+OK.*POP3", "POP3"},
        {"%* OK.*IMAP", "IMAP"},
        {"Dovecot", "Dovecot"}
    }
    
    for _, p in ipairs(patterns) do
        if banner:find(p[1]) then
            return p[2]
        end
    end
    
    return "unknown"
end

-- Clean banner for display
local function clean_banner(banner)
    if not banner then return "" end
    banner = banner:gsub("[\r\n]+", " ")
    banner = banner:gsub("%s+", " ")
    banner = banner:match("^%s*(.-)%s*$")
    if #banner > 80 then
        banner = banner:sub(1, 77) .. "..."
    end
    return banner
end

-- Main function
local function main()
    local ports = parse_ports(ports_arg)
    
    print("GoCat Banner Grabber")
    print("")
    print("Target: " .. target)
    print("Ports: " .. table.concat(ports, ", "))
    print("")
    print("Grabbing banners...")
    print("")
    print("PORT      SERVICE       BANNER")
    print("----      -------       ------")
    
    local found = 0
    
    for _, port in ipairs(ports) do
        -- Try to grab banner
        local banner, err = net.banner_grab(target, port)
        
        if banner and banner ~= "" then
            local service = identify_service(banner, port)
            local clean = clean_banner(banner)
            print(string.format("%-9d %-13s %s", port, service, clean))
            found = found + 1
        end
    end
    
    print("")
    if found > 0 then
        print(string.format("Grabbed %d banner(s)", found))
    else
        print("No banners grabbed (ports may be closed or filtered)")
    end
end

main()
