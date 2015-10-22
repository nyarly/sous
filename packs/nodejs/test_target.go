package nodejs

import (
	"fmt"

	"github.com/opentable/sous/tools/docker"
)

type TestTarget struct {
	*NodeJSTarget
}

func NewTestTarget(np *NodePackage) *TestTarget {
	return &TestTarget{NewNodeJSTarget("test", np)}
}

func (t *TestTarget) DependsOn() []string { return nil }

func (t *TestTarget) RunAfter() []string { return []string{"compile"} }

func (t *TestTarget) Desc() string {
	return "The NodeJS test target builds your Docker image using `npm install`. When you invoke the container, it simply runs `npm test` to execute your test script defined in `package.json:scripts.test`"
}

func (t *TestTarget) Check() error {
	if len(t.PackageJSON.Scripts.Test) == 0 {
		return fmt.Errorf("package.json does not specify a test script")
	}
	return nil
}

func (t *TestTarget) Dockerfile() *docker.Dockerfile {
	df := baseDockerfile(t.PackageJSON)
	df.AddRun("cd " + wd + " && npm install")
	df.AddLabel("com.opentable.tests", "true")
	df.CMD = []string{"npm", "test"}
	return df
}
