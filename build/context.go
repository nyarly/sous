package build

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	. "github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/git"
)

type Context struct {
	Git                  *git.Info
	BuildNumber          int
	DockerRegistry       string
	Host, FullHost, User string
}

func (bc *Context) IsCI() bool {
	return bc.User == "ci"
}

func GetContext() *Context {
	gitInfo := git.GetInfo()
	return &Context{
		Git:            gitInfo,
		BuildNumber:    getBuildNumber(gitInfo),
		DockerRegistry: "docker.otenv.com",
		Host:           Cmd("hostname"),
		FullHost:       Cmd("hostname", "-f"),
		User:           getUser(),
	}
}

func (bc *Context) CanonicalPackageName() string {
	c := bc.Git.CanonicalName()
	p := strings.Split(c, "/")
	return p[len(p)-1]
}

func getBuildNumber(git *git.Info) int {
	if n, ok := tryGetBuildNumberFromEnv(); ok {
		Logf("got build number %d from $BUILD_NUMBER")
		return n
	}
	return getBuildNumberFromHomeDirectory(git)
}

func buildingInCI() bool {
	return os.Getenv("TEAMCITY_VERSION") != ""
}

func getUser() string {
	if buildingInCI() {
		return "ci"
	}
	return Cmd("whoami")
}

func getBuildNumberFromHomeDirectory(git *git.Info) int {
	buildNumDir := fmt.Sprintf("~/.ot/build_numbers/%s", git.CanonicalName())
	dir.EnsureExists(buildNumDir)
	filePath := fmt.Sprintf("%s/%s", buildNumDir, git.CommitSHA)
	bns, ok := file.ReadString(filePath)
	if !ok {
		file.Write(1, filePath)
		return 1
	}
	bn, err := strconv.Atoi(bns)
	if err != nil {
		Dief("unable to parse build number %s (from %s) as int: %s",
			bns, filePath, err)
	}
	bn++
	file.Write(bn, filePath)
	return bn
}

func tryGetBuildNumberFromEnv() (int, bool) {
	envBN := os.Getenv("BUILD_NUMBER")
	if envBN != "" {
		n, err := strconv.Atoi(envBN)
		if err != nil {
			Dief("Unable to parse $BUILD_NUMBER (%s) to int: %s", envBN, err)
		}
		return n, true
	}
	return 0, false
}
