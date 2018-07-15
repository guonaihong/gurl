package json

import (
	"bytes"
	gojson "encoding/json"
	"github.com/yuin/gopher-lua"
)

func format(L *lua.LState) int {
	j := L.CheckString(1)
	out := &bytes.Buffer{}
	err := gojson.Indent(out, []byte(j), "", "  ")
	if err != nil {
		L.ArgError(1, "json expected: "+err.Error())
		return 0
	}

	L.Push(lua.LString(out.String()))
	return 1
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"format": format,
	})

	L.Push(mod)
	return 1
}
