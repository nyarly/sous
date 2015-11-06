package nodejs

import (
	"fmt"
	"strconv"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cmd"
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

// DependsOn returns a slice of core.Targets which must be run before this target
// is able to.
func (t *CompileTarget) DependsOn() []core.Target { return nil }

// RunAfter returns a list of target names which, if supported by this project,
// must be run before this target is run. You can think of this as "optional"
// dependencies.
// NOTE: This will probably be deprecated very soon.
func (t *CompileTarget) RunAfter() []string { return nil }

// Desc returns a human-friendly description of what this target does.
func (t *CompileTarget) Desc() string {
	return "generates artifacts for injection into a production container"
}

// Check returns an error if there are potential problems running this target
// in the current context.
func (t *CompileTarget) Check() error {
	return nil
}

// Dockerfile returns a configured *docker.Dockerfile which is used by Sous
// to build new Docker images when needed.
func (t *CompileTarget) Dockerfile() *docker.Dockerfile {
	df := t.NodeJSPack.baseDockerfile(t.Name())
	// This is a non-portable container, since it includes the UID of the
	// logged-in user.
	uid := cmd.Stdout("id", "-u")
	gid := cmd.Stdout("id", "-g")
	username := cmd.Stdout("whoami")
	// Just use the username for group name, it doesn't matter as long as
	// the IDs are right.
	df.AddRun(fmt.Sprintf("groupadd -g %s %s", gid, username))
	// Explanation of some of the below useradd flags:
	//   -M means do not create home directory, which we do not need
	//   --no-log-init means do not create a 32G sparse file (which Docker commit
	//       cannot handle properly, and tries to create a non-sparse 32G file.
	df.AddRun(fmt.Sprintf("useradd --no-log-init -M --uid %s --gid %s %s", uid, gid, username))
	df.AddRun("npm install -g npm@2")
	return df
}

// This image does not get stale because of any changes to the project itself,
// only when the user ID of the currently logged-in account differs from that
// of the account which created the image, see the Dockerfile above.
func (t *CompileTarget) ImageIsStale(c *core.Context) (bool, string) {
	// TODO: Make the container stale if the user ID has changed.
	return false, ""
}

// This container does not get stale unless the working directory, or the
// artifact path is changed.
func (t *CompileTarget) ContainerIsStale(c *core.Context) (bool, string) {
	// TODO: Detect if the wd or artifact paths have changed, return true
	// if either of those are true.
	return false, ""
}

// TODO: Flesh out the concept of stale artifacts and implement this throughout
// the build chain.
func (t *CompileTarget) ArtifactIsStale(c *core.Context) (bool, string) {
	return true, "we don't currently have a way to check whether it's stale or not"
}

// ImageTag return the tag that should be applied to the next image we build.
func (t *CompileTarget) ImageTag(c *core.Context) string {
	return strconv.Itoa(c.BuildNumber())
}

// ContainerName returns the name that will be given to the next container we
// build. This does not have to change for each build, Sous will automatically
// deleted any pre-existing containers with this name before creating a new one.
func (t *CompileTarget) ContainerName(c *core.Context) string {
	return fmt.Sprintf("%s_reusable_builder", c.CanonicalPackageName())
}

// DockerRun returns a configured *docker.Run, which is used to create a new
// container when the old one is stale or does not exist.
func (t *CompileTarget) DockerRun(c *core.Context) *docker.Run {
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
	run.Command = "npm install"
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
	return fmt.Sprintf("%s-%s-%s-%d", c.CanonicalPackageName(), c.AppVersion, c.Git.CommitSHA, c.BuildNumber())
}
