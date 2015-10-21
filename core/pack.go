package core

// Pack describes a project type based on a particular dev stack.
// It is guaranteed that Detect() will be called before any of
// Problems(), ProjectDesc(), and Targets(). Therefore you can
// use the Detect() step to store internal state inside the pack
// if that is useful.
type Pack interface {
	// Name returns a short constant string naming the pack
	Name() string
	// Desc returns a longer constant string describing the pack
	Desc() string
	// Detect is called to check if the current project is of
	// the type this pack knows how to build. It should return
	// a descriptive error if this pack does not think it can
	// work with the current project.
	Detect() error
	// Problems is called to do a more thorough check on the
	// current project to highlight any potential problems
	// with running the various targets against it.
	Problems() []string
	// ProjectDesc returns a description of the current project.
	// It should include important information such as stack
	// name, runtime version, application version, etc.
	ProjectDesc() string
	// Targets returns a slice of all targets this pack is able
	// to build.
	Targets() []Target
}

// CompiledPack wraps a Pack and adds some validation and
// helper methods.
type CompiledPack struct {
	Pack
	targets Targets
}

// GetTarget allows getting a target by name, returns false as
// the second return value if that target is not defined.
func (p *CompiledPack) GetTarget(name string) (Target, bool) {
	if t, ok := p.Targets()[name]; ok {
		return t, true
	}
	return nil, false
}

// Targets is a lazily initialised map of validated targets.
func (p *CompiledPack) Targets() map[string]Target {
	if p.targets == nil {
		ts := p.Pack.Targets()
		p.targets = Targets{}
		for _, t := range ts {
			p.targets.Add(t)
		}
	}
	return p.targets
}
