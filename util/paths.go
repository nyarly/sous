package util

import (
	"fmt"
	"os"
)

func ResolvePath(pathFormat string, a ...interface{}) string {
	path := fmt.Sprintf(pathFormat, a...)
	if path[0:2] == "~/" {
		home := os.Getenv("HOME")
		if home == "" {
			Dief("unable to resolve path beginning ~/; $HOME not set")
		}
		path = home + "/" + path[2:]
	}
	return path
}

func EnsureDirExists(pathFormat string, a ...interface{}) {
	path := ResolvePath(pathFormat, a...)
	s, err := os.Stat(path)
	if err == nil {
		if s.IsDir() {
			return
		} else {
			Dief("%s exists and is not a directory", path)
		}
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0777); err != nil {
			Dief("unable to make directory %s; %s", path, err)
		}
		return
	}
	Dief("unable to stat or create directory %s", path)
}
