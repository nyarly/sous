package nodejs

import (
	"fmt"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/docker"
)

type TestTarget struct {
	*NodeJSTarget
}

func NewTestTarget(pack *Pack) *TestTarget {
	return &TestTarget{NewNodeJSTarget("test", pack)}
}

func (t *TestTarget) DependsOn() []core.Target { return nil }

func (t *TestTarget) RunAfter() []string { return []string{"compile"} }

func (t *TestTarget) Desc() string {
	return "The NodeJS test target builds your Docker image using `npm install`. When you invoke the container, it simply runs `npm test` to execute your test script defined in `package.json:scripts.test`"
}

func (t *TestTarget) Check() error {
	if len(t.NodeJSPack.PackageJSON.Scripts.Test) == 0 {
		return fmt.Errorf("package.json does not specify a test script")
	}
	return nil
}

func (t *TestTarget) Dockerfile(c *core.Context) *docker.Dockerfile {
	df := t.NodeJSPack.baseDockerfile(t.Name())
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
	df.AddRun("npm install -g npm@2")
	df.Entrypoint = []string{"/build-prep.bash"}
	return df
}

func (t *TestTarget) ContainerName(c *core.Context) string {
	return fmt.Sprintf("%s_test-container", c.CanonicalPackageName())
}

func (t *TestTarget) DockerRun(c *core.Context) *docker.Run {
	containerName := t.ContainerName(c)
	run := docker.NewRun(c.DockerTag())
	run.Name = containerName
	run.AddEnv("ARTIFACT_NAME", t.artifactName(c))
	uid := cmd.Stdout("id", "-u")
	gid := cmd.Stdout("id", "-g")
	artifactOwner := fmt.Sprintf("%s:%s", uid, gid)
	run.AddEnv("ARTIFACT_OWNER", artifactOwner)
	artDir := t.artifactDir(c)
	dir.EnsureExists(artDir)
	run.AddVolume(artDir, "/artifacts")
	run.AddVolume(c.WorkDir, "/wd")
	run.Command = "npm install && npm test"
	return run
}

// This container does not get stale unless the working directory, or the
// artifact path is changed.
func (t *TestTarget) ContainerIsStale(c *core.Context) (bool, string) {
	// TODO: Detect if the wd or artifact paths have changed, return true
	// if either of those are true.
	return false, ""
}

func (t *TestTarget) artifactPath(c *core.Context) string {
	return fmt.Sprintf("%s/%s.tar.gz", t.artifactDir(c), t.artifactName(c))
}

func (t *TestTarget) artifactDir(c *core.Context) string {
	return c.FilePath("artifacts")
}

func (t *TestTarget) artifactName(c *core.Context) string {
	return fmt.Sprintf("%s-%s-%s-%d", c.CanonicalPackageName(), c.BuildVersion, c.Git.CommitSHA, c.BuildNumber())
}
