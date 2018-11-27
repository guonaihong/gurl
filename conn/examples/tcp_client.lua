local socket = require("socket")
local s = socket.new()

local err = s:connect("127.0.0.1:1234")
if err ~= nil then
    print(err)
    return
end

local _, err = s:write("hello world 1 2 3")
if err ~= nil then
    print("write", err)
    return
end

local len = s:read(4)

local body, err = s:read(s:ntohl(len))
if err ~= nil then
    print("read", err)
    return
end

print(body)
