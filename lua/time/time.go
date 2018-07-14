package time

import (
	"github.com/yuin/gopher-lua"
	gotime "time"
)

func sleep(L *lua.LState) int {
	n1 := L.CheckInt64(1)
	t := L.CheckString(2)

	n := gotime.Duration(n1)
	switch t {
	case "ms":
		n *= gotime.Millisecond
	case "s":
		n *= gotime.Second
	case "m":
		n *= gotime.Minute
	case "h":
		n *= gotime.Hour
	}
	gotime.Sleep(n)
	return 1
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"sleep": sleep,
	})

	L.Push(mod)
	return 1
}
