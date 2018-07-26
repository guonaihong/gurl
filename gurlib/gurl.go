package gurlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/NaihongGuo/flag"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FormVal struct {
	Tag   string
	Fname string
	Body  []byte
}

const ADD_LINE = 1 << 62

type GurlCore struct {
	Method string `json:"method,omitempty"`

	J   []string `json:"J,omitempty"`
	F   []string `json:"F,omitempty"`
	H   []string `json:"H,omitempty"` // http header
	Url string   `json:"url,omitempty"`
	O   string   `json:"o,omitempty"`

	Jfa []string `json:"Jfa,omitempty"`

	FormCache []FormVal `json:"-"`

	Body       []byte `json:"body,omitempty"`
	Flag       int
	V          bool `json:"-"`
	A          string
	NotParseAt bool
}

func parseVal(bodyJson map[string]interface{}, key, val string) {
	if val == "{}" {
		bodyJson[key] = map[string]interface{}{}
		return
	}

	f, err := strconv.ParseFloat(val, 0)
	if err == nil {
		bodyJson[key] = f
		return
	}

	i, err := strconv.ParseInt(val, 0, 0)
	if err == nil {
		bodyJson[key] = i
		return
	}

	b, err := strconv.ParseBool(val)
	if err == nil {
		bodyJson[key] = b
		return
	}

	bodyJson[key] = val
}

func parseVal2(bodyJson map[string]interface{}, key, val string) {
	bodyJson[key] = val
}

func toJson(J []string, bodyJson map[string]interface{}) {
	for _, v := range J {
		pos := strings.Index(v, ":")
		if pos == -1 {
			bodyJson[v] = ""
			continue
		}

		key := v[:pos]
		val := v[pos+1:]

		if pos := strings.Index(key, "."); pos != -1 {
			keys := strings.Split(key, ".")

			parseValCb := parseVal2
			if strings.HasPrefix(val, "=") {
				val = val[1:]
				parseValCb = parseVal
			}

			type jsonObj map[string]interface{}

			curMap := bodyJson

			for i, v := range keys {
				if len(keys)-1 == i {
					parseValCb(curMap, v, val)
					break
				}

				vv, ok := curMap[v]
				if !ok {
					vv = jsonObj{}
					curMap[v] = vv
				}

				curMap = vv.(jsonObj)

			}
			continue
		}

		if len(val) == 0 {
			bodyJson[key] = ""
			continue
		}

		if val[0] != '=' {
			bodyJson[key] = val
			continue
		}

		if len(key) == 1 {
			continue
		}

		val = val[1:]
		parseVal(bodyJson, key, val)

	}
}

func (g *GurlCore) form(F []string, fm *[]FormVal) {

	fileds := [2]string{}
	formVals := []FormVal{}

	for _, v := range F {

		fileds[0], fileds[1] = "", ""

		pos := strings.Index(v, "=")
		if pos == -1 {
			continue
		}

		fileds[0], fileds[1] = v[:pos], v[pos+1:]

		if !g.NotParseAt && strings.HasPrefix(fileds[1], "@") {
			fname := fileds[1][1:]

			fd, err := os.Open(fname)
			if err != nil {
				log.Fatalf("open file fail:%v\n", err)
			}

			body, err2 := ioutil.ReadAll(fd)
			if err != nil {
				log.Fatalf("read body fail:%v\n", err2)
			}

			formVals = append(formVals, FormVal{Tag: fileds[0], Fname: fname, Body: body})

			fd.Close()
		} else {
			formVals = append(formVals, FormVal{Tag: fileds[0], Body: []byte(fileds[1])})
		}

		//F[i] = fileds[0]
	}

	*fm = append(*fm, formVals...)
}

func jsonFromAppend(JF []string, fm *[]FormVal) {

	JFMap := map[string][]string{}
	fileds := [2]string{}
	formVals := []FormVal{}

	for _, v := range JF {

		fileds[0], fileds[1] = "", ""

		pos := strings.Index(v, "=")
		if pos == -1 {
			continue
		}

		fileds[0], fileds[1] = v[:pos], v[pos+1:]

		v, _ := JFMap[fileds[0]]
		JFMap[fileds[0]] = append(v, fileds[1])
	}

	for k, v := range JFMap {

		bodyJson := map[string]interface{}{}

		toJson(v, bodyJson)

		body, err := json.Marshal(&bodyJson)

		if err != nil {
			log.Fatalf("marsahl fail:%s\n", err)
			return
		}

		formVals = append(formVals, FormVal{Tag: k, Body: body})
	}

	*fm = append(*fm, formVals...)
}

