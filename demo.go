package main

import (
	"os"
)

var forDemo = []byte(`
# for demo
---
root:
  no_send: yes
  H:
  - my-http-header:value
  url: http://127.0.0.1:18080/
  set:
  - $text[...]=1
  - $text[...]=2
  - $text[...]=3
  - $text[...]=4
  - $text[...]=5
  - $text[...]=6
  - $text[...]=7
  - $text[...]=8
  - $text[...]=9

child:
- for:
    range: $text
    k: $i

    H:
    - "$root_header"
    - session-id:$uuid()
    F:
    - text=$text[$i]
    url: "$root_url"

`)

var ifDemo = []byte(`
# if demo
---
root:
  no_send: yes
  H:
    - my-http-header:value
  url: http://127.0.0.1:18080/

child:
- H:
  - "$root_header"
  - session-id:$uuid()
  F:
  - text=hello world
  url: "$root_url"
  if:
    cond: $number_eq($http_code, 200)
    format: "test ok"
`)

var genMap map[string][]byte

func init() {
	genMap = map[string][]byte{
		"for": forDemo,
		"if":  ifDemo,
	}
}

func demoGen(name string) []byte {
	v, ok := genMap[name]
	if ok {
		return v
	}

	return []byte("Unkown")
}

func demoUsage(name string) {
	v := demoGen(name)
	os.Stdout.Write(v)
}
