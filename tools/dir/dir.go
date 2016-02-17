package dir

import (
	"os"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/path"
)

func Exists(pathFormat string, a ...interface{}) bool {
	path := path.Resolve(pathFormat, a...)
	s, err := os.Stat(path)
	if err == nil {
		if s.IsDir() {
			return true
		} else {
			return false
		}
	}
	if !os.IsNotExist(err) {
		cli.Fatalf("unable to stat path %s", path)
	}
	return false
}

func EnsureExists(pathFormat string, a ...interface{}) {
	path := path.Resolve(pathFormat, a...)
	s, err := os.Stat(path)
	if err == nil {
		if s.IsDir() {
			return
		} else {
			cli.Fatalf("%s exists and is not a directory", path)
		}
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0777); err != nil {
			cli.Fatalf("unable to make directory %s; %s", path, err)
		}
		return
	}
	cli.Fatalf("unable to stat or create directory %s", path)
}

func Current() string {
	wd, err := os.Getwd()
	if err != nil {
		cli.Fatalf("%s", err)
	}
	return wd
}
