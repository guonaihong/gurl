package pipe

import (
	"github.com/guonaihong/gurl/gurlib"
	"log"
	"sync"
)

type Chan struct {
	ch   chan string
	done chan string
}

func Main(name string, args []string, subMain func(gurlib.Message, string, []string)) {

	var wg sync.WaitGroup
	var cmds [][]string

	prevPos := 0
	for k, v := range args {
		if v == "|" {
			cmds = append(cmds, args[prevPos:k])
			prevPos = k + 1
		}
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if len(cmds) == 0 {
		cmds = [][]string{args}
	}

	if prevPos != 0 && prevPos < len(args) {
		cmds = append(cmds, args[prevPos:])
	}

	var channel []*Chan
	wg.Add(len(cmds))
	defer wg.Wait()

	for k, v := range cmds {
		channel = append(channel, &Chan{
			done: make(chan string),
			ch:   make(chan string, 1000),
		})

		go func(ch []*Chan, k int, v []string) {
			defer func() {
				wg.Done()
			}()

			m := gurlib.Message{
				Out:     ch[k].ch,
				OutDone: ch[k].done,
				K:       k,
			}

			if k > 0 {
				m.In = ch[k-1].ch
				m.InDone = ch[k-1].done
			}

			subMain(m, name, v)
		}(channel, k, v)
	}

}
