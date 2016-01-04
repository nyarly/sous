package deploy

import "github.com/opentable/sous/config"

type State struct {
	config.Config
	EnvironmentDefs EnvDefs
	Datacentres     Datacentres
	Manifests       Manifests
	Contracts       Contracts
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
	Name, Desc         string
	SingularityURL     string
	DockerRegistryHost string
	Env                DatacentreEnv
}

type DatacentreEnv map[string]string

type Manifests map[string]Manifest

type Manifest struct {
	App         App
	Deployments Deployments
}

type App struct {
	SourceRepo, Owner, Kind string
}

type Deployments map[string]Deployment

type Deployment struct {
	Instance                  Instance
	SourceTag, SourceRevision string
	Environment               map[string]string
}

type Instance struct {
	Count  int
	CPUs   float32
	Memory string
}

type MemorySize string

type Contracts map[string]Contract

type Contract struct {
	Name, Desc    string
	StartServers  []string
	Servers       map[string]TestServer
	Preconditions []Check
	Checks        []Check
}

type TestServer struct {
	Name, Desc    string
	DefaultValues map[string]string
	Export        []string
	Docker        DockerServer
}

type DockerServer struct {
	Image         string
	Env           map[string]string
	Options, Args []string
}

type GetHTTPAssertion struct {
	URL, ResponseBodyContains, ResponseJSONContains string
	ResponseStatusCode                              int
	AnyResponse                                     bool
}

type Check struct {
	Name, Desc, GET, BodyContains string
	StatusCode                    int
	StatusCodeRange               []int
	BodyContainsJSON              interface{}
}
