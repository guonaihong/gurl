package wsurl

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/guonaihong/flag"
	"github.com/guonaihong/gurl/gurlib"
	"github.com/guonaihong/gurl/input"
	"github.com/guonaihong/gurl/task"
	url2 "github.com/guonaihong/gurl/wsurl/url"
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

type wsCmd struct {
	task.Task
	firstSendAfter string
	A              string
	header         []string
	conf           string
	kArgs          string
	sendRate       string
	url            string
	data           string
	lastData       string
	reqHeader      http.Header
	mt             int
	closeMessage   bool
	bench          bool
	report         *Report
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

	if len(w.A) > 0 {
		w.reqHeader.Set("User-Agent", w.A)
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

/*
func cancelled(message gurlib.Message) bool {
	select {
	case <-message.InDone:
		return true
	default:
		return false
	}
}
*/

/*
func (ws *wsCmd) LuaMain(message gurlib.Message) {

	conf := ws.conf
	kargs := ws.kArgs
	all, err := ioutil.ReadFile(conf)
	if err != nil {
		fmt.Printf("ERROR:%s\n", err)
		os.Exit(1)
	}

	wg := sync.WaitGroup{}
	work := ws.work
	c := ws.c

	defer func() {
		wg.Wait()
		close(message.Out)
		close(message.OutDone)
	}()

	for i := 0; i < c; i++ {

		wg.Add(1)
		go func(id int) {

			defer wg.Done()

				l := NewLuaEngine(kargs)
				l.L.SetGlobal("in_ch", lua.LChannel(message.In))
				l.L.SetGlobal("out_ch", lua.LChannel(message.Out))

			for {

				if ws.n != 0 {
					select {
					case _, ok := <-work:
						if !ok {
							return
						}
					}
				} else {
					if cancelled(message) && len(message.In) == 0 {
						return
					}
				}

				err = l.L.DoString(string(all))
				if err != nil {
					fmt.Printf("run lua script fail:%s\n", err)
					os.Exit(1)
				}
			}

		}(i)
	}
}
*/

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
	wb int
	rb int
}

func (ws *wsCmd) one() (rv wsResult, err error) {
	var c *wsClient
	c, err = newWsClient(ws.url, ws.reqHeader)
	if err != nil {
		//fmt.Printf("new ws client fail %s\n", err)
		return
	}
	defer c.Close()

	mt := ws.mt

	if len(ws.firstSendAfter) > 0 {
		if t := gurlib.ParseTime(ws.firstSendAfter); int(t) > 0 {
			time.Sleep(t)
		}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, m, err := c.ReadMessage()
			if err != nil {
				//fmt.Println("read fail:", err)
				return
			}

			rv.rb += len(m)
			if !ws.bench {
				os.Stdout.Write(m)
			}
		}
	}()

	data := ws.data
	wb := ws.write(c, mt, data)
	rv.wb += wb

	if len(ws.lastData) > 0 {
		wb = ws.write(c, mt, ws.lastData)
		rv.wb += wb
	}

	if ws.closeMessage {
		err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			fmt.Printf("write close message \n")
		}
	}

	<-done
	return
}

func CmdErr(err error) {
	if err == nil {
		return
	}

	if noerr, ok := err.(*net.OpError); ok {
		if scerr, ok := noerr.Err.(*os.SyscallError); ok {
			if scerr.Err == syscall.ECONNREFUSED {
				fmt.Printf("wsurl: (7) couldn't connect to host\n")
				os.Exit(7)
			}
		}
	}

	fmt.Printf("%s\n", err)
}

func (ws *wsCmd) Init() {
	if ws.bench {
		ws.report = NewReport(ws.C, ws.N, ws.url)
		if len(ws.Duration) > 0 {
			if t := gurlib.ParseTime(ws.Duration); int(t) > 0 {
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
}

func (ws *wsCmd) SubProcess(work chan string) {
	for range work {
		taskNow := time.Now()
		rv, err := ws.one()
		if err != nil {
			if ws.report != nil {
				ws.report.AddErr()
			} else {
				CmdErr(err)
			}
			continue
		}

		if ws.report != nil {
			ws.report.Add(taskNow, rv.rb, rv.wb)
		}
	}
}

func Main(message gurlib.Message, argv0 string, argv []string) {
	commandlLine := flag.NewFlagSet(argv0, flag.ExitOnError)
	an := commandlLine.Int("an", 1, "Number of requests to perform")
	ac := commandlLine.Int("ac", 1, "Number of multiple requests to make")
	sendRate := commandlLine.String("send-rate", "", "How many bytes of data in seconds")
	rate := commandlLine.Int("rate", 0, "Requests per second")
	duration := commandlLine.String("duration", "", "Duration of the test")
	bench := commandlLine.Bool("bench", false, "Run benchmarks test")
	conf := commandlLine.String("K, config", "", "lua script")
	firstSendAfter := commandlLine.String("fsa, first-send-after", "", "Wait for the first time before sending")
	kargs := commandlLine.String("kargs", "", "Command line parameters passed to the configuration file")
	URL := commandlLine.String("url", "", "Specify a URL to fetch")
	headers := commandlLine.StringSlice("H, header", []string{}, "Pass custom header LINE to server (H)")
	binary := commandlLine.Bool("binary", false, "Send binary messages instead of utf-8")
	listen := commandlLine.String("l", "", "Listen mode, websocket echo server")
	data := commandlLine.String("d, data", "", "Data to be send per connection")
	lastData := commandlLine.String("ld, last-data", "", "Last message sent to be connection")
	closeMessage := commandlLine.Bool("close", false, "Send close message")
	readStream := commandlLine.Bool("rs, read-stream", false, "Read data from the stream")

	inputMode := commandlLine.Bool("input", false, "open input mode")
	inputRead := commandlLine.String("input-read", "", "open input file")
	inputFields := commandlLine.String("input-fields", " ", "sets the field separator")
	inputRenameKey := commandlLine.String("input-renkey", "", "Rename the default key")

	commandlLine.Author("guonaihong https://github.com/guonaihong/wsurl")
	commandlLine.Parse(argv)

	if *inputMode {
		input.Main(*inputRead, *inputFields, *inputRenameKey, message)
		return
	}

	if len(*conf) > 0 {
		if _, err := os.Stat(*conf); os.IsNotExist(err) {
			fmt.Printf("%s\n", err)
			return
		}
	}

	wscmd := &wsCmd{
		Task: task.Task{
			Duration:   *duration,
			N:          *an,
			Work:       make(chan string, 1000),
			ReadStream: *readStream,
			Message:    message,
			Rate:       *rate,
			C:          *ac,
		},
		firstSendAfter: *firstSendAfter,
		header:         *headers,
		conf:           *conf,
		kArgs:          *kargs,
		sendRate:       *sendRate,
		reqHeader:      make(map[string][]string, 3),
		mt:             websocket.TextMessage,
		data:           *data,
		lastData:       *lastData,
		closeMessage:   *closeMessage,
		bench:          *bench,
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
	if len(*conf) > 0 {
		//wscmd.LuaMain(message)
		return
	}

	if *URL == "" {
		commandlLine.Usage()
		return
	}
	wscmd.url = url2.ModifyUrl(*URL)

	wscmd.Task.Processer = wscmd
	wscmd.Task.RunMain()
}
