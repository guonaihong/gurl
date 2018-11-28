package input

import (
	"encoding/json"
	"os"
	"strings"
)

type StreamFile struct {
	JsonOut chan string
	file    *os.File
}

func ReadFile(fileName string, fieldSeparator string) (*StreamFile, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	sf := StreamFile{
		JsonOut: make(chan string, 10),
		file:    file,
	}

	scanner := bufio.NewScanner(file)

	go func() {

		for scanner.Scan() {

			ls := strings.Split(scanner.Text())
			m := make(map[string]string)
			for k, v := range ls {
				m[fmt.Sprintf("rf.col.%d", k)] = v
			}

			all, err := json.Marshal(m)
			if err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}

			sf.JsonOut <- string(all)
		}

		close(sf.JsonOut)
	}()

	return &sf, nil
}

func (sf *StreamFile) Close() {
	sf.file.Close()
}
