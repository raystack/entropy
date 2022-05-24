package module_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/odpf/entropy/core/mocks"
	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/pkg/errors"
)

func TestRegistry_Register(t *testing.T) {
	t.Parallel()
	reg := module.NewRegistry()

	t.Run("FirstRegistration_NoError", func(t *testing.T) {
		t.Parallel()
		desc := module.Descriptor{
			Kind:   "foo",
			Module: &mocks.Module{},
		}
		assert.NoError(t, reg.Register(desc))
	})

	t.Run("SecondRegistration_Conflict", func(t *testing.T) {
		t.Parallel()
		desc := module.Descriptor{
			Kind:   "foo",
			Module: &mocks.Module{},
		}

		err := reg.Register(desc)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.ErrConflict))
	})
}
