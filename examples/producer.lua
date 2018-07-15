local cmd = require("cmd")

local flag = cmd.new()
local opt = flag
            :opt_str("l, list", "", "open audio list file")
            :parse(gurl_cmd)

local list = opt.l
if list == "" then
    list = opt.list
end

print("producer:lua script start:open "..list)
local f = assert(io.open(list, "r"))

while true do
    local line = f:read("*l")
    if not line then break end
    out_ch:send(line)
end
