package lib

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"testing"
)

func TestRegister(t *testing.T) {
	vm := otto.New()

	Register(vm)

	_, err := vm.Run(`
	all = gurl_readfile("./slice.js");

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
