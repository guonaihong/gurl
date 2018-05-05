package gurlib

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestExecSlice(t *testing.T) {
	handle := func(w http.ResponseWriter, r *http.Request) {
		io.Copy(os.Stdout, r.Body)
	}

	server := httptest.NewServer(http.HandlerFunc(handle))

	ExecSlice([]string{
		"gurl",
		"-H",
		"appkey:www",
		"-H",
		"session:sid",
		"-F",
		"text=good",
		"-url",
		server.URL,
	})
}
