package config

type Config struct {
	DockerRegistry    string
	DockerLabelPrefix string
	GlobalDockerTags  map[string]string
	Packs             *Packs
}

type Packs struct {
	NodeJS *NodeJSConfig
}

type NodeJSConfig struct {
	NPMMirrorURL                   string
	NodeVersionsToDockerBaseImages map[string]string
	DockerTags                     map[string]string
	AvailableNPMVersions           []string
}
