package module

//go:generate mockery --name=Store -r --case underscore --with-expecter --structname ModuleStore --filename=module_store.go --output=../mocks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/odpf/entropy/pkg/errors"
)

type Store interface {
	GetModule(ctx context.Context, urn string) (*Module, error)
	ListModules(ctx context.Context, project string) ([]Module, error)
	CreateModule(ctx context.Context, m Module) error
	UpdateModule(ctx context.Context, m Module) error
	DeleteModule(ctx context.Context, urn string) error
}

// Module represents all the data needed to initialize a particular module.
type Module struct {
	URN       string    `json:"urn"`
	Name      string    `json:"name"`
	Project   string    `json:"project"`
	Spec      Spec      `json:"spec"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Spec struct {
	Configs json.RawMessage `json:"configs"`
}

// Descriptor is a module descriptor that represents supported actions, resource-kind
// the module can operate on, etc.
type Descriptor struct {
	Kind          string                                     `json:"kind"`
	Actions       []ActionDesc                               `json:"actions"`
	Dependencies  map[string]string                          `json:"dependencies"`
	DriverFactory func(conf json.RawMessage) (Driver, error) `json:"-"`
}

func (Module) Validate() error {
	return nil
}

func (desc Descriptor) validateDependencies(dependencies map[string]ResolvedDependency) error {
	for key, wantKind := range desc.Dependencies {
		resolvedDep, found := dependencies[key]
		if !found {
			return errors.ErrInvalid.
				WithMsgf("kind '%s' needs resource of kind '%s' at key '%s'", desc.Kind, wantKind, key)
		} else if wantKind != resolvedDep.Kind {
			return errors.ErrInvalid.
				WithMsgf("value for '%s' must be of kind '%s', not '%s'", key, wantKind, resolvedDep.Kind)
		}
	}
	return nil
}

func (desc Descriptor) validateActionReq(spec ExpandedResource, req ActionRequest) error {
	kind := spec.Resource.Kind

	actDesc := desc.findAction(req.Name)
	if actDesc == nil {
		return errors.ErrInvalid.WithMsgf("action '%s' is not valid on kind '%s'", req.Name, kind)
	}

	return actDesc.validateReq(req)
}

func (desc Descriptor) findAction(name string) *ActionDesc {
	for _, action := range desc.Actions {
		if action.Name == name {
			return &action
		}
	}
	return nil
}
