package module_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/odpf/entropy/core/mocks"
	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/pkg/errors"
)

func TestRegistry_Register(t *testing.T) {
	mod := &mocks.Module{}
	mod.EXPECT().
		Describe().
		Return(module.Desc{Kind: "foo"}).
		Twice()

	reg := module.NewRegistry()

	t.Run("FirstRegistration_NoError", func(t *testing.T) {
		assert.NoError(t, reg.Register(mod))
	})

	t.Run("SecondRegistration_Conflict", func(t *testing.T) {
		err := reg.Register(mod)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.ErrConflict))
	})

	mod.AssertExpectations(t)
}

func TestRegistry_Resolve(t *testing.T) {
	const knownKind = "foo"

	mod := &mocks.Module{}
	mod.EXPECT().
		Describe().
		Return(module.Desc{Kind: knownKind}).
		Twice()

	reg := module.NewRegistry()
	require.NoError(t, reg.Register(mod))

	t.Run("UnknownKind", func(t *testing.T) {
		resolvedMod, err := reg.Resolve("non-existent-kind")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, errors.ErrNotFound))
		assert.Nil(t, resolvedMod)
	})

	t.Run("RegisteredKind", func(t *testing.T) {
		resolvedMod, err := reg.Resolve(knownKind)
		assert.NoError(t, err)
		assert.NotNil(t, resolvedMod)
	})

}
