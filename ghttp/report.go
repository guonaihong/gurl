package ghttp

import (
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

type result struct {
	time       float64
	statusCode int
}

type Report struct {
	allResult   chan result
	waitQuit    chan struct{}
	quit        chan struct{}
	allTimes    []float64
	statusCodes map[int]int
	startNow    time.Time
	addr        string
	laddr       string
	serverName  string
	port        string
	path        string
	c           int
	n           int
	recvN       int
	step        int
	length      int
	duration    time.Duration
	doneNum     int32
	weNum       int32
	totalRead   int32
	totalBody   int32
}

func NewReport(c, n int, url string) *Report {
	step := 0
	if n > 150 {
		if step = n / 10; step < 100 {
			step = 100
		}
	}

	r := &Report{
		allResult:   make(chan result, 1000),
		quit:        make(chan struct{}, 1),
		waitQuit:    make(chan struct{}, 1),
		laddr:       url,
		startNow:    time.Now(),
		c:           c,
		n:           n,
		step:        step,
		statusCodes: make(map[int]int, 2),
	}

	r.parseUrl()

	return r
}

func (r *Report) SetDuration(t time.Duration) {
	r.duration = t
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

	r.allResult <- result{
		time:       float64(time.Now().Sub(now)) / float64(time.Millisecond),
		statusCode: resp.StatusCode,
	}
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

	atomic.AddInt32(&r.totalBody, int32(bodyN))
	atomic.AddInt32(&r.totalRead, int32(hN))
	atomic.AddInt32(&r.totalRead, int32(bodyN))
}

func (r *Report) report() {

	timeTake := time.Now().Sub(r.startNow)
	allTimes := r.allTimes

	fmt.Printf("\n\n")
	fmt.Printf("Server Software:        %s\n", r.serverName)
	fmt.Printf("Server Hostname:        %s\n", r.addr)
	fmt.Printf("Server Port:            %s\n", r.port)
	fmt.Printf("\n")

	fmt.Printf("Document Path:          %s\n", r.path)
	fmt.Printf("Document Length:        %d bytes\n", r.length)
	fmt.Printf("\n")

	fmt.Printf("Status Codes:          ")
	for k, v := range r.statusCodes {
		fmt.Printf(" %d:%d  ", k, v)
	}
	fmt.Printf("[code:count]\n")

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

	sort.Slice(allTimes, func(i, j int) bool {
		return allTimes[i] < allTimes[j]
	})

	if len(allTimes) > 1 {
		fmt.Printf("Percentage of the requests served within a certain time (ms)\n")
		fmt.Printf("  50%%    %0.2fms\n", allTimes[int(float64(len(allTimes))*0.5)])
		fmt.Printf("  66%%    %0.2fms\n", allTimes[int(float64(len(allTimes))*0.66)])
		fmt.Printf("  75%%    %0.2fms\n", allTimes[int(float64(len(allTimes))*0.75)])
		fmt.Printf("  80%%    %0.2fms\n", allTimes[int(float64(len(allTimes))*0.80)])
		fmt.Printf("  90%%    %0.2fms\n", allTimes[int(float64(len(allTimes))*0.90)])
		fmt.Printf("  95%%    %0.2fms\n", allTimes[int(float64(len(allTimes))*0.95)])
		fmt.Printf("  98%%    %0.2fms\n", allTimes[int(float64(len(allTimes))*0.98)])
		fmt.Printf("  99%%    %0.2fms\n", allTimes[int(float64(len(allTimes))*0.99)])
		fmt.Printf(" 100%%    %0.2fms\n", allTimes[len(allTimes)-1])
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

func genTimeStr(now time.Time) string {
	year, month, day := now.Date()
	hour, min, sec := now.Clock()

	return fmt.Sprintf("%4d-%02d-%02d %02d:%02d:%02d.%06d",
		year,
		month,
		day,
		hour,
		min,
		sec,
		now.Nanosecond()/1e3,
	)
}

func (r *Report) StartReport() {
	go func() {
		defer func() {
			fmt.Printf("  Finished  %15d requests\n", r.recvN)
			r.waitQuit <- struct{}{}
		}()

		if r.step > 0 {
			for {
				select {
				case _, ok := <-r.quit:
					if !ok {
						return
					}
				case v := <-r.allResult:
					r.recvN++
					if r.step > 0 && r.recvN%r.step == 0 {
						now := time.Now()

						fmt.Printf("    Opened %15d connections: [%s]\n",
							r.recvN, genTimeStr(now))
					}

					r.allTimes = append(r.allTimes, v.time)
					r.statusCodes[v.statusCode]++
				}
			}
		} else {
			begin := time.Now()
			interval := r.duration / 10

			if interval == 0 {
				interval = time.Second
			}
			nTick := time.NewTicker(interval)
			count := 1
			for {
				select {
				case <-nTick.C:
					now := time.Now()

					fmt.Printf("  Completed %15d requests [%s]\n",
						r.recvN, genTimeStr(now))

					count++
					next := begin.Add(time.Duration(count * int(interval)))
					if newInterval := next.Sub(time.Now()); newInterval > 0 {
						nTick = time.NewTicker(newInterval)
					} else {
						nTick = time.NewTicker(time.Millisecond * 100)
					}
				case v, ok := <-r.allResult:
					if !ok {
						return
					}

					r.recvN++
					r.allTimes = append(r.allTimes, v.time)
					r.statusCodes[v.statusCode]++
				case _, ok := <-r.quit:
					if !ok {
						return
					}
				}
			}
		}

	}()

}

func (r *Report) Wait() {
	close(r.quit)
	<-r.waitQuit
	r.report()
}
