package nodejs

import (
	"fmt"

	"github.com/opentable/sous/build"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
)

var Pack = &build.Pack{
	Name:   "NodeJS",
	Detect: detect,
	Features: map[string]*build.Feature{
		"build": &build.Feature{
			Detect: func(c *build.Context) (*build.AppInfo, error) {
				var np *NodePackage
				if !file.ReadJSON(&np, "package.json") {
					return nil, fmt.Errorf("no file named package.json")
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
	if !file.Exists("package.json") {
		return fmt.Errorf("no package.json file found")
	}
	return nil
}
