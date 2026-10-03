-- GoCat Service Detector
-- Usage: gocat script run service_detector.lua <target>
-- Example: gocat script run service_detector.lua 192.168.1.1

-- Parse arguments
local target = arg[1]

-- Validate
if not target then
    print("GoCat Service Detector")
    print("======================")
    print("")
    print("Usage: gocat script run service_detector.lua <target>")
    print("")
    print("Detects common services running on a target host.")
    print("")
    print("Examples:")
    print("  gocat script run service_detector.lua 192.168.1.1")
    print("  gocat script run service_detector.lua scanme.nmap.org")
    return
end

-- Service definitions with detection methods
local services = {
    {port = 21, name = "FTP", probe = nil},
    {port = 22, name = "SSH", probe = nil},
    {port = 23, name = "Telnet", probe = nil},
    {port = 25, name = "SMTP", probe = "EHLO test\r\n"},
    {port = 53, name = "DNS", probe = nil},
    {port = 80, name = "HTTP", probe = "GET / HTTP/1.0\r\n\r\n"},
    {port = 110, name = "POP3", probe = nil},
    {port = 143, name = "IMAP", probe = nil},
    {port = 443, name = "HTTPS", probe = nil},
    {port = 445, name = "SMB", probe = nil},
    {port = 3306, name = "MySQL", probe = nil},
    {port = 3389, name = "RDP", probe = nil},
    {port = 5432, name = "PostgreSQL", probe = nil},
    {port = 6379, name = "Redis", probe = "PING\r\n"},
    {port = 8080, name = "HTTP-Proxy", probe = "GET / HTTP/1.0\r\n\r\n"},
    {port = 27017, name = "MongoDB", probe = nil}
}

-- Check if port is open and get banner
local function check_service(svc)
    local conn, err = net.connect(target, svc.port, "tcp")
    if not conn then
        return nil
    end
    
    local banner = ""
    
    -- Send probe if defined
    if svc.probe then
        net.send(conn, svc.probe)
    end
    
    -- Try to receive banner
    banner, err = net.receive(conn, 1024)
    net.close(conn)
    
    return banner or ""
end

-- Main function
local function main()
    print("GoCat Service Detector")
    print("======================")
    print("")
    print("Target: " .. target)
    print("")
    print("Detecting services...")
    print("")
    print("PORT      SERVICE       STATUS    INFO")
    print("----      -------       ------    ----")
    
    local found = 0
    
    for _, svc in ipairs(services) do
        local banner = check_service(svc)
        
        if banner then
            found = found + 1
            local info = banner:gsub("[\r\n]", " "):sub(1, 40)
            if info == "" then info = "(no banner)" end
            print(string.format("%-9d %-13s OPEN      %s", svc.port, svc.name, info))
        end
    end
    
    print("")
    print(string.format("Detected %d service(s)", found))
end

main()
