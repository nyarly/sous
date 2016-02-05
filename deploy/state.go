package deploy

type State struct {
	Config
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
