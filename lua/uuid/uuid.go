package uuid

import (
	"github.com/satori/go.uuid"
	"github.com/yuin/gopher-lua"
)

func newv1(L *lua.LState) int {
	u1 := uuid.Must(uuid.NewV1())
	L.Push(lua.LString(u1.String()))
	return 1
}

/*
func newv2(L *lua.LState) int {
	u1 := uuid.Must(uuid.NewV2())
	L.Push(lua.LString(u1.String()))
	return 1
}
*/

/*
func newv3(L *lua.LState) int {
	u1 := uuid.Must(uuid.NewV3())
	L.Push(lua.LString(u1.String()))
	return 1
}
*/

func newv4(L *lua.LState) int {
	u1 := uuid.Must(uuid.NewV4())
	L.Push(lua.LString(u1.String()))
	return 1
}

/*
func newv5(L *lua.LState) int {
	u1 := uuid.Must(uuid.NewV5())
	L.Push(lua.LString(u1.String()))
	return 1
}
*/

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"newv1": newv1,
		//"newv2": newv2,
		//"newv3": newv3,
		"newv4": newv4,
		//"newv5": newv5,
	})

	L.Push(mod)
	return 1
}
