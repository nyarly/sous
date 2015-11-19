package golang

import (
	"fmt"

	"github.com/opentable/sous/config"
	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/file"
)

type Pack struct {
	Config *config.GoConfig
}

func New(c *config.GoConfig) *Pack {
	return &Pack{Config: c}
}

func (p *Pack) Name() string {
	return "Go"
}

func (p *Pack) Desc() string {
	return "Go Build Pack"
}

func (p *Pack) Detect() error {
	if len(file.Find("*.go")) == 0 {
		return fmt.Errorf("No file matching *.go found")
	}
	return nil
}

func (p *Pack) Problems() core.ErrorCollection {
	return core.ErrorCollection{}
}

func (p *Pack) AppVersion() string {
	return "0.0.0"
}

func (p *Pack) AppDesc() string {
	return "a Go project"
}

func (p *Pack) Targets() []core.Target {
	return []core.Target{
		NewCompileTarget(p),
		NewAppTarget(p),
		NewTestTarget(p),
	}
}

func (p *Pack) String() string {
	return p.Name()
}

func (p *Pack) baseImageTag(target string) string {
	baseImageTag, ok := p.Config.AvailableVersions.GetBaseImageTag(
		p.Config.DefaultGoVersion, target)
	if !ok {
		cli.Fatalf("Go build pack misconfigured, default version %s not available for target %s",
			p.Config.DefaultGoVersion, target)
	}
	return baseImageTag
}
