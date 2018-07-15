local cmd = require("cmd")

local flag = cmd.new()
local opt = flag
            :opt_str("f, file", "", "open audio file")
            :parse(gurl_cmd)

if #opt["file"] == 0 then
    opt:usage()
    return
end

local file = io.open(opt.file, "wb")
if not file then
    print("bye bye")
    return
end

exit = false
while not exit do
    channel.select(
    {"|<-", in_ch, function(ok, v)
        if not ok then
            print("channel closed")
            exit = true
        else
            --print("received:", v)
            file:write(v)
        end
    end}
    )
end

io.close(file)
