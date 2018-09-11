package main

import (
	"fmt"
	"github.com/guonaihong/flag"
	"github.com/guonaihong/gurl/gurlib"
	url2 "github.com/guonaihong/gurl/gurlib/url"
	"github.com/yuin/gopher-lua"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	DefaultConnections = 10000
)

type GurlCmd struct {
	duration string
	rate     int
	c        int
	n        int
	conf     string
	KArgs    string
	cronExpr string
	work     chan struct{}
	wg       sync.WaitGroup
	bench    bool
	*gurlib.Gurl
}

/*
func (cmd *GurlCmd) Cron(client *http.Client) {
	cron := cron.New()
	conf := cmd.conf
	cronExpr := cmd.cronExpr
	kargs := cmd.kargs

	defer cron.Stop()

	var js *gurlib.JsEngine
	if len(conf) > 0 {
		js = gurlib.NewJsEngine(client)
	}

	cmd.MemInit()
	cron.AddFunc(cronExpr, func() {
		if len(conf) > 0 {
			all, err := ioutil.ReadFile(conf)
			if err != nil {
				os.Exit(1)
			}

			js.VM.Set("gurl_args", conf+""+kargs)
			js.VM.Run(string(all))
			return
		}

		_, err := cmd.Send()
		CmdErr(err)
	})

	cron.Run()
}
*/

