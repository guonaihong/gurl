package cond

import (
	"core"
)

type For struct {
	Range string `json:"range"`
	K     string `json:"k"`
	V     string `json:"v"`
	core.Base
}
