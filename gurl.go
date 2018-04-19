package main

import (
	"core"
	"gen"
	"github.com/NaihongGuo/flag"
	"github.com/robfig/cron"
	"gurl"
	"strings"
)

func modifyUrl(u string) string {

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
	cronExpr := flag.String("cron", "", "Cron expression")
	conf := flag.String("K", "", "Read config from FILE")
	output := flag.String("o", "", "Write to FILE instead of stdout")
	method := flag.String("X", "", "Specify request command to use")
	genName := flag.String("gen", "", "Generate the default yml configuration file(The optional value is for, if)")
	toJson := flag.StringSlice("J", []string{}, `Turn key=value into {"key": "value"})`)
	url := flag.String("url", "", "Specify a URL to fetch")

	flag.Parse()

	as := flag.Args()
	Url := *url
	if *url == "" && len(as) == 0 && len(*conf) == 0 && len(*genName) == 0 {
		flag.Usage()
		return
	}

	if len(*genName) > 0 {
		gen.Usage(*genName)
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
				Url:    Url,
			},
		},
	}

	multiGurl := gurl.MultiGurl{
		Cmd: c,
	}

	if len(*conf) > 0 {
		c.ConfigInit(*conf, &multiGurl.ConfFile)
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

	gurl.MultiGurlInit(&multiGurl)
	multiGurl.Send()
}
