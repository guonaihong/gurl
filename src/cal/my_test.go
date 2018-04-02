package cal

import (
	"strings"
	"testing"
)

func TestStack(t *testing.T) {
	s := Stack{}
	s.Push("hello")
	s.Push("world")

	for {
		v, err := s.Pop()
		if err != nil {
			break
		}
		t.Logf("value is %s\n", v)
	}
}

func TestQueue(t *testing.T) {
	q := QueueNew(100)
	for i := 0; i < 2; i++ {
		for j := 0; ; j++ {
			err := q.Put(i)
			if err != nil {
				t.Logf("queue put:%s:j(%d), i(%d)\n", err, j, i)
				break
			}
		}

		for j := 0; ; j++ {
			_, err := q.Get()
			if err != nil {
				t.Logf("queue get:%s:j(%d), i(%d)\n", err, j, i)
				break
			}
		}
	}
}

func TestProcess(t *testing.T) {
	expr := ExprNew()

	//eStr := "( 1 + 2 ) * 3"
	//eStr := "5 + ( ( 1 + 2 ) * 4 ) - 3"
	//eStr := "1 - 2"
	eStr := "2 / 1"
	eslice := strings.Split(eStr, " ")
	t.Logf("in(%#v)\n", eslice)
	expr.Process(eslice)
	v, _ := expr.Operator.Top()
	result, _ := v.(int)
	t.Logf("result:%d\n", result)
}
