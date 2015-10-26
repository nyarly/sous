package nodejs

import (
	"fmt"

	"github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/docker"
)

type AppTarget struct {
	*NodeJSTarget
}

func NewAppTarget(pack *Pack) *AppTarget {
	return &AppTarget{NewNodeJSTarget("app", pack)}
}

func (t *AppTarget) DependsOn() []string { return nil }

func (t *AppTarget) RunAfter() []string { return []string{"compile"} }

func (t *AppTarget) Desc() string {
	return "The NodeJS app target uses the contents of your package.json:scripts.start as the main command to start your application inside the container. If your pack supports the 'compile' target, the artifacts from there are first copied to the /srv/app directory inside the container. Otherwise `npm install --production` will be called inside the container (you can customise this by providing a special `installProduction` script inside your package.json)."
}

func (t *AppTarget) Check() error {
	if len(t.Pack.PackageJSON.Scripts.Start) == 0 {
		return fmt.Errorf("package.json does not specify a start script")
	}
	return nil
}

func (t *AppTarget) Dockerfile() *docker.Dockerfile {
	np := t.Pack.PackageJSON
	df := t.Pack.baseDockerfile(np.Version)
	if np.Scripts.InstallProduction != "" {
		df.AddRun(np.Scripts.InstallProduction)
	} else {
		df.AddRun("npm install --production")
	}
	// Pick out the contents of NPM start to invoke directly (using npm start in
	// production shields the app from signals, which are required to be handled by
	// the app itself to do graceful shutdown.
	df.CMD = tools.Whitespace.Split(np.Scripts.Start, -1)
	return df
}
