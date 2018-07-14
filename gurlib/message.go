package gurlib

import (
	"github.com/yuin/gopher-lua"
)

type Message struct {
	In  chan lua.LValue
	Out chan lua.LValue
}
