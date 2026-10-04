-- GoCat SSL/TLS Client
-- Usage: gocat script run ssl_client.lua <host> [port]
-- Example: gocat script run ssl_client.lua google.com 443

local host = arg[1]
local port = tonumber(arg[2]) or 443

-- Validate
if not host then
    print("GoCat SSL/TLS Client")
    print("")
    print("Usage: gocat script run ssl_client.lua <host> [port]")
    print("")
    print("Arguments:")
    print("  host    Target hostname (required)")
    print("  port    Port number (default: 443)")
    print("")
    print("Examples:")
    print("  gocat script run ssl_client.lua google.com")
    print("  gocat script run ssl_client.lua github.com 443")
    print("  gocat script run ssl_client.lua smtp.gmail.com 465")
    return
end

-- Main function
local function main()
    print("GoCat SSL/TLS Client")
    print("")
    print("Target: " .. host .. ":" .. port)
    print("")
    
    -- Make HTTPS request
    local url = "https://" .. host
    if port ~= 443 then
        url = url .. ":" .. port
    end
    
    print("Connecting via HTTPS...")
    print("")
    
    local resp, err = http.get(url)
    
    if err then
        print("Connection failed: " .. err)
        print("")
        print("This could mean:")
        print("  - Host doesn't support HTTPS on this port")
        print("  - Certificate validation failed")
        print("  - Host is unreachable")
        return
    end
    
    if resp then
        print("SSL/TLS connection successful.")
        print("")
        print("HTTP Status: " .. (resp.status or "N/A"))
        print("")
        
        if resp.headers then
            print("Security Headers:")
            local security_headers = {
                "Strict-Transport-Security",
                "Content-Security-Policy",
                "X-Frame-Options",
                "X-Content-Type-Options",
                "X-XSS-Protection"
            }
            
            for _, header in ipairs(security_headers) do
                if resp.headers[header] then
                    print("  " .. header .. ": " .. resp.headers[header]:sub(1, 60))
                end
            end
            
            print("")
            print("Server Info:")
            if resp.headers["Server"] then
                print("  Server: " .. resp.headers["Server"])
            end
        end
        
        print("")
        print("Response body length: " .. (resp.body and #resp.body or 0) .. " bytes")
    end
end

main()
