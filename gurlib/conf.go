package gurlib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	VARIABLE = "var"
	constant = "const"
	FUNCVAL  = "func"
)

type Status struct {
	wordStart       int
	isVariable      bool
	isSlice         bool
	left            int
	n               int
	curlyBracesLeft bool
}

type Conf struct {
	rootMap map[string]interface{}
	funcMap map[string]func(v *FuncVal) error
}

type Val struct {
	Type string
	v    interface{}
}

type FuncVal struct {
	FuncName string
	CallArgs []string
	RvArgs   Val
	Parent   map[string]interface{}
}

func ConfNew(rootMap map[string]interface{}) *Conf {

	conf := &Conf{}

	conf.rootMap = rootMap

	conf.funcMap = map[string]func(v *FuncVal) error{
		"num":            conf.Num,
		"uuid":           conf.GenUUID,
		"read_file":      conf.ReadFile,
		"number_eq":      conf.NumberEq,
		"number_ne":      conf.NumberNe,
		"string_eq":      conf.StringEq,
		"string_ne":      conf.StringNe,
		"format_json":    conf.FormatJson,
		"shell_exec":     conf.ShellExec,
		"shell_exec_str": conf.ShellExecStr,
		"string_fields":  conf.StringFields,
	}

	return conf
}

func (c *Conf) AddFunc(name string, funcVal func(v *FuncVal) error) {
	if name == "" {
		return
	}

	c.funcMap[name] = funcVal
}

func (c *Conf) findFunc(funcName string) bool {
	_, ok := c.funcMap[funcName]
	return ok
}

func (c *Conf) callFunc(val *FuncVal) (*Val, bool) {
	cb, ok := c.funcMap[val.FuncName]
	if !ok {
		return nil, ok
	}

	err := cb(val)
	if err != nil {
		panic(fmt.Sprintf("func(%s) args(%s) ", val.FuncName, val.CallArgs) +
			err.Error())
	}

	return &val.RvArgs, true
}

func (c *Conf) findVal(s string, parent map[string]interface{}) (interface{}, bool) {
	v, ok := c.rootMap[s]

	if !ok {
		v, ok = parent[s]
	}
	return v, ok
}

func (c *Conf) ReadFile(v *FuncVal) error {
	v.CallArgs[0] = c.ParseString([]byte(v.CallArgs[0]), v.Parent, false)
	fd, err := os.Open(v.CallArgs[0])
	if err != nil {
		return err
	}

	defer fd.Close()

	all, err := ioutil.ReadAll(fd)
	if err != nil {
		return err
	}
	v.RvArgs = Val{Type: VARIABLE, v: string(all)}
	return nil
}

func (c *Conf) GenUUID(v *FuncVal) error {
	u1 := uuid.Must(uuid.NewV4())
	v.RvArgs = Val{Type: VARIABLE, v: u1.String()}

	return nil
}

func (c *Conf) StringEq(v *FuncVal) error {
	if len(v.CallArgs) < 2 {
		return errors.New("Too few parameters for calling $$string_eq()")
	}

	v.CallArgs[0] = c.ParseString([]byte(v.CallArgs[0]), v.Parent, false)

	rv := true
	for _, arg := range v.CallArgs[1:] {

		arg = c.ParseString([]byte(arg), v.Parent, false)

		if v.CallArgs[0] != arg {
			rv = false
			break
		}
	}

	v.RvArgs = Val{Type: VARIABLE, v: rv}
	return nil
}

func (c *Conf) StringNe(v *FuncVal) error {
	err := c.StringEq(v)
	if err != nil {
		return err
	}
	v.RvArgs.v = !v.RvArgs.v.(bool)
	return nil
}

func ParseExpr(s string, e *Expr) (rv []string) {

	n := 0

	for i, l := 0, len(s); i < l; i++ {

		n = i

		for ; i < l && s[i] >= '0' && s[i] <= '9'; i++ {
		}

		if n != i {
			rv = append(rv, s[n:i])
			//continue
		}

		if i < l && e.IsOperator(s[i]) {
			rv = append(rv, s[i:i+1])
			continue
		}
	}
	return
}

