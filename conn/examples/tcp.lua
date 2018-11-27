local flag = require("flag").new()
local socket = require("socket")
local time = require("time")
local strings = require("strings")
local log = require("log")
local uuid = require("uuid")
local json = require("json")

local mylog = log.new("debug")
local opt = flag
            :opt_str("f, file", "", "open audio file")
            :opt_str("a, addr", "", "Remote service address")
            :parse(get_cmd_args())

if opt["addr"] ~= nil and #opt["addr"] == 0 then
    flag:usage()
    return
end

function send_asr(fname, addr)
    local f = assert(io.open(fname, "rb"))
    local block = 8000
    local sessionid = uuid:newv4()
    local s = socket.new()

    local err = s:connect(addr)
    if err ~= nil then
        mylog:error("<"..sessionid.."> connect", err, "\n")
        return
    end

    while true do
        local bytes = f:read(block)
        if not bytes then 
            mylog:info("<"..sessionid..">read file bye bye\n")
            break 
        end

        --print(#bytes)
        --s:write(s:n2big_bytes(#bytes))
        local _, err = s:write(bytes, "1s")
        if err ~= nil then
            mylog:error("<"..sessionid..">write bytes", err, "\n")
            break
        end

        local len, err = s:read(4, "1s")
        if err ~= nil then
            mylog:info("<"..sessionid..">read 4 byte", err, "\n")
            break
        end

        local body, err = s:read(s:ntohl(len))
        if err ~= nil then
            mylog:info("<"..sessionid..">", err, "\n")
            break
        end

        --mylog:debug(body, "\n")
        time.sleep("250ms")
    end

    mylog:info("<"..sessionid..">first loop end, start send end\n")
    -- send eof message
    s:write(strings.rep("=", 64))
    f:close()

    mylog:info("<"..sessionid..">--------------------------write end message\n")
    while true do
        local len, err = s:read(4)
        if err ~= nil then 
            mylog:warn("<"..sessionid..">", err, "\n")
            break 
        end

        local body2 , err = s:read(s:ntohl(len))

        if err ~= nil then
            mylog:warn("<"..sessionid..">", err, "\n")
            break
        end

        body = body2
        mylog:debug("<"..sessionid..">", body, "\n")
    end

    mylog:info("<"..sessionid..">", "result#", body, "\n")
    if body == nil then
        return
    end

    local str = fname.." "..body.."\n"
    channel.select(
    {"<-|", out_ch, str, function(data)
        --print(data)
    end},
    {"default", function()
        print("default action")
    end}
    )
    s:close()
end

if opt["file"] ~= nil and #opt["file"] ~= 0 then
    --print("file:"..opt["file"])
    --print("addr:"..opt["addr"])
    send_asr(opt["file"], opt["addr"])
    return
end

while not exit do
    channel.select(
    {"|<-", in_ch, function(ok, v)
        if not ok then
            print("channel closed")
            exit = true
        else
            print("received:", v)
            line = strings.split(v, "\t")
            send_asr(line[1], opt.addr)
        end
    end}
    )
end
