package main

import (
	"github.com/guonaihong/flag"
	"github.com/guonaihong/gurl"
	"github.com/guonaihong/gurl/conn"
	"github.com/guonaihong/gurl/wsurl"
	"os"
)

func main() {
	parent := flag.NewParentCommand(os.Args[0])

	parent.SubCommand("http", "Use the http subcommand", func() {
		gurl.Main(os.Args[0], parent.Args())
	})

	parent.SubCommand("ws, websocket", "Use the websocket subcommand", func() {
		wsurl.Main(os.Args[0], parent.Args())
	})

	parent.SubCommand("tcp, udp", "Use the tcp or udp subcommand", func() {
		conn.Main(os.Args[0], parent.Args())
	})

	parent.Parse(os.Args[1:])
}
