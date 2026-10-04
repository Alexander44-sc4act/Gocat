-- GoCat DNS Lookup
-- Usage: gocat script run dns_lookup.lua <domain>
-- Example: gocat script run dns_lookup.lua google.com

-- Parse arguments
local domain = arg[1]

-- Validate
if not domain then
    print("GoCat DNS Lookup")
    print("")
    print("Usage: gocat script run dns_lookup.lua <domain>")
    print("")
    print("Arguments:")
    print("  domain    Domain name to lookup")
    print("")
    print("Examples:")
    print("  gocat script run dns_lookup.lua google.com")
    print("  gocat script run dns_lookup.lua github.com")
    print("  gocat script run dns_lookup.lua example.org")
    return
end

-- Try to connect to get IP (simple DNS resolution via connection)
local function resolve_host(host)
    -- Try common ports to resolve
    local ports = {80, 443, 22}
    
    for _, port in ipairs(ports) do
        local conn, err = net.connect(host, port, "tcp")
        if conn then
            net.close(conn)
            return true
        end
    end
    
    return false
end

-- Main function
local function main()
    print("GoCat DNS Lookup")
    print("")
    print("Domain: " .. domain)
    print("")
    
    -- Check if domain resolves by trying to connect
    print("Checking DNS resolution...")
    print("")
    
    -- Try HTTP to get more info
    local url = "http://" .. domain
    local resp, err = http.get(url)
    
    if resp then
        print("Domain resolves successfully!")
        print("")
        print("HTTP Response Status: " .. (resp.status or "N/A"))
        
        if resp.headers then
            print("")
            print("Server Headers:")
            if resp.headers["Server"] then
                print("  Server: " .. resp.headers["Server"])
            end
            if resp.headers["X-Powered-By"] then
                print("  X-Powered-By: " .. resp.headers["X-Powered-By"])
            end
            if resp.headers["Content-Type"] then
                print("  Content-Type: " .. resp.headers["Content-Type"])
            end
        end
    else
        -- Try HTTPS
        url = "https://" .. domain
        resp, err = http.get(url)
        
        if resp then
            print("Domain resolves successfully! (HTTPS)")
            print("")
            print("HTTPS Response Status: " .. (resp.status or "N/A"))
        else
            print("Could not connect to domain")
            print("Error: " .. (err or "unknown"))
        end
    end
    
    print("")
    
    -- Port scan common ports
    print("Scanning common ports...")
    local common_ports = "22,80,443"
    local open = net.scan(domain, common_ports)
    
    if open and #open > 0 then
        print("")
        print("Open ports: " .. table.concat(open, ", "))
    else
        print("No common ports open or host unreachable")
    end
end

main()
