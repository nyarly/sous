package util

import (
	"os/exec"
	"regexp"
	"strings"
)

var Whitespace = regexp.MustCompile("[ \\t\\r\\n]+")

func Cmd(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	out, err := cmd.Output()
	if err != nil {
		Dief("Could not run %s: %s", command, err)
	}
	return strings.Trim(string(out), " \t\r\n")
}

func CmdLines(command string, args ...string) []string {
	out := Cmd(command, args...)
	rawLines := strings.Split(out, "\n")
	lines := make([]string, len(rawLines))
	for i, line := range rawLines {
		lines[i] = TrimWhitespace(line)
	}
	return lines
}

func CmdTable(command string, args ...string) [][]string {
	lines := CmdLines(command, args...)
	rows := make([][]string, len(lines))
	for i, line := range lines {
		rows[i] = Whitespace.Split(line, -1)
	}
	return rows
}
