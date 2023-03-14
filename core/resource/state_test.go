package resource_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goto/entropy/core/resource"
)

func TestState_IsTerminal(t *testing.T) {
	t.Parallel()
	state := resource.State{Status: resource.StatusPending}
	assert.False(t, state.IsTerminal())

	state = resource.State{Status: resource.StatusCompleted}
	assert.True(t, state.IsTerminal())
}

func TestState_InDeletion(t *testing.T) {
	t.Parallel()
	state := resource.State{Status: resource.StatusPending}
	assert.False(t, state.InDeletion())

	state = resource.State{Status: resource.StatusDeleted}
	assert.True(t, state.InDeletion())
}

func TestState_Clone(t *testing.T) {
	t.Parallel()
	originalState := resource.State{
		Status:     resource.StatusPending,
		Output:     []byte(`{"foo": "bar"}`),
		ModuleData: []byte(`{"msg": "Hello!"}`),
	}
	clonedState := originalState.Clone()

	assert.EqualValues(t, originalState, clonedState)

	// mutation should not reflect back in original state.
	clonedState.Output[0] = '['
	assert.Equal(t, string(clonedState.Output), `["foo": "bar"}`)
	assert.Equal(t, string(originalState.Output), `{"foo": "bar"}`)

	clonedState.ModuleData[0] = '#'
	assert.Equal(t, string(originalState.ModuleData), `{"msg": "Hello!"}`)
	assert.Equal(t, string(clonedState.ModuleData), `#"msg": "Hello!"}`)
}
