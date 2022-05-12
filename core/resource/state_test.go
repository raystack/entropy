package resource_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/odpf/entropy/core/resource"
)

func TestState_IsTerminal(t *testing.T) {
	state := resource.State{Status: resource.StatusPending}
	assert.False(t, state.IsTerminal())

	state = resource.State{Status: resource.StatusCompleted}
	assert.True(t, state.IsTerminal())
}

func TestState_InDeletion(t *testing.T) {
	state := resource.State{Status: resource.StatusPending}
	assert.False(t, state.InDeletion())

	state = resource.State{Status: resource.StatusDeleted}
	assert.True(t, state.InDeletion())
}

func TestState_Clone(t *testing.T) {
	originalState := resource.State{
		Status: resource.StatusPending,
		Output: map[string]interface{}{
			"foo": "bar",
		},
		ModuleData: []byte("hello world"),
	}
	clonedState := originalState.Clone()

	assert.EqualValues(t, originalState, clonedState)

	// mutation should not reflect back in original state.
	clonedState.Output["foo"] = "modified-value"
	assert.Equal(t, originalState.Output["foo"], "bar")

	clonedState.ModuleData[0] = '#'
	assert.Equal(t, string(originalState.ModuleData), "hello world")
	assert.Equal(t, string(clonedState.ModuleData), "#ello world")
}
