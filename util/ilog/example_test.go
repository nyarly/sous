package ilog

import "log"

type Thing struct {
	logFunc, debugFunc func(...interface{})
	n                  int
}

func NewThing() *Thing {
	return &Thing{
		logFunc:   log.Println,
		debugFunc: log.Println,
	}
}

func (t *Thing) SetLogFunc(f func(...interface{}))   { t.log = f }
func (t *Thing) SetDebugFunc(f func(...interface{})) { t.debug = f }

func (t *Thing) log(v ...interface{}) {
	if t.logFunc != nil {
		t.logFunc(v...)
	}
}

func (t *Thing) debug(v ...interface{}) {
	if t.debugFunc != nil {
		t.debugFunc(v...)
	}
}

func (t *Thing) DoSomething(what string) {
	t.n++
	t.log("I am " + what + "ing")
	t.debug("I've done", t.n, "things")
}

func Example() {
	t := NewThing()
	t.DoSomething("jump")

	// TODO: Finish this.
}
