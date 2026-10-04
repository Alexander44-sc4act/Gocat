-- WebSocket Client Example for GoCat
-- Demonstrates WebSocket connections

local ws = require("websocket")
local json = require("json")
local time = require("time")

-- Configuration
local config = {
    url = arg[1] or "wss://echo.websocket.org",
    timeout = 10000,
    ping_interval = 30000
}

-- Message handler
local function on_message(msg)
    print("[<] Received: " .. msg)
end

-- Error handler
local function on_error(err)
    print("[!] Error: " .. err)
end

-- Close handler
local function on_close(code, reason)
    print(string.format("[*] Connection closed: %d - %s", code or 0, reason or ""))
end

-- Main function
function main()
    print("GoCat WebSocket Client")
    print("")
    print("Connecting to: " .. config.url)
    
    -- Connect to WebSocket server
    local conn, err = ws.connect(config.url, {
        timeout = config.timeout,
        headers = {
            ["User-Agent"] = "GoCat/1.0"
        }
    })
    
    if not conn then
        print("[-] Connection failed: " .. (err or "unknown error"))
        return
    end
    
    print("[+] Connected!")
    print("")
    
    -- Set handlers
    conn:on_message(on_message)
    conn:on_error(on_error)
    conn:on_close(on_close)
    
    -- Send test messages
    local messages = {
        "Hello, WebSocket!",
        json.encode({type = "ping", timestamp = time.now()}),
        "GoCat WebSocket Test"
    }
    
    for _, msg in ipairs(messages) do
        print("[>] Sending: " .. msg)
        conn:send(msg)
        time.sleep(1000)
    end
    
    -- Wait for responses
    print("")
    print("[*] Waiting for responses...")
    time.sleep(3000)
    
    -- Close connection
    conn:close(1000, "Normal closure")
    print("")
    print("[*] Done!")
end

main()
