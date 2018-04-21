package gurl

import (
	"bytes"
	"cond"
	"conf"
	"core"
	"fmt"
	"github.com/NaihongGuo/flag"
	"github.com/ghodss/yaml"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Gurl struct {
	http.Client `json:"-"`

	GurlCore
}

type GurlCore struct {
	core.Base
	For *cond.For `json:"for,omitempty"`
}

type MultiGurl struct {
	http.Client `json:"-"`

	ConfFile

	*conf.Conf
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

	rootMap := core.SaveVar{}

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

	m.Conf = conf.New(rootMap)
	c := m.Conf

	for _, v := range m.Root.Set {
		c.Parse([]byte(v), nil, true)
	}

	for k, _ := range m.Func {
		m.Func[k].Root = m
		m.AddFunc(&m.Func[k])
	}

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

func (m *MultiGurl) ChildInitSend(base *core.Base, valMap core.SaveVar) {

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
	Name     string
	Args     []string
	cond.For `json:"for"`
	core.Base
	Root interface{} `json:"-"`
}

func (f *Func) GurlFunc(v *conf.FuncVal) error {
	if len(f.Args) > len(v.CallArgs) {
		fmt.Printf("v.CallArgs:%v\n", v.CallArgs)
		panic("func " + f.Name + " args must is " + strconv.Itoa(len(f.Args)))
	}

	m := f.Root.(*MultiGurl)
	c := m.Conf
	rangeMap := core.SaveVar{}

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

func (m *MultiGurl) RunFor(c *conf.Conf, For *cond.For, rangeMap core.SaveVar) {
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

		m.RunFor(c, For, core.SaveVar{})
		m.ChildInitSend(&m.Child[j].Base, nil)
	}
}

func BaseParse(g *core.Base, c *conf.Conf, rangeMap core.SaveVar) {
	var (
		newHeader []string
		parentMap core.SaveVar
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

func BaseSend(b *core.Base, client *http.Client, c *conf.Conf, valMap core.SaveVar) {
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

	var curVal core.SaveVar

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

		b.NextMap = core.SaveVar{
			"parent_http_body": string(body),
			"parent_http_code": fmt.Sprintf("%d", rsp.StatusCode),
		}

		curVal = core.SaveVar{
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

	if e := <-errChan; e != nil {
		fmt.Printf("error:%s\n", e)
		return nil, err
	}

	gurlRsp.Body, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
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

	commandlLine.Parse(cmd[1:])

	as := commandlLine.Args()

	transport := http.Transport{
		DisableKeepAlives: true,
	}

	g := Gurl{
		Client: http.Client{
			Transport: &transport,
		},
		GurlCore: GurlCore{
			Base: core.Base{
				Method: *method,
				F:      *forms,
				H:      *headers,
				O:      *output,
				Url:    as[0],
			},
		},
	}

	g.MemInit()

	return g.sendExec()
}
