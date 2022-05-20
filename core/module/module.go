package module

//go:generate mockery --name=Module -r --case underscore --with-expecter --structname Module --filename=module.go --output=../mocks

import (
	"context"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

// Module is responsible for achieving desired external system states based
// on a resource in Entropy.
type Module interface {
	// Plan SHOULD validate the action on the current version of the resource,
	// return the resource with config/status/state changes (if any) applied.
	// Plan SHOULD NOT have side effects on anything other than the resource.
	Plan(ctx context.Context, spec Spec, act ActionRequest) (*resource.Resource, error)

	// Sync is called repeatedly by Entropy core until the returned state is
	// a terminal status. Module implementation is free to execute an action
	// in a single Sync() call or split into steps for better feedback to the
	// end-user about the progress.
	// Sync can return state in resource.StatusDeleted to indicate resource
	// should be removed from the Entropy storage.
	Sync(ctx context.Context, spec Spec) (*resource.State, error)
}

// Spec represents the context for Plan() or Sync() invocations.
type Spec struct {
	Resource     resource.Resource             `json:"resource"`
	Dependencies map[string]ResolvedDependency `json:"dependencies"`
}

type ResolvedDependency struct {
	Kind   string          `json:"kind"`
	Output resource.Output `json:"output"`
}

// Descriptor is a module descriptor that represents supported actions, resource-kind
// the module can operate on, etc.
type Descriptor struct {
	Kind         string            `json:"kind"`
	Actions      []ActionDesc      `json:"actions"`
	Dependencies map[string]string `json:"dependencies"`
	Module       Module            `json:"-"`
}

func (desc Descriptor) validateDependencies(dependencies map[string]ResolvedDependency) error {
	for key, resolvedDep := range dependencies {
		wantKind, found := desc.Dependencies[key]
		if !found {
			return errors.ErrInvalid.
				WithMsgf("unwanted dependency '%s' (kind '%s')", key, resolvedDep.Kind)
		} else if wantKind != resolvedDep.Kind {
			return errors.ErrInvalid.
				WithMsgf("value for '%s' must be from kind '%s', not '%s'", key, wantKind, resolvedDep.Kind)
		}
	}
	return nil
}

func (desc Descriptor) validateActionReq(spec Spec, req ActionRequest) error {
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
