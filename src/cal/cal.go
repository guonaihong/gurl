package cal

import (
	"errors"
	"fmt"
	"strconv"
)

type Stack []interface{}

func (s *Stack) Push(v interface{}) {
	*s = append(*s, v)
}

func (s *Stack) Pop() (interface{}, error) {
	if len(*s) == 0 {
		return nil, errors.New("empty stack")
	}
	v := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return v, nil
}

func (s *Stack) Top() (interface{}, error) {
	if len(*s) == 0 {
		return nil, errors.New("empty stack")
	}
	return (*s)[len(*s)-1], nil
}

func (s *Stack) Len() int {
	return len(*s)
}

type Queue struct {
	Head  int
	Tail  int
	Meet  bool
	Queue []interface{}
}

func QueueNew(size int) *Queue {
	q := Queue{}

	if size <= 0 {
		size = 1
	}

	q.Queue = make([]interface{}, size)
	return &q
}

func (q *Queue) Put(v interface{}) error {
	if q.Head == q.Tail && q.Meet {
		return errors.New("queue full")
	}

	q.Queue[q.Head] = v
	q.Head++

	if q.Head == len(q.Queue) {
		q.Head = 0
	}

	q.Meet = true

	return nil
}

func (q *Queue) Get() (interface{}, error) {
	if q.Head == q.Tail && q.Meet == false {
		return nil, errors.New("empty queue")
	}

	v := q.Queue[q.Tail]
	q.Tail++

	if q.Tail == len(q.Queue) {
		q.Tail = 0
	}

	if q.Head == q.Tail {
		q.Meet = false
	}

	return v, nil
}

func (q *Queue) Cap() int {
	return len(q.Queue)
}

type Expr struct {
	Operator Stack
	Output   *Queue
}

func ExprNew() *Expr {

	e := &Expr{}
	e.Output = QueueNew(100)
	return e
}

func (e *Expr) IsOperator(b byte) bool {
	switch b {
	case '+', '-', '*', '/', '%', '(', ')':
		return true
	}

	return false
}

func (e *Expr) Priority(b byte) int {

	switch b {

	case '+', '-':
		return 1
	case '*', '/', '%':
		return 2
	case '(':
		return 3
	}

	return 0
}

func (e *Expr) ToSuffix(expr []string) {

	for _, exp := range expr {

		if len(exp) == 1 && e.IsOperator(exp[0]) {
			v, err := e.Operator.Top()

			if err != nil {
				e.Operator.Push(exp[0])
				continue
			}

			if exp[0] == ')' {
				for {
					v, err := e.Operator.Pop()
					if err != nil {
						panic("not found (")
					}

					vv := v.(byte)
					if vv == '(' {
						break
					}
					e.Output.Put(vv)
				}
				continue
			}

			stackVal := e.Priority(v.(byte))
			expVal := e.Priority(exp[0])

			switch {
			case expVal > stackVal:
				e.Operator.Push(exp[0])
			case expVal == stackVal:
				if exp[0] != '(' {
					v, err := e.Operator.Pop()
					if err != nil {
						panic("find ( pop fail")
					}

					e.Output.Put(v)
					e.Operator.Push(exp[0])
				} else {
					e.Operator.Push(exp[0])
				}
			default:
				e.Operator.Push(exp[0])
			}

			continue
		}

		e.Output.Put(exp)
	}

	for e.Operator.Len() > 0 {
		v, err := e.Operator.Pop()
		if err != nil {
			break
		}
		e.Output.Put(v)
	}

}

func (e *Expr) PrintQueue() {
	for {
		v, err := e.Output.Get()
		if err != nil {
			break
		}

		switch v.(type) {
		case byte:
			fmt.Printf("%c\n", v)
		case string:
			fmt.Printf("%s\n", v)

		}
	}
}

func (e *Expr) PrintStack() {
	for {
		v, err := e.Operator.Pop()
		if err != nil {
			break
		}

		switch v.(type) {
		case int:
			fmt.Printf("%d\n", v)
		}
	}
}

func (e *Expr) Process(exp []string) {
	e.ToSuffix(exp)
	e.CalVal()
}

func (e *Expr) CalVal() {

	for {

		v, err := e.Output.Get()
		if err != nil {
			break
		}

		b, ok := v.(byte)
		if ok {
			switch b {
			case '+', '-', '*', '/', '%':
				if e.Operator.Len() < 2 {
					panic("error args 0, must is 2")
				}

				v1, _ := e.Operator.Pop()
				n1 := v1.(int)
				v2, _ := e.Operator.Pop()
				n2 := v2.(int)

				switch b {
				case '+':
					e.Operator.Push(n2 + n1)
				case '-':
					e.Operator.Push(n2 - n1)
				case '*':
					e.Operator.Push(n2 * n1)
				case '/':
					e.Operator.Push(n2 / n1)
				case '%':
					e.Operator.Push(n2 % n1)
				}
			default:
				panic("unkown: " + fmt.Sprintf("%c", b))
			}

			continue
		}

		s, ok := v.(string)
		if ok {
			n1, err := strconv.Atoi(s)
			if err != nil {
				panic("error args to int fail")
			}

			e.Operator.Push(n1)
			continue
		}

		if i, ok := v.(string); ok {
			e.Operator.Push(i)
			continue
		}
	}

	if e.Operator.Len() > 1 {
		panic("error expr result argument > 1")
	}

}
