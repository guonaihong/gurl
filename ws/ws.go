package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/guonaihong/flag"
	"github.com/guonaihong/gurl/core"
	"github.com/guonaihong/gurl/input"
	"github.com/guonaihong/gurl/output"
	"github.com/guonaihong/gurl/report"
	"github.com/guonaihong/gurl/task"
	"github.com/guonaihong/gurl/utils"
	url2 "github.com/guonaihong/gurl/ws/url"
	_ "io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"
)

var upgrader = websocket.Upgrader{}

type wsClient struct {
	*websocket.Conn
	url string
}

// 命令wsmd由几部分的数据组成
// task.Task, 负责并发控制
// wsCmdData(业务数据), 存放具体的数据，为并发模块提供燃料
// bool类型
// 报表, bench模式下，输出报表供人观看

type wsCmdData struct {
	packet         []string
	firstSendAfter string
	userAgent      string
	header         []string
	sendRate       string
	url            string
	output         string
	reqHeader      http.Header
}

type wsCmd struct {
	*task.Task

	wsCmdData

	mt           int
	closeMessage bool
	bench        bool

	writeStream bool
	merge       bool

	outFd  *os.File
	report *report.Report
}

func (w *wsCmd) headersAdd() {

	for _, v := range w.header {

		headers := strings.Split(v, ":")

		if len(headers) != 2 {
			continue
		}

		headers[0] = strings.TrimSpace(headers[0])
		headers[1] = strings.TrimSpace(headers[1])

		w.reqHeader.Add(headers[0], headers[1])
	}

	if len(w.userAgent) > 0 {
		w.reqHeader.Set("User-Agent", w.userAgent)
	}

	w.reqHeader.Set("Accept", "*/*")
	//req.Header.Set("Host", req.URL.Host)

}

func newWsClient(u string, header http.Header) (*wsClient, error) {
	u1, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	c, _, err := websocket.DefaultDialer.Dial(u1.String(), header)
	if err != nil {
		return nil, err
	}

	wsc := &wsClient{
		url:  u,
		Conn: c,
	}

	return wsc, nil
}

func (w *wsClient) Close() {
	w.Conn.Close()
}

func (ws *wsCmd) write(c *wsClient, mt int, data string) (rv int) {
	if !strings.HasPrefix(data, "@") {
		c.WriteMessage(mt, []byte(data))
		rv += len(data)
	} else {
		fd, err := os.Open(data[1:])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		defer fd.Close()

		rate := &rate{}
		genRate(ws.sendRate, &rate)
		bufsize := 4096 * 2
		if rate != nil && rate.B > 0 {
			bufsize = rate.B
		}

		buf := make([]byte, bufsize)
		for {
			n, err := fd.Read(buf)
			if err != nil {
				break
			}

			rv += n
			err = c.WriteMessage(mt, buf[:n])
			if err != nil {
				return
			}

			if rate != nil && rate.T > 0 {
				time.Sleep(time.Duration(rate.T))
			}
		}
	}
	return
}

func (ws *wsCmd) webSocketEcho(addr string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}

			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}

	})

	fmt.Println(http.ListenAndServe(addr, nil))
}

type rate struct {
	B int
	T int64
}

func parseTime(s string) time.Duration {
	s = strings.ToLower(s)

	rv := int64(0)

	fmt.Sscanf(s, "%d", &rv)
	switch {
	case strings.HasSuffix(s, "ms"):
		rv = rv * int64(time.Millisecond)
	case strings.HasSuffix(s, "s"):
		rv = rv * int64(time.Second)
	}
	return time.Duration(rv)
}

func genRate(rateStr string, rv **rate) {
	rates := strings.Split(rateStr, "/")

	if len(rates) != 2 {
		return
	}

	rates[0] = strings.ToLower(rates[0])
	rates[1] = strings.ToLower(rates[1])

	r := rate{}
	fmt.Sscanf(rates[0], "%d", &r.B)
	fmt.Sscanf(rates[1], "%d", &r.T)
	switch {
	case strings.HasSuffix(rates[0], "b"):
	case strings.HasSuffix(rates[0], "kb"):
		r.B *= 1024
	case strings.HasSuffix(rates[0], "mb"):
		r.B *= 1024 * 1024
	}

	switch {
	case strings.HasSuffix(rates[1], "ms"):
		r.T = r.T * int64(time.Millisecond)
	case strings.HasSuffix(rates[1], "s"):
		r.T = r.T * int64(time.Second)
	}

	if r.B <= 0 {
		return
	}

	if r.T <= 0 {
		return
	}

	*rv = &r
}

