package gurlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func Cmd2Js(g *Gurl) {
	buf := bytes.Buffer{}
	buf.WriteString(`
var program = gurl_flag();
var flag = program
	.option("url", "", "Remote service address")
	.parse()

`)
	buf.WriteString("var cmd = ")

	all, err := json.MarshalIndent(&g.GurlCore, "", "    ")
	if err != nil {
		fmt.Printf("encode fail:%s\n", err)
		return
	}

	buf.Write(all)
	buf.WriteString(";\n")
	buf.WriteString(`
if (flag.url.length > 0) {
	cmd.url = flag.url;
}

var http = gurl_http();
var rsp  = http.send(cmd);
console.log(rsp.body);
`)
	io.Copy(os.Stdout, &buf)
}

func joinCmdOpt(m map[string]interface{}) {
	buf := &bytes.Buffer{}

	buf.WriteString("gurl")
	for k, v := range m {
		switch strings.ToLower(k) {
		case "h":
			h, ok := v.([]string)
			if ok {
				for _, v := range h {
					buf.WriteString(" -H " + v)
				}
			}
		case "method":
			method, ok := v.(string)
			if ok {
				buf.WriteString(" -X " + method)
			}
		case "f":
			f, ok := v.([]string)
			if ok {
				for _, v := range f {
					buf.WriteString(" -F " + v)
				}
			}

		case "url":
			url, ok := v.(string)
			if ok {
				buf.WriteString(" -url " + url)
			}

		case "o":
			o, ok := v.(string)
			if ok {
				buf.WriteString(" -o " + o)
			}

		case "jfa":
			jia, ok := v.([]string)
			if ok {
				for _, v := range jia {
					buf.WriteString(" -Jfa " + v)
				}
			}
		}
	}

	buf.WriteString("\n")
	io.Copy(os.Stdout, buf)
}

func Js2Cmd(conf string) {
	all, err := ioutil.ReadFile(conf)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	vm := otto.New()

	vm.Run(string(all))
	if value, err := vm.Get("cmd"); err == nil {
		if o, err := value.Export(); err == nil {
			if m, ok := o.(map[string]interface{}); ok {
				//fmt.Printf("", s, err)
				joinCmdOpt(m)
			}
		}
	}
}
