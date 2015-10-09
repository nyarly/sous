package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	. "github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/cli"
)

type CMD struct {
	Name                     string
	Args, Env                []string
	EchoStdout, EchoStderr   bool
	Stdout, Stderr           *bytes.Buffer
	WriteStdout, WriteStderr io.Writer
}

func New(name string, args ...string) *CMD {
	return &CMD{
		Name:   name,
		Args:   args,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
}

func (C *CMD) execute() (code int, err error) {
	c := exec.Command(C.Name, C.Args...)
	c.Stdout = C.Stdout
	c.Stderr = C.Stderr
	if C.EchoStdout {
		c.Stdout = io.MultiWriter(os.Stdout, c.Stdout)
	}
	if C.EchoStderr {
		c.Stderr = io.MultiWriter(os.Stderr, c.Stderr)
	}
	if C.WriteStdout != nil {
		c.Stdout = io.MultiWriter(C.WriteStdout, c.Stdout)
	}
	if C.WriteStderr != nil {
		c.Stderr = io.MultiWriter(C.WriteStderr, c.Stderr)
	}
	if err := c.Start(); err != nil {
		cli.Fatalf("Unable to begin command execution; %s", err)
	}
	if C.EchoStdout || C.EchoStderr {
		cli.Logf("shell> %s", C)
	}
	err = c.Wait()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), err
			}
		}
		cli.Fatalf("Command failed, unable to get exit code: %s", C)
	}
	return 0, nil
}

func Stdout(command string, args ...string) string {
	return New(command, args...).Out()
}

func ExitCode(command string, args ...string) int {
	return New(command, args...).ExitCode()
}

func JSON(v interface{}, command string, args ...string) {
	New(command, args...).JSON(v)
}

func (c *CMD) EchoAll() {
	c.EchoStdout = true
	c.EchoStderr = true
	c.Run()
}

func EchoAll(command string, args ...string) {
	New(command, args...).EchoAll()
}

func (c *CMD) Run() {
	c.Out()
}

func (c *CMD) Out() string {
	if _, err := c.execute(); err != nil {
		cli.Fatalf("Error running %s; %s", c, err)
	}
	return TrimWhitespace(c.Stdout.String())
}

// CmdLines returns whitespace-stripped lines, and removes empty lines.
func (c *CMD) OutLines() []string {
	out := c.Out()
	rawLines := strings.Split(out, "\n")
	lines := []string{}
	for _, line := range rawLines {
		trimmed := TrimWhitespace(line)
		if len(trimmed) != 0 {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

// CmdTable is similar to CmdLines, but further splits each line by whitespace.
func (c *CMD) OutTable() [][]string {
	lines := c.OutLines()
	rows := make([][]string, len(lines))
	for i, line := range lines {
		rows[i] = Whitespace.Split(line, -1)
	}
	return rows
}

func Table(command string, args ...string) [][]string {
	return New(command, args...).OutTable()
}

func Lines(command string, args ...string) []string {
	return New(command, args...).OutLines()
}

func (c *CMD) ExitCode() int {
	code, _ := c.execute()
	return code
}

func (c *CMD) String() string {
	args := strings.Join(c.Args, " ")
	env := strings.Join(c.Env, " ")
	if len(env) != 0 {
		env = env + " "
	}
	return fmt.Sprintf("%s%s %s", env, c.Name, args)
}

func (c *CMD) JSON(v interface{}) {
	o := c.Out()
	if err := json.Unmarshal([]byte(o), &v); err != nil {
		cli.Fatalf("Unable to parse JSON from %s as %T: %s", c, v, err)
	}
	if v == nil {
		cli.Fatalf("Unmarshalled nil")
	}
}