func (c *Conf) Num(v *FuncVal) error {
	expr := ExprNew()

	eslice := ParseExpr(v.CallArgs[0], expr)
	expr.Process(eslice)
	rv, _ := expr.Operator.Pop()
	v.RvArgs = Val{Type: VARIABLE, v: rv}
	return nil
}

func (c *Conf) NumberEq(v *FuncVal) error {
	if len(v.CallArgs) < 2 {
		return errors.New("Too few parameters for calling $$number_eq()")
	}

	v.CallArgs[0] = c.ParseString([]byte(v.CallArgs[0]), v.Parent, false)
	n, err := strconv.ParseInt(v.CallArgs[0], 0, 0)
	if err != nil {
		return err
	}

	rv := true
	for _, arg := range v.CallArgs[1:] {

		arg = c.ParseString([]byte(arg), v.Parent, false)
		n1, err := strconv.ParseInt(arg, 0, 0)
		if err != nil {
			return err
		}

		if n != n1 {
			rv = false
			break
		}
	}

	v.RvArgs = Val{Type: VARIABLE, v: rv}
	return nil
}

func (c *Conf) NumberNe(v *FuncVal) error {
	err := c.NumberEq(v)
	if err != nil {
		return err
	}

	v.RvArgs.v = !v.RvArgs.v.(bool)
	return nil
}

func (c *Conf) FormatJson(v *FuncVal) error {
	if len(v.CallArgs) == 0 {
		return errors.New("Too few parameters for calling $$format_json()")
	}

	var prettyJson bytes.Buffer

	v.CallArgs[0] = c.ParseString([]byte(v.CallArgs[0]), v.Parent, false)

	err := json.Indent(&prettyJson, []byte(v.CallArgs[0]), "", "  ")
	if err != nil {
		return err
	}

	v.RvArgs = Val{Type: VARIABLE, v: prettyJson.String()}

	return nil
}

func (c *Conf) StringFields(v *FuncVal) error {
	if len(v.CallArgs) == 0 {
		return errors.New("Too few parameters for calling $$format_json()")
	}

	rv := strings.Fields(v.CallArgs[0])

	v.RvArgs = Val{Type: VARIABLE, v: rv}

	return nil
}

func (c *Conf) ShellExecStr(v *FuncVal) error {
	if len(v.CallArgs) == 0 {
		return errors.New("Too few parameters for calling $$shell_exec()")
	}

	execNames := v.CallArgs
	var cmd *exec.Cmd

	if len(execNames) == 1 {
		cmd = exec.Command(execNames[0])
	} else {
		cmd = exec.Command(execNames[0], execNames[1:]...)
	}

	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	v.RvArgs = Val{Type: VARIABLE, v: out.String()}
	return nil
}

func (c *Conf) ShellExec(v *FuncVal) error {
	if len(v.CallArgs) == 0 {
		return errors.New("Too few parameters for calling $$shell_exec()")
	}

	execNames := v.CallArgs
	var cmd *exec.Cmd

	if len(execNames) == 1 {
		cmd = exec.Command(execNames[0])
	} else {
		cmd = exec.Command(execNames[0], execNames[1:]...)
	}

	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return nil
}

func (c *Conf) parseFuncArgs(s string, funcName string, i *int) *FuncVal {

	j := *i
	start := j

	arg := ""
	var val *FuncVal

	val = &FuncVal{FuncName: funcName}

	ok := c.findFunc(funcName)
	if !ok {
		panic("not find function " + funcName + "()")
	}

	brackets := 0
	first := true
	for j < len(s) {

		if s[j] == ' ' {
			goto next
		}

		if s[j] == '(' {
			brackets++
			if first {
				start = j + 1
			}
			first = false
			goto next
		}

		if s[j] == ',' && brackets == 1 {
			arg = strings.TrimSpace(s[start:j])
			j++
			start = j
		}

		if len(arg) > 0 {
			val.CallArgs = append(val.CallArgs, arg)
			arg = ""
		}

		if s[j] == ')' {
			brackets--
			if brackets != 0 {
				goto next
			}

			arg = strings.TrimSpace(s[start:j])

			if len(arg) > 0 {
				val.CallArgs = append(val.CallArgs, arg)
			}

			*i = j + 1

			return val
		}
	next:
		j++
	}

	return nil
}

