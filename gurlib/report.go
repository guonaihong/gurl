package gurlib

import (
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

type Report struct {
	percent    chan int
	addr       string
	laddr      string
	serverName string
	port       string
	path       string
	percents   []int
	c          int
	n          int
	doneNum    int32
	recvN      int
	weNum      int32
	totalRead  int32
	totalBody  int32
	step       int
	length     int
	startNow   time.Time
	quit       chan struct{}
}

func NewReport(c, n int, url string) *Report {
	step := 0
	if n > 150 {
		if step = n / 10; step < 100 {
			step = 100
		}
	}

	r := &Report{
		percent:  make(chan int, 1000),
		quit:     make(chan struct{}, 1),
		laddr:    url,
		startNow: time.Now(),
		c:        c,
		n:        n,
		step:     step,
	}

	r.parseUrl()

	return r
}

func (r *Report) AddErrNum() {
	atomic.AddInt32(&r.weNum, 1)
}

func (r *Report) Cal(now time.Time, resp *Response) {
	if r.serverName == "" {
		r.serverName = resp.Header.Get("Server")
	}

	atomic.AddInt32(&r.doneNum, 1)
	r.calBody(resp)

	r.percent <- int(time.Now().Sub(now) / time.Millisecond)
}

func (r *Report) calBody(resp *Response) {

	bodyN := len(resp.Body)

	r.length = bodyN

	hN := len(resp.Status)
	hN += len(resp.Proto)
	hN += 1 //space
	hN += 2 //\r\n
	for k, v := range resp.Header {
		hN += len(k)

		for _, hv := range v {
			hN += len(hv)
		}
		hN += 2 //:space
		hN += 2 //\r\n
	}

	hN += 2

	atomic.AddInt32(&r.totalRead, int32(hN))
	atomic.AddInt32(&r.totalRead, int32(bodyN))
}

func (r *Report) report() {

	timeTake := time.Now().Sub(r.startNow)
	percents := r.percents

	fmt.Printf("\n\n")
	fmt.Printf("Server Software:        %s\n", r.serverName)
	fmt.Printf("Server Hostname:        %s\n", r.addr)
	fmt.Printf("Server Port:            %s\n", r.port)
	fmt.Printf("\n")

	fmt.Printf("Document Path:          %s\n", r.path)
	fmt.Printf("Document Length:        %d bytes\n", r.length)
	fmt.Printf("\n")

	fmt.Printf("Concurrency Level:      %d\n", r.c)
	fmt.Printf("Time taken for tests:   %.3f seconds\n", timeTake.Seconds())
	fmt.Printf("Complete requests:      %v\n", r.recvN)
	fmt.Printf("Failed requests:        %v\n", r.doneNum-int32(r.recvN))
	if r.weNum > 0 {
		fmt.Printf("Write errors:           %v\n", r.weNum)
	}

	fmt.Printf("Total transferred:      %d bytes\n", r.totalRead)
	fmt.Printf("HTML transferred:       %v bytes\n", r.totalBody)
	fmt.Printf("Requests per second:    %.2f [#/sec] (mean)\n",
		float64(r.doneNum)/timeTake.Seconds())
	fmt.Printf("Time per request:       %.3f [ms] (mean)\n",
		float64(r.c)*float64(timeTake)/float64(time.Millisecond)/float64(r.doneNum))
	fmt.Printf("Time per request:       %.3f [ms] (mean, across all concurrent requests)\n",
		float64(timeTake)/float64(time.Millisecond)/float64(r.doneNum))
	fmt.Printf("Transfer rate:          %.2f [Kbytes/sec] received\n",
		float64(r.totalRead)/float64(1000)/timeTake.Seconds())

	sort.Slice(percents, func(i, j int) bool {
		return percents[i] < percents[j]
	})

	if len(percents) > 1 {
		fmt.Printf("Percentage of the requests served within a certain time (ms)\n")
		fmt.Printf("  50%%    %d\n", percents[int(float64(len(percents))*0.5)])
		fmt.Printf("  66%%    %d\n", percents[int(float64(len(percents))*0.66)])
		fmt.Printf("  75%%    %d\n", percents[int(float64(len(percents))*0.75)])
		fmt.Printf("  80%%    %d\n", percents[int(float64(len(percents))*0.80)])
		fmt.Printf("  90%%    %d\n", percents[int(float64(len(percents))*0.90)])
		fmt.Printf("  95%%    %d\n", percents[int(float64(len(percents))*0.95)])
		fmt.Printf("  98%%    %d\n", percents[int(float64(len(percents))*0.98)])
		fmt.Printf("  99%%    %d\n", percents[int(float64(len(percents))*0.99)])
		fmt.Printf(" 100%%    %d\n", percents[len(percents)-1])
	}
}

func (r *Report) parseUrl() {

	addr := r.laddr
	if pos := strings.Index(addr, "http://"); pos != -1 {
		addr = addr[pos+7:]
	}

	if pos := strings.Index(addr, "/"); pos != -1 {
		r.path = addr[pos:]
		addr = addr[:pos]
	}

	if pos := strings.Index(addr, ":"); pos != -1 {
		r.port = addr[pos+1:]
		addr = addr[:pos]
	}

	fmt.Printf("Benchmarking %s (be patient)\n", addr)
}

func (r *Report) StartReport() {
	go func() {
		defer func() {
			if r.step > 0 {
				fmt.Printf("Finished %d requests\n", r.recvN)
			}
			r.quit <- struct{}{}
		}()

		for v := range r.percent {

			r.recvN++
			if r.step > 0 && r.recvN%r.step == 0 {
				fmt.Printf("Completed %d requests\n", r.recvN)
			}

			r.percents = append(r.percents, v)
		}

	}()

}

func (r *Report) Wait() {
	close(r.percent)
	<-r.quit
	r.report()
}
