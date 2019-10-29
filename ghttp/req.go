package ghttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TylerBrock/colorjson"
	"github.com/fatih/color"
	color2 "github.com/guonaihong/gurl/color"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
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

	Json     []string `json:"Json,omitempty"`
	FormData []string `json:"FormData,omitempty"`
	Header   []string `json:"Header,omitempty"` // http header
	Url      string   `json:"url,omitempty"`
	Output   string   `json:"o,omitempty"`

	Jfa []string `json:"Jfa,omitempty"`

	FormCache []FormVal `json:"-"`

	Body       []byte `json:"body,omitempty"`
	Flag       int
	Verbose    bool `json:"-"`
	UserAgent  string
	Color      bool
	Query      []string
	NotParseAt map[string]struct{}
}

func CopyAndNew(g *GurlCore) *GurlCore {
	return &GurlCore{
		Method:    g.Method,
		Json:      append([]string{}, g.Json...),
		FormData:  append([]string{}, g.FormData...),
		Header:    append([]string{}, g.Header...),
		Url:       g.Url,
		Output:    g.Output,
		Jfa:       append([]string{}, g.Jfa...),
		FormCache: append([]FormVal{}, g.FormCache...),
		Body:      append([]byte{}, g.Body...),
		Flag:      g.Flag,
		Verbose:   g.Verbose,
		UserAgent: g.UserAgent,
	}
}

func (g *GurlCore) AddFormStr(FormData []string) {
	if len(FormData) == 0 {
		return
	}

	if g.NotParseAt == nil {
		g.NotParseAt = make(map[string]struct{}, 10)
	}

	oldLen := len(g.FormData)
	for i := 0; i < len(FormData); i++ {
		g.NotParseAt["FormData"+fmt.Sprintf("%d", oldLen+i)] = struct{}{}
	}

	g.FormData = append(g.FormData, FormData...)
}

func (g *GurlCore) AddJsonFormStr(Jfa []string) {
	if len(Jfa) == 0 {
		return
	}

	if g.NotParseAt == nil {
		g.NotParseAt = make(map[string]struct{}, 10)
	}
	oldLen := len(g.Jfa)
	for i := 0; i < len(Jfa); i++ {
		g.NotParseAt["Jfa"+fmt.Sprintf("%d", oldLen+i)] = struct{}{}
	}

	g.Jfa = append(g.Jfa, Jfa...)
}

func (g *GurlCore) formNotParseAt(idx int) bool {
	_, ok := g.NotParseAt[fmt.Sprintf("FormData%d", idx)]
	return ok
}

func (g *GurlCore) jsonFormNotParseAt(idx int) bool {
	_, ok := g.NotParseAt[fmt.Sprintf("Jfa%d", idx)]
	return ok
}

func parseVal(bodyJson map[string]interface{}, key, val string, notParseAt bool) {
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

	bodyJson[key] = parseAt(val, notParseAt)
}

func parseVal2(bodyJson map[string]interface{}, key, val string, notParseAt bool) {
	bodyJson[key] = parseAt(val, notParseAt)
}

