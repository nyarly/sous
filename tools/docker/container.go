package docker

import (
	"fmt"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/cmd"
)

type Container interface {
	CID() string
	Name() string
	Image() string
	String() string
	Kill() error
	Remove() error
	Start() error
	Exists() bool
	Running() bool
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
	if c.name != "" {
		return len(cmd.Lines("docker", "ps", "-a", "--filter", "name="+c.name)) > 1
	} else if c.cid != "" {
		return len(cmd.Lines("docker", "ps", "-a", "--filter", "id="+c.cid)) > 1
	}
	cli.Fatalf("Sous Programmer Error: Container has neither CID nor Name")
	return false
}

func (c *container) Running() bool {
	if c.name != "" {
		return len(cmd.Lines("docker", "ps", "--filter", "name="+c.name)) > 1
	} else if c.cid != "" {
		return len(cmd.Lines("docker", "ps", "--filter", "id="+c.cid)) > 1
	}
	cli.Fatalf("Sous Programmer Error: Container has neither CID nor Name")
	return false
}

func (c *container) Image() string {
	var dc []DockerContainer
	cmd.JSON(&dc, "docker", "inspect", c.Name())
	if len(dc) == 0 {
		cli.Fatalf("Container %s does not exist.", c)
	}
	if len(dc) != 1 {
		cli.Fatalf("Multiple containers match %s", c)
	}
	return dc[0].Image
}

type DockerContainer struct {
	ID, Name, Image string
}

func (c *container) Kill() error {
	if ex := cmd.ExitCode("docker", "kill", c.effectiveName()); ex != 0 {
		return fmt.Errorf("Unable to kill docker container %s", c)
	}
	return nil
}

func (c *container) Remove() error {
	if ex := cmd.ExitCode("docker", "rm", c.effectiveName()); ex != 0 {
		return fmt.Errorf("Unable to remove docker container %s", c)
	}
	return nil
}

func (c *container) Start() error {
	if ex := cmd.ExitCode("docker", "start", c.effectiveName()); ex != 0 {
		return fmt.Errorf("Unable to start docker container %s", c)
	}
	return nil
}

func (c *container) Wait() error {
	if ex := cmd.ExitCode("docker", "wait", c.effectiveName()); ex != 0 {
		return fmt.Errorf("Unable to wait on docker container %s", c)
	}
	return nil
}

func (c *container) String() string {
	return c.effectiveName()
}

func (c *container) effectiveName() string {
	if c.cid == "" {
		return c.name
	}
	return c.cid
}
