package main

import (
	"github.com/guonaihong/wsurl"
	"os"
)

func main() {
	wsurl.Main(os.Args[0], os.Args[1:])
}
