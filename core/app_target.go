package core

import (
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
)

type AppTarget struct {
	Context      *Context
	Buildpack    *RunnableBuildpack
	Command      string
	ArtifactPath string
}

func NewAppTarget(bp *RunnableBuildpack, c *Context) *AppTarget {
	return &AppTarget{c, bp, "-command set by compile target-",
		"-artifact set by compile target-"}
}

func (t *AppTarget) Name() string { return "compile" }

func (t *AppTarget) DependsOn() []Target { return nil }

func (t *AppTarget) SetState(name, value string) {
	if name != "artifact" {
		return
	}
	t.ArtifactPath = value
}

func (t *AppTarget) String() string { return t.Name() }

func (t *AppTarget) Desc() string {
	return "generates artifacts for injection into a production container"
}

func (t *AppTarget) Check() error { return nil }

func (t *AppTarget) Dockerfile(c *TargetContext) *docker.File {
	image, err := c.BaseImage(c.WorkDir, "app")
	if err != nil {
		cli.Fatal(err)
	}
	df := &docker.File{From: image}
	df.Maintainer = c.User
	df.ADD(t.ArtifactPath)
	df.WORKDIR("/srv/app")
	df.CMD(t.Command)
	return df
}

// DockerRun returns a configured *docker.Run, which is used to create a new
// container when the old one is stale or does not exist.
func (t *AppTarget) DockerRun(tc *TargetContext) *docker.Run {
	r := docker.NewRun(tc.DockerTag())
	return r
}