func (c *Conf) ParseBool(s []byte, parent map[string]interface{}, parseAssign bool) bool {
	if bytes.Equal([]byte("yes"), bytes.TrimSpace(s)) {
		return true
	}

	rv := c.Parse(s, parent, parseAssign)
	rvs, ok := rv.(bool)

	if ok {
		return rvs
	}

	return false
}

func (c *Conf) ParseString(s []byte, parent map[string]interface{}, parseAssign bool) string {
	rv := c.Parse(s, parent, parseAssign)

	if rvs, ok := rv.(string); ok {

		return rvs
	}

	//TODO
	if rvs, ok := rv.(int); ok {
		return strconv.Itoa(rvs)
	}

	return string(s)
}

func (c *Conf) ParseSlice(s []byte, parent map[string]interface{}, parseAssign bool) []string {
	rv := c.Parse(s, parent, parseAssign)
	rvslice, ok := rv.([]string)

	if ok {
		return rvslice
	}

	return []string{string(s)}
}

func (c *Conf) ParseName(s []byte) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '$' {
			return strings.TrimSpace(string(s[i+1:]))
		}
	}
	return string(s)
}

func (c *Conf) Parse(s []byte, parent map[string]interface{}, parseAssign bool) (val interface{}) {

	if !parseAssign {
		return c.ParseCore(s, parent, parseAssign)
	}

	const (
		NOT_FOUND = int64(-3)
		FOUND     = int64(-2)
		IOTA      = int64(-1)
	)

	var (
		i           = 0
		err         error
		wStart      = 0
		wEnd        = 0
		word        string
		isVariable  = false
		slicePos    = NOT_FOUND
		lefbrackets = -1
		isEqual     bool
	)

	for i = 0; i < len(s); {
		v := s[i]

		if !isVariable && v == '$' {
			isVariable = true
			wStart = i + 1
		}

		if !isVariable {
			goto next
		}

		if isEqual && v != ' ' && v != '=' {
			// Skip the space at the beginning of value
			wEnd = i
			goto isVar
		}

		if slicePos == NOT_FOUND && v == '[' {
			lefbrackets = i + 1
			slicePos = FOUND
			wEnd = i
		}

		if slicePos == FOUND && v == ']' {
			if string(bytes.TrimSpace(s[lefbrackets:i])) == "..." {
				slicePos = IOTA
			} else {
				indexVal := c.Parse(s[lefbrackets:i], parent, false)
				if _, ok := indexVal.(int); !ok {
					slicePos, err = strconv.ParseInt(indexVal.(string), 0, 0)
					if err != nil {
						panic(err)
					}
				}
			}
		}

		if v == '=' {
			if slicePos == NOT_FOUND {
				// variable
				wEnd = i
			}

			word = string(bytes.TrimSpace(s[wStart:wEnd]))

			// skip ...]=
			if slicePos == IOTA {
				wEnd = i + 1
			}

			isEqual = true
		}

	next:
		i++
	}

	if len(word) == 0 {
		wEnd = 0
	}
isVar:
	val = c.ParseCore(s[wEnd:], parent, parseAssign)

	if len(word) > 0 {
		//fmt.Printf("word(%s), val(%s)\n", word, val)
		var oldSlice []string
		if old, ok := c.rootMap[word]; ok {
			oldSlice, _ = old.([]string)
		}

		switch {

		case slicePos == IOTA:
			oldSlice = append(oldSlice, val.(string))
			//fmt.Printf("word(%s) val=(%s) oldslice(%s)\n", word, val, oldSlice)

		case slicePos >= 0:

			switch {

			case int(slicePos) < len(oldSlice)-1:
				oldSlice[slicePos] = val.(string)

			case int(slicePos) == len(oldSlice)-1:
				val = append(oldSlice, val.(string))

			default:

				oldSlice = append(oldSlice,
					make([]string, int(slicePos)-len(oldSlice)+1)...)

				oldSlice[slicePos] = val.(string)
			}
		}

		if slicePos != NOT_FOUND {
			val = oldSlice
		}
		//TODO
		c.rootMap[word] = val
	}

	return val
}

func getWord(token bytes.Buffer, status *Status) (word string, isVar bool) {

	word = token.String()
	if word[0] == '$' {
		word = word[1:]
		isVar = true
	}

	if status.curlyBracesLeft {
		word = word[1:]

		status.curlyBracesLeft = false
	}

	if word[len(word)-1] == '(' {
		word = word[:len(word)]
	}
	return
}

