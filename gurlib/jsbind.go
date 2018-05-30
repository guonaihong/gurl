package gurlib

import (
	"bytes"
	"encoding/json"
	//"flag"
	"fmt"
	"github.com/guonaihong/flag"
	"github.com/robertkrimen/otto"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type JsEngine struct {
	VM *otto.Otto
	c  *http.Client
}

func NewJsEngine(c *http.Client) *JsEngine {
	js := &JsEngine{
		VM: otto.New(),
		c:  c,
	}

	register(js.VM, js)
	return js
}

type GurlHttp struct {
	js *JsEngine
}

func (j *JsEngine) GurlHttp(call otto.FunctionCall) otto.Value {

	h := GurlHttp{
		js: j,
	}

	m := map[string]interface{}{
		"send": h.Send,
	}

	result, _ := j.VM.ToValue(m)
	return result
}

func (h *GurlHttp) Send(call otto.FunctionCall) otto.Value {

	j := h.js

	o, err := call.Argument(0).Export()
	if err != nil {
		fmt.Printf("err:%v\n", o)
		return otto.Value{}
	}

	m := o.(map[string]interface{})

	g := Gurl{
		Client: j.c,
	}

	g.MemInit()
	for k, v := range m {
		switch strings.ToLower(k) {
		case "h":
			h, ok := v.([]string)
			if ok {
				g.H = h
			}
		case "method":
			method, ok := v.(string)
			if ok {
				g.Method = method
			}
		case "f":
			f, ok := v.([]string)
			if ok {
				form(f, &g.FormCache)
			}

		case "url":
			url, ok := v.(string)
			if ok {
				url = ModifyUrl(url)
				g.Url = url
			}

		case "o":
			o, ok := v.(string)
			if ok {
				g.O = o
			}

		case "d":
			body, ok := v.(string)
			if ok {
				g.Body = []byte(body)
				parseBody(&g.Body)
			}
		case "mf":
			mf, ok := v.([]string)
			if ok {
				formCache := []FormVal{}
				for _, v := range mf {

					parseMF(v, &formCache)
				}

				//fmt.Printf("--->%#v\n", formCache)
				g.GurlCore.FormCache =
					append(g.GurlCore.FormCache, formCache...)
			}
		}
	}

	rsp, _ := g.sendExec(j.c)
	for k := range m {
		delete(m, k)
	}

	m["status_code"] = rsp.StatusCode
	m["body"] = string(rsp.Body)
	m["err"] = rsp.Err

	result, err := j.VM.ToValue(m)
	if err != nil {
		fmt.Printf("--->err:%s\n", err)
	}
	return result
}

func JsReadFile(call otto.FunctionCall) otto.Value {
	f := call.Argument(0).String()

	all, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err.Error())
	}

	result, _ := otto.ToValue(string(all))
	return result
}

func JsSleep(call otto.FunctionCall) otto.Value {
	t := call.Argument(0).String()
	t = strings.TrimSpace(t)
	tv := 0
	company := time.Second
	companyStr := ""

	company = time.Second
	fmt.Sscanf(t, "%d%s", &tv, &companyStr)
	switch companyStr {
	case "ms":
		company = time.Millisecond
	case "s":
		company = time.Second
	case "m":
		company = time.Minute
	case "h":
		company = time.Hour
	}

	time.Sleep(time.Duration(tv) * company)
	return otto.Value{}
}

func JsUUID(call otto.FunctionCall) otto.Value {
	u1 := uuid.Must(uuid.NewV4())
	result, _ := otto.ToValue(u1.String())
	return result
}

func JsExit(call otto.FunctionCall) otto.Value {
	code, err := call.Argument(0).ToInteger()
	if err != nil {
		return otto.Value{}
	}

	os.Exit(int(code))
	return otto.Value{}
}

func JsFjson(call otto.FunctionCall) otto.Value {
	all := call.Argument(0).String()

	out := &bytes.Buffer{}
	err := json.Indent(out, []byte(all), "", "  ")
	if err != nil {
		return call.Argument(0)
	}

	result, _ := otto.ToValue(out.String())
	return result
}

type opt struct {
	resultArgs []*string
	optName    []string
}

type GurlFlag struct {
	commandlLine *flag.FlagSet
	js           *JsEngine
	opt
	cmd []string
	set bool
}

func (j *JsEngine) GurlFlag(call otto.FunctionCall) otto.Value {
	f := GurlFlag{
		js: j,
	}

	m := map[string]interface{}{
		"option": f.Option,
		"parse":  f.Parse,
		"usage":  f.Usage,
	}

	result, _ := j.VM.ToValue(m)
	return result
}

func (f *GurlFlag) Option(call otto.FunctionCall) otto.Value {
	name := call.Argument(0).String()
	value := call.Argument(1).String()
	usage := call.Argument(2).String()

	if f.set == false {
		if original, err := f.js.VM.Get("gurl_args"); err == nil {
			cmd := strings.Split(original.String(), " ")
			f.cmd = cmd
		}

		f.set = true
	}

	if f.commandlLine == nil {
		f.commandlLine = flag.NewFlagSet(f.cmd[0], flag.ExitOnError)
	}

	outValue := ""
	f.commandlLine.StringVar(&outValue, name, value, usage)

	f.resultArgs = append(f.resultArgs, &outValue)
	f.optName = append(f.optName, name)

	m := map[string]interface{}{
		"option": f.Option,
		"parse":  f.Parse,
		"usage":  f.Usage,
	}

	result, _ := f.js.VM.ToValue(m)
	return result
}

func (f *GurlFlag) Usage() {
	f.commandlLine.Usage()
}

func (f *GurlFlag) Parse() otto.Value {
	m := map[string]interface{}{}
	f.commandlLine.Parse(f.cmd[1:])

	for k, v := range f.resultArgs {
		if pos := strings.Index(f.optName[k], ","); pos != -1 {
			ks := strings.Split(f.optName[k], ",")
			for _, kv := range ks {
				m[strings.TrimSpace(kv)] = *v
			}
		}
		m[f.optName[k]] = *v
	}

	result, err := f.js.VM.ToValue(m)
	if err != nil {
		fmt.Printf("-->err:%s\n", err)
	}
	return result
}

func register(vm *otto.Otto, js *JsEngine) {
	vm.Set("gurl_readfile", JsReadFile)
	vm.Set("gurl_sleep", JsSleep)
	vm.Set("gurl_uuid", JsUUID)
	vm.Set("gurl_fjson", JsFjson)
	vm.Set("gurl_exit", JsExit)
	vm.Set("gurl_http", js.GurlHttp)
	vm.Set("gurl_flag", js.GurlFlag)
}
