package core

type Pack struct {
	Name                  string
	Detect                func() error
	Targets               map[string]*Target
	CompatibleProjectDesc func() string
	CheckCompatibility    func() []string
}
