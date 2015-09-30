package file

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/path"
)

func Write(data []byte, pathFormat string, a ...interface{}) {
	path := path.Resolve(pathFormat, a...)
	err := ioutil.WriteFile(path, data, 0777)
	if err != nil {
		cli.Fatalf("unable to write file %s; %s", path, err)
	}
}

func WriteString(data interface{}, pathFormat string, a ...interface{}) {
	s := fmt.Sprint(data)
	Write([]byte(s), pathFormat, a...)
}

func WriteJSON(data interface{}, pathFormat string, a ...interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		cli.Fatalf("Unable to marshal %T object to JSON: %s", data, err)
	}
	Write(b, pathFormat, a...)
}

func Exists(path string) bool {
	i, err := os.Stat(path)
	if err == nil {
		return !i.IsDir()
	}
	if !os.IsNotExist(err) {
		cli.Fatalf("Unable to determine if file exists at '%s'; %s", path, err)
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
		cli.Fatalf("Unable to parse JSON in %s as %T: %s", path, v, err)
	}
	if v == nil {
		cli.Fatalf("Unmarshalled nil")
	}
	return true
}

func Read(pathFormat string, a ...interface{}) ([]byte, bool, string) {
	path := path.Resolve(pathFormat, a...)
	contents, err := ioutil.ReadFile(path)
	if err == nil {
		if len(contents) == 0 {
			cli.Fatalf("%s is zero-length", path)
		}
		return contents, true, path
	}
	if os.IsNotExist(err) {
		return nil, false, path
	}
	cli.Fatalf("Unable to read file %s: %s", path, err)
	return nil, false, path
}
