package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func WriteFile(data interface{}, pathFormat string, a ...interface{}) {
	path := ResolvePath(pathFormat, a...)
	dataStr := fmt.Sprint(data)
	err := ioutil.WriteFile(path, []byte(dataStr), 0777)
	if err != nil {
		Dief("unable to write file %s; %s", path, err)
	}
}

func ReadFileString(pathFormat string, a ...interface{}) (string, bool) {
	b, err, _ := ReadFile(pathFormat, a...)
	return string(b), err
}

func ReadFileJSON(v interface{}, pathFormat string, a ...interface{}) bool {
	b, exists, path := ReadFile(pathFormat, a...)
	if !exists {
		return false
	}
	if err := json.Unmarshal(b, &v); err != nil {
		Dief("Unable to parse JSON in %s as %T: %s", path, v, err)
	}
	return true
}

func ReadFile(pathFormat string, a ...interface{}) ([]byte, bool, string) {
	path := ResolvePath(pathFormat, a...)
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
