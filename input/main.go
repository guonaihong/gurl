package input

import (
	"fmt"
	"github.com/guonaihong/gurl/gurlib"
	"os"
)

func Main(fileName string, fields string, replaceKey string, message gurlib.Message) {
	out, err := ReadFile(fileName, fields, replaceKey)
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
