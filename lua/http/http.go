package http

import (
	_ "fmt"
	"github.com/guonaihong/gurl/gurlib"
	"github.com/guonaihong/gurl/gurlib/url"
	"github.com/yuin/gopher-lua"
	"net/http"
)

type HTTP struct {
	*http.Client
}

func New(client *http.Client) *HTTP {
	h := &HTTP{}
	h.Client = client
	return h
}

func (h *HTTP) send(L *lua.LState) int {

	g := gurlib.Gurl{}
	reqArgs := L.ToTable(1)

	/*
		reqArgs.ForEach(func(k lua.LValue, value lua.LValue) {
			fmt.Printf("k:(%#v) #### v:%#v\n", k, value)
		})
	*/

	header := reqArgs.RawGet(lua.LString("H"))
	mf := reqArgs.RawGet(lua.LString("MF"))
	f := reqArgs.RawGet(lua.LString("F"))
	urlStr := reqArgs.RawGet(lua.LString("url"))
	o := reqArgs.RawGet(lua.LString("o"))
	method := reqArgs.RawGet(lua.LString("X"))
	body := reqArgs.RawGet(lua.LString("body"))

	switch reqUrl := urlStr.(type) {
	case lua.LString:
		g.Url = url.ModifyUrl(reqUrl.String())
	}

	switch reqMethod := method.(type) {
	case lua.LString:
		g.Method = reqMethod.String()
	}

	switch reqO := o.(type) {
	case lua.LString:
		g.O = reqO.String()
	}

	switch reqBody := body.(type) {
	case lua.LString:
		g.Body = []byte(reqBody.String())
	}

	switch reqH := header.(type) {
	case *lua.LTable:
		var gH []string
		reqH.ForEach(func(_ lua.LValue, value lua.LValue) {
			gH = append(gH, value.String())
		})
		g.H = gH
	}

	switch reqMF := mf.(type) {
	case *lua.LTable:
		reqMF.ForEach(func(_ lua.LValue, value lua.LValue) {
			gurlib.ParseMF(value.String(), &g.FormCache)
		})
	}

	switch reqF := f.(type) {
	case *lua.LTable:
		var fs []string
		reqF.ForEach(func(_ lua.LValue, value lua.LValue) {
			fs = append(fs, value.String())
		})
		g.F = fs
	}
	g.ParseInit()

	rsp, err := g.SendExec(h.Client)
	if err != nil {
		L.ArgError(1, "http:send expected:"+err.Error())
		return 0
	}

	tb := L.CreateTable(0, 3)

	tb.RawSetH(lua.LString("status_code"), lua.LNumber(rsp.StatusCode))
	tb.RawSetH(lua.LString("body"), lua.LString(rsp.Body))
	tb.RawSetH(lua.LString("err"), lua.LString(rsp.Err))

	L.Push(tb)
	return 1
}

func (h *HTTP) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"send": h.send,
	})

	//RegisterHTTPType(mod, L)
	L.Push(mod)
	return 1
}
