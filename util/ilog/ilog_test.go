package ilog

import (
	"fmt"
	"os"
	"testing"
)

type (
	A struct {
		log, debug func(...interface{})
		Name       string
	}
)

func defaultLogFunc(v ...interface{}) { fmt.Fprintln(os.Stderr, v...) }
func funcOrDefault(f func(...interface{})) func(...interface{}) {
	if f != nil {
		return f
	}
	return defaultLogFunc
}
func (a *A) SetLogFunc(f func(...interface{}))   { a.log = funcOrDefault(f) }
func (a *A) SetDebugFunc(f func(...interface{})) { a.debug = funcOrDefault(f) }

func (a *A) DoSomething(what string) {
	a.log(fmt.Sprintf("%s is %sing"))
}

func Test_TextOnly_Messages(t *testing.T) {
	writeMessage := func(m Message) { t.Log(m.Text) }

	a := &A{Name: "Alf"}

	w := NewWatcher(1024, writeMessage, writeMessage)
	w.Watch(a)

	a.DoSomething("jump")

}
