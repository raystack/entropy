package module

//go:generate mockery --name=Store -r --case underscore --with-expecter --structname ModuleStore --filename=module_store.go --output=../mocks
//go:generate mockery --name=Registry -r --case underscore --with-expecter --structname ModuleRegistry --filename=module_registry.go --output=../mocks

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/goto/entropy/pkg/errors"
)

// Module represents all the data needed to initialize a particular module.
type Module struct {
	URN       string          `json:"urn"`
	Name      string          `json:"name"`
	Project   string          `json:"project"`
	Configs   json.RawMessage `json:"configs"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Descriptor is a module descriptor that represents supported actions, resource-kind
// the module can operate on, etc.
type Descriptor struct {
	Kind          string                                     `json:"kind"`
	Actions       []ActionDesc                               `json:"actions"`
	Dependencies  map[string]string                          `json:"dependencies"`
	DriverFactory func(conf json.RawMessage) (Driver, error) `json:"-"`
}

// Registry is responsible for installing and managing module-drivers as per
// module definitions provided.
type Registry interface {
	GetDriver(ctx context.Context, mod Module) (Driver, Descriptor, error)
}

// Store is responsible for persisting modules defined for each project.
type Store interface {
	GetModule(ctx context.Context, urn string) (*Module, error)
	ListModules(ctx context.Context, project string) ([]Module, error)
	CreateModule(ctx context.Context, m Module) error
	UpdateModule(ctx context.Context, m Module) error
	DeleteModule(ctx context.Context, urn string) error
}

func (mod *Module) sanitise(isCreate bool) error {
	if mod.Name == "" {
		return errors.ErrInvalid.WithMsgf("name must be set")
	}

	if mod.Project == "" {
		return errors.ErrInvalid.WithMsgf("project must be set")
	}

	if isCreate {
		mod.URN = generateURN(mod.Name, mod.Project)
		mod.CreatedAt = time.Now()
		mod.UpdatedAt = mod.CreatedAt
	}
	mod.URN = strings.TrimSpace(mod.URN)
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

func (desc Descriptor) validateActionReq(res ExpandedResource, req ActionRequest) error {
	kind := res.Resource.Kind

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
