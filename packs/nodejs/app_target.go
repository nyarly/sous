package nodejs

import (
	"fmt"
	"path"

	"github.com/opentable/sous/core"
	"github.com/opentable/sous/tools"
	"github.com/opentable/sous/tools/cli"
	"github.com/opentable/sous/tools/docker"
	"github.com/opentable/sous/tools/file"
)

type AppTarget struct {
	*NodeJSTarget
	artifactPath string
}

func NewAppTarget(pack *Pack) *AppTarget {
	return &AppTarget{NewNodeJSTarget("app", pack), ""}
}

func (t *AppTarget) DependsOn() []core.Target {
	return []core.Target{
		NewCompileTarget(t.Pack),
	}
}

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

func (t *AppTarget) PreDockerBuild(c *core.Context) {
	if t.artifactPath == "" {
		cli.Fatalf("Artifact path not set by compile target.")
	}
	if !file.Exists(t.artifactPath) {
		cli.Fatalf("Artifact not at %s", t.artifactPath)
	}
	filename := path.Base(t.artifactPath)
	localArtifact := filename
	file.TemporaryLink(t.artifactPath, localArtifact)
	t.artifactPath = localArtifact
}

func (t *AppTarget) Dockerfile() *docker.Dockerfile {
	if t.artifactPath == "" {
		// Actually, it is first set by compile target, then the PreDockerBuild
		// step links it into the WD and resets artifactPath to a local, relative
		// path.
		t.artifactPath = "<¡ artifact path set by compile target !>"
	}
	np := t.Pack.PackageJSON
	df := t.Pack.baseDockerfile(np.Version)
	// Since the artifact is tar.gz, docker automatically unpacks it.
	df.Add = []docker.Add{docker.Add{Files: []string{t.artifactPath}, Dest: "/srv/app/"}}
	// Pick out the contents of NPM start to invoke directly (using npm start in
	// production shields the app from signals, which are required to be handled by
	// the app itself to do graceful shutdown.
	df.CMD = tools.Whitespace.Split(np.Scripts.Start, -1)
	return df
}

func (t *AppTarget) SetState(fromTarget string, state interface{}) {
	if fromTarget != "compile" {
		return
	}
	m, ok := state.(map[string]string)
	if !ok {
		cli.Fatalf("app target got a %T from compile target, expected map[string]string", state)
	}
	artifactPath, ok := m["artifactPath"]
	if !ok {
		cli.Fatalf("app target got %+v from compile target; expected key 'artifactPath'", m)
	}
	t.artifactPath = artifactPath
}
