package gurlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/NaihongGuo/flag"
	"github.com/ghodss/yaml"
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

type If struct {
	Cond   string   `json:"cond,omitempty"`
	Format string   `json:"format,omitempty"`
	Set    []string `json:"set,omitempty"`
}

type Cond struct {
	If   If `json:"if,omitempty"`
	Else If `json:"else,omitempty"`
}
type SaveVar map[string]interface{}

type FormVal struct {
	Tag   string
	Fname string
	Body  []byte
}

type Base struct {
	Include string   `json:"include,omitempty"`
	NoSend  bool     `json:"no_send,omitempty"`
	Set     []string `json:"set,omitempty"`
	Method  string   `json:"method,omitempty"`

	J   []string `json:"J,omitempty"`
	F   []string `json:"F,omitempty"`
	H   []string `json:"H,omitempty"` // http header
	Url string   `json:"url,omitempty"`
	O   string   `json:"o,omitempty"`

	Jfa []string `json:"Jfa,omitempty"`

	RunF   []string `json:"-"`
	RunH   []string `json:"-"`
	RunUrl string   `json:"-"`
	RunO   string   `json:"-"`
	RunJfa []string `json:"-"`

	Next    []Base  `json:"next,omitempty"`
	Parent  *Base   `json:"-"`
	NextMap SaveVar `json:"-"`

	FormCache []FormVal `json:"-"`

	Body string `json:"body,omitempty"`
	Cond `json:"-"`
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

func form(F []string, fm *[]FormVal) {

	fileds := [2]string{}
	formVals := []FormVal{}

	for _, v := range F {

		fileds[0], fileds[1] = "", ""

		pos := strings.Index(v, "=")
		if pos == -1 {
			continue
		}

		fileds[0], fileds[1] = v[:pos], v[pos+1:]

		//fileds[1] = strings.TrimLeft(fileds[1], " ")

		if strings.HasPrefix(fileds[1], "@") {
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

		//RunF[i] = fileds[0]
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

func (b *Base) MemInit() {

	if len(b.J) > 0 {
		bodyJson := map[string]interface{}{}

		toJson(b.J, bodyJson)

		body, err := json.Marshal(&bodyJson)
		if err != nil {
			log.Fatalf("marsahl fail:%s\n", err)
			return
		}

		b.Body = string(body)
	}

	b.FormCache = []FormVal{}

	if len(b.RunJfa) > 0 {
		jsonFromAppend(b.RunJfa, &b.FormCache)
	}

	if len(b.RunF) > 0 {
		form(b.RunF, &b.FormCache)
	}
}

func (b *Base) Multipart(client *http.Client) (rsp *http.Response) {

	var req *http.Request

	req, errChan, err := b.MultipartNew()
	if err != nil {
		fmt.Printf("multipart new fail:%s\n", err)
		return
	}

	b.HeadersAdd(req)

	c := client
	rsp, err = c.Do(req)
	if err != nil {
		fmt.Printf("client do fail:%s\n", err)
		return
	}

	if e := <-errChan; e != nil {
		fmt.Printf("error:%s\n", e)
	}

	return rsp
}

func (b *Base) MultipartNew() (*http.Request, chan error, error) {

	var err error

	pipeReader, pipeWriter := io.Pipe()
	errChan := make(chan error, 10)
	writer := multipart.NewWriter(pipeWriter)

	go func() {

		defer pipeWriter.Close()

		var part io.Writer

		for _, fv := range b.FormCache {

			k := fv.Tag

			fname := fv.Fname

			if len(fname) == 0 {
				part, err = writer.CreateFormField(k)
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
	req, err = http.NewRequest(b.Method, b.RunUrl, pipeReader)
	if err != nil {
		fmt.Printf("http neq request:%s\n", err)
		return nil, errChan, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, errChan, nil
}

func (b *Base) HeadersAdd(req *http.Request) {

	for _, v := range b.RunH {

		headers := strings.Split(v, ":")

		if len(headers) != 2 {
			continue
		}

		headers[0] = strings.TrimSpace(headers[0])
		headers[1] = strings.TrimSpace(headers[1])

		req.Header.Add(headers[0], headers[1])
	}
}

func (b *Base) WriteFile(rsp *http.Response, body []byte) {
	fd, err := os.Create(b.RunO)
	if err != nil {
		return
	}
	defer fd.Close()

	if len(body) > 0 {
		fd.Write(body)
	}

	io.Copy(fd, rsp.Body)
}

func (b *Base) BodyRequest(client *http.Client) (rsp *http.Response) {

	var (
		err error
		req *http.Request
	)

	req, err = http.NewRequest(b.Method, b.RunUrl, strings.NewReader(b.Body))
	if err != nil {
		return
	}

	b.HeadersAdd(req)

	c := client

	rsp, err = c.Do(req)
	if err != nil {
		return
	}

	return rsp
}

func (b *Base) NotMultipart(client *http.Client) (rsp *http.Response) {

	var (
		err error
		req *http.Request
	)

	req, err = http.NewRequest(b.Method, b.RunUrl, nil)
	if err != nil {
		return
	}

	b.HeadersAdd(req)

	c := client

	rsp, err = c.Do(req)
	if err != nil {
		return
	}

	return rsp
}

type Gurl struct {
	http.Client `json:"-"`

	GurlCore
}

type GurlCore struct {
	Base
	For *For `json:"for,omitempty"`
}

type MultiGurl struct {
	http.Client `json:"-"`

	ConfFile

	*Conf
}

type ConfFile struct {
	Cmd   Gurl   `json:"cmd"`
	Root  Gurl   `json:"root"`
	Child []Gurl `json:"child"`
	Func  []Func `json:"func"`
}

type Response struct {
	Rsp  *http.Response
	Body []byte
}

func MultiGurlInit(m *MultiGurl) {

	//TODO
	//m.Cmd.Base.MemInit()
	m.Root.Base.MemInit()

	rootMap := SaveVar{}

	url := m.Cmd.Url
	H := m.Cmd.H
	if url == "" {
		url = m.Root.Url
	}

	if url != "" {
		rootMap["root_url"] = url
	}

	if len(H) == 0 {
		rootMap["root_header"] = m.Root.H
	}

	m.Conf = ConfNew(rootMap)
	c := m.Conf

	for _, v := range m.Root.Set {
		c.Parse([]byte(v), nil, true)
	}

	for k, _ := range m.Func {
		m.Func[k].Root = m
		m.AddFunc(&m.Func[k])
	}

}

func (m *MultiGurl) GenCmd() {
	cmd := m.Cmd

	out := bytes.Buffer{}

	out.WriteString("gurl")

	if len(cmd.Method) > 0 {
		out.WriteString(" -X " + cmd.Method)
	}

	for _, v := range cmd.H {
		out.WriteString(" -H " + v)
	}

	for _, v := range cmd.J {
		out.WriteString(" -J " + v)
	}

	for _, v := range cmd.Jfa {
		out.WriteString(" -Jia " + v)
	}

	for _, v := range cmd.F {
		out.WriteString(" -F " + v)
	}

	if len(cmd.Url) > 0 {
		out.WriteString(" -url " + cmd.Url)
	}

	io.Copy(os.Stdout, &out)
}

func (m *MultiGurl) GenYaml(opt string) {
	var (
		cmd   = 1 << 0
		root  = 1 << 1
		child = 1 << 2
		fn    = 1 << 3
	)

	opts := strings.Split(opt, ",")
	out := bytes.Buffer{}
	//out := os.Stdout
	mask := 0

	for _, v := range opts {
		switch v {
		case "cmd":
			mask |= cmd
		case "root":
			mask |= root
		case "child":
			mask |= child
		case "func":
			mask |= fn
		case "all":
			mask = cmd | root | child | fn
			goto next
		}
	}

next:

	for i := 0; i < 3; i++ {
		o := (mask & (1 << uint(i))) > 0
		if !o {
			continue
		}

		var v interface{}
		switch {
		case i == 0:
			v = struct {
				Cmd *Gurl `json:"cmd,omitempty"`
			}{&m.Cmd}
		case i == 1:
			v = struct {
				Root *Gurl `json:"root,omitempty"`
			}{&m.Root}
		case i == 2:
			v = struct {
				Child []Gurl `json:"child,omitempty"`
			}{m.Child}
		case i == 3:
			v = struct {
				Func []Func `json:"func,omitempty"`
			}{m.Func}
		}

		j, err := yaml.Marshal(v)
		out.Write(j)
		if err != nil {
			panic(err.Error())
		}
	}

	io.Copy(os.Stdout, &out)
}

func (m *MultiGurl) ChildInitSend(base *Base, valMap SaveVar) {

	BaseParse(base, m.Conf, valMap)
	base.MemInit()
	BaseSend(base, &m.Client, m.Conf, valMap)

	v2 := base
	for len(v2.Next) > 0 {

		v2.Next[0].Parent = v2

		BaseParse(&v2.Next[0], m.Conf, valMap)
		v2.Next[0].MemInit()
		BaseSend(&v2.Next[0], &m.Client, m.Conf, valMap)

		v2 = &v2.Next[0]
	}
}

type Func struct {
	Name string
	Args []string
	For  `json:"for"`
	Base
	Root interface{} `json:"-"`
}

func (f *Func) GurlFunc(v *FuncVal) error {
	if len(f.Args) > len(v.CallArgs) {
		fmt.Printf("v.CallArgs:%v\n", v.CallArgs)
		panic("func " + f.Name + " args must is " + strconv.Itoa(len(f.Args)))
	}

	m := f.Root.(*MultiGurl)
	c := m.Conf
	rangeMap := SaveVar{}

	for k, v := range v.CallArgs {
		key := c.ParseName([]byte(f.Args[k]))
		rangeMap[key] = v
	}

	m.RunFor(m.Conf, &f.For, rangeMap)

	m.ChildInitSend(&f.Base, rangeMap)
	return nil
}

func (m *MultiGurl) AddFunc(f *Func) {
	m.Conf.AddFunc(f.Name, f.GurlFunc)
}

func (m *MultiGurl) RunFor(c *Conf, For *For, rangeMap SaveVar) {
	if len(For.Range) > 0 {
		rangeSlice := c.ParseSlice(
			[]byte(For.Range), nil, true)

		key := c.ParseName([]byte(For.K))
		val := c.ParseName([]byte(For.V))

		for k, v := range rangeSlice {
			rangeMap[key] = k
			rangeMap[val] = v

			m.ChildInitSend(&For.Base, rangeMap)

			for _, vv := range For.Set {
				c.Parse([]byte(vv), rangeMap, true)
			}
		}

	}
}

func (m *MultiGurl) Send() {

	c := m.Conf

	BaseSend(&m.Cmd.Base, &m.Client, m.Conf, nil)

	m.ChildInitSend(&m.Cmd.Base, nil)

	BaseSend(&m.Root.Base, &m.Client, m.Conf, nil)

	for j, _ := range m.Child {

		For := m.Child[j].For

		m.RunFor(c, For, SaveVar{})
		m.ChildInitSend(&m.Child[j].Base, nil)
	}
}

func BaseParse(g *Base, c *Conf, rangeMap SaveVar) {
	var (
		newHeader []string
		parentMap SaveVar
	)

	if g.Parent != nil {
		parentMap = g.Parent.NextMap
	}

	if len(rangeMap) > 0 {

		//merge parentMap and rangeMap
		for k, v := range parentMap {
			if _, ok := rangeMap[k]; ok {
				continue
			}
			rangeMap[k] = v
		}
		parentMap = rangeMap
	}

	g.RunF = append([]string{}, g.F...)
	g.RunJfa = append([]string{}, g.Jfa...)
	g.RunUrl = g.Url
	g.RunH = g.H
	g.RunO = g.O

	g.RunUrl = c.ParseString([]byte(g.Url), parentMap, true)

	for j, _ := range g.RunF {

		g.RunF[j] = g.F[j]
		g.RunF[j] = c.ParseString([]byte(g.F[j]), parentMap, true)
	}

	isRootH := false
try:
	for _, hv := range g.RunH {
		if isRootH == true {
			if strings.TrimSpace(hv) == "$root_header" {
				hs := c.ParseSlice([]byte(hv), parentMap, true)
				newHeader = append(newHeader, hs...)
				continue
			}

			hv = c.ParseString([]byte(hv), parentMap, true)
			newHeader = append(newHeader, hv)
			continue
		}

		if strings.TrimSpace(hv) == "$root_header" {
			if isRootH == false {
				isRootH = true
				goto try
			}
		}
	}

	if isRootH {
		g.RunH = newHeader
	}
}

func (g *Gurl) ConfigInit(config string, cf *ConfFile) error {
	fd, err := os.Open(config)
	if err != nil {
		fmt.Printf("open config file fail:%v\n", err)
		return err
	}

	defer fd.Close()
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		fmt.Printf("read config fail:%v\n", err)
		return err
	}

	err = yaml.Unmarshal(data, cf)
	if err != nil {
		fmt.Printf("wrong configuration file format:%v\n", err)
		return err
	}

	return nil
}

//todo reflect copy
func MergeCmd(cfCmd *Gurl, cmd *Gurl, tactics string) {
	switch tactics {
	case "append":
		if len(cmd.Url) > 0 {
			cfCmd.Url = cmd.Url
		}
	case "set":
		*cfCmd = *cmd

	}
}

func BaseSend(b *Base, client *http.Client, c *Conf, valMap SaveVar) {
	var (
		rsp     *http.Response
		body    []byte
		err     error
		needVar bool
	)

	if b.NoSend {
		return
	}

	if len(b.Method) == 0 {
		b.Method = "GET"
		if len(b.FormCache) > 0 {
			b.Method = "POST"
		}
	}

	if len(b.Body) > 0 {
		rsp = b.BodyRequest(client)
	} else if len(b.FormCache) > 0 {
		rsp = b.Multipart(client)
	} else {
		rsp = b.NotMultipart(client)
	}

	if rsp == nil {
		return
	}

	var curVal SaveVar

	defer rsp.Body.Close()

	if len(b.If.Cond) > 0 {
		needVar = true //TODO
	}

	ifVal := false
	if len(b.Next) > 0 && len(b.RunUrl) > 0 || needVar {

		body, err = ioutil.ReadAll(rsp.Body)
		if err != nil {
			return
		}

		b.NextMap = SaveVar{
			"parent_http_body": string(body),
			"parent_http_code": fmt.Sprintf("%d", rsp.StatusCode),
		}

		curVal = SaveVar{
			"http_body": string(body),
			"http_code": fmt.Sprintf("%d", rsp.StatusCode),
		}

		if b.Parent != nil {
			for k, v := range b.Parent.NextMap {
				curVal[k] = v
			}
		}

		//merge for cycle valmap and curVal
		for k, v := range valMap {
			curVal[k] = v
		}

		if c.ParseBool([]byte(b.If.Cond), curVal, false) {
			ifVal = true
			body = []byte(c.ParseString([]byte(b.If.Format), curVal, false))

		} else if len(b.Else.Format) > 0 {
			body = []byte(c.ParseString([]byte(b.Else.Format), curVal, false))
		}
	}

	if len(b.RunO) > 0 {
		b.WriteFile(rsp, body)
		goto last
	}

	if len(body) > 0 {
		os.Stdout.Write(body)
		goto next
	}

	io.Copy(os.Stdout, rsp.Body)

next:
	fmt.Printf("\n")
last:
	var set []string

	if ifVal && len(b.If.Set) > 0 {
		set = b.If.Set
	} else if ifVal == false && len(b.Else.Set) > 0 {
		set = b.Else.Set
	}

	for _, v := range set {
		//TODO
		c.Parse([]byte(v), curVal, true)
	}
}

func (g *Gurl) writeBytes(rsp *http.Response, all []byte) {
	fd, err := os.Create(g.RunO)
	if err != nil {
		return
	}
	defer fd.Close()

	fd.Write(all)
}

func (g *Gurl) NotMultipartExec() (*Response, error) {
	var rsp *http.Response
	var req *http.Request
	var err error

	req, err = http.NewRequest(g.Method, g.RunUrl, nil)
	if err != nil {
		return nil, err
	}

	gurlRsp := &Response{}
	g.HeadersAdd(req)

	rsp, err = g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	gurlRsp.Rsp = rsp
	gurlRsp.Body, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	if len(g.RunO) > 0 {
		g.writeBytes(rsp, gurlRsp.Body)
	}

	return gurlRsp, nil
}

func (g *Gurl) MultipartExec() (*Response, error) {

	var rsp *http.Response
	var req *http.Request

	req, errChan, err := g.MultipartNew()
	if err != nil {
		fmt.Printf("multipart new fail:%s\n", err)
		return nil, err
	}

	gurlRsp := &Response{}
	g.HeadersAdd(req)

	rsp, err = g.Client.Do(req)
	if err != nil {
		fmt.Printf("client do fail:%s:URL(%s)\n", err, req.URL)
		return nil, err
	}

	gurlRsp.Rsp = rsp

	defer rsp.Body.Close()

	if err := <-errChan; err != nil {
		fmt.Printf("error:%s\n", err)
		return nil, err
	}

	gurlRsp.Body, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		fmt.Printf("ioutil.Read:%s\n", err)
		return nil, err
	}

	if len(g.RunO) > 0 {

		g.writeBytes(rsp, gurlRsp.Body)
	}

	return gurlRsp, nil
}

func (g *Gurl) sendExec() (*Response, error) {
	if len(g.Method) == 0 {
		g.Method = "GET"
		if len(g.FormCache) > 0 {
			g.Method = "POST"
		}
	}

	if len(g.FormCache) > 0 {
		return g.MultipartExec()
	}

	return g.NotMultipartExec()
}

// TODO
func ToArgs(cmd string) []string {
	return nil
}

// TODO
func ExecString(cmd string) *Response {
	return nil
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
		Client: http.Client{
			Transport: &transport,
		},
		GurlCore: GurlCore{
			Base: Base{
				Method: *method,
				F:      *forms,
				H:      *headers,
				O:      *output,
				Url:    u,
			},
		},
	}

	g.RunUrl = g.Url
	g.RunF = g.F
	g.RunH = g.H
	g.MemInit()

	formCache := []FormVal{}
	for _, v := range *memForms {

		pos := strings.Index(v, "=")
		if pos == -1 {
			continue
		}

		fv := FormVal{}

		fv.Tag = v[:pos]
		fv.Body = []byte(v[pos+1:])
		fv.Fname = "test"
		formCache = append(formCache, fv)
	}

	g.GurlCore.FormCache = append(g.GurlCore.FormCache, formCache...)

	return g.sendExec()
}
