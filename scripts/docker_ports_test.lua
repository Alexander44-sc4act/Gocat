-- Docker Ports Test - Testing known open Docker ports
-- From ss output: 9090 and 3000 are open via docker-proxy

local CONFIG = {
    host = "127.0.0.1",
    known_open_ports = {9090, 3000, 53}, -- From ss command output
    known_closed_ports = {80, 443, 22},   -- Should be closed on localhost
    delay = 0.01
}

function test_known_open_ports()
    log("info", "Docker ports true positive test")
    log("info", "Testing ports that ss shows as OPEN (docker-proxy + dnscrypt-proxy)")
    log("info", "Expected: 9090, 3000, 53 should be OPEN")
    
    local true_positives = {}
    local false_negatives = {}
    
    for i, port in ipairs(CONFIG.known_open_ports) do
        log("info", "Testing known OPEN port " .. port)
        
        local conn, err = connect(CONFIG.host, port, "tcp")
        if conn then
            log("info", "True positive: port " .. port .. " detected as open")
            table.insert(true_positives, port)
            close(conn)
        else
            log("error", "Port " .. port .. " should be open but connection failed: " .. (err or "unknown"))
            table.insert(false_negatives, port)
        end
        
        sleep(CONFIG.delay)
    end
    
    log("info", "")
    log("info", "Known open ports results:")
    log("info", "True positives: " .. #true_positives .. "/" .. #CONFIG.known_open_ports)
    log("info", "False negatives: " .. #false_negatives .. "/" .. #CONFIG.known_open_ports)
    
    if #false_negatives == 0 then
        log("info", "All known open ports detected.")
    else
        log("warn", "Some open ports were missed (false negatives)")
        for _, port in ipairs(false_negatives) do
            log("warn", "  - Port " .. port .. " was missed")
        end
    end
    
    return true_positives, false_negatives
end

function test_known_closed_ports()
    log("info", "")
    log("info", "Known closed ports test")
    log("info", "Testing ports that should be CLOSED")
    log("info", "Expected: 80, 443, 22 should be CLOSED")
    
    local true_negatives = {}
    local false_positives = {}
    
    for i, port in ipairs(CONFIG.known_closed_ports) do
        log("info", "Testing known CLOSED port " .. port)
        
        local conn, err = connect(CONFIG.host, port, "tcp")
        if conn then
            log("error", "Port " .. port .. " reported as open but should be closed")
            table.insert(false_positives, port)
            close(conn)
        else
            log("info", "True negative: port " .. port .. " detected as closed")
            table.insert(true_negatives, port)
        end
        
        sleep(CONFIG.delay)
    end
    
    log("info", "")
    log("info", "Known closed ports results:")
    log("info", "True negatives: " .. #true_negatives .. "/" .. #CONFIG.known_closed_ports)
    log("info", "False positives: " .. #false_positives .. "/" .. #CONFIG.known_closed_ports)
    
    return true_negatives, false_positives
end

function calculate_overall_accuracy()
    log("info", "")
    log("info", "Comprehensive accuracy test")
    
    local tp, fn = test_known_open_ports()
    local tn, fp = test_known_closed_ports()
    
    local total_tests = #CONFIG.known_open_ports + #CONFIG.known_closed_ports
    local correct_predictions = #tp + #tn
    local accuracy = (correct_predictions / total_tests) * 100
    
    log("info", "")
    log("info", "Final accuracy metrics:")
    log("info", "True Positives (TP): " .. #tp .. " - Correctly identified OPEN ports")
    log("info", "True Negatives (TN): " .. #tn .. " - Correctly identified CLOSED ports")
    log("info", "False Positives (FP): " .. #fp .. " - Incorrectly identified as OPEN")
    log("info", "False Negatives (FN): " .. #fn .. " - Missed OPEN ports")
    log("info", "")
    log("info", "Total tests: " .. total_tests)
    log("info", "Correct predictions: " .. correct_predictions)
    log("info", "Overall accuracy: " .. string.format("%.1f%%", accuracy))
    
    if accuracy >= 95 then
        log("info", "Scanner accuracy is high.")
    elseif accuracy >= 80 then
        log("info", "Scanner accuracy is acceptable.")
    else
        log("warn", "Scanner accuracy needs improvement")
    end
    
    return accuracy
end

-- Run comprehensive tests
log("info", "Starting GoCat Comprehensive Accuracy Test...")
log("info", "Testing against real system port status (from ss output)")

local final_accuracy = calculate_overall_accuracy()

log("info", "")
log("info", "Scanner validation complete.")
log("info", "Final accuracy score: " .. string.format("%.1f%%", final_accuracy))