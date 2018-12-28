package gurl

import (
	"encoding/json"
	"fmt"
	"github.com/guonaihong/flag"
	"github.com/guonaihong/gurl/gurlib"
	url2 "github.com/guonaihong/gurl/gurlib/url"
	"github.com/guonaihong/gurl/input"
	"github.com/guonaihong/gurl/output"
	"github.com/guonaihong/gurl/task"
	"io"
	_ "io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	DefaultConnections = 10000
)

type GurlCmd struct {
	task.Task
	cronExpr    string
	bench       bool
	writeStream bool
	merge       bool
	report      *gurlib.Report
	*gurlib.Gurl
}

func parse(val map[string]string, g *gurlib.Gurl, inJson string) {

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
	for k, v := range g.J {
		g.J[k] = r.Replace(v)
	}

	for k, v := range g.F {
		g.F[k] = r.Replace(v)
	}

	for k, v := range g.H {
		g.H[k] = r.Replace(v)
	}

	for k, v := range g.Jfa {
		g.Jfa[k] = r.Replace(v)
	}

	if len(g.A) > 0 {
		g.A = r.Replace(g.A)
	}
	g.Url = r.Replace(g.Url)
	g.O = r.Replace(g.O)
	g.Body = []byte(r.Replace(string(g.Body)))
}

func (cmd *GurlCmd) Init() {
	cmd.Gurl.ParseInit()
	if cmd.bench {
		cmd.report = gurlib.NewReport(cmd.C, cmd.N, cmd.Gurl.Url) // todo
		if len(cmd.Duration) > 0 {
			if t := gurlib.ParseTime(cmd.Duration); int(t) > 0 {
				cmd.report.SetDuration(t) // todo
			}
		}
		cmd.report.StartReport()
	}
}

func (cmd *GurlCmd) WaitAll() {
	if cmd.report != nil {
		cmd.report.Wait()
	}

	close(cmd.Message.Out)
}

//todo
func (cmd *GurlCmd) streamWriteJson(rsp *gurlib.Response, err error, inJson map[string]string) {
	m := map[string]interface{}{}
	m["err"] = ""
	m["status_code"] = fmt.Sprintf("%d", rsp.StatusCode)
	m["body"] = string(rsp.Body)
	m["header"] = rsp.Header

	if err != nil {
		m["err"] = err.Error()
	}

	output.WriteStream(m, inJson, cmd.merge, cmd.Message)
	//todo
}

