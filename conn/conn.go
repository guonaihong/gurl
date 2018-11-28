// Copyright 2018 The guonaihong Authors. All rights reserved.
// Use of this source code is governed by a apache-style
// license that can be found in the LICENSE file.

package conn

import (
	"fmt"
	"github.com/guonaihong/flag"
	"github.com/guonaihong/gurl/gurlib"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func cmdErr(err error) {
	if err == nil {
		return
	}

	if noerr, ok := err.(*net.OpError); ok {
		if scerr, ok := noerr.Err.(*os.SyscallError); ok {
			if scerr.Err == syscall.ECONNREFUSED {
				fmt.Printf("wsurl: (7) couldn't connect to host\n")
				os.Exit(7)
			}
		}
	}

	fmt.Printf("%s\n", err)
}

func (c *Conn) Producer() {
	work, n := c.work, c.n

	if len(c.duration) > 0 {
		if t := ParseTime(c.duration); int(t) > 0 {
			n, c.n = -1, -1
		}
	}

	go func() {

		defer close(work)
		if n >= 0 {

			for i := 0; i < n; i++ {
				work <- struct{}{}
			}

			return
		}

		for {
			work <- struct{}{}
		}
		return

	}()
}

func Main(message gurlib.Message, argv0 string, argv []string) {
	commandlLine := flag.NewFlagSet(argv0, flag.ExitOnError)

	listen := commandlLine.Bool("l", false, "Listen mode, for inbound connects")
	rate := commandlLine.Int("rate", 0, "Requests per second")
	duration := commandlLine.String("duration", "", "Duration of the test")
	z := commandlLine.Bool("z", false, "Zero-I/O mode [used for scanning]")
	k := commandlLine.Bool("k", false, "Keep inbound sockets open for multiple connects")
	u := commandlLine.Bool("u", false, "UDP mode")
	U := commandlLine.Bool("U", false, "Use UNIX domain socket")
	w := commandlLine.String("w", "", "Timeout for connects and final net reads")
	conf := commandlLine.String("K", "", "Read lua config from FILE")
	mode := commandlLine.String("M, mode", "c2c", "Set the server to read data from stdin or connect, and write to stdout or connect")
	LType := commandlLine.String("lt, ltype", "", "Add data header type")
	sendRate := commandlLine.String("send-rate", "", "How many bytes of data in seconds")
	connSleep := commandlLine.String("conn-sleep", "", "Sleep for a while before connecting")
	data := commandlLine.String("d, data", "", "Send data to the peer")
	bench := commandlLine.Bool("bench", false, "Run benchmarks test")
	confArgs := commandlLine.String("kargs", "", "Command line parameters passed to the configuration file")
	an := commandlLine.Int("an", 1, "Number of requests to perform")
	ac := commandlLine.Int("ac", 1, "Number of multiple requests to make")
	commandlLine.Author("guonaihong https://github.com/guonaihong/conn")
	commandlLine.Parse(argv)

	var err error
	defer func() {
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}()

	conn := Conn{
		TimeOut:   ParseTime(*w),
		ConnSleep: ParseTime(*connSleep),
		Multiple:  *k,
		Mode:      *mode,
		LType:     *LType,
		SendRate:  *sendRate,
		Data:      *data,
		KArgs:     *confArgs,
		n:         *an,
		c:         *ac,
		Conf:      *conf,
		bench:     *bench,
		duration:  *duration,
		rate:      *rate,
		work:      make(chan struct{}, 1000),
	}

	if len(*conf) > 0 {
		if _, err := os.Stat(*conf); os.IsNotExist(err) {
			fmt.Printf("%s\n", err)
			return
		}

		//conn.LuaMain(message)
		return
	}

	conn.NetType = "tcp"

	if *u {
		conn.NetType = "udp"
	}

	as := commandlLine.Args()
	if len(as) == 0 {
		commandlLine.Usage()
		return
	}

	conn.Addr = ":" + as[0]

	if *U {
		if len(as) < 1 {
			commandlLine.Usage()
			return
		}

		conn.NetType = "unix"
		conn.Addr = as[0]
	}

	if *z == true {
		if len(as) < 2 {
			commandlLine.Usage()
			return
		}

		err = CheckPort(as[0], as[1])
		return
	}

	if *listen {

		if conn.NetType == "unix" {
			syscall.Unlink(conn.Addr)
		}

		if *u {
			conn.ListenUdp()
			return
		}

		if *k {
			conn.ListenTcp()
			return
		}

		conn.ListenTcp()

		return
	}

	conn.Producer()

	host := as[0]
	portOrPath := as[0]
	if len(as) == 2 {
		portOrPath = as[1]
	} else if len(as) == 1 && *U {
		host = ""
	}

	conn.main(host, portOrPath)
}

func (c *Conn) main(host, portOrPath string) {

	var report *Report
	wg := sync.WaitGroup{}

	C := c.c
	n := c.n
	work := c.work

	sig := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)

	signal.Notify(sig, os.Interrupt)

	if c.bench {
		report = NewReport(c.c, c.n, c.Addr)
		report.Start()
	}

	begin := time.Now()

	interval := 0
	if c.rate > 0 {
		interval = int(time.Second) / c.rate
	}

	if len(c.duration) > 0 {
		if t := ParseTime(c.duration); int(t) > 0 {
			wg.Add(1)

			if report != nil {
				report.SetDuration(t)
			}
			workTimeout := make(chan struct{}, 1000)
			work = workTimeout

			ticker := time.NewTicker(t)
			go func() {

				defer func() {
					close(workTimeout)
					for range workTimeout {
					}
					wg.Done()
				}()

				for {
					select {
					case <-ticker.C:
						return
					case workTimeout <- struct{}{}:
					}

				}
			}()
		}
	}

	if interval > 0 {
		count := 0
		oldwork := work
		work = make(chan struct{}, 1000)
		wg.Add(1)
		go func() {
			defer func() {
				close(work)
				wg.Done()
			}()

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

				work <- struct{}{}
				if count++; count == n {
					return
				}
			}
		}()
	}

	for i := 0; i < C; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for range work {
				taskNow := time.Now()
				rv, err := c.OnceClient(host, portOrPath)
				if err != nil && err != io.EOF {
					if report != nil {
						report.AddErr()
					} else {
						cmdErr(err)
					}
					continue
				}

				if report != nil {
					report.Add(taskNow, rv.rb, rv.wb)
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

end:
	for {
		select {
		case <-sig:
			if report != nil {
				report.Wait()
			}
			break end
		case <-done:
			if report != nil {
				report.Wait()
			}
			break end
		}
	}
}
