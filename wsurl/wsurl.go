package wsurl

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/guonaihong/flag"
	"github.com/guonaihong/gurl/gurlib"
	url2 "github.com/guonaihong/gurl/wsurl/url"
	"github.com/yuin/gopher-lua"
	_ "io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var upgrader = websocket.Upgrader{}

type wsClient struct {
	*websocket.Conn
	url string
}

type wsCmd struct {
	firstSendAfter string
	A              string
	header         []string
	conf           string
	kArgs          string
	sendRate       string
	url            string
	data           string
	lastData       string
	duration       string
	reqHeader      http.Header
	rate           int
	mt             int
	n              int
	c              int
	work           chan struct{}
	closeMessage   bool
	bench          bool
}

func ParseTime(t string) (rv time.Duration) {

	t0 := 0
	for k, _ := range t {
		v := int(t[k])
		switch {
		case v >= '0' && v <= '9':
			t0 = t0*10 + (v - '0')
		case v == 's':
			rv += time.Duration(t0) * time.Second
			t0 = 0
		case v == 'm':
			if k+1 < len(t) && t[k+1] == 's' {
				rv += time.Duration(t0) * time.Millisecond
				t0 = 0
				k++
				continue
			}
			rv += time.Duration(t0*60) * time.Second
			t0 = 0
		case v == 'h':
			rv += time.Duration(t0*60*60) * time.Second
			t0 = 0
		case v == 'd':
			rv += time.Duration(t0*60*60*24) * time.Second
			t0 = 0
		case v == 'w':
			rv += time.Duration(t0*60*60*24*7) * time.Second
			t0 = 0
		case v == 'M':
			rv += time.Duration(t0*60*60*24*7*31) * time.Second
			t0 = 0
		case v == 'y':
			rv += time.Duration(t0*60*60*24*7*31*365) * time.Second
			t0 = 0
		}
	}

	return
}

func (ws *wsCmd) Producer() {
	work, n := ws.work, ws.n

	if len(ws.duration) > 0 {
		if t := ParseTime(ws.duration); int(t) > 0 {
			n, ws.n = -1, -1
		}
	}

	go func() {

		defer close(work)
		if n >= 0 {

			for i := 0; i < n; i++ {
				work <- struct{}{}
			}

			return
		}

		for {
			work <- struct{}{}
		}
		return

	}()
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
		if t := ParseTime(ws.firstSendAfter); int(t) > 0 {
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

func (ws *wsCmd) main() {

	var report *Report
	wg := sync.WaitGroup{}
	c := ws.c
	n := ws.n
	work := ws.work

	sig := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)

	signal.Notify(sig, os.Interrupt)

	if ws.bench {
		report = NewReport(ws.c, ws.n, ws.url)
		report.Start()
	}

	begin := time.Now()

	interval := 0
	if ws.rate > 0 {
		interval = int(time.Second) / ws.rate
	}

	if len(ws.duration) > 0 {
		if t := ParseTime(ws.duration); int(t) > 0 {
			wg.Add(1)

			if report != nil {
				report.SetDuration(t)
			}
			workTimeout := make(chan struct{}, 1000)
			work = workTimeout

			ticker := time.NewTicker(t)
			go func() {

				defer func() {
					close(workTimeout)
					for range workTimeout {
					}
					wg.Done()
				}()

				for {
					select {
					case <-ticker.C:
						return
					case workTimeout <- struct{}{}:
					}

				}
			}()
		}
	}

	if interval > 0 {
		count := 0
		oldwork := work
		work = make(chan struct{}, 1000)
		wg.Add(1)
		go func() {
			defer func() {
				close(work)
				wg.Done()
			}()

			for {
				next := begin.Add(time.Duration(count * interval))
				time.Sleep(next.Sub(time.Now()))

				select {
				case _, ok := <-oldwork:
					if !ok {
						return
					}
				default:
				}

				work <- struct{}{}
				if count++; count == n {
					return
				}
			}
		}()
	}
	for i := 0; i < c; i++ {

		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for range work {
				taskNow := time.Now()
				rv, err := ws.one()
				if err != nil {
					if report != nil {
						report.AddErr()
					} else {
						CmdErr(err)
					}
					continue
				}

				if report != nil {
					report.Add(taskNow, rv.rb, rv.wb)
				}
			}

		}(i)
	}

	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

end:
	for {
		select {
		case <-sig:
			if report != nil {
				report.Wait()
			}
			break end
		case <-done:
			if report != nil {
				report.Wait()
			}
			break end
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

	commandlLine.Author("guonaihong https://github.com/guonaihong/wsurl")
	commandlLine.Parse(argv)

	if len(*conf) > 0 {
		if _, err := os.Stat(*conf); os.IsNotExist(err) {
			fmt.Printf("%s\n", err)
			return
		}
	}

	wscmd := &wsCmd{
		firstSendAfter: *firstSendAfter,
		header:         *headers,
		conf:           *conf,
		kArgs:          *kargs,
		sendRate:       *sendRate,
		rate:           *rate,
		reqHeader:      make(map[string][]string, 3),
		mt:             websocket.TextMessage,
		data:           *data,
		lastData:       *lastData,
		duration:       *duration,
		n:              *an,
		c:              *ac,
		closeMessage:   *closeMessage,
		bench:          *bench,
		work:           make(chan struct{}, 1000),
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
	wscmd.main()
}

type Message struct {
	In      chan lua.LValue
	Out     chan lua.LValue
	InDone  chan lua.LValue
	OutDone chan lua.LValue
	K       int
}

type Chan struct {
	ch   chan lua.LValue
	done chan lua.LValue
}
