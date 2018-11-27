package main

import (
	"github.com/guonaihong/conn"
	"os"
)

func main() {
	conn.Main(os.Args[0], os.Args[1:])
}
