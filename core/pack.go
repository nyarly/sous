package core

type Pack struct {
	Name, Desc         string
	Detect             func() (packInfo interface{}, err error)
	MakeTargets        func(packInfo interface{}) []Target
	ProjectDesc        func(packinfo interface{}) string
	CheckCompatibility func(packInfo interface{}) []string

	packInfo interface{}
	targets  Targets
}

func (p *Pack) GetTarget(name string) (Target, bool) {
	if t, ok := p.Targets()[name]; ok {
		return t, true
	}
	return nil, false
}

// Targets is a lazily initialised map of targets.
func (p *Pack) Targets() map[string]Target {
	if p.targets == nil {
		ts := p.MakeTargets(p.packInfo)
		p.targets = Targets{}
		for _, t := range ts {
			p.targets.Add(t)
		}
	}
	return p.targets
}
