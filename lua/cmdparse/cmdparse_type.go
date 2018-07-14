package cmdparse

import (
	_ "fmt"
	"github.com/guonaihong/flag"
	"github.com/yuin/gopher-lua"
	"strconv"
	"strings"
)

type CmdParse struct {
	commandlLine *flag.FlagSet
	opt          map[string]interface{}
}

const luaCmdParseTypeName = "cmd_parse"

func RegisterCmdParseType(module *lua.LTable, L *lua.LState) {
	mt := L.NewTypeMetatable(luaCmdParseTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), cmdParseMethods))
}

func newCmdParse(L *lua.LState) int {
	cmd := &CmdParse{}
	ud := L.NewUserData()
	ud.Value = cmd
	L.SetMetatable(ud, L.GetTypeMetatable(luaCmdParseTypeName))
	L.Push(ud)
	return 1
}

func (c *CmdParse) init() {
	if c.commandlLine == nil {
		c.commandlLine = flag.NewFlagSet("", flag.ExitOnError)
	}
	if c.opt == nil {
		c.opt = make(map[string]interface{}, 1)
	}
}

func (c *CmdParse) returnThis(L *lua.LState) {
	ud := L.NewUserData()
	ud.Value = c
	L.SetMetatable(ud, L.GetTypeMetatable(luaCmdParseTypeName))
	L.Push(ud)
}

func optStr(L *lua.LState) int {
	cmdParse := checkCmdParse(L)
	name := L.CheckString(2)
	value := L.CheckString(3)
	usage := L.CheckString(4)

	cmdParse.init()

	outValue := ""
	cmdParse.commandlLine.StringVar(&outValue, name, value, usage)

	cmdParse.opt[name] = &outValue
	cmdParse.returnThis(L)
	return 1
}

func optInt(L *lua.LState) int {
	cmdParse := checkCmdParse(L)
	name := L.CheckString(2)
	value := L.CheckString(3)
	usage := L.CheckString(4)

	cmdParse.init()

	var err error
	valueInt, err := strconv.Atoi(value)

	if err != nil {
		return 1
	}

	outValue := 0
	cmdParse.commandlLine.IntVar(&outValue, name, valueInt, usage)

	cmdParse.opt[name] = &outValue
	cmdParse.returnThis(L)
	return 1
}

func parse(L *lua.LState) int {
	//TODO set argv[0] to commandline
	cmdParse := checkCmdParse(L)
	name := L.CheckString(2)

	opts := strings.Fields(name)
	m := map[string]interface{}{}
	cmdParse.commandlLine.Parse(opts)

	for k, v := range cmdParse.opt {
		if pos := strings.Index(k, ","); pos != -1 {
			ks := strings.Split(k, ",")
			for _, kv := range ks {
				m[strings.TrimSpace(kv)] = v
			}
		}
	}

	tb := L.CreateTable(0, len(m))
	for k, v := range m {
		if sp, ok := v.(*string); ok {
			tb.RawSetH(lua.LString(k), lua.LString(*sp))
		} else if ip, ok := v.(*int); ok {
			tb.RawSetH(lua.LString(k), lua.LNumber(*ip))
		}
	}
	L.Push(tb)
	return 1
}

func usage(L *lua.LState) int {
	cmdParse := checkCmdParse(L)
	cmdParse.commandlLine.Usage()
	return 1
}

func version(L *lua.LState) int {
	cmdParse := checkCmdParse(L)
	version := L.CheckString(2)

	cmdParse.init()

	cmdParse.commandlLine.Version(version)

	cmdParse.returnThis(L)
	return 1
}

func author(L *lua.LState) int {
	cmdParse := checkCmdParse(L)
	author := L.CheckString(2)

	cmdParse.init()

	cmdParse.commandlLine.Author(author)

	cmdParse.returnThis(L)
	return 1
}

var cmdParseMethods = map[string]lua.LGFunction{
	"version": version,
	"author":  author,
	"parse":   parse,
	"opt_str": optStr,
	"opt_int": optInt,
	"usage":   usage,
}

func checkCmdParse(L *lua.LState) *CmdParse {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*CmdParse); ok {
		return v
	}

	L.ArgError(1, "cmdParse expected")
	return nil
}
