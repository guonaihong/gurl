package core

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Cond   string   `json:"cond"`
	Format string   `json:"format"`
	Set    []string `json:"set"`
}

type Cond struct {
	If   If `json:"if"`
	Else If `json:"else"`
}
type SaveVar map[string]interface{}

type FormVal struct {
	Tag   string
	Fname string
	Body  []byte
}

type Base struct {
	Include string   `json:"include"`
	NoSend  bool     `json:"no_send"`
	Set     []string `json:"set"`
	Method  string   `json:"method"`

	J   []string `json:"J"`
	F   []string `json:"F"`
	H   []string // http header
	Url string   `json:"url"`
	O   string   `json:"o"`

	RunF   []string `json:"-"`
	RunH   []string
	RunUrl string
	RunO   string

	Next    []Base  `json:"next"`
	Parent  *Base   `json:"-"`
	NextMap SaveVar `json:"-"`

	FormCache []FormVal

	Body string
	Cond
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

	var (
		fileds   [2]string
		formVals []FormVal
	)

	for _, v := range F {
		fileds[0] = ""
		fileds[1] = ""

		pos := strings.Index(v, "=")
		if pos == -1 {
			continue
		}

		fileds[0] = v[:pos]
		fileds[1] = v[pos+1:]

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

	if len(b.F) > 0 {
		form(b.F, &b.FormCache)
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
