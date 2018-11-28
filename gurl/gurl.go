package main

import (
	"github.com/guonaihong/flag"
	"github.com/guonaihong/gurl"
	"github.com/guonaihong/gurl/conn"
	"github.com/guonaihong/gurl/pipe"
	"github.com/guonaihong/gurl/wsurl"
	"os"
)

func main() {
	parent := flag.NewParentCommand(os.Args[0])

	parent.SubCommand("http", "Use the http subcommand", func() {
		pipe.Main(os.Args[0], parent.Args(), gurl.Main)
	})

	parent.SubCommand("ws, websocket", "Use the websocket subcommand", func() {
		pipe.Main(os.Args[0], parent.Args(), wsurl.Main)
	})

	parent.SubCommand("tcp, udp", "Use the tcp or udp subcommand", func() {
		pipe.Main(os.Args[0], parent.Args(), conn.Main)
	})

	parent.Parse(os.Args[1:])
}