func (cmd *GurlCmd) SubProcess(work chan string) {
	g := *cmd.Gurl //这里是copy不是操作指针
	g0 := gurlib.Gurl{Client: g.Client}
	g0.GurlCore = *gurlib.CopyAndNew(&g.GurlCore)
	var inJson map[string]string

	for v := range work {
		if len(v) > 0 && v[0] == '{' {
			inJson = map[string]string{}
			g.GurlCore = *gurlib.CopyAndNew(&g.GurlCore)

			parse(inJson, &g, v)
			//fmt.Printf("read work:%s\n", v)
		}

		taskNow := time.Now()
		rsp, err := g.Send()
		if cmd.writeStream {
			cmd.streamWriteJson(rsp, err, inJson)
		}

		if err != nil {
			if cmd.report != nil {
				cmd.report.AddErrNum()
			} else {
				CmdErr(err)
			}
			continue
		}

		if cmd.report != nil {
			cmd.report.Cal(taskNow, rsp)
		}

		if len(v) > 0 && v[0] == '{' {
			g = g0
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

func Main(message gurlib.Message, argv0 string, argv []string) {
	commandlLine := flag.NewFlagSet(argv0, flag.ExitOnError)

	headers := commandlLine.StringSlice("H, header", []string{}, "Pass custom header LINE to server (H)")
	forms := commandlLine.StringSlice("F, form", []string{}, "Specify HTTP multipart POST data (H)")
	formStrings := commandlLine.StringSlice("form-string", []string{}, "Specify HTTP multipart POST data (H)")
	jfa := commandlLine.StringSlice("Jfa", []string{}, "Specify HTTP multipart POST json data (H)")
	jfaStrings := commandlLine.StringSlice("Jfa-string", []string{}, "Specify HTTP multipart POST json data (H)")
	cronExpr := commandlLine.String("cron", "", "Cron expression")
	outputFileName := commandlLine.String("o, output", "stdout", "Write to FILE instead of stdout")
	oflag := commandlLine.String("oflag", "", "Control the way you write(append|line|trunc)")
	method := commandlLine.String("X, request", "", "Specify request command to use")
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
	userAgent := commandlLine.String("A, user-agent", "gurl", "Send User-Agent STRING to server")
	duration := commandlLine.String("duration", "", "Duration of the test")
	connectTimeout := commandlLine.String("connect-timeout", "", "Maximum time allowed for connection")

	readStream := commandlLine.Bool("r, read-stream", false, "Read data from the stream")
	writeStream := commandlLine.Bool("w, write-stream", false, "Write data from the stream")
	merge := commandlLine.Bool("m, merge", false, "Combine the output results into the output")

	inputMode := commandlLine.Bool("I, input-model", false, "open input mode")
	inputRead := commandlLine.String("R, input-read", "", "open input file")
	inputFields := commandlLine.String("input-fields", " ", "sets the field separator")
	inputSetKey := commandlLine.String("skey, input-setkey", "", "Set a new name for the default key")

	outputMode := commandlLine.Bool("O, output-mode", false, "open output mode")
	outputKey := commandlLine.String("wkey, write-key", "", "Key that can be write")
	outputWrite := commandlLine.String("W, output-write", "", "open output file")

	commandlLine.Author("guonaihong https://github.com/guonaihong/gurl")
	commandlLine.Parse(argv)

	if *inputMode {
		input.Main(*inputRead, *inputFields, *inputSetKey, message)
		return
	}

	if *outputMode {
		output.WriteFile(*outputWrite, *outputKey, message)
		return
	}

	if *listen != "" {
		httpEcho(*listen)
		return
	}

	as := commandlLine.Args()
	Url := *URL
	if *URL == "" && len(as) == 0 && !*bench {
		commandlLine.Usage()
		return
	}

	if len(as) > 0 {
		Url = as[0]
	}

	Url = url2.ModifyUrl(Url)

	if *cpus > 0 {
		runtime.GOMAXPROCS(*cpus)
	}

	/*
		dialer := &net.Dialer{
			Timeout: gurlib.ParseTime(*connectTimeout),
		}
	*/

	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: *conns,
			//Dial:                dialer.Dial,
		},
		Timeout: gurlib.ParseTime(*connectTimeout),
	}

	g := gurlib.Gurl{
		Client: &client,
		GurlCore: gurlib.GurlCore{
			Method: *method,
			F:      *forms,
			H:      *headers,
			O:      *outputFileName,
			J:      *toJson,
			Jfa:    *jfa,
			Url:    Url,
			Flag:   toFlag(*outputFileName, *oflag),
			Body:   []byte(*data),
			V:      *verbose,
			A:      *userAgent,
		},
	}

	g.AddFormStr(*formStrings)
	g.AddJsonFormStr(*jfaStrings)

	cmd := GurlCmd{
		Task: task.Task{
			Duration:   *duration,
			N:          *an,
			Work:       make(chan string, 1000),
			ReadStream: *readStream,
			Message:    message,
			Rate:       *rate,
			C:          *ac,
		},

		writeStream: *writeStream,
		merge:       *merge,
		cronExpr:    *cronExpr,
		Gurl:        &g,
		bench:       *bench,
	}

	/*
		if len(*cronExpr) > 0 {
			cmd.Cron(&client)
		}
	*/

	cmd.Producer()

	if *bench {
		g.O = ""
	}

	cmd.Task.Processer = &cmd
	cmd.Task.RunMain()
}
