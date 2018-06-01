var program = gurl_flag();
var flag = program
   .option("l, list", "", "list file")
   .parse()

if (!flag.hasOwnProperty("l")) {
    flag.Usage();
    gurl.exit(1)
}

var all = gurl_readfile(flag.list)
var lines = all.split("\n");
var message = gurl_message();

for (var l in lines) {
    if (lines[l].trim().length == 0) {
        continue;
    }
    
    var body = [];
    if (lines[l].indexOf("\t") != -1) {
	body = lines[l].split("\t");
    } else if (lines[l].indexOf("##") != -1) {
	body = lines[l].split("##");
    }

    //console.log("--->", body.join("###"));
    message.write({
        "voice":body[0],
        "text":body[body.length-1],
    });

    //console.log("filename->", body[0], "text->", body[body.length-1]);
}
