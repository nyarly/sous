package nodejs

import (
	"fmt"
	"strings"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/version"
)

var Pack = &core.Pack{
	Name:   "NodeJS",
	Detect: detect,
	CompatibleProjectDesc: func() string {
		var np *NodePackage
		if !file.ReadJSON(&np, "package.json") {
			cli.Fatalf("no file named package.json")
		}
		return fmt.Sprintf("a NodeJS %s project named %s v%s",
			np.Engines.Node, np.Name, np.Version)
	},
	CheckCompatibility: func() []string {
		c := []string{}
		var np *NodePackage
		if !file.ReadJSON(&np, "package.json") {
			cli.Fatalf("no file named package.json")
		}
		if np.Engines.Node == "" {
			c = append(c, "unable to determine NodeJS version, missing engines.node in package.json")
		} else {
			r := version.Range(np.Engines.Node)
			if v := r.BestMatchFrom(AvailableNodeVersions()); v == nil {
				f := "node version range (%s) not supported (pick from %s)"
				m := fmt.Sprintf(f, r.Original, strings.Join(AvailableNodeVersions().Strings(), ", "))
				c = append(c, m)
			}
		}
		if np.Version == "" {
			c = append(c, "no version specified in package.json:version")
		}
		return c
	},
	Features: map[string]*core.Feature{
		"build": &core.Feature{
			Detect: func(c *core.Context) (*core.AppInfo, error) {
				var np *NodePackage
				if !file.ReadJSON(&np, "package.json") {
					return nil, fmt.Errorf("no file named package.json")
				}
				if len(np.Scripts.Start) == 0 {
					return nil, fmt.Errorf("package.json does not specify a start script")
				}
				return &core.AppInfo{
					Version: np.Version,
					Data:    np,
				}, nil
			},
			MakeDockerfile: func(i *core.AppInfo) *docker.Dockerfile {
				return buildNodeJS(i.Data.(*NodePackage))
			},
		},
		"test": &core.Feature{
			Detect: func(c *core.Context) (*core.AppInfo, error) {
				var np *NodePackage
				if !file.ReadJSON(&np, "package.json") {
					return nil, fmt.Errorf("no file named package.json")
				}
				if len(np.Scripts.Test) == 0 {
					return nil, fmt.Errorf("package.json does not specify a test script")
				}
				return &core.AppInfo{
					Version: np.Version,
					Data:    np,
				}, nil
			},
			MakeDockerfile: func(i *core.AppInfo) *docker.Dockerfile {
				return testNodeJS(i.Data.(*NodePackage))
			},
		},
	},
}

func detect() error {
	var np *NodePackage
	if !file.ReadJSON(np, "package.json") {
		return fmt.Errorf("no package.json file found")
	}
	return nil
}
