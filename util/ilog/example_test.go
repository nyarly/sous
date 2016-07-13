package ilog

import (
	"fmt"
	"log"
	"os"
)

type Thing struct {
	log, debug func(...interface{})
	n          int
}

func NewThing() *Thing {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)
	return &Thing{
		log:   log.Println,
		debug: log.Println,
	}
}

func (t *Thing) SetLogFunc(f func(...interface{}))   { t.log = f }
func (t *Thing) SetDebugFunc(f func(...interface{})) { t.debug = f }

func (t *Thing) DoSomething(what string) {
	t.n++
	t.log("I am " + what + "ing")
	t.debug("Call", t.n)
}

func Example() {
	t := NewThing()
	t.DoSomething("jump")
	writeInfo := func(m Message) { fmt.Println(m.Text) }
	writeDebug := func(m Message) { fmt.Println(m.Text) }
	w := NewWatcher(100, writeInfo, writeDebug)
	w.Watch(t)
	w.WatchDebug(t)
	t.DoSomething("sing")

	w.CloseWait()
	// Output:
	// I am jumping
	// Call 1
	// I am singing
	// Call 2
}
