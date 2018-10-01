package gurl

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
	L    *lua.LState
	args string
}

func (l *LuaEngine) getCmdArgs(L *lua.LState) int {
	l.L.Push(lua.LString(l.args)) /* push result */
	return 1
}

func NewLuaEngine(client *http.Client, kargs string) *LuaEngine {
	L := lua.NewState()
	engine := &LuaEngine{L: L, args: kargs}
	L.PreloadModule("flag", cmdparse.Loader)
	L.PreloadModule("http", myhttp.New(client).Loader)
	L.PreloadModule("uuid", uuid.Loader)
	L.PreloadModule("time", time.Loader)
	L.PreloadModule("json", json.Loader)
	L.PreloadModule("log", log.Loader)
	L.PreloadModule("strings", strings.Loader)

	L.SetGlobal("get_cmd_args", L.NewFunction(engine.getCmdArgs))
	//socket.RegisterSocketType(L)
	return engine
}
