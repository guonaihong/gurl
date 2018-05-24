
var flag = gurl_flag_parse(
        gurl_args,
        ["f", "", "audio file name"]
);

if (!flag.hasOwnProperty("f")) {
    console.log("not set -f")
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

var suffix = getSuffix(flag.f);
suffix = suffix == "wav"  ? "pcm" : suffix;

var config = {
    H :[
        "appkey:haha",
    ],
    url:'http://192.168.6.128:24987/asr/' + suffix
};

try {
    config.url =  gurl_url
} catch(e) {
    console.log(e)
}

var files = [
    flag.f
];

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
        var rsp = gurl_send({
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

    var rsp = gurl_send({
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
        console.log("<" +sessionId + ">" + gurl_fjson(rsp.body))
    }
}

for (var fname in files) {
    console.log("---->" + files[fname]);
    try {
        var suffix = getSuffix(files[fname]);
        var setp   = 200;
        if (suffix == "pcm") {
            suffix =  8000;
        }
        slice_one(files[fname], suffix, "250ms")
    } catch(e) {
        console.log("call slice_one fail " + e);
        gurl_exit(1)
    }
}
