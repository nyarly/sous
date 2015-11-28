package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/yaml"
)

func Parse(configDir string) (*State, error) {
	configFile := fmt.Sprintf("%s/config.yaml", configDir)
	var state State
	if err := parseYAMLFile(configFile, &state); err != nil {
		return nil, err
	}
	dcs, err := parseDatacentres(fmt.Sprintf("%s/datacentres", configDir))
	if err != nil {
		return nil, err
	}
	state.Datacentres = dcs
	manifestsDir := fmt.Sprintf("%s/manifests", configDir)
	manifests, err := parseManifests(manifestsDir)
	if err != nil {
		return nil, err
	}
	state.Manifests = manifests
	return &state, nil
}

func parseDatacentres(datacentresDir string) (Datacentres, error) {
	dcs := Datacentres{}
	err := walkYAMLDir(datacentresDir, func(path string) error {
		var dc Datacentre
		if err := parseYAMLFile(path, &dc); err != nil {
			return err
		}
		dcs[dc.Name] = &dc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dcs, nil
}

func parseYAMLFile(f string, v interface{}) error {
	if !file.Exists(f) {
		return fmt.Errorf("%s not found", f)
	}
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(b, v); err != nil {
		return fmt.Errorf("unable to parse %s as %T: %s", f, v, err)
	}
	return nil
}

func walkYAMLDir(d string, fn func(path string) error) error {
	files, err := filepath.Glob(d + "/*.yaml")
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := fn(f); err != nil {
			return err
		}
	}
	return nil
}

func parseManifests(manifestsDir string) (Manifests, error) {
	manifests := Manifests{}
	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}
		manifest, err := parseManifest(manifestsDir, path)
		if err != nil {
			return err
		}
		manifests[manifest.App.SourceRepo] = manifest
		return nil
	}
	if err := filepath.Walk(manifestsDir, fn); err != nil {
		return nil, err
	}
	return manifests, nil
}

func parseManifest(manifestsDir, path string) (*Manifest, error) {
	manifest := Manifest{}
	if err := parseYAMLFile(path, &manifest); err != nil {
		return nil, err
	}
	relPath, err := filepath.Rel(manifestsDir, path)
	if err != nil {
		return nil, err
	}
	// Check manifest SourceRepo matches path
	expectedSourceRepo := strings.TrimSuffix(relPath, ".yaml")
	if manifest.App.SourceRepo != expectedSourceRepo {
		return nil, fmt.Errorf("SourceRepo was %q; want %q (%s)\nREST:%+v",
			manifest.App.SourceRepo, expectedSourceRepo, path, manifest)
	}
	return &manifest, nil
}

type State struct {
	EnvironmentDefs EnvDefs
	Datacentres     Datacentres
	Manifests       Manifests
}

type EnvDefs map[string]*EnvDef

type EnvDef map[string]*VarDef

type VarDef struct {
	Type       VarType
	Name, Desc string
	Automatic  bool
}

type VarType string

const (
	URL_VARTYPE    = VarType("url")
	INT_VARTYPE    = VarType("int")
	STRING_VARTYPE = VarType("string")
)

type Datacentres map[string]*Datacentre

type Datacentre struct {
	Name string
	Env  DatacentreEnv
}

type DatacentreEnv map[string]string

type Manifests map[string]*Manifest

type Manifest struct {
	App         App
	Deployments Deployments
}

type App struct {
	SourceRepo, Owner, Kind string
}

type Deployments map[string]*Deployment

type Deployment struct {
	Instance                  *Instance
	SourceTag, SourceRevision string
}

type Instance struct {
	Count  string
	CPUs   string
	Memory string
}

type MemorySize string
