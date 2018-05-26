package main

import (
	"fmt"
	"github.com/guonaihong/flag"
	"github.com/guonaihong/gurl/gurlib"
	"github.com/robfig/cron"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	DefaultConnections = 10000
)

func modifyUrl(u string) string {
	if len(u) == 0 {
		return u
	}

	if len(u) > 0 && u[0] == ':' {
		return "http://127.0.0.1" + u
	}

	if len(u) > 0 && u[0] == '/' {
		return "http://127.0.0.1:80" + u
	}

	if !strings.HasPrefix(u, "http") {
		return "http://" + u
	}

	return u
}

type GurlCmd struct {
	c        int
	n        int
	conf     string
	cronExpr string
	confArgs string
	work     chan struct{}
	wg       sync.WaitGroup
	*gurlib.Gurl
}

func (cmd *GurlCmd) Cron(client *http.Client) {
	cron := cron.New()
	conf := cmd.conf
	cronExpr := cmd.cronExpr
	confArgs := cmd.confArgs

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

			js.VM.Set("gurl_args", conf+""+confArgs)
			js.VM.Run(string(all))
			return
		}

		cmd.Send()
	})

	cron.Run()
}

func (cmd *GurlCmd) Producer() {
	work := cmd.work
	n := cmd.n

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
				g.Send()
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

func (cmd *GurlCmd) jsConfMain() {

	c := cmd.c
	conf := cmd.conf
	wg := &cmd.wg
	work := cmd.work
	confArgs := cmd.confArgs
	g := cmd.Gurl

	defer func() {
		wg.Wait()
		os.Exit(0)
	}()

	all, err := ioutil.ReadFile(conf)
	if err != nil {
		os.Exit(1)
	}

	for i := 0; i < c; i++ {
		wg.Add(1)

		go func() {
			js := gurlib.NewJsEngine(g.Client)
			js.VM.Set("gurl_args", conf+" "+confArgs)
			if len(g.Url) > 0 {
				js.VM.Set("gurl_url", g.Url)
			}

			defer wg.Done()

			for range work {
				_, err := js.VM.Run(string(all))
				if err != nil {
					fmt.Printf("%s\n", err)
					os.Exit(1)
				}
			}
		}()
	}

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

func main() {

	headers := flag.StringSlice("H, header", []string{}, "Pass custom header LINE to server (H)")
	forms := flag.StringSlice("F, form", []string{}, "Specify HTTP multipart POST data (H)")
	jfa := flag.StringSlice("Jfa", []string{}, "Specify HTTP multipart POST json data (H)")
	cronExpr := flag.String("cron", "", "Cron expression")
	conf := flag.String("K, config", "", "Read js config from FILE")
	confArgs := flag.String("Kargs", "", "Command line parameters passed to the configuration file")
	output := flag.String("o, output", "stdout", "Write to FILE instead of stdout")
	oflag := flag.String("oflag", "", "Control the way you write(append|line|trunc)")
	method := flag.String("X, request", "", "Specify request command to use")
	gen := flag.Bool("gen", false, "Generate the default js configuration file")
	toJson := flag.StringSlice("J", []string{}, `Turn key:value into {"key": "value"})`)
	url := flag.String("url", "", "Specify a URL to fetch")
	an := flag.Int("an", 1, "Number of requests to perform")
	ac := flag.Int("ac", 1, "Number of multiple requests to make")
	bench := flag.Bool("bench", false, "Run benchmarks test")
	conns := flag.Int("conns", DefaultConnections, "Max open idle connections per target host")
	cpus := flag.Int("cpus", 0, "Number of CPUs to use")
	echo := flag.String("echo", "", "HTTP echo server")
	data := flag.String("d, data", "", "HTTP POST data")
	verbose := flag.Bool("v, verbose", false, "Make the operation more talkative")
	agent := flag.String("A, user-agent", "gurl", "Send User-Agent STRING to server")

	flag.Parse()

	if *echo != "" {
		httpEcho(*echo)
	}

	as := flag.Args()
	Url := *url
	if *url == "" && len(as) == 0 && len(*conf) == 0 && !*gen && !*bench {
		flag.Usage()
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

	Url = modifyUrl(Url)

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
			gurlib.Js2Cmd(*conf)
			return
		}

		gurlib.Cmd2Js(&g)
		return
	}

	cmd := GurlCmd{
		c:        *ac,
		n:        *an,
		conf:     *conf,
		cronExpr: *cronExpr,
		confArgs: *confArgs,
		Gurl:     &g,
		work:     make(chan struct{}, 1000),
	}

	if len(*cronExpr) > 0 {
		cmd.Cron(&client)
	}

	cmd.Producer()

	if len(*conf) > 0 {
		g.O = ""
		cmd.jsConfMain()

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
