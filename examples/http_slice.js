
var config = {
    H :[
        "appkey:haha",
    ],
    url:'http://192.168.6.128:24987/asr/opus'
    //url:'http://192.168.5.132:24987/asr/opus'
};


var files = [
    "good.opus"
];

slice_one = function(fname, step, time){

    var xnumber = 0;
    var sessionId = gurl_uuid();
    var all = gurl_readfile(fname);

    console.log("<" + sessionId + ">", "start");
    for (var i = 0, l = gurl_len(all); i < l; i += step) {
        var end  = i + step;
        if (end > gurl_len(all)) {
            end = gurl_len(all);
        }

        console.log("<" + sessionId + ">","i = ", i, "end = ", end);
        var rsp = gurl({
            H : [
                config.H[0],
                "X-Number:" + xnumber,
                "session-id:" + sessionId,
            ],

            MF : [
                "voice=" + gurl_extract(all, i, end), 
            ],
            url : config.url
        });

        if (rsp.err != "") {
            console.log("rsp error is " + rsp.err);
            return
        }

        if (rsp.status_code === 200) {
            gurl_fjson(rsp.body)
        }

        gurl_sleep(time);
        xnumber++;
    }

    var rsp = gurl({
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
        slice_one(files[fname], 200, "250ms")
    } catch(e) {
        console.log("call slice_one fail " + e);
    }
}
