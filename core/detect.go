package core

import (
	"github.com/opentable/sous/tools/cli"
)

// DetectProjectType invokes Detect() for each registered pack.
//
// If a single pack is found to match, it returns that pack along with
// the object returned from its detect func. This object is subsequently
// passed into the detect step for each target supported by the pack.
func (c *Context) DetectProjectType(packs Buildpacks) *RunnableBuildpack {
	var pack *RunnableBuildpack
	for _, p := range packs {
		rbp, err := p.Detect(c.WorkDir)
		if err != nil {
			if packErr, ok := err.(BuildpackError); ok {
				cli.Fatal(packErr)
			}
			continue
		}
		if pack != nil {
			cli.Fatalf("multiple project types detected")
		}
		pack = rbp
	}
	return pack
}
