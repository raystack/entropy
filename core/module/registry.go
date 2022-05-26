package module

import (
	"context"
	"reflect"
	"sync"

	"github.com/xeipuuv/gojsonschema"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

// Registry maintains a list of supported/enabled modules.
type Registry struct {
	mu         sync.RWMutex
	collection map[string]Descriptor
}

func NewRegistry() *Registry {
	return &Registry{
		collection: map[string]Descriptor{},
	}
}

// Register adds a module to the registry.
func (mr *Registry) Register(desc Descriptor) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	if v, exists := mr.collection[desc.Kind]; exists {
		return errors.ErrConflict.
			WithMsgf("module '%s' is already registered for kind '%s'", reflect.TypeOf(v), desc.Kind)
	}

	for i, action := range desc.Actions {
		if action.ParamSchema == "" {
			continue
		}

		loader := gojsonschema.NewStringLoader(action.ParamSchema)

		schema, err := gojsonschema.NewSchema(loader)
		if err != nil {
			return errors.ErrInvalid.
				WithMsgf("parameter schema for action '%s' is not valid", action.Name).
				WithCausef(err.Error())
		}
		desc.Actions[i].schema = schema
	}

	mr.collection[desc.Kind] = desc
	return nil
}

func (mr *Registry) get(kind string) (Descriptor, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	desc, found := mr.collection[kind]
	return desc, found
}

func (mr *Registry) Plan(ctx context.Context, spec Spec, act ActionRequest) (*resource.Resource, error) {
	kind := spec.Resource.Kind

	desc, found := mr.get(kind)
	if !found {
		return nil, errors.ErrInvalid.WithMsgf("kind '%s' is not valid", kind)
	} else if err := desc.validateDependencies(spec.Dependencies); err != nil {
		return nil, err
	} else if err := desc.validateActionReq(spec, act); err != nil {
		return nil, err
	}

	return desc.Module.Plan(ctx, spec, act)
}

func (mr *Registry) Sync(ctx context.Context, spec Spec) (*resource.State, error) {
	kind := spec.Resource.Kind

	desc, found := mr.get(kind)
	if !found {
		return nil, errors.ErrInvalid.WithMsgf("kind '%s' is not valid", kind)
	} else if err := desc.validateDependencies(spec.Dependencies); err != nil {
		return nil, err
	}

	return desc.Module.Sync(ctx, spec)
}

func (mr *Registry) Log(ctx context.Context, spec Spec, filter map[string]string) (<-chan LogChunk, error) {
	kind := spec.Resource.Kind

	desc, found := mr.get(kind)
	if !found {
		return nil, errors.ErrInvalid.WithMsgf("kind '%s' is not valid", kind)
	}

	lg, supported := desc.Module.(Loggable)
	if !supported {
		return nil, errors.ErrUnsupported.WithMsgf("log streaming not supported for kind '%s'", kind)
	}

	return lg.Log(ctx, spec, filter)
}
