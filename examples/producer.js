
var message = gurl_message();

for (var i = 0; i < 1000; i++) {
    message.write({
        "file": (i + ""),
        //"file": String(i + "").repeat(10),
    })
}
