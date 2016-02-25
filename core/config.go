package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/dir"
	"github.com/opentable/sous/tools/file"
)

type Props map[string]string

func Properties() Props {
	var c Props
	file.ReadJSON(&c, propertiesFilePath())
	if c == nil {
		c = map[string]string{}
	}
	return c
}

var s *State

func Load() *Config {
	if s == nil {
		if !file.ReadJSON(&s, "~/.sous/config") {
			if err := Update(); err != nil {
				cli.Fatalf("Unable to load config: %s", err)
			}
			if !file.ReadJSON(&s, "~/.sous/config") {
				cli.Fatalf("Unable to read %s", "~/.sous/config")
			}
		}
	}
	return &s.Config
}

func Update() error {
	s = nil
	p := cli.BeginProgress("Updating config")
	Set("last-update-check", time.Now().Format(time.RFC3339))
	props := Properties()
	serverURL := props["sous-server"]
	if serverURL == "" {
		p.Done("Failed")
		return fmt.Errorf("sous-server not set; use `sous config sous-server http://your.sous.server`")
	}
	var c *State
	if err := getJSON(&c, "%s/state", serverURL); err != nil {
		p.Done("Failed")
		return err
	}
	file.WriteJSON(c, "~/.sous/config")
	p.Done("Done")
	return nil
}

func Set(name, value string) {
	c := Properties()
	c[name] = value
	save(c)
}

func save(c Props) {
	file.WriteJSON(c, propertiesFilePath())
}

func propertiesFilePath() string {
	return dir.Resolve("~/.sous/properties")
}
func configFilePath() string {
	return dir.Resolve("~/.sous/config")
}

func getJSON(v interface{}, urlFormat string, a ...interface{}) error {
	jsonURL := fmt.Sprintf(urlFormat, a...)
	r, err := http.Get(jsonURL)
	if err != nil {
		return err
	}
	if r.StatusCode != 200 {
		return fmt.Errorf("%s returned HTTP status code %d",
			jsonURL, r.StatusCode)
	}

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return fmt.Errorf("Unable to parse JSON from %s as %T: %s", jsonURL, v, err)
	}
	return nil
}
