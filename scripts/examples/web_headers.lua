-- GoCat Web Headers Analyzer
-- Usage: gocat script run web_headers.lua <url>
-- Example: gocat script run web_headers.lua https://google.com

local url = arg[1]

-- Validate
if not url then
    print("GoCat Web Headers Analyzer")
    print("")
    print("Usage: gocat script run web_headers.lua <url>")
    print("")
    print("Analyzes HTTP headers and security configuration.")
    print("")
    print("Examples:")
    print("  gocat script run web_headers.lua https://google.com")
    print("  gocat script run web_headers.lua https://github.com")
    print("  gocat script run web_headers.lua http://example.com")
    return
end

-- Add protocol if missing
if not url:match("^https?://") then
    url = "https://" .. url
end

-- Security headers to check
local security_headers = {
    {name = "Strict-Transport-Security", desc = "HSTS - Forces HTTPS"},
    {name = "Content-Security-Policy", desc = "CSP - Prevents XSS"},
    {name = "X-Frame-Options", desc = "Clickjacking protection"},
    {name = "X-Content-Type-Options", desc = "MIME sniffing protection"},
    {name = "X-XSS-Protection", desc = "XSS filter"},
    {name = "Referrer-Policy", desc = "Referrer information control"},
    {name = "Permissions-Policy", desc = "Feature permissions"},
    {name = "Cross-Origin-Opener-Policy", desc = "COOP"},
    {name = "Cross-Origin-Resource-Policy", desc = "CORP"},
    {name = "Cross-Origin-Embedder-Policy", desc = "COEP"}
}

-- Main function
local function main()
    print("GoCat Web Headers Analyzer")
    print("")
    print("URL: " .. url)
    print("")
    print("Fetching headers...")
    print("")
    
    local resp, err = http.get(url)
    
    if err then
        print("Error: " .. err)
        return
    end
    
    if not resp then
        print("Error: No response")
        return
    end
    
    print("HTTP Status: " .. (resp.status or "N/A"))
    print("")
    
    -- Server info
    print("Server Information:")
    print("-------------------")
    if resp.headers then
        local server_headers = {"Server", "X-Powered-By", "Via"}
        for _, h in ipairs(server_headers) do
            if resp.headers[h] then
                print("  " .. h .. ": " .. resp.headers[h])
            end
        end
    end
    print("")
    
    -- Security headers analysis
    print("Security Headers Analysis:")
    print("--------------------------")
    
    local present = 0
    local missing = 0
    
    for _, sh in ipairs(security_headers) do
        local value = resp.headers and resp.headers[sh.name]
        if value then
            present = present + 1
            local display = value:sub(1, 50)
            if #value > 50 then display = display .. "..." end
            print(string.format("  [+] %s", sh.name))
            print(string.format("      %s", display))
        else
            missing = missing + 1
            print(string.format("  [-] %s (missing)", sh.name))
        end
    end
    
    print("")
    print("Summary:")
    print("--------")
    print(string.format("  Present: %d/%d headers", present, #security_headers))
    print(string.format("  Missing: %d/%d headers", missing, #security_headers))
    
    local score = math.floor((present / #security_headers) * 100)
    print(string.format("  Security Score: %d%%", score))
    
    if score >= 80 then
        print("  Rating: GOOD")
    elseif score >= 50 then
        print("  Rating: MODERATE")
    else
        print("  Rating: NEEDS IMPROVEMENT")
    end
end

main()
