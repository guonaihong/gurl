package http

import (
	"bytes"
	"encoding/binary"
	"github.com/yuin/gopher-lua"
	"io"
	"net"
	_ "strconv"
)

type HTTP struct {
	*http.Client
}

const luaHTTPTypeName = "http"

func RegisterHTTPType(module *lua.LTable, L *lua.LState) {
	mt := L.NewTypeMetatable(luaHTTPTypeName)
	//L.SetGlobal("http", mt)
	//L.SetField(mt, "new", L.NewFunction(newHTTP))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), httpMethods))
}

func newHTTP(L *lua.LState) int {
	http := &HTTP{}
	ud := L.NewUserData()
	ud.Value = http
	L.SetMetatable(ud, L.GetTypeMetatable(luaHTTPTypeName))
	L.Push(ud)
	return 1
}

func connect(L *lua.LState) int {
	s := checkHTTP(L)

	addr := L.CheckString(2)
	var err error
	s.Conn, err = net.Dial("tcp", addr)
	if err != nil {
		L.ArgError(1, err.Error())
		return 0
	}

	return 1
}

func httpClose(L *lua.LState) int {
	s := checkHTTP(L)
	s.Conn.Close()
	return 1
}

var httpMethods = map[string]lua.LGFunction{
	"connect": connect,
	"close":   httpClose,
}

func checkHTTP(L *lua.LState) *HTTP {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*HTTP); ok {
		return v
	}

	L.ArgError(1, "http expected")
	return nil
}
