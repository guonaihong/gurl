package gurlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func Cmd2Js(g *Gurl) {
	buf := bytes.Buffer{}

	buf.WriteString("var cmd = ")

	all, err := json.MarshalIndent(&g.GurlCore, "", "   ")
	if err != nil {
		fmt.Printf("encode fail:%s\n", err)
		return
	}

	buf.Write(all)
	buf.WriteString(";\n")
	buf.WriteString("gurl(cmd);\n")
	io.Copy(os.Stdout, &buf)
}
