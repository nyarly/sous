package docker

import (
	"github.com/opentable/sous/tools/cmd"
)

func Build(tag string) {
	cmd.EchoAll("docker", "build", "-t", tag, ".")
}

func Run(tag string) {
	cmd.EchoAll("docker", "run", tag)
}
