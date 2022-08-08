package module_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/odpf/entropy/core/mocks"
	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/pkg/errors"
)

func TestRegistry_Register(t *testing.T) {
	t.Parallel()
	reg := module.NewRegistry(nil)

	t.Run("FirstRegistration_NoError", func(t *testing.T) {
		t.Parallel()
		desc := module.Descriptor{
			Kind: "foo",
			DriverFactory: func(conf json.RawMessage) (module.Driver, error) {
				return &mocks.ModuleDriver{}, nil
			},
		}
		assert.NoError(t, reg.Register(desc))
	})

	t.Run("SecondRegistration_Conflict", func(t *testing.T) {
		t.Parallel()
		desc := module.Descriptor{
			Kind: "foo",
			DriverFactory: func(conf json.RawMessage) (module.Driver, error) {
				return &mocks.ModuleDriver{}, nil
			},
		}

		err := reg.Register(desc)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.ErrConflict))
	})
}
