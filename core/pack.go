package core

type Pack struct {
	Name               string
	Detect             func() (packInfo interface{}, err error)
	Targets            map[string]*Target
	ProjectDesc        func(packinfo interface{}) string
	CheckCompatibility func(packInfo interface{}) []string
}
