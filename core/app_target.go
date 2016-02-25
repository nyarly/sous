package core

import (
	"fmt"
	"os/user"

	"github.com/opentable/sous/deploy"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
)

type AppTarget struct {
	Context   *Context
	Buildpack *deploy.Buildpack
}

func NewAppTarget(bp *deploy.Buildpack, c *Context) *AppTarget {
	return &AppTarget{c, bp}
}

func (t *AppTarget) Name() string { return "compile" }

func (t *AppTarget) DependsOn() []Target { return nil }

func (t *AppTarget) String() string { return t.Name() }

func (t *AppTarget) Desc() string {
	return "generates artifacts for injection into a production container"
}

func (t *AppTarget) Check() error { return nil }

func (t *AppTarget) Dockerfile(c *TargetContext) *docker.File {
	image, err := t.Buildpack.BaseImage(c.WorkDir, t.Name())
	if err != nil {
		cli.Fatal(err)
	}
	df := &docker.File{From: image}
	// This is a non-portable container, since it includes the UID of the
	// logged-in user. This is necessary to ensure the user in the container
	// can write files accessible to the user invoking the container on the
	// host.
	u, err := user.Current()
	if err != nil {
		cli.Fatalf("unable to get current user: %s", err)
	}
	// Just use the username for group name, it doesn't matter as long as
	// the IDs are right.
	df.RUN("groupadd", "-g", u.Gid, u.Username)
	// Explanation of some of the below useradd flags:
	//   -M means do not create home directory, which we do not need
	//   --no-log-init means do not create a 32G sparse file (which Docker commit
	//       cannot handle properly, and tries to create a non-sparse 32G file.)
	df.RUN("useradd", "--no-log-init", "-M", "--uid", u.Uid, "--gid", u.Gid, u.Username)

	df.USER(u.Username)
	return df
}

// DockerRun returns a configured *docker.Run, which is used to create a new
// container when the old one is stale or does not exist.
func (t *AppTarget) DockerRun(tc *TargetContext) *docker.Run {
	r := docker.NewRun(tc.DockerTag())
	r.AddEnv("PROJ_NAME", tc.CanonicalPackageName())
	r.AddEnv("PROJ_VERSION", "0.0.0") // TODO: Get project version from TargetContext
	r.AddEnv("PROJ_REVISION", tc.Git.CommitSHA)
	r.AddEnv("PROJ_DIRTY", YESorNO(tc.Git.Dirty))
	r.AddEnv("BASE_DIR", fmt.Sprintf("/source"))
	r.AddEnv("REPO_DIR", tc.CanonicalPackageName())
	r.AddEnv("REPO_WORKDIR", tc.Git.RepoWorkDirPathOffset)

	artifactDir := GetEmptyArtifactDir(tc)
	r.AddEnv("ARTIFACT_DIR", artifactDir)

	//uid := cmd.Stdout("id", "-u")
	//gid := cmd.Stdout("id", "-g")
	//artifactOwner := fmt.Sprintf("%s:%s", uid, gid)
	//run.AddEnv("ARTIFACT_OWNER", artifactOwner)
	r.AddVolume(artifactDir, "/mnt/artifacts")
	r.AddVolume(tc.Git.Dir, "/mnt/repo")
	return r
}
