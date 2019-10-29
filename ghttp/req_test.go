package ghttp

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TODO 更多功能测试代码
type testHeader struct {
	H1 string `header:"h1"`
	H2 string `header:"h2"`
}

func Test_Req_Header(t *testing.T) {
	router := func() *gin.Engine {
		router := gin.Default()

		need := testHeader{H1: "v1", H2: "v2"}
		got := testHeader{}

		router.GET("/test.header", func(c *gin.Context) {
			err := c.ShouldBindHeader(&got)
			assert.NoError(t, err)
			assert.Equal(t, need, got)
		})

		return router
	}()

	ts := httptest.NewServer(http.HandlerFunc(router.ServeHTTP))

	type testHeader struct {
		Sid  string `header:"sid"`
		Code int
	}

	g := Gurl{Client: http.DefaultClient}

	g.Color = true
	g.GurlCore.Method = "GET"
	g.GurlCore.Header = []string{"h1:v1", "h2:v2"}
	g.GurlCore.Url = ts.URL + "/test.header"

	rsp, err := g.Send()

	assert.NoError(t, err)
	assert.Equal(t, rsp.StatusCode, 200)
}
