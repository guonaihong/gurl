package conn

import (
	"github.com/guonaihong/glua/lib/cmdparse"
	"github.com/guonaihong/glua/lib/json"
	"github.com/guonaihong/glua/lib/log"
	"github.com/guonaihong/glua/lib/socket"
	"github.com/guonaihong/glua/lib/strings"
	"github.com/guonaihong/glua/lib/time"
	"github.com/guonaihong/glua/lib/uuid"
	"github.com/yuin/gopher-lua"
)

type LuaEngine struct {
	L    *lua.LState
	args string
}

func (l *LuaEngine) getCmdArgs(L *lua.LState) int {
	l.L.Push(lua.LString(l.args)) /* push result */
	return 1
}

func NewLuaEngine(kargs string) *LuaEngine {
	L := lua.NewState()
	engine := &LuaEngine{L: L, args: kargs}
	L.PreloadModule("socket", socket.Loader)
	L.PreloadModule("flag", cmdparse.Loader)
	L.PreloadModule("time", time.Loader)
	L.PreloadModule("strings", strings.Loader)
	L.PreloadModule("log", log.Loader)
	L.PreloadModule("uuid", uuid.Loader)
	L.PreloadModule("json", json.Loader)

	L.SetGlobal("get_cmd_args", L.NewFunction(engine.getCmdArgs))
	//socket.RegisterSocketType(L)
	return engine
}
