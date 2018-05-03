package lib

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"strings"
	"time"
)

/*
	vm.Set("gurl_readfile", func(call otto.FunctionCall) otto.Value {

		o, err := call.Argument(0).Export()
		if err != nil {
			fmt.Printf("err:%v\n", o)
		}

		//m := o.(map[string]interface{})
		fmt.Printf("Hello, %s.\n", o)

		result, _ := otto.ToValue("hello")
		return result
	})
*/
func JsReadFile(call otto.FunctionCall) otto.Value {
	f := call.Argument(0).String()

	all, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err.Error())
	}

	result, _ := otto.ToValue(string(all))
	return result
}

func JsLen(call otto.FunctionCall) otto.Value {
	a := call.Argument(0).String()

	result, _ := otto.ToValue(len(a))
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
	return otto.Value{}
}

func Register(vm *otto.Otto) {
	vm.Set("gurl_readfile", JsReadFile)
	vm.Set("gurl_len", JsLen)
	vm.Set("gurl_sleep", JsSleep)
	//vm.Set("gurl_uuid", JsUUID)
}
