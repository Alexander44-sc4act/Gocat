-- Health Checker Example for GoCat
-- Monitors service health and availability

local net = require("net")
local http = require("http")
local time = require("time")
local json = require("json")

-- Configuration
local config = {
    interval = 10000,  -- 10 seconds
    timeout = 5000,    -- 5 seconds
    retries = 3,
    services = {
        {
            name = "Web Server",
            type = "http",
            url = "http://localhost:8080/health",
            expected_status = 200
        },
        {
            name = "Database",
            type = "tcp",
            host = "localhost",
            port = 5432
        },
        {
            name = "Redis",
            type = "tcp",
            host = "localhost",
            port = 6379
        },
        {
            name = "API Gateway",
            type = "http",
            url = "http://localhost:3000/api/health",
            expected_status = 200
        }
    }
}

-- Service status
local status = {}

-- Check HTTP service
local function check_http(service)
    local start = time.now()
    local resp, err = http.get(service.url, {
        timeout = config.timeout
    })
    local latency = time.now() - start
    
    if err then
        return false, err, latency
    end
    
    if resp.status ~= service.expected_status then
        return false, string.format("unexpected status: %d", resp.status), latency
    end
    
    return true, nil, latency
end

-- Check TCP service
local function check_tcp(service)
    local start = time.now()
    local addr = service.host .. ":" .. service.port
    local conn, err = net.dial("tcp", addr, {
        timeout = config.timeout
    })
    local latency = time.now() - start
    
    if err then
        return false, err, latency
    end
    
    conn:close()
    return true, nil, latency
end

-- Check a service with retries
local function check_service(service)
    local check_func = service.type == "http" and check_http or check_tcp
    
    for attempt = 1, config.retries do
        local ok, err, latency = check_func(service)
        if ok then
            return {
                healthy = true,
                latency = latency,
                attempts = attempt
            }
        end
        
        if attempt < config.retries then
            time.sleep(1000)  -- Wait 1 second before retry
        end
    end
    
    local ok, err, latency = check_func(service)
    return {
        healthy = false,
        error = err,
        latency = latency,
        attempts = config.retries
    }
end

local function print_status()
    local now = os.date("%Y-%m-%d %H:%M:%S")
    
    print("")
    print("Service health status:")
    print(string.format("Time: %s", now))
    
    local healthy_count = 0
    local total_count = #config.services
    
    for _, service in ipairs(config.services) do
        local s = status[service.name] or {healthy = false}
        local icon = s.healthy and "[OK]" or "[FAIL]"
        local state = s.healthy and "HEALTHY" or "UNHEALTHY"
        local latency = s.latency and string.format("%.0fms", s.latency) or "N/A"
        
        if s.healthy then
            healthy_count = healthy_count + 1
        end
        
        print(string.format("%s %-20s  %-10s  Latency: %-8s",
            icon, service.name, state, latency))
        
        if s.error then
            print(string.format("Error: %s",
                string.sub(s.error, 1, 47)))
        end
    end
    
    print("Summary: %d/%d services healthy")
        healthy_count, total_count))
end

-- Main monitoring loop
function main()
    print("GoCat Health Checker")
    print("")
    print("Monitoring " .. #config.services .. " services")
    print("Check interval: " .. (config.interval / 1000) .. " seconds")
    print("Press Ctrl+C to stop")
    
    while true do
        -- Check all services
        for _, service in ipairs(config.services) do
            status[service.name] = check_service(service)
        end
        
        print_status()
        
        -- Wait for next check
        time.sleep(config.interval)
    end
end

main()
