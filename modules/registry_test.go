package modules_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goto/entropy/core/mocks"
	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/modules"
	"github.com/goto/entropy/pkg/errors"
)

func TestRegistry_GetDriver(t *testing.T) {
	t.Parallel()

	reg := &modules.Registry{}
	require.NoError(t, reg.Register(module.Descriptor{
		Kind: "foo",
		DriverFactory: func(_ json.RawMessage) (module.Driver, error) {
			return &mocks.ModuleDriver{}, nil
		},
	}))
	require.NoError(t, reg.Register(module.Descriptor{
		Kind: "error_generating_kind",
		DriverFactory: func(_ json.RawMessage) (module.Driver, error) {
			return nil, errors.ErrInvalid
		},
	}))

	t.Run("UnknownKind", func(t *testing.T) {
		driver, _, err := reg.GetDriver(context.Background(), module.Module{
			URN:     "orn:entropy:module:prj:unknown_kind",
			Name:    "unknown_kind",
			Project: "prj",
			Configs: nil,
		})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.ErrNotFound))
		assert.Nil(t, driver)
	})

	t.Run("KnownKind_DriverFactory_Error", func(t *testing.T) {
		driver, _, err := reg.GetDriver(context.Background(), module.Module{
			URN:     "orn:entropy:module:prj:error_generating_kind",
			Name:    "error_generating_kind",
			Project: "prj",
			Configs: nil,
		})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.ErrInvalid))
		assert.Nil(t, driver)
	})

	t.Run("KnownKind_Success", func(t *testing.T) {
		driver, _, err := reg.GetDriver(context.Background(), module.Module{
			URN:     "orn:entropy:module:prj:foo",
			Name:    "foo",
			Project: "prj",
			Configs: nil,
		})
		assert.NoError(t, err)
		assert.NotNil(t, driver)
	})
}

func TestRegistry_Register(t *testing.T) {
	t.Parallel()
	reg := &modules.Registry{}

	t.Run("FirstRegistration_NoError", func(t *testing.T) {
		t.Parallel()
		desc := module.Descriptor{
			Kind: "foo",
			DriverFactory: func(conf json.RawMessage) (module.Driver, error) {
				return &mocks.ModuleDriver{}, nil
			},
			Actions: []module.ActionDesc{
				{
					Name:        "get_stuff_done",
					Description: "This action gets stuff done",
					ParamSchema: `{}`,
				},
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

	t.Run("NewRegistration_InvalidParameterSchema", func(t *testing.T) {
		t.Parallel()
		desc := module.Descriptor{
			Kind: "Nova",
			DriverFactory: func(conf json.RawMessage) (module.Driver, error) {
				return &mocks.ModuleDriver{}, nil
			},
			Actions: []module.ActionDesc{
				{
					Name:        "get_stuff_done",
					Description: "This action gets stuff done",
					ParamSchema: `this is not valid json schema`,
				},
			},
		}
		got := reg.Register(desc)
		assert.Error(t, got)
		assert.True(t, errors.Is(got, errors.ErrInvalid), cmp.Diff(got, errors.ErrInvalid))
	})
}
