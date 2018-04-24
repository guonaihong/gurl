package gurlib

import (
	"testing"
)

func TestExecSlice2(t *testing.T) {
	rsp, err := ExecSlice([]string{
		"gurl",
		"-X",
		"POST",
		"-F",
		"text=good",
		"-o",
		"tst.log",
		"http://127.0.0.1:5002",
	})

	t.Logf("err(%v), rsp(%#v)\n", err, rsp)
}

func TestExecSlice(t *testing.T) {
	rsp, err := ExecSlice([]string{
		"gurl",
		"-F",
		"text=good",
		"http://127.0.0.1:5002",
	})

	t.Logf("err(%v), rsp(%#v)\n", err, rsp)

}
