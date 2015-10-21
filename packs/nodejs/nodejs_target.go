package nodejs

import "github.com/opentable/sous/core"

// NodeJSTarget is the base for all NodeJS targets
type NodeJSTarget struct {
	*core.TargetBase
	PackageJSON *NodePackage
}

// NewNodeJSTarget creates a new NodeJSTarget based on a known target name
// from core.
func NewNodeJSTarget(name string, np *NodePackage) *NodeJSTarget {
	return &NodeJSTarget{core.MustGetTargetBase(name), np}
}
