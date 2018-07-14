package http

import (
	"github.com/guonaihong/gurl/gurlib"
	"github.com/yuin/gopher-lua"
	"net/http"
)

func New(client *http.Client) *HTTP {
	h := &HTTP{}
	h.Client = client
	return h
}

func (h *HTTP) send(L *lua.LState) int {

	reqArgs := L.ToTable(1)

	h := regArgs.RawGet(lua.LString("H"))
	mf := reqArgs.RawGet(lua.LString("MF"))
	url := reqArgs.RawGet(lua.LString("url"))
	o := reqArgs.RawGet(lua.LString("o"))
	method := reqArgs.RawGet(lua.LString("X"))
	body := reqArgs.RawGet(lua.LString("body"))

	switch reqUrl := url.(type) {
	case lua.LString:
		url = ModifyUrl(url)
		g.Url = url
	}

	switch reqMethod := method.(type) {
	case lua.LString:
		g.Method = reqMethod
	}

	switch reqO := o.(type) {
	case lua.LString:
		g.O = reqO
	}

	switch reqBody := body.(type) {
	case lua.LString:
		g.Body = []byte(body)
	}

	switch reqH := h.(type) {
	case *lua.LTable:
		var gH []string
		reqH.ForEach(func(_ lua.LValue, value lua.LValue) {
			gH = append(gH, value.String())
		})
		g.H = gH
	}

	switch reqMF := h.(type) {
	case *lua.LTable:
		var gMF []string
		reqMF.ForEach(func(_ lua.LValue, value lua.LValue) {
			gMF = append(gMF, value.String())
		})
	}
	g := gurlib.Gurl{}

	g.MemInit()

	rsp, _ := g.sendExec(h.Client)

	tb := L.CreateTable(0, len(m))

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

	RegisterHTTPType(mod, L)
	L.Push(mod)
	return 1
}
