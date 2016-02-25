package nodejs

import (
	"testing"

	"github.com/opentable/sous/core"
)

func TestCompileTarget(t *testing.T) {
	target := interface{}(&CompileTarget{})
	if _, ok := target.(core.ImageIsStaler); !ok {
		t.Errorf("%T does not implement ImageIsStaler", target)
	}
	if _, ok := target.(core.Stater); !ok {
		t.Errorf("%T does not implement Stater", target)
	}
}

func TestAppTarget(t *testing.T) {
	target := interface{}(&AppTarget{})
	if _, ok := target.(core.SetStater); !ok {
		t.Errorf("%T does not implement SetStater", target)
	}
	// TODO: Use this method of static checking instead of tests for
	// checking interface implementation.
	var _ core.ContainerTarget = &AppTarget{}
	if _, ok := target.(core.ContainerTarget); !ok {
		t.Errorf("%T does not implement ContainerTarget", target)
	}
}
