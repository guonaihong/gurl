
var program = gurl_flag();
var flag = program
   .option("f, file", "", "pcm file")
   .option("url", "", "http url")
   .option("ak, appkey", "", "appkey")
   .option("p", "", "open consumer mode")
   .parse()

if (!flag.hasOwnProperty("f")) {
    flag.Usage()
    gurl_exit(0)
}

function getSuffix(audioName) {
    var pos = flag.f.lastIndexOf(".")
    var suffix = "";
    if (pos != -1) {
        return audioName.substring(pos + 1, audioName.length);
    }
    return audioName
}

function getSetp(suffix) {
    var setp   = 200;
    if (suffix == "pcm" || suffix == "wav") {
        setp =  8000;
    }
    return setp
}

var suffix = getSuffix(flag.f);
suffix = suffix == "wav"  ? "pcm" : suffix;

var config = {
    H :[
        "appkey:haha",
    ],
    url:'http://192.168.6.128:24987/asr/' + suffix
};

try {
    config.H[0] =  flag.appkey  != "" ? "appkey:"+flag.appkey : config.H[0];
} catch(e) {
    console.log(e)
}

try {
    config.url =  flag.url  != "" ? flag.url : config.url
} catch(e) {
    console.log(e)
}

var file = flag.f;

var http = gurl_http();
slice_one = function(fname, step, time){

    var xnumber = 0;
    var sessionId = gurl_uuid();
    var all = gurl_readfile(fname);


    var i, l;
    i = 0, l = all.length;
    console.log("<" + sessionId + ">", "start", "audio.length",
            l, "step", step);

    for (i = 0, l = all.length; i < l; i += step) {
        var end  = i + step;
        if (end > l) {
            end = l;
        }

        console.log("<" + sessionId + ">","i = ", i, "end = ", end);
        var rsp = http.send({
            H : [
                config.H[0],
                "X-Number:" + xnumber,
                "session-id:" + sessionId,
            ],

            MF : [
                "voice=" + all.slice(i, end), 
            ],
            url : config.url
        });

        if (rsp.err != "") {
            console.log("rsp error is " + rsp.err);
            return
        }

        if (rsp.status_code === 200) {
            gurl_fjson(rsp.body)
        } else {
            console.log("error http code", rsp.status_code)
        }

        gurl_sleep(time);
        xnumber++;
    }

    var rsp = http.send({
        H : [
            config.H[0],
            "X-Number:" + xnumber +"$",
            "session-id:" + sessionId,
        ],

        url : config.url,
    });

    if (rsp.err != "") {
        console.log("rsp error is " + rsp.err);
        return
    }

    console.log("rsp.err:", rsp.err, "rsp.status_code", rsp.status_code);
    if (rsp.status_code == 200) {
        console.log("[filename] " +fname+ "<" +sessionId + ">" + gurl_fjson(rsp.body))
    }
}


if (flag.hasOwnProperty("p") && flag.p.length > 0) {

    do {
        var message = gurl_message();
        var obj = message.read();
        if (obj.text.trim().length == 0) {
            continue;
        }
        console.log("file name---->(" + obj.voice, ") text-->", obj.text);
        var file = obj.voice;
        var suffix = getSuffix(file);
        var step = getSetp(suffix);
        slice_one(obj.voice, step, "249ms");
    }while(false);

} else {
    try {
        var suffix = getSuffix(file);
        var setp = getSetp(suffix);
        slice_one(file, setp, "250ms")
    } catch(e) {
        console.log("call slice_one fail " + e);
        gurl_exit(1)
    }
}
