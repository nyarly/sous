package core

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/opentable/sous/config"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/git"
	"github.com/opentable/sous/tools/path"
)

type Context struct {
	Git                  *git.Info
	WorkDir              string
	Action               string
	DockerRegistry       string
	Host, FullHost, User string
	BuildState           *BuildState
	AppVersion           string
	PackInfo             interface{}
	changes              *Changes
}

func (bc *Context) IsCI() bool {
	return bc.User == "ci"
}

func GetContext(action string) *Context {
	var c = config.Load()
	registry := c.DockerRegistry
	gitInfo := git.GetInfo()
	bs := GetBuildState(action, gitInfo)
	wd, err := os.Getwd()
	if err != nil {
		cli.Fatalf("Unable to get current working directory: %s", err)
	}
	return &Context{
		Git:            gitInfo,
		WorkDir:        wd,
		Action:         action,
		DockerRegistry: registry,
		Host:           cmd.Stdout("hostname"),
		FullHost:       cmd.Stdout("hostname", "-f"),
		User:           getUser(),
		BuildState:     bs,
	}
}

func (c *Context) DockerTag() string {
	return c.DockerTagForBuildNumber(c.BuildNumber())
}

func (c *Context) BuildNumber() int {
	return c.BuildState.CurrentCommit().BuildNumber
}

func (c *Context) PrevDockerTag() string {
	return c.DockerTagForBuildNumber(c.BuildNumber() - 1)
}

func (c *Context) DockerTagForBuildNumber(n int) string {
	if c.AppVersion == "" {
		cli.Fatalf("AppVersion not set")
	}
	name := c.CanonicalPackageName()
	if c.Action != "build" {
		name += "_" + c.Action
	}
	repo := fmt.Sprintf("%s/%s", c.User, name)
	buildNumber := strconv.Itoa(n)
	if c.User != "teamcity" {
		buildNumber = c.Host + "-" + buildNumber
	}
	tag := fmt.Sprintf("v%s-%s-%s",
		c.AppVersion, c.Git.CommitSHA[0:8], buildNumber)
	// e.g. on local dev machine:
	//   some.registry.com/username/widget-factory:v0.12.1-912eeeab-host-1
	return fmt.Sprintf("%s/%s:%s", c.DockerRegistry, repo, tag)
}

func (c *Context) ChangesSinceLastBuild() *Changes {
	cc := c.BuildState.CurrentCommit()
	if c.changes == nil {
		c.changes = &Changes{
			NoBuiltImage:       !c.LastBuildImageExists(),
			NewCommit:          c.BuildState.CommitSHA != c.BuildState.LastCommitSHA,
			WorkingTreeChanged: cc.TreeHash != cc.OldTreeHash,
			SousUpdated:        cc.SousHash != cc.OldSousHash,
		}
	}
	return c.changes
}

type Changes struct {
	NoBuiltImage, NewCommit, WorkingTreeChanged, SousUpdated bool
}

func (c *Changes) Any() bool {
	return c.NoBuiltImage || c.NewCommit || c.WorkingTreeChanged || c.SousUpdated
}

func (c *Context) LastBuildImageExists() bool {
	return docker.ImageExists(c.PrevDockerTag())
}

func (s *BuildState) CurrentCommit() *Commit {
	return s.Commits[s.CommitSHA]
}

func (bc *Context) Commit() {
	if bc.IsCI() {
		return
	}
	bc.BuildState.Commit()
}

func (bc *Context) CanonicalPackageName() string {
	c := bc.Git.CanonicalName()
	p := strings.Split(c, "/")
	return p[len(p)-1]
}

func buildingInCI() bool {
	return os.Getenv("TEAMCITY_VERSION") != ""
}

func getUser() string {
	if buildingInCI() {
		return "ci"
	}
	return cmd.Stdout("id", "-un")
}

func (c *Context) IncrementBuildNumber() {
	if !buildingInCI() {
		c.BuildState.CurrentCommit().BuildNumber++
	}
}

func (s *BuildState) Commit() {
	file.WriteJSON(s, s.path)
}

func (c *Context) SaveFile(content, name string) {
	file.WriteString(content, c.FilePath(name))
}

func (c *Context) FilePath(name string) string {
	return path.Resolve(c.BaseDir() + "/" + name)
}

func (c *Context) BaseDir() string {
	return path.BaseDir(c.BuildState.path)
}

func tryGetBuildNumberFromEnv() (int, bool) {
	envBN := os.Getenv("BUILD_NUMBER")
	if envBN != "" {
		n, err := strconv.Atoi(envBN)
		if err != nil {
			cli.Fatalf("Unable to parse $BUILD_NUMBER (%s) to int: %s", envBN, err)
		}
		return n, true
	}
	return 0, false
}
