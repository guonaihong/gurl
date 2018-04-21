package main

import (
	"core"
	"demo"
	_ "fmt"
	"github.com/NaihongGuo/flag"
	"github.com/robfig/cron"
	"gurl"
	"strings"
	"sync"
	_ "unsafe"
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
	conf := flag.String("K", "", "Read config from FILE")
	output := flag.String("o", "", "Write to FILE instead of stdout")
	method := flag.String("X", "", "Specify request command to use")
	demoName := flag.String("demo", "", "Generate the default yml configuration file(The optional value is for, if)")
	gen := flag.String("gen", "", "Generate the default yml configuration file(The optional value is cmd, root, child, func, all)")
	toJson := flag.StringSlice("J", []string{}, `Turn key:value into {"key": "value"})`)
	url := flag.String("url", "", "Specify a URL to fetch")
	an := flag.Int("an", 1, "Number of requests to perform")
	ac := flag.Int("ac", 1, "Number of multiple requests to make")

	flag.Parse()

	as := flag.Args()
	Url := *url
	if *url == "" && len(as) == 0 && len(*conf) == 0 && len(*demoName) == 0 && len(*gen) == 0 {
		flag.Usage()
		return
	}

	if len(*demoName) > 0 {
		demo.Usage(*demoName)
		return
	}

	if len(as) > 0 {
		Url = as[0]
	}

	Url = modifyUrl(Url)

	c := gurl.Gurl{
		GurlCore: gurl.GurlCore{
			Base: core.Base{
				Method: *method,
				F:      *forms,
				H:      *headers,
				O:      *output,
				J:      *toJson,
				Jfa:    *jfa,
				Url:    Url,
			},
		},
	}

	multiGurl := gurl.MultiGurl{}

	if len(*conf) > 0 {
		c.ConfigInit(*conf, &multiGurl.ConfFile)
		gurl.MergeCmd(&multiGurl.ConfFile.Cmd, &c, "append")
	} else {
		gurl.MergeCmd(&multiGurl.ConfFile.Cmd, &c, "set")
	}

	//fmt.Printf("%v\n", multiGurl.ConfFile)
	if len(*gen) > 0 {
		multiGurl.GenYaml(*gen)
		return
	}

	if len(*cronExpr) > 0 {
		cron := cron.New()

		defer cron.Stop()

		cron.AddFunc(*cronExpr, func() {

			gurl.MultiGurlInit(&multiGurl)
			multiGurl.Send()
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

	//fmt.Printf("-->%d\n", unsafe.Sizeof(multiGurl))
	for i, c := 0, *ac; i < c; i++ {
		wg.Add(1)
		m := multiGurl
		go func(m *gurl.MultiGurl) {
			defer wg.Done()
			for range work {
				gurl.MultiGurlInit(m)
				m.Send()
			}
		}(&m)
	}

	wg.Wait()
}
