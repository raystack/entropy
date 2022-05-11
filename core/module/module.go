package module

//go:generate mockery --name=Module -r --case underscore --with-expecter --structname Module --filename=module.go --output=../mocks

import (
	"context"

	"github.com/odpf/entropy/core/resource"
)

const (
	CreateAction = "create"
	UpdateAction = "update"
)

// Module is responsible for achieving desired external system states based
// on a resource in Entropy.
type Module interface {
	// Describe should return a descriptor with information about the module.
	// This descriptor will be used for discovery of supported actions, etc.
	Describe() Desc

	// Plan SHOULD validate the action on the current version of the resource,
	// return the resource with config/status/state changes (if any) applied.
	// Plan SHOULD NOT have side effects on anything other than the resource.
	Plan(ctx context.Context, spec Spec, act ActionRequest) (*resource.Resource, error)

	// Sync is called repeatedly by Entropy core until the returned state has
	// StatusCompleted. Module implementation is free to execute an action in
	// a single Sync() call or split into multiple steps for better feedback
	// to the end-user about the progress.
	Sync(ctx context.Context, spec Spec) (*resource.Output, error)
}

// Desc is a module descriptor that represents supported actions, resource-kind
// the module can operate on, etc.
type Desc struct {
	Kind    string       `json:"kind"`
	Actions []ActionDesc `json:"actions"`
}

// ActionDesc is a descriptor for an action supported by a module.
type ActionDesc struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParamSchema string `json:"param_schema"`
}

// ActionRequest describes an invocation of action on module.
type ActionRequest struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

// Spec represents the context for Plan() or Sync() invocations.
type Spec struct {
	Resource     resource.Resource          `json:"resource"`
	Dependencies map[string]resource.Output `json:"dependencies"`
}
