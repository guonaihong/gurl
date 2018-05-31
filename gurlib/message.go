package gurlib

import (
	"fmt"
	"github.com/robertkrimen/otto"
)

type MessageChan struct {
	P chan map[string]interface{}
	C chan map[string]interface{}
}

type Message struct {
	*MessageChan
	js *JsEngine
}

func MessageNew(message *MessageChan) *Message {
	return &Message{
		MessageChan: message,
	}
}

func MessageChanNew() *MessageChan {
	return &MessageChan{
		P: make(chan map[string]interface{}, 10),
		C: make(chan map[string]interface{}, 10),
	}
}

func (m *MessageChan) Close() {
	close(m.P)
	close(m.C)
}

func (m *Message) Read(call otto.FunctionCall) otto.Value {
	v := <-m.C
	result, err := m.js.VM.ToValue(v)
	if err != nil {
		fmt.Printf("-->err:%s\n", err)
		return otto.Value{}
	}

	return result
}

func (m *Message) Write(call otto.FunctionCall) otto.Value {
	o, err := call.Argument(0).Export()
	if err != nil {
		return otto.Value{}
	}

	v, ok := o.(map[string]interface{})
	if ok {
		m.P <- v
	}
	return otto.Value{}
}
