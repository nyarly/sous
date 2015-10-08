package config

type Config struct {
	DockerRegistry string
	Packs          *Packs
}

type Packs struct {
	NodeJS *NodeJSConfig
}

type NodeJSConfig struct {
	NPMMirrorURL                   string
	NodeVersionsToDockerBaseImages map[string]string
	AvailableNPMVersions           []string
}
