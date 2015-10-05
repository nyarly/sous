package dockermachine

import (
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/cmd"
)

func Installed() bool {
	return cmd.ExitCode("docker-machine") == 0
}

func VMs() []string {
	return cmd.Lines("docker-machine", "ls", "-q")
}

func RunningVMs() []string {
	list := []string{}
	for _, v := range VMs() {
		if cmd.Stdout("docker-machine", "status", v) == "Running" {
			list = append(list, v)
		}
	}
	return list
}

type DMInfo struct {
	Driver DMInfoDriver
}

type DMInfoDriver struct {
	IPAddress string
}

func HostIP(vm string) string {
	var dmi *DMInfo
	cli.Logf("HOSTIP")
	cmd.JSON(&dmi, "docker-machine", "inspect", vm)
	cli.Logf("HOSTIP: %+v", dmi)
	return dmi.Driver.IPAddress
}
