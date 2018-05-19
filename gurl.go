package main

import (
	"fmt"
	"github.com/NaihongGuo/flag"
	"github.com/guonaihong/gurl/gurlib"
	"github.com/robfig/cron"
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

func cmdBenchMain(c, n int, url string,
	work chan struct{}, wg *sync.WaitGroup,
	g *gurlib.Gurl) {

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

func cmdMain(c int, work chan struct{}, wg *sync.WaitGroup,
	g *gurlib.Gurl) {

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

func jsConfMain(c int, conf string, work chan struct{},
	wg *sync.WaitGroup, g *gurlib.Gurl) {

	defer func() {
		wg.Wait()
		os.Exit(0)
	}()

	all, _ := ioutil.ReadFile(conf)
	for i := 0; i < c; i++ {
		wg.Add(1)

		go func() {
			js := gurlib.NewJsEngine(g.Client)

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

func main() {

	headers := flag.StringSlice("H", []string{}, "Pass custom header LINE to server (H)")
	forms := flag.StringSlice("F", []string{}, "Specify HTTP multipart POST data (H)")
	jfa := flag.StringSlice("Jfa", []string{}, "Specify HTTP multipart POST json data (H)")
	cronExpr := flag.String("cron", "", "Cron expression")
	conf := flag.String("K", "", "Read js config from FILE")
	output := flag.String("o", "stdout", "Write to FILE instead of stdout")
	method := flag.String("X", "", "Specify request command to use")
	gen := flag.Bool("gen", false, "Generate the default js configuration file")
	toJson := flag.StringSlice("J", []string{}, `Turn key:value into {"key": "value"})`)
	url := flag.String("url", "", "Specify a URL to fetch")
	an := flag.Int("an", 1, "Number of requests to perform")
	ac := flag.Int("ac", 1, "Number of multiple requests to make")
	bench := flag.Bool("bench", false, "Run benchmarks test")
	conns := flag.Int("conns", DefaultConnections, "Max open idle connections per target host")
	cpus := flag.Int("cpus", 0, "Number of CPUs to use")

	flag.Parse()

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

	if len(*cronExpr) > 0 {

		cron := cron.New()

		defer cron.Stop()

		var js *gurlib.JsEngine
		if len(*conf) > 0 {
			js = gurlib.NewJsEngine(&client)
		}

		g.MemInit()
		cron.AddFunc(*cronExpr, func() {
			if len(*conf) > 0 {
				all, _ := ioutil.ReadFile(*conf)
				js.VM.Run(string(all))
				return
			}

			g.Send()
		})

		cron.Run()
	}

	work := make(chan struct{}, 1000)
	wg := sync.WaitGroup{}

	go func() {

		defer close(work)
		if *an >= 0 {

			for i, n := 0, *an; i < n; i++ {
				work <- struct{}{}
			}

			return
		}

		for {
			work <- struct{}{}
		}

	}()

	if len(*conf) > 0 {
		g.O = ""
		jsConfMain(*ac, *conf, work, &wg, &g)

		if *bench {
			//TODO
		}
	}

	if *bench {
		g.O = ""
		cmdBenchMain(*ac, *an, Url, work, &wg, &g)
	}

	cmdMain(*ac, work, &wg, &g)
}
