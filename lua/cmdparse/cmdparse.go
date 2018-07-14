package cmdparse

import (
	"github.com/yuin/gopher-lua"
)

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"new": newCmdParse,
	})

	RegisterCmdParseType(mod, L)
	L.Push(mod)
	return 1
}
