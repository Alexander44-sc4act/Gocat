-- GoCat HTTP Client
-- Usage: gocat script run http_client.lua <url> [method] [data]
-- Example: gocat script run http_client.lua https://httpbin.org/get
-- Example: gocat script run http_client.lua https://httpbin.org/post POST '{"key":"value"}'

local url = arg[1]
local method = arg[2] or "GET"
local data = arg[3] or ""

-- Validate
if not url then
    print("GoCat HTTP Client")
    print("")
    print("Usage: gocat script run http_client.lua <url> [method] [data]")
    print("")
    print("Arguments:")
    print("  url       Target URL (required)")
    print("  method    HTTP method (default: GET)")
    print("  data      Request body for POST/PUT")
    print("")
    print("Examples:")
    print("  gocat script run http_client.lua https://httpbin.org/get")
    print("  gocat script run http_client.lua https://httpbin.org/ip")
    print("  gocat script run http_client.lua https://api.github.com")
    print("  gocat script run http_client.lua https://httpbin.org/post POST '{\"test\":true}'")
    return
end

local function format_headers(headers)
    if not headers then return "" end
    local result = {}
    for k, v in pairs(headers) do
        table.insert(result, string.format("  %s: %s", k, v))
    end
    table.sort(result)
    return table.concat(result, "\n")
end

local function format_body(body, max_len)
    if not body then return "(empty)" end
    max_len = max_len or 2000
    if #body > max_len then
        return body:sub(1, max_len) .. "\n... (truncated, " .. #body .. " bytes total)"
    end
    return body
end

-- Main function
local function main()
    print("GoCat HTTP Client")
    print("")
    print("Request:")
    print("  Method: " .. method:upper())
    print("  URL: " .. url)
    if data ~= "" then
        print("  Body: " .. data:sub(1, 100))
    end
    print("")
    print("Sending request...")
    print("")
    
    local resp, err
    
    method = method:upper()
    
    if method == "GET" then
        resp, err = http.get(url)
    elseif method == "POST" then
        resp, err = http.post(url, data)
    elseif method == "PUT" then
        resp, err = http.put(url, data)
    elseif method == "DELETE" then
        resp, err = http.delete(url)
    else
        resp, err = http.request(method, url, nil, data)
    end
    
    if err then
        print("Error: " .. err)
        return
    end
    
    if not resp then
        print("Error: No response received")
        return
    end
    
    print("Response:")
    print("---------")
    print("Status: " .. (resp.status or "unknown"))
    print("")
    
    if resp.headers then
        print("Headers:")
        print(format_headers(resp.headers))
        print("")
    end
    
    print("Body:")
    print(format_body(resp.body))
end

main()
