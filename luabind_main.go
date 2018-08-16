package main

import (
	myhttp "github.com/guonaihong/gurl/lua/http"

	"github.com/guonaihong/glua/lib/cmdparse"
	"github.com/guonaihong/glua/lib/json"
	"github.com/guonaihong/glua/lib/log"
	"github.com/guonaihong/glua/lib/strings"
	"github.com/guonaihong/glua/lib/time"
	"github.com/guonaihong/glua/lib/uuid"
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
	L.PreloadModule("log", log.Loader)
	L.PreloadModule("strings", strings.Loader)
	//socket.RegisterSocketType(L)
	return engine
}
