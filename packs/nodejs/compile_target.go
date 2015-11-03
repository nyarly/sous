package nodejs

import (
	"fmt"
	"strconv"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/docker"
)

type CompileTarget struct {
	*NodeJSTarget
	state map[string]string
}

func NewCompileTarget(pack *Pack) *CompileTarget {
	return &CompileTarget{NewNodeJSTarget("compile", pack), nil}
}

func (t *CompileTarget) DependsOn() []core.Target { return nil }

func (t *CompileTarget) RunAfter() []string { return nil }

func (t *CompileTarget) Desc() string {
	return "The NodeJS compile target invokes `npm install` inside the container, and zips up the resultant app dir"
}

func (t *CompileTarget) Check() error {
	return nil
}

func (t *CompileTarget) Dockerfile() *docker.Dockerfile {
	df := t.Pack.baseDockerfile(t.Name())
	df.AddRun("npm install -g npm@2")
	return df
}

// This image does not get stale because of any changes to the project itself.
// Everything is stale when Sous or its configuration is updated, or when the
// relevant Docker base image is updated.
func (t *CompileTarget) ImageIsStale(c *core.Context) bool {
	return false
}

// This container does not get stale unless the working directory is changed, but
// we currently don't have a way to check this.
// TODO: Record WD changes in context so we can invalidate the container when the
// WD changes.
func (t *CompileTarget) ContainerIsStale(c *core.Context) bool {
	return false
}

func (t *CompileTarget) ImageTag(c *core.Context) string {
	return strconv.Itoa(c.BuildNumber())
}

func (t *CompileTarget) ContainerName(c *core.Context) string {
	return fmt.Sprintf("%s_reusable_builder", c.CanonicalPackageName())
}

// Run first checks if a container with the right name has already been built. If so,
// it re-uses that container (note: this container is built exactly once per project,
// per configuration per change or upgrade to sous, not when source code generally,
// nor even dependencies change.
//
// It builds a stateful container with the NPM cache that implies, which is re-used
// for every build of this project. It's basically a caching layer. It is based on the
// exact same OS and Arch as the production containers, but with additional build tools
// which enable the building of complex dependencies.
func (t *CompileTarget) DockerRun(c *core.Context) *docker.Run {
	containerName := t.ContainerName(c)
	run := docker.NewRun(c.DockerTag())
	run.Name = containerName
	run.AddEnv("ARTIFACT_NAME", t.artifactName(c))
	artDir := t.artifactDir(c)
	dir.EnsureExists(artDir)
	run.AddVolume(artDir, "/artifacts")
	run.AddVolume(c.WorkDir, "/wd")
	run.Command = "npm install"
	return run
}

func (t *CompileTarget) State(c *core.Context) interface{} {
	return map[string]string{
		"artifactPath": t.artifactPath(c),
	}
}

func (t *CompileTarget) artifactPath(c *core.Context) string {
	return fmt.Sprintf("%s/%s.tar.gz", t.artifactDir(c), t.artifactName(c))
}

func (t *CompileTarget) artifactDir(c *core.Context) string {
	return c.FilePath("artifacts")
}

func (t *CompileTarget) artifactName(c *core.Context) string {
	return fmt.Sprintf("%s-%s-%s-%d", c.CanonicalPackageName(), c.AppVersion, c.Git.CommitSHA, c.BuildNumber())
}
