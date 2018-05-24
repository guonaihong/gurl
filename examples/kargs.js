var o = gurl_flag_parse(gurl_args,
        ["f", "", "pcm file"],
        ["dir", "", "open dir"]); 

if (!o.hasOwnProperty("f") || !o.hasOwnProperty("dir")) {

    console.log("not found f, dir");
}

console.log("-->" + o.f);
console.log("-->" + o.dir);

