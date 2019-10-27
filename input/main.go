package input

import (
	"fmt"
	"github.com/guonaihong/gurl/core"
	"os"
)

func Main(fileName string, fields string, renameKey string, message core.Message) {
	out, err := ReadFile(fileName, fields, renameKey)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	defer close(message.Out)

	for {
		select {
		case v, ok := <-out.JsonOut:
			if !ok {
				return
			}
			message.Out <- v
		}
	}
}
