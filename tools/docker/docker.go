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

// Build builds the dockerfile in the specified directory and returns the image ID
func Build(dir, tag string) string {
	dir = path.Resolve(dir)
	c := cmd.New("docker", "build", "-t", tag, dir)
	c.EchoStdout = true
	c.EchoStderr = true
	return c.Out()
}

func Run(tag string) int {
	c := cmd.New("docker", "run", tag)
	c.EchoStdout = true
	c.EchoStderr = true
	return c.ExitCode()
}
