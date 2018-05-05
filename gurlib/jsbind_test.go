package gurlib

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewJsEngine(t *testing.T) {
	js := NewJsEngine(&http.Client{})

	all, err := ioutil.ReadFile("test.js")
	if err != nil {
		panic(err.Error())
	}

	_, err = js.VM.Run(string(all))
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func TestJsExtract(t *testing.T) {
	js := NewJsEngine(nil)
	_, err := js.VM.Run(`
	var str = "123456";
	var s2 = gurl_extract(str, 0, 3);
	console.log(s2);
	`)

	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func TestJsGurl(t *testing.T) {
	handle := func(w http.ResponseWriter, r *http.Request) {
		io.Copy(os.Stdout, r.Body)
	}

	// server.URL
	server := httptest.NewServer(http.HandlerFunc(handle))
	js := NewJsEngine(&http.Client{})

	fmt.Printf("url ->%s\n", server.URL)
	s := fmt.Sprintf("var url = \"%s\";\n", server.URL)
	s1 := `
var xnumber = 0;
var sessionId = "12342";
var rsp = gurl({
	H : [
	"X-Number:" + xnumber,
	"session-id:" + sessionId,
	],

	MF : [
	"voice=" + gurl_extract("good", 0,  4), 
	],

	url: url
});

console.log("http status code:" + rsp.status_code);
console.log("http body size:" + gurl_len(rsp.body));
console.log("http err:" + rsp.err);
	`

	//fmt.Printf("%s\n", s+s1)

	_, err := js.VM.Run(s + s1)

	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func TestRegister(t *testing.T) {

	c := http.Client{}
	js := NewJsEngine(&c)
	_, err := js.VM.Run(`
	sessionId = "sid";
	xnumber = 0;
	var rsp = gurl({
		H : [
		"X-Number:" + xnumber,
		"session-id:" + sessionId,
		],

		/*
		MF : [
		"voice=" + gurl_copy(all, i, i + 4096), 
		]
		*/
	});
	all = gurl_readfile("./test.js");

	console.log(gurl_uuid())
	gurl_sleep("250ms")
	var xnumber = 0;
	for (var i = 0, l = gurl_len(all); i < l; i += 4096) {
		console.log(l)
	}

	var config = {
		H : [
			"appkey:xuqo7pqagqx5gvdbqyfybrusfosbbkjjtfvsr5qx"
		],
		url:'http://192.168.6.128:24987/asr/opus'
	};

	// abc = 2 + 2;
	// console.log("The value of abc is " + abc); // 4
	var files = [
		"good.pcm",
		"good.pcm",
		"good.pcm",
		"good.pcm"
	];

	for (var fname in files) {
		console.log(files[fname]);
	}

	console.log(config.url);
	`)

	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
