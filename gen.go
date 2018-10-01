package gurl

import (
	"bytes"
	"fmt"
	"github.com/guonaihong/gurl/gurlib"
	"github.com/yuin/gopher-lua"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func Cmd2Lua(g *gurlib.Gurl) {
	var out bytes.Buffer
	cmd := `
local cmd = require("cmd")
local http = require("http")
local flag = cmd.new()
local opt = flag
            :opt_str("url", "", "Remote service address")
            :parse(get_cmd_args())

`
	out.WriteString(cmd)
	out.WriteString("http_data = {\n")
	out.WriteString("    H = {\n")
	for _, v := range g.H {
		out.WriteString(`        "` + v + `",` + "\n")
	}
	out.WriteString("    },\n")

	out.WriteString("    F = {\n")
	for _, v := range g.F {
		out.WriteString(`        "` + v + `",` + "\n")
	}
	out.WriteString("    },\n")

	out.WriteString("    url = \"" + g.Url + "\"\n")

	out.WriteString("}\n")

	out.WriteString(`
if #opt["url"] ~= 0 then
    http_data.url = opt.url
end
`)
	out.WriteString("\n" + `http.send(http_data)` + "\n")
	io.Copy(os.Stdout, &out)
}

func Lua2Cmd(conf string, kargs string) {
	all, err := ioutil.ReadFile(conf)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	L := NewLuaEngine(&http.Client{}, kargs)

	err = L.L.DoString(string(all))
	if err != nil {
		//fmt.Printf("lua2cmd:%s\n", err)
		//return
	}

	buf := &bytes.Buffer{}
	buf.WriteString("gurl")

	var tb *lua.LTable
	hd := L.L.GetGlobal("http_data")
	switch tb1 := hd.(type) {
	case *lua.LTable:
		tb = tb1
	}

	if tb == nil {
		return
	}

	tb.ForEach(func(k lua.LValue, v lua.LValue) {
		switch strings.ToLower(k.String()) {
		case "h":
			switch reqH := v.(type) {
			case *lua.LTable:
				reqH.ForEach(func(_ lua.LValue, value lua.LValue) {
					buf.WriteString(" -H " + value.String())
				})
			}
		case "method":
			buf.WriteString(" -X " + v.String())
		case "f":
			switch reqF := v.(type) {
			case *lua.LTable:
				reqF.ForEach(func(_ lua.LValue, value lua.LValue) {
					buf.WriteString(" -F " + value.String())
				})
			}

		case "url":
			buf.WriteString(" -url " + v.String())

		case "o":
			buf.WriteString(" -o " + v.String())

		case "jfa":
			switch jfa := v.(type) {
			case *lua.LTable:
				jfa.ForEach(func(_ lua.LValue, value lua.LValue) {
					buf.WriteString(" -Jfa " + value.String())
				})
			}

		}

	})

	buf.WriteString("\n")
	io.Copy(os.Stdout, buf)
}
