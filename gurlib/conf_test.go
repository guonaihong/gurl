package gurlib

import (
	"fmt"
	"testing"
)

func TestParseBool(t *testing.T) {
	valMap := map[string]interface{}{
		"http_code":        "200",
		"parent_http_code": "200",
	}

	ss := []string{
		"$number_eq($parent_http_code, $http_code)",
	}

	c := New(valMap)
	for _, v := range ss {
		val := c.ParseBool([]byte(v), nil, true)
		fmt.Printf("parse bool result:v(%s), rv(%t)\n", v, val)
	}
}

func TestParse(t *testing.T) {
	ss := []string{
		"$parent_url/eval/mp3",
		"http://$base_url/eval/mp3",
		"http://$base_url/eval/mp3",
		"$parent_url/eval/mp3",
		"$parent_url/eval/mp3",
		"session-id:$uuid()+:sh",
		"$uuid()-area+bb",
		"<$uuid()>",
		"$parent_url",
		"$chinese_eval_url",
		"$text=$parent_http_body",
	}

	valMap := map[string]interface{}{
		"parent_url":       "http://127.0.0.1:5001",
		"base_url":         "127.0.0.1:5001",
		"chinese_eval_url": "http://127.0.0.1:14987/eval/pcm",
		"parent_http_body": "{'score':22}",
	}

	c := New(valMap)
	for _, v := range ss {
		val := c.Parse([]byte(v), nil, true)
		fmt.Printf("%s\n", val)
	}
}

func TestParseSlice(t *testing.T) {
	ss := []string{
		"$chinese[$i]",
		"$chinese[$i]",
	}

	valMap := map[string]interface{}{
		"chinese": []string{"1111", "2222"},
	}

	c := New(valMap)
	for k, v := range ss {
		valMap["i"] = k
		//valMap["i"] = fmt.Sprintf("%d", k)
		val := c.Parse([]byte(v), valMap, true)
		fmt.Printf("%s---> %v\n", v, val)
	}
}
