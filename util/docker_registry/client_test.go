package docker_registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistries(t *testing.T) {
	assert := assert.New(t)

	rs := NewRegistries()
	r := &registry{}
	assert.NoError(rs.AddRegistry("x", r))
	assert.Equal(rs.GetRegistry("x"), r)
	assert.NoError(rs.DeleteRegistry("x"))
	assert.Nil(rs.GetRegistry("x"))
}