func (cmd *GurlCmd) Producer() {
	work, n := cmd.work, cmd.n

	if len(cmd.duration) > 0 {
		if t := gurlib.ParseTime(cmd.duration); int(t) > 0 {
			cmd.n = -1
		}
	}

	go func() {

		defer close(work)
		if cmd.n >= 0 {

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

func (cmd *GurlCmd) main() {

	defer os.Exit(0)

	var report *gurlib.Report
	c, n := cmd.c, cmd.n
	g := cmd.Gurl
	url := g.Url
	work, wg := cmd.work, &cmd.wg

	g.ParseInit()

	sig := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)

	signal.Notify(sig, os.Interrupt)

	begin := time.Now()

	interval := 0
	if cmd.rate > 0 {
		interval = int(time.Second) / cmd.rate
	}

	if cmd.bench {
		report = gurlib.NewReport(c, n, url)
	}

	if len(cmd.duration) > 0 {
		if t := gurlib.ParseTime(cmd.duration); int(t) > 0 {
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
				rsp, err := g.Send()
				if err != nil {
					if report != nil {
						report.AddErrNum()
					} else {
						CmdErr(err)
					}
					continue
				}

				if report != nil {
					report.Cal(taskNow, rsp)
				}
			}

		}(i)
	}

	if report != nil {
		report.StartReport()
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

func CmdErr(err error) {
	if err == nil {
		return
	}

	if uerr, ok := err.(*url.Error); ok {
		if noerr, ok := uerr.Err.(*net.OpError); ok {
			if scerr, ok := noerr.Err.(*os.SyscallError); ok {
				if scerr.Err == syscall.ECONNREFUSED {
					fmt.Printf("gurl: (7) couldn't connect to host\n")
					os.Exit(7)
				}
			}
		}
	}

	fmt.Printf("%s\n", err)
}

func httpEcho(addr string) {

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		io.Copy(os.Stdout, req.Body)
		w.Header().Add("Server", "gurl-server")
		req.Body.Close()
		return
	})

	fmt.Println(http.ListenAndServe(addr, nil))
}

func toFlag(output, str string) (flag int) {

	if output != "stdout" && output != "stderr" {
		flag |= os.O_CREATE | os.O_RDWR
	}

	flags := strings.Split(str, "|")
	for _, v := range flags {
		switch v {
		case "create":
			flag |= os.O_CREATE
		case "append":
			flag |= os.O_APPEND
		case "line":
			flag |= gurlib.ADD_LINE
		case "trunc":
			flag |= os.O_TRUNC
		}
	}

	return flag
}

func cancelled(message gurlib.Message) bool {
	select {
	case <-message.InDone:
		return true
	default:
		return false
	}
}

func (cmd *GurlCmd) LuaMain(message gurlib.Message) {

	conf := cmd.conf
	kargs := cmd.KArgs
	all, err := ioutil.ReadFile(conf)
	if err != nil {
		fmt.Printf("ERROR:%s\n", err)
		os.Exit(1)
	}

	wg := sync.WaitGroup{}

	work := cmd.work

	c := cmd.c

	defer func() {
		wg.Wait()
		close(message.Out)
		close(message.OutDone)
	}()

	for i := 0; i < c; i++ {

		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			l := NewLuaEngine(cmd.Client, kargs)
			l.L.SetGlobal("in_ch", lua.LChannel(message.In))
			l.L.SetGlobal("out_ch", lua.LChannel(message.Out))

			for {
				if cmd.n != 0 {
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

func gurlMain(message gurlib.Message, argv0 string, argv []string) {
	commandlLine := flag.NewFlagSet(argv0, flag.ExitOnError)

	headers := commandlLine.StringSlice("H, header", []string{}, "Pass custom header LINE to server (H)")
	forms := commandlLine.StringSlice("F, form", []string{}, "Specify HTTP multipart POST data (H)")
	formStrings := commandlLine.StringSlice("form-string", []string{}, "Specify HTTP multipart POST data (H)")
	jfa := commandlLine.StringSlice("Jfa", []string{}, "Specify HTTP multipart POST json data (H)")
	jfaStrings := commandlLine.StringSlice("Jfa-string", []string{}, "Specify HTTP multipart POST json data (H)")
	cronExpr := commandlLine.String("cron", "", "Cron expression")
	conf := commandlLine.String("K, config", "", "lua script")
	kargs := commandlLine.String("kargs", "", "Command line parameters passed to the configuration file")
	output := commandlLine.String("o, output", "stdout", "Write to FILE instead of stdout")
	oflag := commandlLine.String("oflag", "", "Control the way you write(append|line|trunc)")
	method := commandlLine.String("X, request", "", "Specify request command to use")
	gen := commandlLine.Bool("gen", false, "Generate the default lua script")
	toJson := commandlLine.StringSlice("J", []string{}, `Turn key:value into {"key": "value"})`)
	URL := commandlLine.String("url", "", "Specify a URL to fetch")
	an := commandlLine.Int("an", 1, "Number of requests to perform")
	ac := commandlLine.Int("ac", 1, "Number of multiple requests to make")
	rate := commandlLine.Int("rate", 0, "Requests per second")
	bench := commandlLine.Bool("bench", false, "Run benchmarks test")
	conns := commandlLine.Int("conns", DefaultConnections, "Max open idle connections per target host")
	cpus := commandlLine.Int("cpus", 0, "Number of CPUs to use")
	listen := commandlLine.String("l", "", "Listen mode, HTTP echo server")
	data := commandlLine.String("d, data", "", "HTTP POST data")
	verbose := commandlLine.Bool("v, verbose", false, "Make the operation more talkative")
	agent := commandlLine.String("A, user-agent", "gurl", "Send User-Agent STRING to server")
	duration := commandlLine.String("duration", "", "Duration of the test")
	connectTimeout := commandlLine.String("connect-timeout", "", "Maximum time allowed for connection")

	commandlLine.Author("guonaihong https://github.com/guonaihong/gurl")
	commandlLine.Parse(argv)

	if *listen != "" {
		httpEcho(*listen)
		return
	}

	as := commandlLine.Args()
	Url := *URL
	if *URL == "" && len(as) == 0 && len(*conf) == 0 && !*gen && !*bench {
		commandlLine.Usage()
		return
	}

	if len(as) > 0 {
		Url = as[0]
	}

	if len(*conf) > 0 {
		if _, err := os.Stat(*conf); os.IsNotExist(err) {
			fmt.Printf("%s\n", err)
			return
		}
	}

	Url = url2.ModifyUrl(Url)

	if *cpus > 0 {
		runtime.GOMAXPROCS(*cpus)
	}

	dialer := &net.Dialer{
		Timeout: gurlib.ParseTime(*connectTimeout),
	}

	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: *conns,
			Dial:                dialer.Dial,
		},
	}

	g := gurlib.Gurl{
		Client: &client,
		GurlCore: gurlib.GurlCore{
			Method: *method,
			F:      *forms,
			H:      *headers,
			O:      *output,
			J:      *toJson,
			Jfa:    *jfa,
			Url:    Url,
			Flag:   toFlag(*output, *oflag),
			Body:   []byte(*data),
			V:      *verbose,
			A:      *agent,
		},
	}

	g.AddFormStr(*formStrings)
	g.AddJsonFormStr(*jfaStrings)

	if *gen {
		if len(*conf) > 0 {
			Lua2Cmd(*conf, *kargs)
			return
		}

		Cmd2Lua(&g)
		return
	}

	cmd := GurlCmd{
		duration: *duration,
		c:        *ac,
		n:        *an,
		rate:     *rate,
		conf:     *conf,
		KArgs:    *kargs,
		cronExpr: *cronExpr,
		Gurl:     &g,
		work:     make(chan struct{}, 1000),
		bench:    *bench,
	}

	if len(*cronExpr) > 0 {
		//cmd.Cron(&client)
	}

	cmd.Producer()

	if len(*conf) > 0 {
		g.O = ""
		cmd.LuaMain(message)

		if *bench {
			//TODO
		}
		return
	}

	if *bench {
		g.O = ""
	}

	cmd.main()
}

type Chan struct {
	ch   chan lua.LValue
	done chan lua.LValue
}

func main() {

	var wg sync.WaitGroup
	var cmds [][]string

	prevPos := 1
	for k, v := range os.Args[1:] {
		if v == "|" {
			cmds = append(cmds, os.Args[prevPos:k+1])
			prevPos = k + 2
		}
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if len(cmds) == 0 {
		cmds = [][]string{os.Args[1:]}
	}

	if prevPos != 1 && prevPos < len(os.Args) {
		cmds = append(cmds, os.Args[prevPos:])
	}

	var channel []*Chan
	wg.Add(len(cmds))
	defer wg.Wait()

	for k, v := range cmds {
		channel = append(channel, &Chan{
			done: make(chan lua.LValue),
			ch:   make(chan lua.LValue, 1000),
		})

		go func(ch []*Chan, k int, v []string) {
			defer func() {
				wg.Done()
			}()

			m := gurlib.Message{
				Out:     ch[k].ch,
				OutDone: ch[k].done,
				K:       k,
			}

			if k > 0 {
				m.In = ch[k-1].ch
				m.InDone = ch[k-1].done
			}

			//fmt.Printf("k=%d, %#v\n", k, m)
			gurlMain(m, os.Args[0], v)
		}(channel, k, v)
	}

}
