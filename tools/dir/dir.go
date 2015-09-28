package dir

import (
	"os"

	. "github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/path"
)

func EnsureExists(pathFormat string, a ...interface{}) {
	path := path.Resolve(pathFormat, a...)
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
