package core

import (
	"github.com/opentable/sous/deploy"
	"github.com/opentable/sous/tools/cli"
)

// DetectProjectType invokes Detect() for each registered pack.
//
// If a single pack is found to match, it returns that pack along with
// the object returned from its detect func. This object is subsequently
// passed into the detect step for each target supported by the pack.
func (c *Context) DetectProjectType(packs deploy.Buildpacks) *deploy.Buildpack {
	var err error
	var pack *deploy.Buildpack
	for _, p := range packs {
		if err = p.Detect(c.WorkDir); err != nil {
			continue
		}
		if pack != nil {
			cli.Fatalf("multiple project types detected")
		}
		pack = &p
	}
	if pack == nil {
		cli.Fatalf("no buildable project detected")
	}
	return pack
}
