package path

import (
	"fmt"
	"os"

	. "github.com/opentable/sous/tools"
)

func Resolve(pathFormat string, a ...interface{}) string {
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
