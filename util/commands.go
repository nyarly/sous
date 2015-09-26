package util

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
)

var Whitespace = regexp.MustCompile("[ \\t\\r\\n]+")

func _cmd(name string, a ...string) (stdout, stderr string, code int, err error) {
	c := exec.Command(name, a...)
	o := &bytes.Buffer{}
	e := &bytes.Buffer{}
	c.Stdout = o
	c.Stderr = e
	if err := c.Start(); err != nil {
		Dief("Unable to begin command execution; %s", err)
	}
	if err := c.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return o.String(), e.String(), status.ExitStatus(), err
			}
		}
		Dief("Command failed, unable to get exit code; Command was '%s %s'", name, strings.Join(a, " "))
	}
	return o.String(), e.String(), 0, nil
}

func Cmd(command string, args ...string) string {
	o, _, _, err := _cmd(command, args...)
	if err != nil {
		Dief("Could not run %s: %s", command, err)
	}
	return TrimWhitespace(o)
}

// CmdLines returns whitespace-stripped lines, and removes empty lines.
func CmdLines(command string, args ...string) []string {
	out := Cmd(command, args...)
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
func CmdTable(command string, args ...string) [][]string {
	lines := CmdLines(command, args...)
	rows := make([][]string, len(lines))
	for i, line := range lines {
		rows[i] = Whitespace.Split(line, -1)
	}
	return rows
}

func CmdExitCode(command string, args ...string) int {
	_, _, code, _ := _cmd(command, args...)
	return code
}
