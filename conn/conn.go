// Copyright 2018 The guonaihong Authors. All rights reserved.
// Use of this source code is governed by a apache-style
// license that can be found in the LICENSE file.

package conn

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/guonaihong/flag"
	"github.com/guonaihong/gurl/gurlib"
	"github.com/guonaihong/gurl/input"
	"github.com/guonaihong/gurl/output"
	"github.com/guonaihong/gurl/report"
	"github.com/guonaihong/gurl/task"
	"io"
	_ "io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Conn struct {
	task.Task
	NetType   string
	Addr      string
	TimeOut   time.Duration
	ConnSleep time.Duration
	Multiple  bool
	Mode      string
	LType     string
	SendRate  string
	Data      string
	Conf      string
	KArgs     string
	bench     bool

	writeStream bool
	merge       bool
	host        string
	portOrPath  string
	url         string
	report      *report.Report
}

func (c *Conn) Init() {
	if c.bench {
		//c.report =
		c.report = report.NewReport(c.C, c.N, c.url)
		if len(c.Duration) > 0 {
			if t := gurlib.ParseTime(c.Duration); int(t) > 0 {
				c.report.SetDuration(t)
			}
		}
		c.report.Start()
	}
}

func (c *Conn) WaitAll() {
	if c.report != nil {
		c.report.Wait()
	}
	close(c.Out)
}

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

func head2buf(h string, body []byte) []byte {
	buf := &bytes.Buffer{}

	switch h {
	case "u64be":
		h := [8]byte{}
		binary.BigEndian.PutUint64(h[:], uint64(len(body)))
		buf.Write(h[:])
	case "u32be":
		h := [4]byte{}
		binary.BigEndian.PutUint32(h[:], uint32(len(body)))
		buf.Write(h[:])
	case "u16be":
		h := [2]byte{}
		binary.BigEndian.PutUint16(h[:], uint16(len(body)))
		buf.Write(h[:])
	}

	buf.Write(body)
	return buf.Bytes()
}

func (g *Conn) Copy(dst io.Writer, src io.Reader) (int64, error) {

	if len(g.LType) == 0 {
		return io.Copy(dst, src)
	}

	buf := make([]byte, 4096)
	total := int64(0)

	for {

		n, err := src.Read(buf)
		if err != nil {
			break
		}

		n, err = dst.Write(head2buf(g.LType, buf[:n]))
		if err != nil {
			break
		}
		total += int64(n)
	}

	return total, nil
}

func cancelled(message gurlib.Message) bool {
	select {
	case <-message.InDone:
		return true
	default:
		return false
	}
}

func (g *Conn) tcpWork(c net.Conn) {
	to := g.TimeOut

	if to > 0 {
		go func() {
			time.Sleep(to)
			c.Close()
			return
		}()
	}
	defer c.Close()
	if g.Mode == "c2c" {
		g.Copy(c, c)
		return
	}

	g.Copy(c, os.Stdin)
	//g.Copy(os.Stdout, c)
}

func (g *Conn) ListenUdp() {
	ls := g.Addr
	var err error
	var udpAddr *net.UDPAddr

	if udpAddr, err = net.ResolveUDPAddr("udp4", ls); err != nil {
		fmt.Println(err)
		return
	}

	con, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Printf("listen log at ", udpAddr, " error:", err)
		return
	}

	defer con.Close()

	for {
		buf := make([]byte, 65535)
		count, _, err := con.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("read message from %v failed\n", con.RemoteAddr())
			return
		}

		os.Stdout.Write(buf[:count])
	}

}

func (g *Conn) ListenTcp() {

	netType, ls := g.NetType, g.Addr

	l, err := net.Listen(netType, ls)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	defer l.Close()

	for {

		c, err2 := l.Accept()

		if err2 != nil {
			fmt.Printf("%s\n", err2)
			continue
		}

		if g.Multiple {
			go g.tcpWork(c)
			continue
		}

		g.tcpWork(c)
	}
}

type rate struct {
	B int
	T int64
}

func genRate(rateStr string, rv **rate) {
	rates := strings.Split(rateStr, "/")

	if len(rates) != 2 {
		return
	}

	rates[0] = strings.ToLower(rates[0])
	rates[1] = strings.ToLower(rates[1])

	r := rate{}
	fmt.Sscanf(rates[0], "%d", &r.B)
	fmt.Sscanf(rates[1], "%d", &r.T)
	switch {
	case strings.HasSuffix(rates[0], "b"):
	case strings.HasSuffix(rates[0], "kb"):
		r.B *= 1024
	case strings.HasSuffix(rates[0], "mb"):
		r.B *= 1024 * 1024
	}

	switch {
	case strings.HasSuffix(rates[1], "ms"):
		r.T = r.T * int64(time.Millisecond)
	case strings.HasSuffix(rates[1], "s"):
		r.T = r.T * int64(time.Second)
	}

	if r.B <= 0 {
		return
	}

	if r.T <= 0 {
		return
	}

	*rv = &r
}

type connResult struct {
	wb int
	rb int
}

