package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	// Ensure dependencies are installed:
	//    - Git
	// Sniff current directory for type:
	//
	//    - NodeJS = package.json
	//    - others later
	gitVersion := cmd("git", "version")
	logf(gitVersion)
	buildContext := getBuildContext()
	tryBuildNodeJS(buildContext)
	dief("no buildable project detected")
}

func writeFile(data interface{}, pathFormat string, a ...interface{}) {
	path := fmt.Sprintf(pathFormat, a...)
	dataStr := fmt.Sprint(data)
	err := ioutil.WriteFile(path, []byte(dataStr), 777)
	if err != nil {
		dief("unable to write file %s; %s", path, err)
	}
}

func readFileString(pathFormat string, a ...interface{}) (string, bool) {
	b, err, _ := readFile(pathFormat, a...)
	return string(b), err
}

func readFileJSON(v interface{}, pathFormat string, a ...interface{}) bool {
	b, exists, path := readFile(pathFormat, a...)
	if !exists {
		return false
	}
	if err := json.Unmarshal(b, &v); err != nil {
		dief("Unable to parse JSON in %s as %T: %s", path, v, err)
	}
	return true
}

func readFile(pathFormat string, a ...interface{}) ([]byte, bool, string) {
	path := fmt.Sprintf(pathFormat, a...)
	contents, err := ioutil.ReadFile(path)
	if err == nil {
		if len(contents) == 0 {
			dief("%s is zero-length", path)
		}
		return contents, true, path
	}
	if os.IsNotExist(err) {
		return nil, false, path
	}
	dief("Unable to read file %s: %s", path, err)
	return nil, false, path
}

func cmd(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	out, err := cmd.Output()
	if err != nil {
		dief("Could not run %s: %s", command, err)
	}
	return strings.Trim(string(out), " \t\r\n")
}

func cmdLines(command string, args ...string) []string {
	out := cmd(command, args...)
	rawLines := strings.Split(out, "\n")
	lines := make([]string, len(rawLines))
	for i, line := range rawLines {
		lines[i] = trimWhitespace(line)
	}
	return lines
}

func cmdTable(command string, args ...string) [][]string {
	lines := cmdLines(command, args...)
	rows := make([][]string, len(lines))
	for i, line := range lines {
		rows[i] = whitespace.Split(line, -1)
	}
	return rows
}

func resolvePath(pathFormat string, a ...interface{}) string {
	path := fmt.Sprintf(pathFormat, a...)
	if path[0:1] == "~/" {
		home := os.Getenv("HOME")
		if home == "" {
			dief("unable to resolve path beginning ~/; $HOME not set")
		}
		path = home + "/" + path[2:]
	}
	return path
}

func ensureDirExists(pathFormat string, a ...interface{}) {
	path := resolvePath(pathFormat, a...)
	s, err := os.Stat(path)
	if err == nil {
		if s.IsDir() {
			return
		} else {
			dief("%s exists and is not a directory", path)
		}
	}
	if os.IsNotExist(err) {
		if err := os.Mkdir(path, 777); err != nil {
			dief("unable to make directory %s; %s", path, err)
		}
		return
	}
	dief("unable to stat or create directory %s", path)
}

var whitespace = regexp.MustCompile("[ \\t\\r\\n]+")

func trimWhitespace(s string) string {
	return strings.Trim(s, "\t\r\n")
}

func dief(format string, a ...interface{}) {
	logf(format, a...)
	os.Exit(1)
}

func logf(format string, a ...interface{}) {
	log.Printf(format, a...)
}

func exitSuccess(format string, a ...interface{}) {
	if len(format) != 0 {
		logf(format, a...)
	}
	os.Exit(0)
}