func ParseBody(Body *[]byte) {
	if bytes.HasPrefix(*Body, []byte("@")) {
		body, err := ioutil.ReadFile(string((*Body)[1:]))
		if err != nil {
			log.Fatalf("%v\n", err)
			return
		}

		*Body = body
	}
}

func (g *GurlCore) ParseInit() {

	if len(g.Body) > 0 {
		ParseBody(&g.Body)
	}

	if len(g.J) > 0 {
		bodyJson := map[string]interface{}{}

		toJson(g.J, bodyJson)

		body, err := json.Marshal(&bodyJson)
		if err != nil {
			log.Fatalf("marsahl fail:%s\n", err)
			return
		}

		g.Body = body
	}

	//g.FormCache = []FormVal{}

	if len(g.Jfa) > 0 {
		jsonFromAppend(g.Jfa, &g.FormCache)
	}

	if len(g.F) > 0 {
		g.form(g.F, &g.FormCache)
	}
}

func (g *GurlCore) MultipartNew() (*http.Request, chan error, error) {

	var err error

	pipeReader, pipeWriter := io.Pipe()
	errChan := make(chan error, 10)
	writer := multipart.NewWriter(pipeWriter)

	go func() {

		defer pipeWriter.Close()

		var part io.Writer

		for _, fv := range g.FormCache {

			k := fv.Tag

			fname := fv.Fname

			if len(fname) == 0 {
				part, err = writer.CreateFormField(k)
				if err != nil {
					fmt.Printf("%s\n", err)
					continue
				}
				part.Write([]byte(fv.Body))
				continue
			}

			body := bytes.NewBuffer(fv.Body)

			part, err = writer.CreateFormFile(k, filepath.Base(fname))
			if err != nil {
				errChan <- err
				return
			}

			if _, err = io.Copy(part, body); err != nil {
				errChan <- err
				return
			}
		}

		errChan <- writer.Close()

	}()

	var req *http.Request
	req, err = http.NewRequest(g.Method, g.Url, pipeReader)
	if err != nil {
		fmt.Printf("http neq request:%s\n", err)
		return nil, errChan, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, errChan, nil
}

func (g *GurlCore) HeadersAdd(req *http.Request) {

	for _, v := range g.H {

		headers := strings.Split(v, ":")

		if len(headers) != 2 {
			continue
		}

		headers[0] = strings.TrimSpace(headers[0])
		headers[1] = strings.TrimSpace(headers[1])

		req.Header.Add(headers[0], headers[1])
	}

	if len(g.A) > 0 {
		req.Header.Set("User-Agent", g.A)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Host", req.URL.Host)

}

func (g *GurlCore) writeHead(rsp *Response, w io.Writer) {

	if !g.V {
		return
	}

	if rsp.Req != nil {
		req := rsp.Req
		path := "/"
		if len(req.URL.Path) > 0 {
			path = req.URL.Path
		}
		fmt.Fprintf(w, "> %s %s %s\r\n", req.Method, path, req.Proto)
		for k, v := range req.Header {
			fmt.Fprintf(w, "> %s: %s\r\n", k, strings.Join(v, ","))
		}

		fmt.Fprint(w, ">\r\n")
	}

	fmt.Fprintf(w, "< %s %s\r\n", rsp.Proto, rsp.Status)
	for k, v := range rsp.Header {
		fmt.Fprintf(w, "< %s: %s\r\n", k, strings.Join(v, ","))
	}
}

func (g *GurlCore) writeBytes(rsp *Response) {
	all := rsp.Body
	if g.O == "stdout" {
		g.writeHead(rsp, os.Stdout)
		os.Stdout.Write(all)
		return
	}

	if g.O == "stderr" {
		g.writeHead(rsp, os.Stderr)
		os.Stderr.Write(all)
		return
	}

	fd, err := os.OpenFile(g.O, g.Flag, 0644)
	if err != nil {
		return
	}
	defer fd.Close()

	g.writeHead(rsp, fd)
	if g.Flag&ADD_LINE > 0 {
		out := &bytes.Buffer{}
		out.Write(all)
		out.Write([]byte("\n"))
		fd.Write(out.Bytes())

		return
	}

	fd.Write(all)
}

type Gurl struct {
	*http.Client `json:"-"`

	GurlCore
}

type Response struct {
	StatusCode int         `json:"status_code"`
	Err        string      `json:"err"`
	Body       []byte      `json:"body"`
	Status     string      `json:"status"`
	Proto      string      `json:"proto"`
	Header     http.Header `json:"header"`
	Req        *http.Request
}

func (g *Gurl) Send() (*Response, error) {
	return g.send(g.Client)
}

func (g *GurlCore) send(client *http.Client) (*Response, error) {
	rsp, err := g.sendExec(client)
	if rsp.Err == "" && len(g.O) > 0 {
		g.writeBytes(rsp)
	}
	return rsp, err
}

func rspCopy(dst *Response, src *http.Response) {
	dst.StatusCode = src.StatusCode
	dst.Status = src.Status
	dst.Proto = src.Proto
	dst.Header = src.Header
	dst.Req = src.Request
}

func (g *GurlCore) GetOrBodyExec(client *http.Client) (*Response, error) {
	var rsp *http.Response
	var req *http.Request
	var err error

	body := bytes.NewBuffer(g.Body)
	req, err = http.NewRequest(g.Method, g.Url, body)
	gurlRsp := &Response{}
	if err != nil {
		return &Response{Err: err.Error()}, err
	}

	g.HeadersAdd(req)

	rsp, err = client.Do(req)
	if err != nil {
		return &Response{Err: err.Error()}, err
	}

	defer rsp.Body.Close()
	gurlRsp.Body, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		return &Response{Err: err.Error()}, err
	}

	rspCopy(gurlRsp, rsp)
	return gurlRsp, nil
}

func (g *GurlCore) MultipartExec(client *http.Client) (*Response, error) {

	var rsp *http.Response
	var req *http.Request

	req, errChan, err := g.MultipartNew()
	if err != nil {
		fmt.Printf("multipart new fail:%s\n", err)
		return &Response{Err: err.Error()}, err
	}

	gurlRsp := &Response{}
	g.HeadersAdd(req)

	rsp, err = client.Do(req)
	if err != nil {
		fmt.Printf("client do fail:%s:URL(%s)\n", err, req.URL)
		return &Response{Err: err.Error()}, err
	}

	defer rsp.Body.Close()

	if err := <-errChan; err != nil {
		fmt.Printf("error:%s\n", err)
		return &Response{Err: err.Error()}, err
	}

	gurlRsp.Body, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		fmt.Printf("ioutil.Read:%s\n", err)
		return &Response{Err: err.Error()}, err
	}

	rspCopy(gurlRsp, rsp)
	return gurlRsp, nil
}

