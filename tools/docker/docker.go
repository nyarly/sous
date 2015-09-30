package docker

import (
	"strings"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/path"
	"github.com/opentable/sous/tools/version"
)

func RequireVersion(r *version.R) {
	vs := cmd.Table("docker", "--version")[0][2]
	vs = strings.Trim(vs, ",")
	v := version.Version(vs)
	if !r.IsSatisfiedBy(v) {
		cli.Fatalf("got docker version %s; want %s", v, r)
	}
}

func RequireDaemon() {
	if c := cmd.ExitCode("docker", "ps"); c != 0 {
		cli.Fatalf("`docker ps` exited with code %d", c)
	}
}

func dockerCmd(args ...string) *cmd.CMD {
	c := cmd.New("docker", args...)
	c.EchoStdout = true
	c.EchoStderr = true
	return c
}

// Build builds the dockerfile in the specified directory and returns the image ID
func Build(dir, tag string) string {
	dir = path.Resolve(dir)
	return dockerCmd("build", "-t", tag, dir).Out()
}

func Run(tag string) int {
	return dockerCmd("run", tag).ExitCode()
}

func Push(tag string) {
	cmd.EchoAll("docker", "push", tag)
}

func ImageExists(tag string) bool {
	rows := cmd.Table("docker", "images")
	for _, r := range rows {
		comp := r[0] + ":" + r[1]
		if comp == tag {
			return true
		}
	}
	return false
}
