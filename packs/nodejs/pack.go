package nodejs

import (
	"fmt"
	"strings"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/version"
)

var Pack = &core.Pack{
	Name: "NodeJS",
	Detect: func() (packInfo interface{}, err error) {
		var np *NodePackage
		if !file.ReadJSON(np, "package.json") {
			return fmt.Errorf("no package.json file found"), nil
		}
		return np, nil
	},
	ProjectDesc: func(packInfo interface{}) string {
		np := packInfo.(*NodePackage)
		return fmt.Sprintf("a NodeJS %s project named %s v%s",
			np.Engines.Node, np.Name, np.Version)
	},
	CheckCompatibility: func(packInfo interface{}) []string {
		c := []string{}
		np := packInfo.(*NodePackage)
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
	Targets: map[string]*core.Target{
		"build": &core.Target{
			Detect: func(c *core.Context, packInfo interface{}) (*core.AppInfo, error) {
				np := packInfo.(*NodePackage)
				if len(np.Scripts.Start) == 0 {
					return nil, fmt.Errorf("package.json does not specify a start script")
				}
				return &core.AppInfo{
					Version: np.Version,
					Data:    np,
				}, nil
			},
			MakeDockerfile: func(i *core.AppInfo, packInfo interface{}) *docker.Dockerfile {
				return buildNodeJS(i.Data.(*NodePackage))
			},
		},
		"test": &core.Target{
			Detect: func(c *core.Context, packInfo interface{}) (*core.AppInfo, error) {
				np := packInfo.(*NodePackage)
				if len(np.Scripts.Test) == 0 {
					return nil, fmt.Errorf("package.json does not specify a test script")
				}
				return &core.AppInfo{
					Version: np.Version,
					Data:    np,
				}, nil
			},
			MakeDockerfile: func(i *core.AppInfo, packInfo interface{}) *docker.Dockerfile {
				return testNodeJS(i.Data.(*NodePackage))
			},
		},
	},
}
