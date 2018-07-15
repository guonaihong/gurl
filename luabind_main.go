package main

import (
	"github.com/guonaihong/gurl/lua/cmdparse"
	myhttp "github.com/guonaihong/gurl/lua/http"
	"github.com/guonaihong/gurl/lua/json"
	"github.com/guonaihong/gurl/lua/time"
	"github.com/guonaihong/gurl/lua/uuid"
	"github.com/yuin/gopher-lua"
	"net/http"
)

type LuaEngine struct {
	L *lua.LState
}

func NewLuaEngine(client *http.Client) *LuaEngine {
	L := lua.NewState()
	engine := &LuaEngine{L: L}
	L.PreloadModule("cmd", cmdparse.Loader)
	L.PreloadModule("http", myhttp.New(client).Loader)
	L.PreloadModule("uuid", uuid.Loader)
	L.PreloadModule("time", time.Loader)
	L.PreloadModule("json", json.Loader)
	//socket.RegisterSocketType(L)
	return engine
}
