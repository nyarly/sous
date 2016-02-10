package core

import (
	"flag"
	"fmt"
	"reflect"
	"time"
)

type Flags struct {
	*flag.FlagSet
	flags map[string]Flag
}

func NewFlags() *Flags {
	return &Flags{flag.NewFlagSet("", flag.ExitOnError), map[string]Flag{}}
}

type Flag struct {
	DefaultValue interface{}
	CreateFlag   func(out interface{})
}

func (f *Flag) Bind(v interface{}) {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr || t.Elem() != reflect.TypeOf(f.DefaultValue) {
		panic(fmt.Sprintf("Bind: received %T; want *%T", v, f.DefaultValue))
	}
	f.CreateFlag(v)
}

func (fs *Flags) Bind(name string, v interface{}) {
	f, ok := fs.flags[name]
	if !ok {
		panic(fmt.Sprintf("flag %s not defined", name))
	}
	f.Bind(v)
}

func (fs *Flags) AddFlag(name, usage string, value interface{}) {
	f := Flag{DefaultValue: value}
	switch v := value.(type) {
	default:
		panic(fmt.Sprintf("AddFlag does not support %T values", value))
	case string:
		f.CreateFlag = func(out interface{}) { o := out.(*string); fs.StringVar(o, name, v, usage) }
	case bool:
		f.CreateFlag = func(out interface{}) { o := out.(*bool); fs.BoolVar(o, name, v, usage) }
	case int:
		f.CreateFlag = func(out interface{}) { o := out.(*int); fs.IntVar(o, name, v, usage) }
	case time.Duration:
		f.CreateFlag = func(out interface{}) { o := out.(*time.Duration); fs.DurationVar(o, name, v, usage) }
	}
	fs.flags[name] = f
}

func init() {
	f := NewFlags()
	f.AddFlag("rebuild", "force rebuild of the target", false)
	f.AddFlag("rebuild-all", "force rebuild of the target and all dependencies", false)
	f.AddFlag("timeout", "per-contract timeout", 5*time.Second)

	var rebuild, rebuildAll bool
	var timeout time.Duration
	f.Bind("rebuild", &rebuild)
	f.Bind("rebuild-all", &rebuildAll)
	f.Bind("timeout", &timeout)

	f.Parse([]string{"-timeout", "7m", "-rebuild", "some", "args"})
	f.PrintDefaults()
	fmt.Printf("Program name: %s; Duration: %s; Rebuild: %v; Remaining args: %+v",
		f.Arg(0), timeout, rebuild, f.Args()[1:])
}
