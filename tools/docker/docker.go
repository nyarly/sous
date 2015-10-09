package docker

import (
	"fmt"
	"strings"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/dockermachine"
	"github.com/opentable/sous/tools/file"
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
		cli.Logf("`docker ps` exited with code %d", c)
		if dockermachine.Installed() {
			vms := dockermachine.RunningVMs()
			if len(vms) != 0 {
				cli.Fatalf(`Tip: eval "$(docker-machine env %s)"`, vms[0])
			}
			vms = dockermachine.VMs()
			switch len(vms) {
			case 0:
				cli.Logf("Tip: you should create a machine using docker-machine")
			case 1:
				start := ""
				if cmd.Stdout("docker-machine", "status", vms[0]) != "Running" {
					start = fmt.Sprintf("docker-machine start %s && ", vms[0])
				}
				cli.Logf(`Tip: %seval "$(docker-machine env %s)"`, start, vms[0])
			default:
				cli.Logf("Tip: start one of your docker machines (%s)",
					strings.Join(vms, ", "))
			}
		}
		cli.Fatal()
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

// BuildFile builds the specified docker file in the context of the specified
// directory.
func BuildFile(dockerfile, dir, tag string) string {
	if !file.Exists(dockerfile) {
		cli.Fatalf("File does not exist: %s")
	}
	dir = path.Resolve(dir)
	dfLocal := ".SousDockerfile"
	if file.Exists(dfLocal) {
		file.Remove(dfLocal)
	}
	file.Link(dockerfile, dfLocal)
	defer file.Remove(dfLocal)
	return dockerCmd("build", "-f", dfLocal, "-t", tag, dir).Out()
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
