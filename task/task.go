package task

import (
	"github.com/guonaihong/gurl/core"
	"github.com/guonaihong/gurl/utils"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Task struct {
	Duration   string
	Work       chan string
	ReadStream bool
	N          int
	C          int
	Rate       int
	core.Message
	Wg sync.WaitGroup
	Processer
}

type Processer interface {
	Init()
	SubProcess(chan string)
	WaitAll()
}

func (T *Task) Producer() {
	work, n := T.Work, T.N

	if T.ReadStream {
		go func() {
			for v := range T.In {
				work <- v
			}
			close(work)
		}()
		return
	}

	if len(T.Duration) > 0 {

		if t := utils.ParseTime(T.Duration); int(t) > 0 {
			T.N = -1

			ticker := time.NewTicker(t)
			go func() {

				defer func() {
					close(work)
					for range work {
					}
				}()

				for {
					select {
					case <-ticker.C:
						return
					case work <- "":
					}

				}

			}()
			return
		}
	}

	go func() {

		defer close(work)
		if T.N >= 0 {

			for i := 0; i < n; i++ {
				work <- ""
			}

			return
		}

		for {
			work <- ""
		}

		return

	}()
}

func (T *Task) RunMain() {

	work, wg := T.Work, &T.Wg

	sig := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	signal.Notify(sig, os.Interrupt)

	T.Init()

	begin := time.Now()

	interval := 0
	if T.Rate > 0 {
		interval = int(time.Second) / T.Rate
	}

	if interval > 0 {
		count := 0
		oldwork := work
		work = make(chan string, 1000)
		wg.Add(1)
		go func() {
			defer func() {
				close(work)
				wg.Done()
			}()

			n := T.N
			for {
				next := begin.Add(time.Duration(count * interval))
				time.Sleep(next.Sub(time.Now()))

				select {
				case _, ok := <-oldwork:
					if !ok {
						return
					}
				default:
				}

				work <- ""
				if count++; count == n {
					return
				}
			}
		}()
	}

	for i, c := 0, T.C; i < c; i++ {

		wg.Add(1)

		go func() {
			defer wg.Done()
			T.SubProcess(work)
		}()
	}

	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

end:
	for {
		select {
		case <-sig:
			T.WaitAll()
			break end
		case <-done:
			T.WaitAll()
			break end
		}
	}
}
