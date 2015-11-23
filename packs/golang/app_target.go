package golang

import (
	"fmt"
	"path"
	"strconv"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/ports"
)

type AppTarget struct {
	*GoTarget
	artifactPath string
}

func NewAppTarget(pack *Pack) *AppTarget {
	return &AppTarget{NewGoTarget("app", pack), ""}
}

func (t *AppTarget) DependsOn() []core.Target {
	return []core.Target{
		NewCompileTarget(t.pack),
	}
}

func (t *AppTarget) RunAfter() []string { return []string{"compile"} }

func (t *AppTarget) Desc() string {
	return "The Go app target simply places a compiled Go binary inside a plain ubuntu base image"
}

// Checking a Go project always passes.
func (t *AppTarget) Check() error {
	return nil
}

func (t *AppTarget) PreDockerBuild(c *core.Context) {
	if t.artifactPath == "" {
		cli.Fatalf("Artifact path not set by compile target.")
	}
	if !file.Exists(t.artifactPath) {
		cli.Fatalf("Artifact not at %s", t.artifactPath)
	}
	filename := path.Base(t.artifactPath)
	localArtifact := filename
	file.TemporaryLink(t.artifactPath, "./"+localArtifact)
	t.artifactPath = localArtifact
}

func (t *AppTarget) Dockerfile(c *core.Context) *docker.Dockerfile {
	if t.artifactPath == "" {
		// Actually, it is first set by compile target, then the PreDockerBuild
		// step links it into the WD and resets artifactPath to a local, relative
		// path.
		t.artifactPath = "<¡ artifact path set by compile target !>"
	}
	df := &docker.Dockerfile{}
	df.From = t.pack.baseImageTag("app")

	// Since the artifact is tar.gz, and the dest is a directory, docker automatically unpacks it.
	df.AddAdd(t.artifactPath, "/srv/app/")
	// Pick out the contents of NPM start to invoke directly (using npm start in
	// production shields the app from signals, which are required to be handled by
	// the app itself to do graceful shutdown.
	df.Entrypoint = []string{
		fmt.Sprintf("/srv/app/%s-%s", c.CanonicalPackageName(), c.BuildVersion),
	}
	return df
}

func (t *AppTarget) SetState(fromTarget string, state interface{}) {
	if fromTarget != "compile" {
		return
	}
	m, ok := state.(map[string]string)
	if !ok {
		cli.Fatalf("app target got a %T from compile target, expected map[string]string", state)
	}
	artifactPath, ok := m["artifactPath"]
	if !ok {
		cli.Fatalf("app target got %+v from compile target; expected key 'artifactPath'", m)
	}
	t.artifactPath = artifactPath
}

func (t *AppTarget) DockerRun(c *core.Context) *docker.Run {
	dr := docker.NewRun(c.DockerTag())
	port0, err := ports.GetFreePort()
	if err != nil {
		cli.Fatalf("Unable to get free port: %s", err)
	}
	dr.AddEnv("PORT0", strconv.Itoa(port0))
	dr.AddEnv("TASK_HOST", core.DivineTaskHost())
	return dr
}

func (t *AppTarget) ContainerName(c *core.Context) string {
	return c.CanonicalPackageName()
}

func (t *AppTarget) ContainerIsStale(c *core.Context) (bool, string) {
	return true, "it is not reusable"
}
