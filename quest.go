package interact

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Question params
type Quest struct {
	Choices
	Default
	parent             *Question
	Response           interface{}
	Options, Err, Msg string
	Resolve BoolFunc
}

// Question entity
type Question struct {
	Quest
	prefix
	choices                bool
	resp                   string
	input interface{}
	parent                 *Interact
	Action                 InterfaceFunc
	Subs []*Question
	After, Before ErrorFunc
}

// Default options
type Default struct {
	Text   interface{}
	Status bool
}

// Choice option
type Choice struct {
	Text     string
	Response interface{}
}

// Choices list and prefix color
type Choices struct {
	Alternatives []Choice
	Color        func(...interface{}) string
}

func (q *Question) answer() interface {}{
	return response{answer:q.Response, input:q.input}
}

func (q *Question) append(p prefix) {
	q.prefix = p
}

func (q *Question) father() model {
	return q.parent
}

func (q *Question) ask() (err error) {
	context := &context{model: q}
	if err := context.method(q.Before); err != nil{
		return err
	}
	if q.prefix.Text != nil{
		q.print(q.prefix.Text, " ")
	}else if q.parent != nil && q.parent.Text != nil {
		q.print(q.parent.Text, " ")
	}
	if q.Msg != "" {
		q.print(q.Msg, " ")
	}
	if q.Options != "" {
		q.print(q.Options, " ")
	}
	if q.Default.Status != false {
		q.print(q.Default.Text, " ")
	}
	if q.Alternatives != nil && len(q.Alternatives) > 0 {
		q.multiple()
	}
	if err = q.wait(); err != nil {
		return q.loop(err)
	}
	if err = q.response(); err != nil {
		return q.loop(err)
	}
	if q.Subs != nil && len(q.Subs) > 0 {
		if q.Resolve != nil {
			if q.Resolve(context){
				for _, s := range q.Subs {
					s.parent = q.parent
					s.ask()
				}
			}
		}else{
			for _, s := range q.Subs {
				s.parent = q.parent
				s.ask()
			}
		}
	}
	if q.Action != nil {
		if err := q.Action(context); err != nil {
			q.print(err, " ")
			return q.ask()
		}
	}
	if err := context.method(q.After); err != nil{
		return err
	}
	return nil
}

func (q *Question) wait() error {
	reader := bufio.NewReader(os.Stdin)
	if q.choices {
		q.print(q.Color("?"), " ", "Answer", " ")
	}
	r, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	q.resp = r[:len(r)-1]
	q.input = q.resp
	return nil
}

func (q *Question) response() error {
	var v interface{}
	var err error

	// dafault response
	if len(q.resp) == 0 && q.Default.Status {
		return nil
	}
	// multiple choice
	if q.choices {
		q.Response, err = strconv.ParseInt(q.resp, 10, 64)
		if err != nil {
			return err
		}
		q.input = int(q.Response.(int64))
		q.Response = q.Alternatives[q.Response.(int64)-1].Response
	}

	switch q.Response.(type) {
	case uint:
		if v, err = strconv.ParseUint(q.resp, 10, 32); err == nil {
			q.Response = uint(v.(uint64))
		}
	case uint8:
		if v, err = strconv.ParseUint(q.resp, 10, 8); err == nil {
			q.Response = uint8(v.(uint64))
		}
	case uint16:
		if v, err = strconv.ParseUint(q.resp, 10, 16); err == nil {
			q.Response = uint16(v.(uint64))
		}
	case uint32:
		if v, err = strconv.ParseUint(q.resp, 10, 32); err == nil {
			q.Response = uint32(v.(uint64))
		}
	case uint64:
		q.Response, err = strconv.ParseUint(q.resp, 10, 64)
	case int:
		if v, err = strconv.ParseInt(q.resp, 10, 32); err == nil {
			q.Response = int(v.(int64))
		}
	case int8:
		if v, err = strconv.ParseInt(q.resp, 10, 8); err == nil {
			q.Response = int8(v.(int64))
		}
	case int16:
		if v, err = strconv.ParseInt(q.resp, 10, 16); err == nil {
			q.Response = int16(v.(int64))
		}
	case int32:
		if v, err = strconv.ParseInt(q.resp, 10, 32); err == nil {
			q.Response = int32(v.(int64))
		}
	case int64:
		q.Response, err = strconv.ParseInt(q.resp, 10, 64)
	case float32:
		if v, err = strconv.ParseFloat(q.resp, 64); err == nil {
			q.Response = float32(v.(float64))
		}
	case float64:
		q.Response, err = strconv.ParseFloat(q.resp, 64)
	case bool:
		if q.resp == "y" || q.resp == "yes" {
			q.Response = true
		} else if q.resp == "n" || q.resp == "no" {
			q.Response = false
		} else {
			q.Response, err = strconv.ParseBool(q.resp)
		}
	case time.Duration:
		if v, err = strconv.ParseUint(q.resp, 10, 64); err == nil {
			q.Response = time.Duration(v.(uint64)) * time.Second
		}
	case string:
	default:
		q.Response = strings.ToLower(strings.TrimSpace(q.resp))
	}
	return err
}

func (q *Question) print(a ...interface{}) {
	if q.parent != nil && q.parent.Writer != nil {
		fmt.Fprint(q.parent.Writer, a...)
	} else {
		fmt.Print(a...)
	}

}

func (q *Question) loop(err error) error {
	if q.Err != "" {
		q.print(q.Err, " ")
	}
	return q.ask()
}

func (q *Question) multiple() error{
	for index, i := range q.Alternatives {
		q.print("\n\t", q.Color(index+1, ") "), i.Text, " ")
	}
	q.choices = true
	q.print("\n")
	return nil
}