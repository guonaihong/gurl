package main

import (
	_ "github.com/guonaihong/flag"
	"github.com/guonaihong/gurl"
	"os"
)

func main() {
	/*
		parent := NewParentCommand(os.Args[0])

		parent.SubCommand("http", "Use the http subcommand", func() {
			t.Logf("call add subcommand")
		})

		parent.SubCommand("ws, websocket", "Use the websocket subcommand", func() {
			t.Logf("call add subcommand")
		})

		parent.SubCommand("tcp, udp", "Use the tcp or udp subcommand", func() {
			t.Logf("call add subcommand")
		})

		parent.Parse(os.Args[1:])
	*/

	gurl.Main(os.Args[0], os.Args[1:])
}