func toJson(Json []string, notParseAts []bool, bodyJson map[string]interface{}) {
	for j, v := range Json {
		pos := strings.Index(v, ":")
		if pos == -1 {
			bodyJson[v] = ""
			continue
		}

		key := v[:pos]
		val := v[pos+1:]

		notParseAt := false
		if len(notParseAts) > 0 {
			notParseAt = notParseAts[j]
		}

		if pos := strings.Index(key, "."); pos != -1 {
			keys := strings.Split(key, ".")

			parseValfn := parseVal2
			if strings.HasPrefix(val, "=") {
				val = val[1:]
				parseValfn = parseVal
			}

			type jsonObj map[string]interface{}

			curMap := bodyJson

			for i, v := range keys {
				if len(keys)-1 == i {
					parseValfn(curMap, v, val, notParseAt)
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
			parseVal2(bodyJson, key, "", notParseAt)
			continue
		}

		if val[0] != '=' {
			parseVal2(bodyJson, key, val, notParseAt)
			continue
		}

		if len(key) == 1 {
			continue
		}

		val = val[1:]
		parseVal(bodyJson, key, val, notParseAt)

	}
}

func (g *GurlCore) form(FormData []string, fm *[]FormVal) {

	fileds := [2]string{}
	formVals := []FormVal{}

	for k, v := range FormData {

		fileds[0], fileds[1] = "", ""

		pos := strings.Index(v, "=")
		if pos == -1 {
			continue
		}

		fileds[0], fileds[1] = v[:pos], v[pos+1:]

		if !g.formNotParseAt(k) && strings.HasPrefix(fileds[1], "@") {
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

		//FormData[i] = fileds[0]
	}

	*fm = append(*fm, formVals...)
}

func (g *GurlCore) jsonFromAppend(JF []string, fm *[]FormVal) {

	JFMap := map[string][]string{}
	notParseAt := map[string][]bool{}
	fileds := [2]string{}
	formVals := []FormVal{}

	for i, v := range JF {

		fileds[0], fileds[1] = "", ""

		pos := strings.Index(v, "=")
		if pos == -1 {
			continue
		}

		fileds[0], fileds[1] = v[:pos], v[pos+1:]

		v, _ := JFMap[fileds[0]]
		JFMap[fileds[0]] = append(v, fileds[1])
		notParseAt[fileds[0]] = append(notParseAt[fileds[0]], g.jsonFormNotParseAt(i))

	}

	for k, v := range JFMap {

		bodyJson := map[string]interface{}{}

		toJson(v, notParseAt[k], bodyJson)

		body, err := json.Marshal(&bodyJson)

		if err != nil {
			log.Fatalf("marsahl fail:%s\n", err)
			return
		}

		formVals = append(formVals, FormVal{Tag: k, Body: body})
	}

	*fm = append(*fm, formVals...)
}

func parseAt(data string, notParseAt bool) interface{} {
	if !notParseAt && strings.HasPrefix(data, "@") {
		body, err := ioutil.ReadFile(data[1:])
		if err != nil {
			log.Fatalf("%v\n", err)
			return ""
		}
		return string(body)
	}

	if len(data) >= 2 {
		b := []byte(data)
		if json.Valid(b) {

			return json.RawMessage(b)
		}
	}

	return data
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

	if len(g.Json) > 0 {
		bodyJson := map[string]interface{}{}

		toJson(g.Json, nil, bodyJson)

		body, err := json.Marshal(&bodyJson)
		if err != nil {
			log.Fatalf("marsahl fail:%s\n", err)
			return
		}

		g.Body = body
	}

	//g.FormCache = []FormVal{}

	if len(g.Jfa) > 0 {
		g.jsonFromAppend(g.Jfa, &g.FormCache)
	}

	if len(g.FormData) > 0 {
		g.form(g.FormData, &g.FormCache)
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
	req, err = http.NewRequest(g.Method, g.Url+g.addQueryString(), pipeReader)
	if err != nil {
		fmt.Printf("http neq request:%s\n", err)
		return nil, errChan, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, errChan, nil
}

func (g *GurlCore) addQueryString() string {
	if len(g.Query) == 0 {
		return ""
	}

	u := url.Values{}
	for _, g := range g.Query {
		qs := strings.Split(g, "=")
		if len(qs) != 2 {
			continue
		}

		u.Add(qs[0], qs[1])
	}

	s := u.Encode()
	if len(u) > 0 {
		return "?" + s
	}

	return ""
}

func (g *GurlCore) HeadersAdd(req *http.Request) {

	for _, v := range g.Header {

		headers := strings.Split(v, ":")

		if len(headers) != 2 {
			continue
		}

		headers[0] = strings.TrimSpace(headers[0])
		headers[1] = strings.TrimSpace(headers[1])

		req.Header.Add(headers[0], headers[1])
	}

	if len(g.UserAgent) > 0 {
		req.Header.Set("User-Agent", g.UserAgent)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Host", req.URL.Host)

}

func (g *GurlCore) writeHead(rsp *Response, w io.Writer) {

	if !g.Verbose {
		return
	}

	keyStart, keyEnd, valStart, valEnd := color2.NewKeyVal(g.Color)

	if rsp.Req != nil {
		req := rsp.Req
		path := "/"
		if len(req.URL.Path) > 0 {
			path = req.URL.RequestURI()
		}

		fmt.Fprintf(w, "> %s %s %s\r\n", req.Method, path, req.Proto)
		for k, v := range req.Header {
			fmt.Fprintf(w, "%s> %s%s: %s%s%s\r\n", keyStart, k, keyEnd,
				valStart, strings.Join(v, ","), valEnd)
		}

		fmt.Fprint(w, ">\r\n")
		fmt.Fprint(w, "\n")
	}

	if g.Color {
		all, colorOk := colorBody(w.(*os.File), g.Body)
		w.Write(all)
		if colorOk {
			fmt.Fprintf(w, "\r\n")
		}
	}

	fmt.Fprintf(w, "< %s %s\r\n", rsp.Proto, rsp.Status)

	for k, v := range rsp.Header {
		fmt.Fprintf(w, "%s< %s%s: %s%s%s\r\n", keyStart, k, keyEnd,
			valStart, strings.Join(v, ","), valEnd)
	}
}

func colorBody(fd *os.File, all []byte) ([]byte, bool) {
	var obj map[string]interface{}
	if len(all) > 0 && all[0] == '{' {
		err := json.Unmarshal(all, &obj)
		if err != nil {
			return all, false
		}

		f := colorjson.NewFormatter()
		f.KeyColor = color.New(color.FgHiBlue)
		f.Indent = 2
		all, _ = f.Marshal(obj)
		all = append(all, '\n')
	}

	return all, true
}

func (g *GurlCore) writeBytes(rsp *Response) (err error) {
	all := rsp.Body
	var fd *os.File

	switch g.Output {
	case "stdout":
		fd = os.Stdout
	case "stderr":
		fd = os.Stderr
	default:
		fd, err = os.OpenFile(g.Output, g.Flag, 0644)
		if err != nil {
			return
		}
	}

	var colorOk bool
	if g.Color {
		all, colorOk = colorBody(fd, all)
	}

	if fd != os.Stdout || fd != os.Stderr {
		defer fd.Close()
	}

	// write http head
	g.writeHead(rsp, fd)

	if g.Flag&ADD_LINE > 0 {
		out := &bytes.Buffer{}
		out.Write(all)
		out.Write([]byte("\n"))
		fd.Write(out.Bytes())

		return
	}

	if colorOk {
		fd.Write([]byte("\n\n"))
	}
	fd.Write(all)
	return nil
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
	if rsp.Err == "" && len(g.Output) > 0 {
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
	req, err = http.NewRequest(g.Method, g.Url+g.addQueryString(), body)
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
		if len(g.FormCache) > 0 || len(g.Body) > 0 {
			g.Method = "POST"
		}
	}

	if len(g.FormCache) > 0 {
		return g.MultipartExec(client)
	}

	// 创建http.NewRequest地方有两个，todo归一化
	return g.GetOrBodyExec(client)
}
