package conn

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/yuin/gopher-lua"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Message struct {
	In      chan lua.LValue
	Out     chan lua.LValue
	InDone  chan lua.LValue
	OutDone chan lua.LValue
}

type Conn struct {
	NetType   string
	Addr      string
	TimeOut   time.Duration
	ConnSleep time.Duration
	Multiple  bool
	Mode      string
	LType     string
	SendRate  string
	duration  string
	Data      string
	Conf      string
	KArgs     string
	rate      int
	c         int
	n         int
	bench     bool
	work      chan struct{}
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

func cancelled(message Message) bool {
	select {
	case <-message.InDone:
		return true
	default:
		return false
	}
}

func (g *Conn) LuaMain(message Message) {

	conf := g.Conf
	kargs := g.KArgs
	all, err := ioutil.ReadFile(conf)
	if err != nil {
		fmt.Printf("ERROR:%s\n", err)
		os.Exit(1)
	}

	wg := sync.WaitGroup{}

	work := make(chan struct{}, 1000)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(work)

		n := g.n
		if n >= 0 {
			for i := 0; i < n; i++ {
				work <- struct{}{}
			}
			return
		}

		for {
			work <- struct{}{}
		}
	}()

	c := g.c
	defer func() {
		wg.Wait()
		close(message.Out)
		close(message.OutDone)
	}()

	for i := 0; i < c; i++ {

		wg.Add(1)
		go func() {
			defer wg.Done()

			l := NewLuaEngine(kargs)
			l.L.SetGlobal("in_ch", lua.LChannel(message.In))
			l.L.SetGlobal("out_ch", lua.LChannel(message.Out))

			for {
				if g.n != 0 {
					select {
					case _, ok := <-work:
						if !ok {
							return
						}
					}
				} else {
					if cancelled(message) && len(message.In) == 0 {
						return
					}
				}

				err = l.L.DoString(string(all))
				if err != nil {
					fmt.Printf("%s\n", err)
					os.Exit(1)
				}
			}
		}()
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

func ParseTime(t string) (rv time.Duration) {

	t0 := 0
	for k, _ := range t {
		v := int(t[k])
		switch {
		case v >= '0' && v <= '9':
			t0 = t0*10 + (v - '0')
		case v == 's':
			rv += time.Duration(t0) * time.Second
			t0 = 0
		case v == 'm':
			if k+1 < len(t) && t[k+1] == 's' {
				rv += time.Duration(t0) * time.Millisecond
				t0 = 0
				k++
				continue
			}
			rv += time.Duration(t0*60) * time.Second
			t0 = 0
		case v == 'h':
			rv += time.Duration(t0*60*60) * time.Second
			t0 = 0
		case v == 'd':
			rv += time.Duration(t0*60*60*24) * time.Second
			t0 = 0
		case v == 'w':
			rv += time.Duration(t0*60*60*24*7) * time.Second
			t0 = 0
		case v == 'M':
			rv += time.Duration(t0*60*60*24*7*31) * time.Second
			t0 = 0
		case v == 'y':
			rv += time.Duration(t0*60*60*24*7*31*365) * time.Second
			t0 = 0
		}
	}

	return
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
