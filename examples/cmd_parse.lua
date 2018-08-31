local flag = require("flag").new()
local opt = flag
            :opt_str("f, file", "", "open audio file")
            :opt_str("a, addr", "", "Remote service address")
            :parse("-f ./tst.pcm -a 127.0.0.1:8080")


if  #opt["f"] == 0 or
    #opt["file"] == 0 or
    #opt["a"] == 0 or
    #opt["addr"] == 0  then

    opt.Usage()

    return
end

for k, v in pairs(opt) do
    print("cmd opt ("..k..") parse value ("..v..")")
end
