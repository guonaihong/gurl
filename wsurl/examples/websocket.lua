local log = require("log").new("debug")
local flag = require("flag").new()
local uuid = require("uuid")
local time = require("time")
local uuid = require("uuid")
local json = require("json")
local websocket = require("websocket")
local strings = require("strings")

local opt = flag
            :opt_str("f, file", "", "open audio file")
            :opt_str("u, url", "", "websocket url")
            :opt_str("appkey", "", "appkey")
            :parse(get_cmd_args())

if #opt["url"] == 0 or #opt["appkey"] == 0 then
    flag:usage()
    return
end

function send_asr(config)
    local fname = config.fname
    local url = config.url
    local f = assert(io.open(fname, "rb"))
    local ws = websocket.new()
    local header = {}

    local sessionid = uuid:newv4()
    table.insert(header, "appkey:"..opt["appkey"])
    table.insert(header, "session-id:"..sessionid)
    table.insert(header, "eof:eof")
    print("url", url)
    local err = ws:connect(url, header)
    if err ~= nil then
        log:error("<"..sessionid.."connect fail", err)
        return
    end

    local block = 8000

    local last_body = ""
    while true do
        local bytes = f:read(block)
        if not bytes then
            log:info("<"..sessionid..">read file bye bye\n")
            break
        end

        ws:write("binary", bytes)

        local mt, body, err = ws:read()
        if err ~= nil then
            log:error("<"..sessionid..">read fail", err, "\n")
            break
        end

        last_data = body
        --log:debug("mt", mt, body, "\n")
        time.sleep(tostring(config.time).."ms")
    end

    -- 发送结束消息
    ws:write("binary", "eof")
    while true do
        local mt, body, err = ws:read()
        if err ~= nil then
            log:error("<"..sessionid..">read fail", err, "\n")
            break
        end

        local _, p = string.find(body, '"isEnd":true')
        if p ~= nil then
            last_body = body
            break
        end

        last_body = body
        --log:debug("mt", mt, body, "\n")
    end

    local str = "[filename]"..fname.."<"..sessionid..">"..last_body.."\n"
    --local str = "[filename]"..fname.."<"..sessionid..">"..json.format(last_body)
    log:debug("last_data", str)
    if #last_body ~= 0 then
        channel.select(
	{"<-|", out_ch, str, function(data)
            print("------------------>lastdata", data)
        end},
	{"default", function()
	    print("default action")
        end})

    end
    ws:close()
end

if #opt["file"] ~= 0 then
    send_asr({
        fname=opt["file"],
        url=opt["url"],
        time=250,
    })
    print("bye bye\n")
    return
end

while not exit do
    local yes = false
    local line = {}

    channel.select(
    {"|<-", in_ch, function(ok, v)
        if not ok then
            print("channel closed")
            exit = true
        else
            print("1.received:", v)
            line = strings.split(v, "\t")
            yes = true
        end
    end}
    )

    if yes then
        print("2. filename"..line[1])
        send_asr({
            fname = line[1],
            step = 8000,
            time = 250,
            url = opt["url"],
        })
    end
end