func (c *Conf) GetVarAndFunc(s string, word string,
	token *bytes.Buffer, status *Status, i *int, vals *[]Val, parent map[string]interface{},
	parseAssign bool) {

	//fmt.Printf("1.token:%s:%d:%d\n", token.String(), i, len(s))
	//fmt.Printf("2.token:%s:%d:%d\n", word, i, len(s))
	ok := c.findFunc(word)
	if ok {
		status.isVariable = false
		token.Reset()

		funcVal := c.parseFuncArgs(string(s), string(s[status.wordStart:*i]), i)
		for k, v := range funcVal.CallArgs {
			funcVal.CallArgs[k] = c.ParseString([]byte(v), parent, parseAssign)
		}

		funcVal.Parent = parent
		v, ok := c.callFunc(funcVal)
		if !ok {
			panic("call func " + funcVal.FuncName)
		}

		*vals = append(*vals, *v)

		// now
		return
	}

	vv, ok := c.findVal(word, parent)
	if !ok {
		panic("not found -->" + word)
	}

	if _, ok := vv.([]string); ok {
		// $slice
		if *i+1 >= len(s) || *i+1 < len(s) && s[*i] != '[' {
			*vals = append(*vals, Val{Type: VARIABLE, v: vv})
			status.isVariable = false
			return
		}
		status.isSlice = true
		//next
		return
	}

	status.isVariable = false

	if v, ok := vv.(int); ok {
		vv = strconv.Itoa(v)
	}

	*vals = append(*vals, Val{Type: VARIABLE, v: vv})
	token.Reset()
	//next
	return
}

func (c *Conf) ParseCore(s []byte, parent map[string]interface{}, parseAssign bool) interface{} {

	var token bytes.Buffer

	var vals []Val

	var out bytes.Buffer

	pos := bytes.Index(s, []byte("$"))

	if pos == -1 {
		return string(s)
	}

	var status Status

	i := 0
	for i = 0; i < len(s); {

		v := s[i]

		if status.isSlice {
			if v != '[' && v != ']' {
				goto next
			}

			if v == '[' {
				status.left = i
				goto next
			}

			word, _ := getWord(token, &status)
			vv, ok := c.findVal(word, parent)
			if !ok {
				panic("not found:" + token.String())
			}

			pos := c.ParseCore(s[status.left+1:i], parent, parseAssign)

			status.n, ok = pos.(int)
			if !ok {
				n1, err := strconv.ParseInt(pos.(string), 0, 0)
				if err != nil {
					panic(err.Error())
				}
				status.n = int(n1)
			}

			slice := vv.([]string)
			vals = append(vals, Val{Type: VARIABLE, v: slice[status.n]})
			token.Reset()
			status.isSlice = false
			status.isVariable = false
			status.n = 0
			goto next
		}

		if status.isVariable == false && v == '$' {

			if token.Len() > 0 {
				vals = append(vals, Val{Type: constant, v: token.String()})
				token.Reset()
			}

			status.wordStart = i + 1

			status.isVariable = true
		}

		if status.isVariable {

			size := 0
			if v >= '0' && v <= '9' || v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z' || v == '_' || v == '$' || v == '{' {
				if v == '{' {
					status.curlyBracesLeft = true
				}

				_, size = utf8.DecodeRune(s[i:])
				if size == 1 {
					token.WriteByte(v)
					goto next
				}
			}

			word, _ := getWord(token, &status)

			c.GetVarAndFunc(string(s), word, &token, &status, &i, &vals, parent, parseAssign)

			if v == '}' {
				goto next
			}

			goto now
		}

		token.WriteByte(v)

	next:
		i++
	now:
	}

	if token.Len() > 0 {
		word, isVar := getWord(token, &status)
		if isVar {
			c.GetVarAndFunc(string(s), word, &token, &status, &i, &vals, parent, parseAssign)
		}

		token := token.String()
		vals = append(vals, Val{Type: constant, v: token})
	}

	for _, v := range vals {

		_, ok := v.v.(string)

		if !ok {
			return v.v
		}

		out.WriteString(v.v.(string))

	}

	//fmt.Printf("out-->%s\n", out.String())

	return out.String()
}
