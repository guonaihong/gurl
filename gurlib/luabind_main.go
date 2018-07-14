package gurlib

import (
	"github.com/guonaihong/conn/lua/cmdparse"
	"github.com/guonaihong/conn/lua/socket"
	"github.com/yuin/gopher-lua"
)

type LuaEngine struct {
	L *lua.LState
}

func NewLuaEngine() *LuaEngine {
	L := lua.NewState()
	engine := &LuaEngine{L: L}
	L.PreloadModule("socket", socket.Loader)
	L.PreloadModule("cmd", cmdparse.Loader)
	//socket.RegisterSocketType(L)
	return engine
}
