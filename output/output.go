package output

import (
	"encoding/json"
	"fmt"
	"github.com/guonaihong/gurl/gurlib"
	"os"
	"strings"
)

func WriteFile(fileName string, onlyWriteKey string, message gurlib.Message) error {
	var fd *os.File
	var err error

	fd = os.Stdout //default
	switch fileName {
	case "stdout", "": //stdout
	case "stderr":
		fd = os.Stderr
	default:
		fd, err = os.Create(fileName)
		if err != nil {
			return err
		}
		defer fd.Close()
	}

	//fmt.Printf("========\n")
	onlyWriteMap := map[string]struct{}{}
	if len(onlyWriteKey) > 0 {
		onlyWriteList := strings.FieldsFunc(onlyWriteKey, func(r rune) bool { return r == ',' })
		for _, v := range onlyWriteList {
			onlyWriteMap[v] = struct{}{}
		}
	}

	for v := range message.In {
		if len(onlyWriteMap) > 0 {
			only := map[string]interface{}{}
			err := json.Unmarshal([]byte(v), &only)
			if err != nil {
				fmt.Printf("writefile:%s, v(%s)\n", err, v)
				continue
			}

			only2 := map[string]interface{}{}
			for k, _ := range onlyWriteMap {
				only2[k] = only[k]
			}

			all, err := json.Marshal(only2)
			if err != nil {
				fmt.Printf("%s\n", err)
				continue
			}
			fd.Write(all)
			continue
		}

		fd.WriteString(v)
	}

	return nil
}
