package report

import (
	"fmt"
	"sort"
	"sync/atomic"
	"time"
)

type result struct {
	time float64
}

type Report struct {
	c          int
	n          int
	url        string
	step       int
	connNum    int
	errN       int32
	readBytes  int64
	writeBytes int64
	allResult  chan result
	allTimes   []float64
	waitQuit   chan struct{}
	quit       chan struct{}
	startNow   time.Time
	duration   time.Duration
}

func NewReport(c, n int, url string) *Report {
	r := Report{}
	r.c, r.n = c, n
	r.url = url

	if n > 150 {
		if r.step = n / 10; r.step < 100 {
			r.step = 100
		}
	}

	r.quit = make(chan struct{})
	r.waitQuit = make(chan struct{})
	r.allResult = make(chan result, 1000)
	r.startNow = time.Now()
	return &r
}

func (r *Report) report() {
	timeTake := time.Now().Sub(r.startNow)

	fmt.Printf("Concurrency Level:        %d\n", r.c)
	fmt.Printf("Time taken for tests:     %v\n", timeTake)
	fmt.Printf("Connected:                %d\n", r.connNum)
	fmt.Printf("Disconnected:             %d\n", r.errN)
	fmt.Printf("Failed:                   %d\n", r.errN)
	//todo:Calculate the websocket protocol header
	fmt.Printf("Total transferred:        %d\n", r.writeBytes)
	fmt.Printf("Total received            %d\n", r.readBytes)
	fmt.Printf("Requests per second:      %d [#/sec] (mean)\n", int(float64(r.connNum)/timeTake.Seconds()))
	fmt.Printf("Time per request:         %.3f [ms] (mean)\n",
		float64(r.c)*float64(timeTake)/float64(time.Millisecond)/float64(r.connNum))
	fmt.Printf("Time per request:         %.3f [ms] (mean, across all concurrent requests)\n",
		float64(timeTake)/float64(time.Millisecond)/float64(r.connNum))
	fmt.Printf("Transfer rate:            %.3f [Kbytes/sec] received\n", float64(r.readBytes)/float64(r.connNum))

	sort.Slice(r.allTimes, func(i, j int) bool {
		return r.allTimes[i] < r.allTimes[j]
	})

	fmt.Printf("\n")
	if len(r.allTimes) > 1 {
		fmt.Printf("Percentage of the requests served within a certain time (ms)\n")
		fmt.Printf("    50%%    %0.2fms\n", r.allTimes[int(float64(len(r.allTimes))*0.5)])
		fmt.Printf("    66%%    %0.2fms\n", r.allTimes[int(float64(len(r.allTimes))*0.65)])
		fmt.Printf("    75%%    %0.2fms\n", r.allTimes[int(float64(len(r.allTimes))*0.75)])
		fmt.Printf("    80%%    %0.2fms\n", r.allTimes[int(float64(len(r.allTimes))*0.80)])
		fmt.Printf("    90%%    %0.2fms\n", r.allTimes[int(float64(len(r.allTimes))*0.90)])
		fmt.Printf("    95%%    %0.2fms\n", r.allTimes[int(float64(len(r.allTimes))*0.95)])
		fmt.Printf("    98%%    %0.2fms\n", r.allTimes[int(float64(len(r.allTimes))*0.98)])
		fmt.Printf("    99%%    %0.2fms\n", r.allTimes[int(float64(len(r.allTimes))*0.99)])
		fmt.Printf("    100%%   %0.2fms\n", r.allTimes[int(float64(len(r.allTimes)-1))])
	}
}

func (r *Report) Add(openTime time.Time, rb int, wb int) {
	atomic.AddInt64(&r.readBytes, int64(rb))
	atomic.AddInt64(&r.writeBytes, int64(wb))
	r.allResult <- result{
		time: float64(time.Now().Sub(openTime) / time.Millisecond),
	}
}

func (r *Report) SetDuration(t time.Duration) {
	r.duration = t
}

func (r *Report) AddErr() {
	atomic.AddInt32(&r.errN, int32(1))
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

func (r *Report) Start() {
	fmt.Printf("Connecting to to %s\n", r.url)
	go func() {

		defer func() {
			fmt.Printf("\n    Finished %d connections\n\n", r.connNum)
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
					r.connNum++
					if r.step > 0 && r.connNum%r.step == 0 {
						now := time.Now()

						fmt.Printf("    Opened %15d connections: [%s]\n",
							r.connNum, genTimeStr(now))
					}

					r.allTimes = append(r.allTimes, v.time)
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
						r.connNum, genTimeStr(now))

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

					r.connNum++
					r.allTimes = append(r.allTimes, v.time)
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
