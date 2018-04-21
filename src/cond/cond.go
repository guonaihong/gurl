package cond

import (
	"core"
)

type For struct {
	Range     string `json:"range,omitempty"`
	K         string `json:"k,omitempty"`
	V         string `json:"v,omitempty"`
	core.Base `json:"-"`
}
