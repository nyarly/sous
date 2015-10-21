package nodejs

import (
	"fmt"
	"strings"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/version"
)

type Pack struct {
	PackageJSON *NodePackage
}

func (p *Pack) Name() string {
	return "NodeJS"
}

func (p *Pack) Desc() string {
	return "NodeJS Build Pack"
}

func (p *Pack) Detect() error {
	if !file.ReadJSON(&p.PackageJSON, "package.json") {
		return fmt.Errorf("no package.json file found")
	}
	return nil
}

func (p *Pack) Problems() []string {
	if p.PackageJSON == nil {
		panic("PackageJSON not set, detect must have failed.")
	}
	np := p.PackageJSON
	c := []string{}
	if np.Engines.Node == "" {
		c = append(c, "unable to determine NodeJS version, missing engines.node in package.json")
	} else {
		r := version.Range(np.Engines.Node)
		if v := r.BestMatchFrom(AvailableNodeVersions()); v == nil {
			f := "node version range (%s) not supported (pick from %s)"
			m := fmt.Sprintf(f, r.Original, strings.Join(AvailableNodeVersions().Strings(), ", "))
			c = append(c, m)
		}
	}
	if np.Version == "" {
		c = append(c, "no version specified in package.json:version")
	}
	return c
}

func (p *Pack) AppVersion() string {
	return p.PackageJSON.Version
}

func (p *Pack) AppDesc() string {
	np := p.PackageJSON
	return fmt.Sprintf("a NodeJS %s project named %s v%s",
		np.Engines.Node, np.Name, np.Version)
}

func (p *Pack) Targets() []core.Target {
	return []core.Target{
		NewAppTarget(p.PackageJSON),
		NewTestTarget(p.PackageJSON),
	}
}

func (p *Pack) String() string {
	return p.Name()
}