type wsResult struct {
	wb       int
	rb       int
	lastBody []byte
}

func (ws *wsCmd) outputFileNew() {

	var err error

	if ws.output != "" {
		switch ws.output {
		case "stdout":
			ws.outFd = os.Stdout
		case "stderr":
			ws.outFd = os.Stderr
		default:
			ws.outFd, err = os.OpenFile(ws.output, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				fmt.Printf("%s\n", err)
			}
		}
	}
}

func (ws *wsCmd) outputFileWrite(m []byte) {
	if ws.outFd != nil {
		ws.outFd.Write(m)
	}
}

func (ws *wsCmd) outputClose() {
	if ws.outFd != nil && ws.outFd != os.Stdout {
		ws.outFd.Close()
	}

}

func (ws *wsCmd) one() (rv wsResult, err error) {

	var c *wsClient
	c, err = newWsClient(ws.url, ws.reqHeader)
	if err != nil {
		return
	}
	defer c.Close()

	mt := ws.mt

	if len(ws.firstSendAfter) > 0 {
		if t := utils.ParseTime(ws.firstSendAfter); int(t) > 0 {
			time.Sleep(t)
		}
	}

	ws.outputFileNew()

	done := make(chan struct{})
	go func() {
		defer close(done)

		for {
			_, m, err := c.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, 1000) {
					fmt.Println("read fail:", err)
				}
				return
			}

			rv.lastBody = m
			rv.rb += len(m)
			if !ws.bench {
				ws.outputFileWrite(m)
			}
		}
	}()

	for _, v := range ws.packet {
		wb := ws.write(c, mt, v)
		rv.wb += wb
	}

	if ws.closeMessage {
		err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			fmt.Printf("write close message \n")
		}
	}

	<-done
	ws.outputClose()
	return
}

func CmdErr(err error) {
	if err == nil {
		return
	}

	if noerr, ok := err.(*net.OpError); ok {
		if scerr, ok := noerr.Err.(*os.SyscallError); ok {
			if scerr.Err == syscall.ECONNREFUSED {
				fmt.Printf("ws: (7) couldn't connect to host\n")
				os.Exit(7)
			}
		}
	}

	fmt.Printf("%s\n", err)
}

func (ws *wsCmd) Init() {
	if ws.bench {
		ws.report = report.NewReport(ws.C, ws.N, ws.url)
		if len(ws.Duration) > 0 {
			if t := utils.ParseTime(ws.Duration); int(t) > 0 {
				ws.report.SetDuration(t)
			}
		}
		ws.report.Start()
	}
}

func (ws *wsCmd) WaitAll() {
	if ws.report != nil {
		ws.report.Wait()
	}
	close(ws.Out)
}

func (ws *wsCmd) parse(val map[string]string, inJson string) {
	err := json.Unmarshal([]byte(inJson), &val)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	i := 0
	rs := make([]string, len(val)*2)
	for k, v := range val {

		rs[i] = "{" + k + "}"
		i++
		rs[i] = v
		i++
	}

	r := strings.NewReplacer(rs...)
	for k, v := range ws.header {
		ws.header[k] = r.Replace(v)
	}

	ws.userAgent = r.Replace(ws.userAgent)
	ws.url = r.Replace(ws.url)
	for k, v := range ws.packet {
		ws.packet[k] = r.Replace(v)
	}

	ws.firstSendAfter = r.Replace(ws.firstSendAfter)
}

func (ws *wsCmd) copyAndNew() *wsCmdData {
	return &wsCmdData{
		firstSendAfter: ws.firstSendAfter,
		userAgent:      ws.userAgent,
		header:         append([]string{}, ws.header...),
		sendRate:       ws.sendRate,
		url:            ws.url,
		packet:         append([]string{}, ws.packet...),
		output:         ws.output,
		reqHeader:      make(map[string][]string, 3),
	}
}

//todo
func (cmd *wsCmd) streamWriteJson(rsp wsResult, err error, inJson map[string]string) {
	m := map[string]interface{}{}
	m["err"] = ""
	m["last_body"] = string(rsp.lastBody)

	if err != nil {
		m["err"] = err.Error()
	}

	output.WriteStream(m, inJson, cmd.merge, cmd.Message)
}

