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
	contractsDir := fmt.Sprintf("%s/contracts", configDir)
	contracts, err := ParseContracts(contractsDir)
	if err != nil {
		return nil, err
	}
	state.Contracts = contracts
	return &state, nil
}

func ParseContracts(contractsDir string) (Contracts, error) {
	contracts := Contracts{}
	serversDir := filepath.Join(contractsDir, "servers")
	servers, err := parseServers(serversDir)
	if err != nil {
		return nil, err
	}
	err = walkYAMLDir(contractsDir, func(path string) error {
		var c Contract
		if err := parseYAMLFile(path, &c); err != nil {
			return err
		}
		contracts[c.Name] = c
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Add servers to contracts...
	for name, contract := range contracts {
		contract.Servers = map[string]TestServer{}
		for _, serverName := range contract.StartServers {
			server, ok := servers[serverName]
			if !ok {
				return nil, fmt.Errorf("Server %q not defined in %q", serverName, serversDir)
			}
			contract.Servers[serverName] = server
			contracts[name] = contract
		}
	}
	return contracts, nil
}

func parseServers(serversDir string) (map[string]TestServer, error) {
	servers := map[string]TestServer{}
	err := walkYAMLDir(serversDir, func(path string) error {
		var s TestServer
		if err := parseYAMLFile(path, &s); err != nil {
			return err
		}
		servers[s.Name] = s
		return nil
	})
	if err != nil {
		return nil, err
	}
	return servers, nil
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
		manifests[manifest.App.SourceRepo] = *manifest
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
