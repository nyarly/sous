package nodejs

import (
	"fmt"
	"strings"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/deploy"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/version"
)

type Pack struct {
	Config      *deploy.NodeJSConfig
	PackageJSON *NodePackage
}

func New(c *deploy.NodeJSConfig) *Pack {
	return &Pack{Config: c}
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
	// This is the place to set defaults
	if p.PackageJSON.Engines.Node == "" {
		p.PackageJSON.Engines.Node = p.Config.DefaultNodeVersion
	}
	return nil
}

func (p *Pack) Problems() core.ErrorCollection {
	if p.PackageJSON == nil {
		panic("PackageJSON not set, detect must have failed.")
	}
	np := p.PackageJSON
	errs := core.ErrorCollection{}
	c := deploy.Load()
	if np.Engines.Node == "" {
		errs.AddWarningf("missing node engine version in package.json, defaulting to node %s; see https://docs.npmjs.com/files/package.json#engines",
			c.Packs.NodeJS.DefaultNodeVersion)
	} else {
		r := version.Range(np.Engines.Node)
		if v := r.BestMatchFrom(p.AvailableNodeVersions()); v == nil {
			f := "node version range (%s) not supported (pick from %s)"
			errs.AddErrorf(f, r.Original, strings.Join((p.AvailableNodeVersions().Strings()), ", "))
		}
	}
	if np.Version == "" {
		errs.AddWarningf("no app version specified in package.json:version")
	}
	return errs
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
		NewAppTarget(p),
		NewTestTarget(p),
		NewCompileTarget(p),
	}
}

func (p *Pack) String() string {
	return p.Name()
}
