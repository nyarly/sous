package sous

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterMap(t *testing.T) {
	assert := assert.New(t)

	s := State{
		Defs: Defs{
			Clusters: Clusters{
				"one": Cluster{},
				"two": Cluster{},
			},
		},
	}

	m := s.ClusterMap()
	assert.Len(m, 2)
	assert.Contains(m, "one")
	assert.Contains(m, "two")

	s.singleCluster = "one"
	m = s.ClusterMap()
	assert.Len(m, 1)
	assert.Contains(m, "one")

}
