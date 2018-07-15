local cmd = require("cmd")
local uuid = require("uuid")
local time = require("time")
local http = require("http")
local json = require("json")
local strings = require("strings")

local flag = cmd.new()
local opt = flag
            :opt_str("f, file", "", "open audio file")
            :opt_str("url", "", "http url")
            :opt_str("appkey", "", "appkey")
            :parse(gurl_cmd)

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
    local f = assert(io.open(config.fname), "rb")

    print("start asr send step:"..config.step.." url "..config.url)

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
            print(json.format(body))
        else
            print("error http code".. rsp["status_code"])
        end

        --time.sleep(config.time, "ms")
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
        local str = "[filename]"..config.fname.."<"..session_id..">"..json.format(body)
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


if opt["f"] ~= nil or opt["file"] ~= nil then
    send_asr({
        appkey = opt.appkey,
        fname = opt["file"],
        step = 8000,
        time = 250,
        url = opt.url,
    })
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
            local v = strings.split(v, "\t")
            send_asr({
                appkey = opt.appkey,
                fname = v[1],
                step = 8000,
                time = 250,
            })
        end
    end}
    )
end
