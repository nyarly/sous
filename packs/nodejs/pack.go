package nodejs

import (
	"fmt"
	"strings"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/version"
)

var Pack = &build.Pack{
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
			c = append(c, "no node version specified in package.json:engines.node")
		} else {
			r := version.Range(np.Engines.Node)
			if v := r.BestMatchFrom(availableNodeVersions); v == nil {
				f := "unable to satisfy node version %s (from package.json:engines.node); avialable versions are: %s"
				m := fmt.Sprintf(f, r.Original, strings.Join(availableNodeVersions.Strings(), " "))
				c = append(c, m)
			}
		}
		if np.Version == "" {
			c = append(c, "no version specified in package.json:version")
		}
		return c
	},
	Features: map[string]*build.Feature{
		"build": &build.Feature{
			Detect: func(c *build.Context) (*build.AppInfo, error) {
				var np *NodePackage
				if !file.ReadJSON(&np, "package.json") {
					return nil, fmt.Errorf("no file named package.json")
				}
				if len(np.Scripts.Start) == 0 {
					return nil, fmt.Errorf("package.json does not specify a start script")
				}
				return &build.AppInfo{
					Version: np.Version,
					Data:    np,
				}, nil
			},
			MakeDockerfile: func(i *build.AppInfo) *docker.Dockerfile {
				return buildNodeJS(i.Data.(*NodePackage))
			},
		},
		"test": &build.Feature{
			Detect: func(c *build.Context) (*build.AppInfo, error) {
				var np *NodePackage
				if !file.ReadJSON(&np, "package.json") {
					return nil, fmt.Errorf("no file named package.json")
				}
				if len(np.Scripts.Test) == 0 {
					return nil, fmt.Errorf("package.json does not specify a test script")
				}
				return &build.AppInfo{
					Version: np.Version,
					Data:    np,
				}, nil
			},
			MakeDockerfile: func(i *build.AppInfo) *docker.Dockerfile {
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
