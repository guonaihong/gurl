package main

import (
	"fmt"
	"github.com/NaihongGuo/flag"
	"github.com/guonaihong/gurl/gurlib"
	"github.com/robfig/cron"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
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

func main() {

	headers := flag.StringSlice("H", []string{}, "Pass custom header LINE to server (H)")
	forms := flag.StringSlice("F", []string{}, "Specify HTTP multipart POST data (H)")
	jfa := flag.StringSlice("Jfa", []string{}, "Specify HTTP multipart POST json data (H)")
	cronExpr := flag.String("cron", "", "Cron expression")
	conf := flag.String("K", "", "Read js config from FILE")
	output := flag.String("o", "", "Write to FILE instead of stdout")
	method := flag.String("X", "", "Specify request command to use")
	gen := flag.String("gen", "", "Generate the default js configuration file")
	toJson := flag.StringSlice("J", []string{}, `Turn key:value into {"key": "value"})`)
	url := flag.String("url", "", "Specify a URL to fetch")
	an := flag.Int("an", 1, "Number of requests to perform")
	ac := flag.Int("ac", 1, "Number of multiple requests to make")

	flag.Parse()

	as := flag.Args()
	Url := *url
	if *url == "" && len(as) == 0 && len(*conf) == 0 && len(*gen) == 0 {
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

	client := http.Client{}

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

	if len(*gen) > 0 {
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

		if *an >= 0 {

			for i, n := 0, *an; i < n; i++ {
				work <- struct{}{}
			}

		} else {

			for {
				work <- struct{}{}
			}
		}

		close(work)
	}()

	if len(*conf) > 0 {

		all, _ := ioutil.ReadFile(*conf)
		for i, c := 0, *ac; i < c; i++ {
			wg.Add(1)
			go func() {
				js := gurlib.NewJsEngine(&client)
				defer wg.Done()
				for range work {
					js.VM.Run(string(all))
				}
			}()
		}

	} else {

		g.MemInit()
		for i, c := 0, *ac; i < c; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range work {
					g.Send()
				}
			}()
		}
	}

	wg.Wait()
}
