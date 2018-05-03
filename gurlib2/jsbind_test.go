package gurlib2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestNewJsEngine(t *testing.T) {
	js := NewJsEngine(&http.Client{})

	all, err := ioutil.ReadFile("slice.js")
	if err != nil {
		panic(err.Error())
	}

	_, err = js.VM.Run(string(all))
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
	all = gurl_readfile("./slice.js");

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
