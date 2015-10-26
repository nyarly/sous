package nodejs

import "github.com/opentable/sous/core"

// NodeJSTarget is the base for all NodeJS targets
type NodeJSTarget struct {
	*core.TargetBase
	Pack *Pack
}

// NewNodeJSTarget creates a new NodeJSTarget based on a known target name
// from core.
func NewNodeJSTarget(name string, pack *Pack) *NodeJSTarget {
	return &NodeJSTarget{core.MustGetTargetBase(name), pack}
}
