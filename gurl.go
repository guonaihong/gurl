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
	c        int
	n        int
	conf     string
	KArgs    string
	cronExpr string
	work     chan struct{}
	wg       sync.WaitGroup
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

func (cmd *GurlCmd) benchMain() {

	c := cmd.c
	n := cmd.n
	g := cmd.Gurl
	url := g.Url
	work := cmd.work
	wg := &cmd.wg

	g.MemInit()
	report := gurlib.NewReport(c, n, url)

	for i := 0; i < c; i++ {

		wg.Add(1)

		go func() {
			defer wg.Done()

			for range work {

				taskNow := time.Now()
				rsp, err := g.Send()
				if err != nil {
					report.AddErrNum()
					continue
				}

				report.Cal(taskNow, rsp)
			}
		}()
	}

	report.StartReport()
	wg.Wait()
	report.Wait()
	os.Exit(0)
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

func (cmd *GurlCmd) main() {

	c := cmd.c
	work := cmd.work
	wg := &cmd.wg
	g := cmd.Gurl

	defer func() {
		wg.Wait()
		os.Exit(0)
	}()

	g.MemInit()
	for i := 0; i < c; i++ {

		wg.Add(1)

		go func() {
			defer wg.Done()
			for range work {
				_, err := g.Send()
				CmdErr(err)
			}
		}()
	}

}

/*
//TODO
func jsConfBenchMain(c, n int, url string,
	conf string, work chan struct{},
	wg *sync.WaitGroup, g *gurlib.Gurl) {

	report := gurlib.NewReport(c, n, url)
	all, _ := ioutil.ReadFile(conf)

	for i := 0; i < c; i++ {
		wg.Add(1)

		go func() {
			js := gurlib.NewJsEngine(g.Client)

			defer wg.Done()

			for range work {

				taskNow := time.Now()
				rsp, err := js.VM.Run(string(all))
				if err != nil {
					report.AddErrNum()
					fmt.Printf("%s\n", err)
					os.Exit(1)
				}

				report.Cal(taskNow, rsp)
			}
		}()
	}

	report.StartReport()
	wg.Wait()
	report.Wait()
	os.Exit(0)
}
*/

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

	for i := 0; i < c; i++ {

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				//message.Quit <- lua.LTrue
				//fmt.Printf("close out:%v, close quit:%v\n", message.Out, message.Quit)
				close(message.Out)
				//close(message.Quit)
			}()

			l := NewLuaEngine(cmd.Client)
			l.L.SetGlobal("gurl_cmd", lua.LString(kargs))
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
				}

				err = l.L.DoString(string(all))
				if err != nil {
					fmt.Printf("%s\n", err)
					os.Exit(1)
				}
			}
		}()
	}

	wg.Wait()
}

func gurlMain(message gurlib.Message, argv0 string, argv []string) {
	commandlLine := flag.NewFlagSet(argv0, flag.ExitOnError)

	headers := commandlLine.StringSlice("H, header", []string{}, "Pass custom header LINE to server (H)")
	forms := commandlLine.StringSlice("F, form", []string{}, "Specify HTTP multipart POST data (H)")
	jfa := commandlLine.StringSlice("Jfa", []string{}, "Specify HTTP multipart POST json data (H)")
	cronExpr := commandlLine.String("cron", "", "Cron expression")
	conf := commandlLine.String("K, config", "", "Read js config from FILE")
	kargs := commandlLine.String("kargs", "", "Command line parameters passed to the configuration file")
	output := commandlLine.String("o, output", "stdout", "Write to FILE instead of stdout")
	oflag := commandlLine.String("oflag", "", "Control the way you write(append|line|trunc)")
	method := commandlLine.String("X, request", "", "Specify request command to use")
	gen := commandlLine.Bool("gen", false, "Generate the default js configuration file")
	toJson := commandlLine.StringSlice("J", []string{}, `Turn key:value into {"key": "value"})`)
	URL := commandlLine.String("url", "", "Specify a URL to fetch")
	an := commandlLine.Int("an", 1, "Number of requests to perform")
	ac := commandlLine.Int("ac", 1, "Number of multiple requests to make")
	bench := commandlLine.Bool("bench", false, "Run benchmarks test")
	conns := commandlLine.Int("conns", DefaultConnections, "Max open idle connections per target host")
	cpus := commandlLine.Int("cpus", 0, "Number of CPUs to use")
	echo := commandlLine.String("echo", "", "HTTP echo server")
	data := commandlLine.String("d, data", "", "HTTP POST data")
	verbose := commandlLine.Bool("v, verbose", false, "Make the operation more talkative")
	agent := commandlLine.String("A, user-agent", "gurl", "Send User-Agent STRING to server")

	commandlLine.Author("guonaihong https://github.com/guonaihong/gurl")
	commandlLine.Parse(argv)

	if *echo != "" {
		httpEcho(*echo)
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

	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: *conns,
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

	if *gen {
		if len(*conf) > 0 {
			gurlib.Lua2Cmd(*conf)
			return
		}

		gurlib.Cmd2Lua(&g)
		return
	}

	cmd := GurlCmd{
		c:        *ac,
		n:        *an,
		conf:     *conf,
		KArgs:    *kargs,
		cronExpr: *cronExpr,
		Gurl:     &g,
		work:     make(chan struct{}, 1000),
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
	}

	if *bench {
		g.O = ""
		cmd.benchMain()
	}

	cmd.main()
}

type Chan struct {
	ch   chan lua.LValue
	quit chan lua.LValue
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

	wg.Add(len(cmds))

	var channel []*Chan

	for k, v := range cmds {
		channel = append(channel, &Chan{
			quit: make(chan lua.LValue, 2),
			ch:   make(chan lua.LValue, 1000),
		})

		go func(ch []*Chan, k int, v []string) {
			defer func() {
				wg.Done()
			}()

			m := gurlib.Message{
				Out: ch[k].ch,
			}

			if k > 0 {
				m.In = ch[k-1].ch
			}

			//fmt.Printf("k=%d, %#v\n", k, m)
			gurlMain(m, os.Args[0], v)
		}(channel, k, v)
	}

	wg.Wait()
}
