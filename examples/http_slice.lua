local cmd = require("cmd")
local uuid = require("uuid")
local time = require("time")
local http = require("http")
local json = require("json")

local flag = cmd.new()
local opt = flag
            :opt_str("f, file", "", "open audio file")
            :opt_str("url", "", "http url")
            :opt_str("appkey", "", "appkey")
            :parse(gurl_cmd)

function send_file(config) 
    local xnumber = 0
    local session_id = uuid:newv4()
    local f = assert(io.open(config.fname), "rb")

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

        if #rsp["err"] == 0 then
            print("rsp error is ".. rsp["err"])
            return
        end

        if rsp["status_code"] == 200 then
            print(rsp["body"])
        else
            console.log("error http code".. rsp["status_code"])
        end

        time.sleep(config.time)
        xnumber = xnumber+1
    end

    local rsp = http.send({
        H = {
            "appkey:"..config.appkey,
            "X-Number:"..xnumber.."$",
            "session-id:"..session_id,
        },

        url = url
    })

    if #rsp["err"] == 0 then
        print("rsp error is " + rsp["err"])
        return
    end

    print("rsp.err:"..rsp["err".."rsp.status_code"..rsp["status_code"]])
    if rsp["status_code"] == 200 then
        print("[filename]"..fname.."<"..session_id..">" +json.format(rsp["body"]))
    end
end

file = flag["f"] or ""
if #file ~= 0 then
    file = flag["file"]
end

if #file ~= 0 then
    send_asr({
    
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
            --send_asr({
            --})
        end
    end}
    )
end
