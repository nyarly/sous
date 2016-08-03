package sous

import (
	"testing"

	"github.com/samsalisbury/semv"
	"github.com/stretchr/testify/assert"
)

func TestInheritingFromGlobal(t *testing.T) {
	assert := assert.New(t)

	m := &Manifest{
		Source: SourceLocation{
			RepoURL:    RepoURL("github.com/opentable/sms-continual-test"),
			RepoOffset: RepoOffset(""),
		},
		Owners: []string{
			"Connect Services",
			"wwade@opentable.com",
		},
		Kind: "http-service",
		Deployments: DeploySpecs{
			"Global": DeploySpec{
				DeployConfig: DeployConfig{
					NumInstances: 1,
					Env:          map[string]string{"SCT_FORGET_ME_NOT_URL": "srvc://forgetmenot/v1"},
					Resources: map[string]string{
						"cpus":   "0.1",
						"memory": "256",
						"ports":  "1",
					},
				},
				Version: semv.MustParse(`0.1.5`),
			},
			"sf-qa-ci": DeploySpec{
				clusterName: "sf-qa-ci",
				DeployConfig: DeployConfig{
					Env: map[string]string{"SCT_ENV": "pp"},
				},
			},
			"prod-sc": DeploySpec{
				clusterName: "prod-sc",
				DeployConfig: DeployConfig{
					Env: map[string]string{"SCT_ENV": "cc"},
				},
			},
		},
	}

	s := &State{
		Defs: Defs{
			Clusters: Clusters{
				"sf-qa-ci": Cluster{
					Kind:    "singularity",
					BaseURL: "http://singularity-qa-sf.otenv.com",
					Env:     EnvDefaults{"OT_DISCO_INIT_URL": "discovery-ci-sf.otenv.com"},
				},
				"prod-sc": Cluster{
					Kind:    "singularity",
					BaseURL: "http://singularity-prod-sc.otenv.com",
					Env:     EnvDefaults{"OT_DISCO_INIT_URL": "discovery-prod-sc.otenv.com"},
				},
			},
		},
	}
	/*
		EnvVars:
		  - Name: OT_DISCO_INIT_URL
		    Type: url
		    Desc: The VIP URL which discovery clients should use to connect to the discovery system.
		  - Name: PORT0
		    Type: int
		    Desc: The port provided to apps

		Resources:
		  - Name: memory
		    Type: Float
		  - Name: cpu
		    Type: Float
		  - Name: ports
		    Type: Integer
		Clusters:
	*/

	deps, err := s.DeploymentsFromManifest(m)
	assert.NoError(err)
	assert.Len(deps, 2)
	qa := deps[0]
	assert.Equal(qa.NumInstances, 1)
	assert.Equal(qa.SourceID.Version.String(), `0.1.5`)
}
