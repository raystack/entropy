package module

import (
	"context"
	"reflect"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func NewRegistry() *Registry {
	return &Registry{
		collection: map[string]Descriptor{},
	}
}

// Registry maintains a list of supported/enabled modules.
type Registry struct {
	collection map[string]Descriptor
}

// Register adds a module to the registry.
func (mr *Registry) Register(desc Descriptor) error {
	if v, exists := mr.collection[desc.Kind]; exists {
		return errors.ErrConflict.
			WithMsgf("module '%s' is already registered for kind '%s'", reflect.TypeOf(v), desc.Kind)
	}

	mr.collection[desc.Kind] = desc
	return nil
}

func (mr *Registry) Plan(ctx context.Context, spec Spec, act ActionRequest) (*resource.Resource, error) {
	kind := spec.Resource.Kind

	_, found := mr.collection[kind]
	if !found {
		return nil, errors.ErrInvalid.WithMsgf("kind '%s' is not valid", kind)
	}

	// TODO: perform action-request validation using the descriptor.
	// TODO: dispatch Plan() to specific module.

	return &spec.Resource, nil
}

func (mr *Registry) Sync(ctx context.Context, spec Spec) (*resource.State, error) {
	kind := spec.Resource.Kind

	_, found := mr.collection[kind]
	if !found {
		return nil, errors.ErrInvalid.WithMsgf("kind '%s' is not valid", kind)
	}

	// TODO: perform dependency validation using the descriptor.
	// TODO: dispatch Sync() to specific module.

	return &spec.Resource.State, nil
}

func (mr *Registry) Log(ctx context.Context, spec Spec, filter map[string]string) (<-chan LogChunk, error) {
	kind := spec.Resource.Kind

	desc, found := mr.collection[kind]
	if !found {
		return nil, errors.ErrInvalid.WithMsgf("kind '%s' is not valid", kind)
	}

	lg, supported := desc.Module.(Loggable)
	if !supported {
		return nil, errors.ErrUnsupported.WithMsgf("log streaming not supported for kind '%s'", kind)
	}

	return lg.Log(ctx, spec, filter)
}
