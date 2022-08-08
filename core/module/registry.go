package module

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/xeipuuv/gojsonschema"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

// Registry maintains a list of supported/enabled modules.
type Registry struct {
	mu          sync.RWMutex
	store       Store
	descriptors map[string]Descriptor
}

func NewRegistry(store Store) *Registry {
	return &Registry{
		store:       store,
		descriptors: map[string]Descriptor{},
	}
}

// Register adds a module to the registry.
func (mr *Registry) Register(desc Descriptor) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	if v, exists := mr.descriptors[desc.Kind]; exists {
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

	mr.descriptors[desc.Kind] = desc
	return nil
}

func (mr *Registry) Plan(ctx context.Context, spec ExpandedResource, act ActionRequest) (*Plan, error) {
	kind := spec.Resource.Kind

	driver, desc, err := mr.initDriver(ctx, kind, spec.Project)
	if err != nil {
		return nil, err
	} else if err := desc.validateDependencies(spec.Dependencies); err != nil {
		return nil, err
	} else if err := desc.validateActionReq(spec, act); err != nil {
		return nil, err
	}

	return driver.Plan(ctx, spec, act)
}

func (mr *Registry) Sync(ctx context.Context, spec ExpandedResource) (*resource.State, error) {
	kind := spec.Resource.Kind

	driver, desc, err := mr.initDriver(ctx, kind, spec.Project)
	if err != nil {
		return nil, err
	} else if err := desc.validateDependencies(spec.Dependencies); err != nil {
		return nil, err
	}

	return driver.Sync(ctx, spec)
}

func (mr *Registry) Log(ctx context.Context, spec ExpandedResource, filter map[string]string) (<-chan LogChunk, error) {
	kind := spec.Resource.Kind

	driver, _, err := mr.initDriver(ctx, kind, spec.Project)
	if err != nil {
		return nil, err
	}

	lg, supported := driver.(Loggable)
	if !supported {
		return nil, errors.ErrUnsupported.WithMsgf("log streaming not supported for kind '%s'", kind)
	}

	return lg.Log(ctx, spec, filter)
}

func (mr *Registry) initDriver(ctx context.Context, kind, project string) (Driver, Descriptor, error) {
	urn := generateURN(kind, project)
	m, err := mr.store.GetModule(ctx, urn)
	if err != nil {
		return nil, Descriptor{}, err
	}

	mr.mu.RLock()
	defer mr.mu.RUnlock()
	desc, found := mr.descriptors[m.Name]
	if !found {
		return nil, Descriptor{}, errors.ErrInvalid.WithMsgf("kind '%s' is not valid", kind)
	}

	driver, err := desc.DriverFactory(m.Spec.Configs)
	if err != nil {
		return nil, Descriptor{}, errors.ErrInternal.WithCausef(err.Error())
	}
	return driver, desc, nil
}

func generateURN(name, project string) string {
	if strings.HasPrefix(name, "orn:entropy:module") {
		return name
	}
	return fmt.Sprintf("orn:entropy:module:%s:%s", project, name)
}