func (cmd *wsCmd) SubProcess(work chan string) {

	var inJson map[string]string

	ws := *cmd
	ws0 := *cmd
	ws0.wsCmdData = *ws.copyAndNew()

	for v := range work {

		if len(v) > 0 && v[0] == '{' {
			inJson = map[string]string{}
			ws.wsCmdData = *ws.copyAndNew()
			ws.parse(inJson, v)
			ws.headersAdd()
		}

		taskNow := time.Now()
		rv, err := ws.one()
		if cmd.writeStream {
			cmd.streamWriteJson(rv, err, inJson)
		}
		//todo Give this judgment a name

		if err != nil {
			if ws.report != nil {
				ws.report.AddErr(err)
			} else {
				CmdErr(err)
			}
			continue
		}

		if ws.report != nil {
			ws.report.Add(taskNow, rv.rb, rv.wb)
		}

		if len(v) > 0 && v[0] == '{' {
			ws = ws0
		}
	}
}

func Main(message core.Message, argv0 string, argv []string) {
	command := flag.NewFlagSet(argv0, flag.ExitOnError)
	an := command.Int("an", 1, "Number of requests to perform")
	ac := command.Int("ac", 1, "Number of multiple requests to make")
	sendRate := command.String("send-rate", "", "How many bytes of data in seconds")
	rate := command.Int("rate", 0, "Requests per second")
	duration := command.String("duration", "", "Duration of the test")
	connectTimeout := command.String("connect-timeout", "", "Maximum time allowed for connection")
	bench := command.Bool("bench", false, "Run benchmarks test")
	outputFileName := command.String("o, output", "stdout", "Write to FILE instead of stdout")
	firstSendAfter := command.String("fsa, first-send-after", "", "Wait for the first time before sending")
	URL := command.String("url", "", "Specify a URL to fetch")
	headers := command.StringSlice("H, header", []string{}, "Pass custom header LINE to server (H)")
	binary := command.Bool("binary", false, "Send binary messages instead of utf-8")
	listen := command.String("l", "", "Listen mode, websocket echo server")
	userAgent := command.String("A, user-agent", "gurl", "Send User-Agent STRING to server")
	closeMessage := command.Bool("close", false, "Send close message")

	readStream := command.Bool("r, read-stream", false, "Read data from the stream")
	writeStream := command.Bool("w, write-stream", false, "Write data from the stream")
	merge := command.Bool("m, merge", false, "Combine the output results into the output")

	inputMode := command.Bool("I, input-model", false, "open input mode")
	inputRead := command.String("R, input-read", "", "open input file")
	inputFields := command.String("input-fields", " ", "sets the field separator")
	inputSetKey := command.String("skey, input-setkey", "", "Set a new name for the default key")

	outputMode := command.Bool("O, output-mode", false, "open output mode")
	outputKey := command.String("wkey, write-key", "", "Key that can be write")
	outputWrite := command.String("W, output-write", "", "open output file")

	packet := command.Opt("p, packet", "Data packet to be send per connection").
		Flags(flag.GreedyMode).
		NewStringSlice([]string{})

	command.Parse(argv)

	if !*inputMode {
		if len(*inputRead) > 0 {
			*inputMode = true
		}
	}

	if *inputMode {
		input.Main(*inputRead, *inputFields, *inputSetKey, message)
		return
	}

	if *outputMode {
		output.WriteFile(*outputWrite, *outputKey, message)
		return
	}

	if len(*connectTimeout) > 0 {
		websocket.DefaultDialer.HandshakeTimeout = utils.ParseTime(*connectTimeout)
	}

	wscmd := &wsCmd{
		Task: &task.Task{
			Duration:   *duration,
			N:          *an,
			Work:       make(chan string, 1000),
			ReadStream: *readStream,
			Message:    message,
			Rate:       *rate,
			C:          *ac,
		},
		wsCmdData: wsCmdData{
			packet:         *packet,
			firstSendAfter: *firstSendAfter,
			header:         *headers,
			sendRate:       *sendRate,
			userAgent:      *userAgent,
			reqHeader:      make(map[string][]string, 3),
			output:         *outputFileName,
		},
		mt:           websocket.TextMessage,
		closeMessage: *closeMessage,
		bench:        *bench,
		writeStream:  *writeStream,
		merge:        *merge,
	}

	if *binary {
		wscmd.mt = websocket.BinaryMessage
	}

	wscmd.headersAdd()

	if len(*listen) > 0 {
		wscmd.webSocketEcho(*listen)
		return
	}

	wscmd.Producer()

	if *URL == "" {
		command.Usage()
		return
	}

	wscmd.url = url2.ModifyUrl(*URL)

	wscmd.Task.Processer = wscmd
	wscmd.Task.RunMain()
}
