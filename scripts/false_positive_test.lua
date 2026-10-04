-- False Positive Test Script
-- Testing ports that are definitely closed to check for false positives

-- Configuration - Test ports that should be closed
local CONFIG = {
    host = "127.0.0.1",
    test_ports = {12345, 54321, 9999, 8888, 7777, 6666}, -- Uncommon ports likely to be closed
    delay = 0.01
}

function test_false_positives()
    log("info", "False positive test")
    log("info", "Testing ports that should be CLOSED to detect false positives")
    log("info", "Expected result: ALL ports should be CLOSED")
    
    local false_positives = {}
    local total_tested = #CONFIG.test_ports
    
    for i, port in ipairs(CONFIG.test_ports) do
        log("info", "[" .. i .. "/" .. total_tested .. "] Testing " .. CONFIG.host .. ":" .. port .. " (should be closed)")
        
        local conn, err = connect(CONFIG.host, port, "tcp")
        if conn then
            -- This would be a false positive!
            log("error", "Port " .. port .. " reported as open but should be closed")
            table.insert(false_positives, port)
            close(conn)
        else
            log("info", "Port " .. port .. " detected as closed (" .. (err or "refused") .. ")")
        end
        
        sleep(CONFIG.delay)
    end
    
    log("info", "")
    log("info", "False positive test results:")
    log("info", "Total ports tested: " .. total_tested)
    log("info", "Expected closed: " .. total_tested)
    log("info", "False positives found: " .. #false_positives)
    
    if #false_positives == 0 then
        log("info", "No false positives detected.")
        log("info", "All tested ports correctly identified as closed")
        log("info", "Scanner accuracy: 100%")
    else
        log("error", "False positives detected:")
        for _, port in ipairs(false_positives) do
            log("error", "  Port " .. port .. " incorrectly reported as open")
        end
        local accuracy = ((total_tested - #false_positives) / total_tested) * 100
        log("warn", "Scanner accuracy: " .. string.format("%.1f%%", accuracy))
    end
    
    return false_positives
end

-- Additional test with random high ports
function test_random_high_ports()
    log("info", "")
    log("info", "Random high port test")
    log("info", "Testing random high ports (30000-65000) for false positives")
    
    local random_ports = {}
    for i = 1, 5 do
        local port = math.random(30000, 65000)
        table.insert(random_ports, port)
    end
    
    local false_positives = {}
    
    for i, port in ipairs(random_ports) do
        log("info", "Testing random port " .. port)
        
        local conn, err = connect(CONFIG.host, port, "tcp")
        if conn then
            log("warn", "Random port " .. port .. " is actually open")
            table.insert(false_positives, port)
            close(conn)
        else
            log("info", "Expected: Port " .. port .. " is closed")
        end
        
        sleep(CONFIG.delay)
    end
    
    log("info", "Random port test: " .. (#random_ports - #false_positives) .. "/" .. #random_ports .. " correctly identified as closed")
    return false_positives
end

-- Run false positive tests
log("info", "Starting GoCat False Positive Detection Tests...")

local fp1 = test_false_positives()
local fp2 = test_random_high_ports()

local total_false_positives = #fp1 + #fp2

log("info", "")
log("info", "False positive test completed.")
log("info", "Total false positives detected: " .. total_false_positives)

if total_false_positives == 0 then
    log("info", "No false positives detected.")
    log("info", "Scanner reported no false positives.")
else
    log("warn", "False positives detected: " .. total_false_positives)
    log("info", "Consider investigating connection logic or timeout settings")
end