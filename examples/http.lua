local log = require("log").new("debug")
local flag = require("flag").new()
local uuid = require("uuid")
local time = require("time")
local http = require("http")
local json = require("json")
local strings = require("strings")

local opt = flag
            :opt_str("f, file", "", "open audio file")
            :opt_str("url", "", "http url")
            :opt_str("appkey", "", "appkey")
            :parse(get_cmd_args())

--for k, v in pairs(opt) do
    --print("k "..k.." v "..v)
--end

if opt["url"] == nil or opt["appkey"] == nil then
    print("flag:url:"..(opt["url"] or "").." flag:appkey:"..(opt["appkey"] or ""))
    flag:usage()
    return
end

function send_asr(config)
    local xnumber = 0
    local session_id = uuid:newv4()

    log:debug("####"..config.name, "####\n")

    local f = assert(io.open(config.name, "r"))

    print("start"..config.name.. "start asr send step:"..config.step.." url "..config.url)

    while true do
        local bytes = f:read(config.step)
        if not bytes then break end
        local rsp = http.send({
            H = {
                "appkey:"..config.appkey,
                "X-Number:"..xnumber,
                "session-id:"..session_id,
            },
            MF = {
                "voice=" .. bytes,
            },
            url = config.url
        })

        --print("bytes ("..bytes..")")
        if #rsp["err"] ~= 0 then
            print("rsp error is ".. rsp["err"])
            return
        end

        if rsp["status_code"] == 200 then
            body = rsp["body"]
            if #rsp["body"] == 0 then
                body = "{}"
            end
            log:debug(json.format(body), "\n")
        else
            print("error http code".. rsp["status_code"])
        end

        time.sleep(tostring(config.time).."ms")
        xnumber = xnumber+1
    end

    local rsp = http.send({
        H = {
            "appkey:"..config.appkey,
            "X-Number:"..xnumber.."$",
            "session-id:"..session_id,
        },

        url = config.url
    })

    if #rsp["err"] ~= 0 then
        print("rsp error is " + rsp["err"])
        return
    end

    --print("rsp.err:"..rsp["err".."rsp.status_code"..rsp["status_code"]])
    if rsp["status_code"] ~= nil and rsp["status_code"] == 200 then
        body = rsp["body"]
        if #body == 0 then
            body = "{}"
        end
        local str = "[filename]"..config.name.."<"..session_id..">"..json.format(body)
        print(str)

        channel.select(
        {"<-|", out_ch, str, function(data)
            --print(data)
        end},
        {"default", function()
            print("default action")
        end}
        )
    end
end


if (opt["f"] ~= nil and #opt["f"] ~= 0 ) or (opt["file"] ~= nil and #opt["file"] ~= 0) then
    send_asr({
        appkey = opt.appkey,
        name = opt["file"],
        step = 8000,
        time = 250,
        url = opt.url,
    })
    return
end

print("================")
while not exit do
    local yes = false
    local line = {}

    channel.select(
    {"|<-", in_ch, function(ok, v)
        if not ok then
            print("channel closed")
            exit = true
        else
            print("###received:", v)
            line = strings.split(v, "\t")
            yes = true
        end
    end}
    )

    if yes then
        print("##########"..line[1].."##")
        send_asr({
            appkey = opt.appkey,
            name = line[1],
            step = 8000,
            time = 250,
            url = opt.url,
        })
    end
end
