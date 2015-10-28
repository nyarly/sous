package nodejs

import (
	"fmt"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/docker"
)

type CompileTarget struct {
	*NodeJSTarget
}

func NewCompileTarget(pack *Pack) *CompileTarget {
	return &CompileTarget{NewNodeJSTarget("compile", pack)}
}

func (t *CompileTarget) DependsOn() []string { return nil }

func (t *CompileTarget) RunAfter() []string { return nil }

func (t *CompileTarget) Desc() string {
	return "The NodeJS compile target invokes `npm install` inside the container, and zips up the resultant app dir"
}

func (t *CompileTarget) Check() error {
	return nil
}

func (t *CompileTarget) Dockerfile() *docker.Dockerfile {
	df := t.Pack.baseDockerfile(t.Name())
	df.CMD = []string{"npm install -g npm@2 && npm install"}
	return df
}

// Stale for this target only rebuilds when Sous itself is updated. This is
// because we want to preserve the same container as long as possible, as it
// builds up a cache, speeding up builds. When Sous itself is updated (either a
// new version of the binary, or the config is changed) we must always re-build
// everything, as base images and policies may have been updated.
func (t *CompileTarget) Stale(c *core.Context) bool {
	return c.ChangesSinceLastBuild().SousUpdated
}

// Run first checks if a container with the right name has already been bcd /wd && ls -lah / &&  ./build.bashuilt. If so,
// it re-uses that container (note: this container is built exactly once per project,
// per configuration par change or upgrade to sous, not when source code generally,
// nor even dependencies change.
//
// It builds a stateful container with the NPM cache that implies, which is re-used
// for every build of this project. It's basically a caching layer. It is based on the
// exact same OS and Arch as the production containers, but with additional build tools
// which enable the building of complex dependencies.
func (t *CompileTarget) DockerRun(c *core.Context) *docker.Run {
	np := t.Pack.PackageJSON
	containerName := fmt.Sprintf("sous-builder_%s", np.Name)
	container := docker.ContainerWithName(containerName)
	if container.Exists() {
		if container.Running() {
			container.Kill()
		}
		container.Remove()
	}
	//if container.Exists() {
	//	if err := container.Start(); err != nil {
	//		cli.Fatalf("ERROR: Failed to start build container: %s", err)
	//	}
	//} else {
	//	cli.Logf("=====> Preparing build container for first run")
	//}
	// docker run --name sous-builder_%s  -v "$PWD:/wd" -v "$HOME/.sous:/artifacts" start-page-builder
	run := docker.NewRun(c.DockerTag())
	run.Name = containerName
	run.AddEnv("ARTIFACT_NAME", c.CanonicalPackageName())
	artDir := c.FilePath("artifacts")
	dir.EnsureExists(artDir)
	run.AddVolume(artDir, "/artifacts")
	run.AddVolume(c.WorkDir, "/wd")
	run.Command = "npm install -g npm@2 && npm install"
	return run
}
