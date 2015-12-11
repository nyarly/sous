package deploy

import (
	"fmt"
	"testing"

	"github.com/opentable/sous/tools/yaml"
)

func TestMergeValidConfig(t *testing.T) {
	stateDir := "_testdata/valid_config"
	state, err := Parse(stateDir)
	assertErrNil(t, err)
	assertErrNil(t, state.Validate())
	merged, err := state.Merge()
	assertErrNil(t, err)
	y, err := yaml.Marshal(merged)
	assertErrNil(t, err)
	fmt.Println(string(y))
}

func assertErrNil(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Got err=\n\t%s\n want nil", err)
	}
}
