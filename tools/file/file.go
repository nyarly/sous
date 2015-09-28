package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/path"
)

func Write(data interface{}, pathFormat string, a ...interface{}) {
	path := path.Resolve(pathFormat, a...)
	dataStr := fmt.Sprint(data)
	err := ioutil.WriteFile(path, []byte(dataStr), 0777)
	if err != nil {
		Dief("unable to write file %s; %s", path, err)
	}
}

func Exists(path string) bool {
	i, err := os.Stat(path)
	if err == nil {
		return !i.IsDir()
	}
	if !os.IsNotExist(err) {
		Dief("Unable to determine if file exists at '%s'; %s", path, err)
	}
	return false
}

func ReadString(pathFormat string, a ...interface{}) (string, bool) {
	b, err, _ := Read(pathFormat, a...)
	return string(b), err
}

func ReadJSON(v interface{}, pathFormat string, a ...interface{}) bool {
	b, exists, path := Read(pathFormat, a...)
	if !exists {
		return false
	}
	if err := json.Unmarshal(b, &v); err != nil {
		Dief("Unable to parse JSON in %s as %T: %s", path, v, err)
	}
	return true
}

func Read(pathFormat string, a ...interface{}) ([]byte, bool, string) {
	path := path.Resolve(pathFormat, a...)
	contents, err := ioutil.ReadFile(path)
	if err == nil {
		if len(contents) == 0 {
			Dief("%s is zero-length", path)
		}
		return contents, true, path
	}
	if os.IsNotExist(err) {
		return nil, false, path
	}
	Dief("Unable to read file %s: %s", path, err)
	return nil, false, path
}