func (c *Conn) OnceClient(host, portOrPath string) (rv connResult, err error) {

	var conn net.Conn
	var rate *rate

	netType := c.NetType
	if c.ConnSleep > 0 {
		time.Sleep(c.ConnSleep)
	}

	if pos := strings.Index(portOrPath, "-"); pos == -1 {

		addr := host + ":" + portOrPath
		if netType == "unix" {
			addr = portOrPath
		}

		if int(c.TimeOut) > 0 {
			conn, err = net.DialTimeout(netType, addr, c.TimeOut)
		} else {
			conn, err = net.Dial(netType, addr)
		}

		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		defer conn.Close()

		var r io.Reader
		r = os.Stdin

		if len(c.Data) > 0 {
			if !strings.HasPrefix(c.Data, "@") {
				r = strings.NewReader(c.Data)
			} else {
				fd, err0 := os.Open(c.Data[1:])
				if err != nil {
					err = err0
					return
				}
				r = fd
				defer fd.Close()
			}
		}

		genRate(c.SendRate, &rate)
		in := make([]byte, 1024*80)
		out := make([]byte, 1024*80)

		if rate != nil && rate.B > 0 {
			in = make([]byte, rate.B)
		}

		n := 0
		for {
			n, err = r.Read(in)
			if n <= 0 {
				break
			}

			if err != nil {
				break
			}

			if rate != nil && rate.T > 0 {
				time.Sleep(time.Duration(rate.T))
			}

			n, err = conn.Write(in[:n])
			if err != nil {
				return
			}
			rv.wb += n
			if c.TimeOut > 0 {
				conn.SetReadDeadline(time.Now().Add(c.TimeOut))
			}
			n, err = conn.Read(out)
			if err != nil {
				break
			}

			rv.rb += n
			if c.TimeOut > 0 {
				conn.SetReadDeadline(time.Time{})
			}
			if !c.bench {
				os.Stdout.Write(out[:n])
			}
		}

		return
	}

	rangePort(host, portOrPath, func(addr string) {

		if int(c.TimeOut) > 0 {
			conn, err = net.DialTimeout(netType, addr, c.TimeOut)
		} else {
			conn, err = net.Dial(netType, addr)
		}

		if err != nil {
			fmt.Printf("%s\n", err)
			return
			//return err
		}
		defer conn.Close()

		io.Copy(conn, os.Stdin)

	})

	return
}

func CheckPort(host, port string) error {

	rangePort(host, port, func(addr string) {

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			//fmt.Printf("%s\n", err)
		} else {
			conn.Close()
			fmt.Printf("Connection to %s port [tcp/mysql] succeeded!\n", addr)
		}
	})
	return nil
}

func rangePort(host, port string, cb func(addr string)) (err error) {
	var start, end int

	ports := strings.Split(port, "-")
	if len(ports) != 2 {
		err = fmt.Errorf("%s port range not valid\n", os.Args[0])
		return
	}

	start, err = strconv.Atoi(ports[0])
	if err != nil {
		err = fmt.Errorf("%s port range not valid\n", os.Args[0])
		return
	}

	end, err = strconv.Atoi(ports[1])
	if err != nil || end > 65535 || end <= 0 {
		err = fmt.Errorf("%s port range not valid\n", os.Args[0])
		return
	}

	for ; start <= end; start++ {
		addr := fmt.Sprintf("%s:%d", host, start)
		cb(addr)
	}
	return nil
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

	commandlLine.Author("guonaihong https://github.com/guonaihong/conn")
	commandlLine.Parse(argv)

	var err error
	defer func() {
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}()

	if *inputMode {
		input.Main(*inputRead, *inputFields, *inputSetKey, message)
		return
	}

	if *outputMode {
		output.WriteFile(*outputWrite, *outputKey, message)
		return
	}

	conn := Conn{
		Task: task.Task{
			Duration:   *duration,
			N:          *an,
			Work:       make(chan string, 1000),
			ReadStream: *readStream,
			Message:    message,
			Rate:       *rate,
			C:          *ac,
		},

		TimeOut:     gurlib.ParseTime(*w),
		ConnSleep:   gurlib.ParseTime(*connSleep),
		Multiple:    *k,
		Mode:        *mode,
		LType:       *LType,
		SendRate:    *sendRate,
		Data:        *data,
		KArgs:       *confArgs,
		Conf:        *conf,
		bench:       *bench,
		writeStream: *writeStream,
		merge:       *merge,
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

	conn.host = host
	conn.portOrPath = portOrPath

	conn.Task.Processer = &conn
	conn.Task.RunMain()
}

func (c *Conn) SubProcess(work chan string) {

	//var inJson map[string]string

	for v := range work {
		taskNow := time.Now()
		rv, err := c.OnceClient(c.host, c.portOrPath)
		if err != nil && err != io.EOF {
			if c.report != nil {
				c.report.AddErr(err)
			} else {
				cmdErr(err)
			}
			continue
		}

		if c.report != nil {
			c.report.Add(taskNow, rv.rb, rv.wb)
		}

		if len(v) > 0 && v[0] == '{' {
		}
	}

}
