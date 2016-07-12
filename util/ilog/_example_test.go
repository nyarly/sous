package ilog_test

import (
	"fmt"
	"os"

	"github.com/opentable/sous/util/ilog"
)

type Goat struct {
	Name string
	Age  int
	ilog.Mixin
}

type Fields map[string]interface{}

func (g *Goat) AddYear() {
	g.DebugData("AddYear called", Fields{"Name": g.Name, "Age": g.Age})
	defer func() {
		g.DebugData("AddYear finished", Fields{"Name": g.Name, "Age": g.Age})
	}()
	g.Age++
	g.Logf("%s aged by one year, from %d to %d", g.Name, g.Age-1, g.Age)
}

func Example() {
	w := ilog.NewWatcher(10, writeLog, writeDebug)
	g := &Goat{Name: "Algernon", Age: 6}
	w.Watch(g)
	w.WatchDebug(g)
	g.AddYear()
	g.AddYear()
	w.CloseWait()
	// Output:
	// DEBUG AddYear called :: map[Age: 6 Name:Algernon]
	// INFO  Algernon aged by one year, from 6 to 7
	// DEBUG AddYear finished :: map[Name:Algernon Age: 7]
	// DEBUG AddYear called :: map[Name:Algernon Age: 7]
	// INFO  Algernon aged by one year, from 7 to 8
	// DEBUG AddYear finished :: map[Name:Algernon Age: 8]
}

func writeLog(message string, data interface{}) {
	fmt.Fprintln(os.Stdout, "INFO ", message)
}

func writeDebug(message string, data interface{}) {
	fmt.Fprintf(os.Stdout, "DEBUG %s :: % +v\n", message, data)
}
