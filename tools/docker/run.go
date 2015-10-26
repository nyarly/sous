package docker

import (
	"fmt"
	"os/exec"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/cmd"
	"github.com/opentable/sous/tools/file"
)

type Run struct {
	Image, Name            string
	Env                    map[string]string
	Net                    string
	StdoutFile, StderrFile string
	inBackground           bool
}

func NewRun(image string) *Run {
	return &Run{
		Image: image,
		Net:   "host",
		Env:   map[string]string{},
	}
}

func (r *Run) AddEnv(key, value string) {
	r.Env[key] = value
}

func (r *Run) Background() *Run {
	r.inBackground = true
	return r
}

func (r *Run) prepareCommand() *cmd.CMD {
	args := []string{"run"}
	if r.inBackground {
		args = append(args, "-d")
	}
	if r.Name != "" {
		args = append(args, "--name", r.Name)
	}
	for k, v := range r.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	if r.Net != "" {
		args = append(args, "--net="+r.Net)
	}
	args = append(args, r.Image)
	c := dockerCmd(args...)
	if r.inBackground {
		c.EchoStdout = false
		c.EchoStderr = false
	}
	return c
}

func (r *Run) ExitCode() int {
	return r.prepareCommand().ExitCode()
}

func (r *Run) Start() (*container, error) {
	r.inBackground = true
	c := r.prepareCommand()
	cid := c.Out()
	tailLogs := exec.Command("docker", "logs", "-f", cid)
	tailLogs.Stdout = file.Create(r.StdoutFile)
	tailLogs.Stderr = file.Create(r.StderrFile)
	err := tailLogs.Start()
	if err != nil {
		cli.Fatalf("Unable to tail logs: %s", err)
	}
	return &container{cid, ""}, nil
}
