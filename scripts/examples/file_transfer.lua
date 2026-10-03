-- File Transfer Example for GoCat
-- Demonstrates file sending and receiving

local net = require("net")
local file = require("file")
local crypto = require("crypto")
local time = require("time")

-- Configuration
local config = {
    mode = arg[1] or "send",  -- "send" or "receive"
    host = arg[2] or "127.0.0.1",
    port = tonumber(arg[3]) or 9000,
    filename = arg[4] or "transfer.dat",
    chunk_size = 64 * 1024,  -- 64KB chunks
    timeout = 30000
}

-- Format bytes
local function format_bytes(bytes)
    local units = {"B", "KB", "MB", "GB"}
    local unit = 1
    while bytes >= 1024 and unit < #units do
        bytes = bytes / 1024
        unit = unit + 1
    end
    return string.format("%.2f %s", bytes, units[unit])
end

-- Calculate file checksum
local function calculate_checksum(filepath)
    local content, err = file.read(filepath)
    if not content then
        return nil, err
    end
    return crypto.sha256(content)
end

-- Send file
local function send_file()
    print("GoCat File Transfer - Sender")
    print("============================")
    print("")
    
    -- Check file exists
    local info, err = file.stat(config.filename)
    if not info then
        print("[-] File not found: " .. config.filename)
        return
    end
    
    local filesize = info.size
    print("File: " .. config.filename)
    print("Size: " .. format_bytes(filesize))
    
    -- Calculate checksum
    print("[*] Calculating checksum...")
    local checksum = calculate_checksum(config.filename)
    print("Checksum: " .. checksum:sub(1, 16) .. "...")
    print("")
    
    -- Connect to receiver
    print("[*] Connecting to " .. config.host .. ":" .. config.port)
    local conn, err = net.dial("tcp", config.host .. ":" .. config.port, {
        timeout = config.timeout
    })
    
    if not conn then
        print("[-] Connection failed: " .. (err or "unknown"))
        return
    end
    
    print("[+] Connected!")
    
    -- Send file metadata
    local metadata = string.format("%s|%d|%s\n", 
        config.filename, filesize, checksum)
    conn:write(metadata)
    
    -- Read file and send
    print("[*] Sending file...")
    local start_time = time.now()
    local sent = 0
    
    local f, err = file.open(config.filename, "r")
    if not f then
        print("[-] Failed to open file: " .. err)
        conn:close()
        return
    end
    
    while sent < filesize do
        local chunk = f:read(config.chunk_size)
        if not chunk or #chunk == 0 then
            break
        end
        
        conn:write(chunk)
        sent = sent + #chunk
        
        -- Progress
        local progress = (sent / filesize) * 100
        io.write(string.format("\r[*] Progress: %.1f%% (%s / %s)", 
            progress, format_bytes(sent), format_bytes(filesize)))
        io.flush()
    end
    
    f:close()
    conn:close()
    
    local elapsed = (time.now() - start_time) / 1000
    local speed = sent / elapsed
    
    print("")
    print("")
    print("[+] Transfer complete!")
    print(string.format("    Time: %.2f seconds", elapsed))
    print(string.format("    Speed: %s/s", format_bytes(speed)))
end

-- Receive file
local function receive_file()
    print("GoCat File Transfer - Receiver")
    print("==============================")
    print("")
    
    -- Listen for connection
    print("[*] Listening on port " .. config.port)
    local listener, err = net.listen("tcp", ":" .. config.port)
    
    if not listener then
        print("[-] Failed to listen: " .. (err or "unknown"))
        return
    end
    
    print("[*] Waiting for connection...")
    local conn, err = listener:accept()
    
    if not conn then
        print("[-] Accept failed: " .. (err or "unknown"))
        listener:close()
        return
    end
    
    print("[+] Connection from: " .. conn:remote_addr())
    
    -- Read metadata
    local metadata = conn:read_line()
    local filename, filesize, checksum = metadata:match("([^|]+)|(%d+)|(%S+)")
    filesize = tonumber(filesize)
    
    print("")
    print("File: " .. filename)
    print("Size: " .. format_bytes(filesize))
    print("Checksum: " .. checksum:sub(1, 16) .. "...")
    print("")
    
    -- Receive file
    print("[*] Receiving file...")
    local start_time = time.now()
    local received = 0
    
    local f, err = file.open(config.filename, "w")
    if not f then
        print("[-] Failed to create file: " .. err)
        conn:close()
        listener:close()
        return
    end
    
    while received < filesize do
        local chunk = conn:read(config.chunk_size)
        if not chunk or #chunk == 0 then
            break
        end
        
        f:write(chunk)
        received = received + #chunk
        
        -- Progress
        local progress = (received / filesize) * 100
        io.write(string.format("\r[*] Progress: %.1f%% (%s / %s)", 
            progress, format_bytes(received), format_bytes(filesize)))
        io.flush()
    end
    
    f:close()
    conn:close()
    listener:close()
    
    local elapsed = (time.now() - start_time) / 1000
    local speed = received / elapsed
    
    print("")
    print("")
    
    -- Verify checksum
    print("[*] Verifying checksum...")
    local local_checksum = calculate_checksum(config.filename)
    
    if local_checksum == checksum then
        print("[+] Checksum verified!")
    else
        print("[-] Checksum mismatch!")
    end
    
    print("")
    print("[+] Transfer complete!")
    print(string.format("    Time: %.2f seconds", elapsed))
    print(string.format("    Speed: %s/s", format_bytes(speed)))
end

-- Main
function main()
    if config.mode == "send" then
        send_file()
    else
        receive_file()
    end
end

main()
