package golang

import (
	"fmt"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/docker"
)

type TestTarget struct {
	*GoTarget
}

func NewTestTarget(pack *Pack) *TestTarget {
	return &TestTarget{NewGoTarget("test", pack)}
}

func (t *TestTarget) DependsOn() []core.Target { return nil }

func (t *TestTarget) RunAfter() []string { return nil }

func (t *TestTarget) Desc() string {
	return "The Go test target executes `go generate && go test ./...`"
}

func (t *TestTarget) Check() error {
	return nil
}

func (t *TestTarget) Dockerfile(c *core.Context) *docker.Dockerfile {
	df := &docker.Dockerfile{}
	df.From = t.pack.baseImageTag("test")
	c.TemporaryLinkResource("build-prep.bash")
	buildPrepContainerPath := "/build-prep.bash"
	df.AddAdd("build-prep.bash", buildPrepContainerPath)
	df.AddRun(fmt.Sprintf("chmod +x %s", buildPrepContainerPath))
	uid := cmd.Stdout("id", "-u")
	gid := cmd.Stdout("id", "-g")
	username := cmd.Stdout("whoami")
	// Just use the username for group name, it doesn't matter as long as
	// the IDs are right.
	groupAdd := fmt.Sprintf("groupadd -g %s %s", gid, username)
	// Explanation of some of the below useradd flags:
	//   -M means do not create home directory, which we do not need
	//   --no-log-init means do not create a 32G sparse file (which Docker commit
	//       cannot handle properly, and tries to create a non-sparse 32G file.)
	userAdd := fmt.Sprintf("useradd --no-log-init -M --uid %s --gid %s %s", uid, gid, username)
	df.AddRun(fmt.Sprintf("%s && %s", groupAdd, userAdd))
	df.Entrypoint = []string{"/build-prep.bash"}
	return df
}

func (t *TestTarget) ContainerName(c *core.Context) string {
	return c.CanonicalPackageName() + "_test"
}

func (t *TestTarget) ContainerIsStale(c *core.Context) (bool, string) {
	return true, "it is not reusable"
}

func (t *TestTarget) DockerRun(c *core.Context) *docker.Run {
	containerName := t.ContainerName(c)
	run := docker.NewRun(c.DockerTag())
	run.Name = containerName
	//run.AddEnv("ARTIFACT_NAME", t.artifactName(c))
	uid := cmd.Stdout("id", "-u")
	gid := cmd.Stdout("id", "-g")
	artifactOwner := fmt.Sprintf("%s:%s", uid, gid)
	run.AddEnv("ARTIFACT_OWNER", artifactOwner)
	artDir := t.artifactDir(c)
	dir.EnsureExists(artDir)
	run.AddVolume(artDir, "/artifacts")
	run.AddVolume(c.WorkDir, "/wd")
	run.Command = fmt.Sprintf("go generate && { [ -d Godeps ] && godep go test ./... || go test ./...; }")
	return run
}

func (t *TestTarget) artifactPath(c *core.Context) string {
	return fmt.Sprintf("%s/%s.tar.gz", t.artifactDir(c), t.artifactName(c))
}

func (t *TestTarget) artifactDir(c *core.Context) string {
	return c.FilePath("artifacts")
}

func (t *TestTarget) artifactName(c *core.Context) string {
	return fmt.Sprintf("%s-%s-%s-%d", c.CanonicalPackageName(), c.AppVersion, c.Git.CommitSHA, c.BuildNumber())
}
