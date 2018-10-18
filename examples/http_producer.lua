local flag = require("flag").new()

local opt = flag
            :opt_str("l, list", "", "open audio list file")
            :parse(get_cmd_args())

local list = opt.l
if list == "" then
    list = opt.list
end

local f = assert(io.open(list, "r"))

while true do
    local line = f:read("*l")
    if not line then break end
    out_ch:send(line)
end
