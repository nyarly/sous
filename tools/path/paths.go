package path

import (
	"fmt"
	"os"

	. "path"

	"github.com/opentable/sous/tools/cli"
)

func Resolve(pathFormat string, a ...interface{}) string {
	path := fmt.Sprintf(pathFormat, a...)
	if path[0:2] == "~/" {
		home := os.Getenv("HOME")
		if home == "" {
			cli.Fatalf("unable to resolve path beginning ~/; $HOME not set")
		}
		path = home + "/" + path[2:]
	}
	return path
}

func BaseDir(path string) string {
	return Dir(path)
}
