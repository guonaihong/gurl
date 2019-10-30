package ghttp

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"syscall"
	"testing"
)

func Test_Ghttp_CmdErr(t *testing.T) {
	_, err := http.Get("http://127.0.0.1:3333")
	assert.True(t, errors.Is(err, syscall.ECONNREFUSED))
}