func (g *GurlCore) SendExec(client *http.Client) (*Response, error) {
	return g.sendExec(client)
}

func (g *GurlCore) sendExec(client *http.Client) (*Response, error) {
	if len(g.Method) == 0 {
		g.Method = "GET"
		if len(g.FormCache) > 0 {
			g.Method = "POST"
		}
	}

	if len(g.FormCache) > 0 {
		return g.MultipartExec(client)
	}

	return g.GetOrBodyExec(client)
}

func ParseMF(mf string, formCache *[]FormVal) {
	pos := strings.Index(mf, "=")
	if pos == -1 {
		return
	}

	fv := FormVal{}

	fv.Tag = mf[:pos]
	fv.Body = []byte(mf[pos+1:])
	fv.Fname = "test"
	*formCache = append(*formCache, fv)
}

func ExecSlice(cmd []string) (*Response, error) {

	commandlLine := flag.NewFlagSet(cmd[0], flag.ExitOnError)
	headers := commandlLine.StringSlice("H", []string{}, "Pass custom header LINE to server (H)")
	forms := commandlLine.StringSlice("F", []string{}, "Specify HTTP multipart POST data (H)")
	output := commandlLine.String("o", "", "Write to FILE instead of stdout")
	method := commandlLine.String("X", "", "Specify request command to use")
	memForms := commandlLine.StringSlice("mF", []string{}, "Specify HTTP multipart POST data (H)")
	url := commandlLine.String("url", "", "Specify a URL to fetch")

	commandlLine.Parse(cmd[1:])

	as := commandlLine.Args()

	transport := http.Transport{
		DisableKeepAlives: true,
	}

	u := *url
	if u == "" {
		u = as[0]
	}

	g := Gurl{
		Client: &http.Client{
			Transport: &transport,
		},
		GurlCore: GurlCore{
			Method: *method,
			F:      *forms,
			H:      *headers,
			O:      *output,
			Url:    u,
		},
	}

	g.ParseInit()

	formCache := []FormVal{}
	for _, v := range *memForms {

		ParseMF(v, &formCache)
	}

	g.GurlCore.FormCache = append(g.GurlCore.FormCache, formCache...)

	return g.sendExec(g.Client)
}
