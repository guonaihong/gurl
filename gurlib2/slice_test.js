
var config = {
    H :[
        "appkey:haha",
    ],
    url:'http://192.168.6.128:24987/asr/opus'
};


var files = [
    "good.pcm"
];

slice_one = function(fname){

    var xnumber = 0;
    var sessionId = gurl_uuid();
    var step = 4
    var all = gurl_readfile(fname);

    for (var i = 0, l = gurl_len(all); i < l; i += step) {
        var end  = i + step;
        if (end > gurl_len(all)) {
            end = gurl_len(all)
        }

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
            console.log("rsp error is " + rsp.err)
        }

        if (rsp.status_code === 200) {
            gurl_format_json(rsp.body)
        }

        gurl_sleep("1s")
        xnumber++;
    }

    var rsp = curl({
        H : [
            config.H[0],
            "X-Number" + xnumber,
        ],

        MF : [
            "voice=" + gurl_extract(all, i, i + step), 
        ],
        url : config.url
    });
}

for (var fname in files) {
    console.log("---->" + files[fname])
    slice_one(files[fname])
}
