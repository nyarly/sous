package golang

import "github.com/opentable/sous/core"

// GoTarget is the base for all Go targets
type GoTarget struct {
	*core.TargetBase
	pack *Pack
}

// NewGoTarget creates a new NodeJSTarget based on a known target name
// from core.
func NewGoTarget(name string, pack *Pack) *GoTarget {
	return &GoTarget{core.MustGetTargetBase(name, pack), pack}
}
