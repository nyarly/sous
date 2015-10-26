package docker

import (
	"fmt"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/cmd"
)

type Container interface {
	CID() string
	Name() string
	String() string
	Kill() error
	Start() error
	Exists() bool
}

type container struct {
	cid, name string
}

func ContainerWithName(name string) Container {
	return &container{"", name}
}

func ContainerWithCID(cid string) Container {
	return &container{cid, ""}
}

func (c *container) CID() string  { return c.cid }
func (c *container) Name() string { return c.name }

func (c *container) Exists() bool {
	t := cmd.Table("docker", "ps", "-a", "--no-trunc")
	var match func([]string) bool
	if c.name != "" {
		match = func(row []string) bool { return row[6] == c.name }
	} else if c.cid != "" {
		match = func(row []string) bool { return row[0] == c.cid }
	} else {
		cli.Fatalf("Sous Programmer Error: Container has neither CID nor Name")
	}
	for _, r := range t {
		if match(r) {
			return true
		}
	}
	return false
}

func (c *container) Kill() error {
	if ex := cmd.ExitCode("docker", "kill", c.cid); ex != 0 {
		return fmt.Errorf("Unable to kill docker container %s", c)
	}
	return nil
}

func (c *container) Start() error {
	if ex := cmd.ExitCode("docker", "start", c.cid); ex != 0 {
		return fmt.Errorf("Unable to start docker container %s", c)
	}
	return nil
}

func (c *container) String() string {
	return c.cid
}
