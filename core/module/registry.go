package module

import (
	"reflect"

	"github.com/odpf/entropy/pkg/errors"
)

func NewRegistry() *Registry {
	return &Registry{
		collection: map[string]Module{},
	}
}

// Registry maintains a list of supported/enabled modules.
type Registry struct {
	collection map[string]Module
}

// Register adds a module to the registry.
func (mr *Registry) Register(m Module) error {
	desc := m.Describe()

	if v, exists := mr.collection[desc.Kind]; exists {
		return errors.ErrConflict.
			WithMsgf("module '%s' is already registered for kind '%s'", reflect.TypeOf(v), desc.Kind)
	}

	mr.collection[desc.Kind] = m
	return nil
}

// Resolve resolves the module for the given kind.
func (mr *Registry) Resolve(kind string) (Module, error) {
	if m, exists := mr.collection[kind]; exists {
		return m, nil
	}
	return nil, errors.ErrNotFound
}
