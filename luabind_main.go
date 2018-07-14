package main

import (
	"github.com/guonaihong/conn/lua/cmdparse"
	"github.com/guonaihong/conn/lua/socket"
	myhttp "github.com/guonaihong/gurl/lua/http"
	"github.com/yuin/gopher-lua"
	"net/http"
)

type LuaEngine struct {
	L *lua.LState
}

func NewLuaEngine(client *http.Client) *LuaEngine {
	L := lua.NewState()
	engine := &LuaEngine{L: L}
	L.PreloadModule("socket", socket.Loader)
	L.PreloadModule("cmd", cmdparse.Loader)
	L.PreloadModule("http", myhttp.New(client).Loader)
	//socket.RegisterSocketType(L)
	return engine
}
