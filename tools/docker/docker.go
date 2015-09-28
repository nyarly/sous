package docker

import (
	. "github.com/opentable/sous/tools"
)

func Build(tag string) {
	Cmd("docker", "build", "-t", tag, ".")
}

func Run(tag string) {
	Cmd("docker", "run", "tag")
}
