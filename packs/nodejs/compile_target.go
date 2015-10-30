package nodejs

import (
	"fmt"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/docker"
)

type CompileTarget struct {
	*NodeJSTarget
}

func NewCompileTarget(pack *Pack) *CompileTarget {
	return &CompileTarget{NewNodeJSTarget("compile", pack)}
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
func (t *CompileTarget) Stale(c *core.Context) bool {
	return false
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
	// TODO: This logic probably should be higher up, and be default behaviour
	// for any target that is .(DockerContainer)
	containerName := t.DockerContainerName(c)
	container := docker.ContainerWithName(containerName)
	if container.Exists() {
		// TODO: Get container image ID, and target dockerfile FROM;
		// Then check if the base image has changed, and destroy/recreate
		// the container if so...
		image := container.Image()
		baseImage := t.Dockerfile().From
		if !docker.BaseImageUpdated(baseImage, image) {
			cli.Logf("Re-using build container %s", container)
			return docker.NewReRun(container)
		}
		cli.Logf("INFO: Base image %s updated; re-creating build container, the first build may take some time.", baseImage)
		if err := container.Remove(); err != nil {
			cli.Fatalf("Unable to remove outdated container %s", container)
		}
	}
	cli.Logf("====> Preparing build container for first use")
	run := docker.NewRun(c.DockerTag())
	run.Name = containerName
	artifactName := fmt.Sprintf("%s-%s-%s", c.CanonicalPackageName(), c.AppVersion, c.Git.CommitSHA)
	run.AddEnv("ARTIFACT_NAME", artifactName)
	artDir := c.FilePath("artifacts")
	dir.EnsureExists(artDir)
	run.AddVolume(artDir, "/artifacts")
	run.AddVolume(c.WorkDir, "/wd")
	run.Command = "npm install"
	return run
}

func (t *CompileTarget) DockerContainerName(c *core.Context) string {
	return fmt.Sprintf("%s_reusable_builder", c.CanonicalPackageName())
}
