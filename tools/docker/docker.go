package docker

import (
	"strings"

	"github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/version"
)

func RequireVersion(r *version.R) {
	vs := cmd.Table("docker", "--version")[0][2]
	vs = strings.Trim(vs, ",")
	v := version.Version(vs)
	if !r.IsSatisfiedBy(v) {
		tools.Dief("got docker version %s; want %s", v, r)
	}
}

func RequireDaemon() {
	if c := cmd.ExitCode("docker", "ps"); c != 0 {
		tools.Dief("`docker ps` exited with code %d", c)
	}
}

func Build(tag string) {
	cmd.EchoAll("docker", "build", "-t", tag, ".")
}

func Run(tag string) {
	cmd.EchoAll("docker", "run", tag)
}
