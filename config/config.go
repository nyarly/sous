package config

import (
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/path"
)

type Config map[string]string

func Load() Config {
	var c Config
	file.ReadJSON(&c, configFilePath())
	if c == nil {
		c = map[string]string{}
	}
	return c
}

func Set(name, value string) {
	c := Load()
	c[name] = value
	save(c)
}

func save(c Config) {
	file.WriteJSON(c, configFilePath())
}

func configFilePath() string {
	return path.Resolve("~/.sous/config")
}
