local cmd = require("cmd")
local flag = cmd.new()
local opt = flag
            :opt_str("f, file", "", "open audio file")
            :opt_str("a, addr", "", "Remote service address")
            :parse("-f ./tst.pcm -a 127.0.0.1:8080")

function tableHasKey(table, key)
    return table[key] ~= nil
end

if (not tableHasKey(opt, "f")) or
    (not tableHasKey(opt, "file")) or
    (not tableHasKey(opt, "a")) or
    (not tableHasKey(opt, "addr")) then

    opt.Usage()

    return
end

for k, v in pairs(opt) do
    print("cmd opt ("..k..") parse value ("..v..")")
end
