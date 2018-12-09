package input

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type StreamFile struct {
	JsonOut chan string
	file    *os.File
}

func ReadFile(fileName string, fieldSeparator string, renameKey string) (*StreamFile, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	sf := StreamFile{
		JsonOut: make(chan string, 10),
		file:    file,
	}

	scanner := bufio.NewScanner(file)

	renameMap := map[string]string{}

	if len(renameKey) > 0 {
		defaultKeys := strings.FieldsFunc(renameKey, func(r rune) bool { return r == ',' })
		for _, v := range defaultKeys {
			if pos := strings.Index(v, "="); pos != -1 {
				renameMap[v[pos+1:]] = v[:pos]
			}
		}
	}

	go func() {

		defer func() {
			sf.file.Close()
			close(sf.JsonOut)
		}()

		for scanner.Scan() {

			ls := strings.Split(scanner.Text(), fieldSeparator)
			m := make(map[string]string)
			for k, v := range ls {
				colName := fmt.Sprintf("rf.col.%d", k)
				newKey, ok := renameMap[colName]
				if ok {
					m[newKey] = v
					continue
				}

				m[colName] = v
			}

			all, err := json.Marshal(m)
			if err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}

			sf.JsonOut <- string(all)
		}

	}()

	return &sf, nil
}
