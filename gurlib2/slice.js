
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

    all = gurl_readfile(fname);
    for (i = 0; i < gurl_len(all); i += step) {
        var rsp = gurl({
            H : [
                config.H[0],
                "X-Number:" + xnumber,
                "session-id:" + sessionId,
            ],

            MF : [
                "voice=" + gurl_copy(all, i, i + step), 
            ]
        });

        if (rsp.err != "") {
            console.log("rsp error is " + rsp.err)
        }

        if (rsp.http_code === 200) {
            gurl_format_json(rsp.http_body)
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
            "voice=" + gurl_copy(all, i, i + step), 
        ]
    });
}

for (var fname in files) {
    console.log("---->" + files[fname])
    slice_one(files[fname])
}
