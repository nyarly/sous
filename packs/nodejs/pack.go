package nodejs

import (
	"fmt"
	"strings"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools/file"
	"github.com/opentable/sous/tools/version"
)

func GetPack() *core.Pack {
	return &core.Pack{
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
		MakeTargets: func(packInfo interface{}) []core.Target {
			np := packInfo.(*NodePackage)
			return []core.Target{
				NewAppTarget(np),
				NewTestTarget(np),
			}
		},
	}
}
