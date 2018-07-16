package gurlib

import (
	"bytes"
	"io"
	"os"
)

func Cmd2Lua(g *Gurl) {
	var out bytes.Buffer
	cmd := `
local cmd = require("cmd")
local http = require("http")
local flag = cmd.new()
local opt = flag
            :opt_str("url", "", "Remote service address")
            :parse(gurl_cmd)

`
	out.WriteString(cmd)
	out.WriteString("local http_data = {\n")
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
print(http_data.url.."###"..opt["url"])
`)
	out.WriteString("\n" + `http.send(http_data)` + "\n")
	io.Copy(os.Stdout, &out)
}

func Lua2Cmd(conf string) {
}
