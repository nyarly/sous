package golang

import (
	"fmt"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/docker"
)

type CompileTarget struct {
	*GoTarget
}

func NewCompileTarget(pack *Pack) *CompileTarget {
	return &CompileTarget{NewGoTarget("compile", pack)}
}

func (t *CompileTarget) DependsOn() []core.Target {
	return nil
}

func (t *CompileTarget) RunAfter() []string { return nil }

func (t *CompileTarget) Desc() string {
	return "The Go compile target generates a single binary file"
}

// Checking a Go project always passes.
func (t *CompileTarget) Check() error {
	return nil
}

func (t *CompileTarget) Dockerfile(c *core.Context) *docker.Dockerfile {
	df := &docker.Dockerfile{}
	df.From = t.pack.baseImageTag(t.Name())
	c.TemporaryLinkResource("build-prep.bash")
	buildPrepContainerPath := "/build-prep.bash"
	df.AddAdd("build-prep.bash", buildPrepContainerPath)
	df.AddRun(fmt.Sprintf("chmod +x %s", buildPrepContainerPath))
	// This is a non-portable image, since it includes the UID of the
	// logged-in user.
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

func (t *CompileTarget) ContainerName(c *core.Context) string {
	return fmt.Sprintf("%s_reusable-builder", c.CanonicalPackageName())
}

func (t *CompileTarget) ContainerIsStale(c *core.Context) (bool, string) {
	return true, "it is not reusable"
}

func (t *CompileTarget) DockerRun(c *core.Context) *docker.Run {
	containerName := t.ContainerName(c)
	run := docker.NewRun(c.DockerTag())
	run.Name = containerName
	run.AddEnv("ARTIFACT_NAME", t.artifactName(c))
	run.AddEnv("REPO_WORKDIR", "/"+c.Git.RepoWorkDirPathOffset)
	uid := cmd.Stdout("id", "-u")
	gid := cmd.Stdout("id", "-g")
	artifactOwner := fmt.Sprintf("%s:%s", uid, gid)
	run.AddEnv("ARTIFACT_OWNER", artifactOwner)
	artDir := t.artifactDir(c)
	dir.EnsureExists(artDir)
	run.AddVolume(artDir, "/artifacts")
	run.AddVolume(c.Git.Dir, "/repo")
	binName := fmt.Sprintf("%s-%s", c.CanonicalPackageName(), c.BuildVersion)
	run.Command = fmt.Sprintf("go build -o %s", binName)
	return run
}

// State returns any interesting state from this target to and dependent targets
// in the build chain.
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
	return fmt.Sprintf("%s-%s-%s-%d", c.CanonicalPackageName(), c.BuildVersion, c.Git.CommitSHA, c.BuildNumber())
}
