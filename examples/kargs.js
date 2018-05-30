var program = gurl_flag();

var o = program
  .option("f, file", "", "pcm file")
  .option("d, dir", "", "open dir")
  .parse()

if (!o.hasOwnProperty("f") || !o.hasOwnProperty("dir")) {

    console.log("not found f, dir");
}

console.log("-->" + o.f);
console.log("-->" + o.dir);

