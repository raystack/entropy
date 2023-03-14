package module

//go:generate mockery --name=Driver -r --case underscore --with-expecter --structname ModuleDriver --filename=driver.go --output=../mocks
//go:generate mockery --name=Loggable -r --case underscore --with-expecter --structname LoggableModule --filename=loggable_module.go --output=../mocks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goto/entropy/core/resource"
)

// Driver is responsible for achieving desired external system states based
// on a resource in Entropy.
type Driver interface {
	// Plan SHOULD validate the action on the current version of the resource,
	// return the resource with config/status/state changes (if any) applied.
	// Plan SHOULD NOT have side effects on anything other than the resource.
	Plan(ctx context.Context, res ExpandedResource, act ActionRequest) (*Plan, error)

	// Sync is called repeatedly by Entropy core until the returned state is
	// a terminal status. Driver implementation is free to execute an action
	// in a single Sync() call or split into steps for better feedback to the
	// end-user about the progress.
	// Sync can return state in resource.StatusDeleted to indicate resource
	// should be removed from the Entropy storage.
	Sync(ctx context.Context, res ExpandedResource) (*resource.State, error)

	// Output returns the current external state of the resource
	// Output should not have any side effects on any external resource
	Output(ctx context.Context, res ExpandedResource) (json.RawMessage, error)
}

// Plan represents the changes to be staged and later synced by module.
type Plan struct {
	Resource      resource.Resource
	ScheduleRunAt time.Time
	Reason        string
}

// Loggable extension of driver allows streaming log data for a resource.
type Loggable interface {
	Driver

	Log(ctx context.Context, res ExpandedResource, filter map[string]string) (<-chan LogChunk, error)
}

// ExpandedResource represents the context for Plan() or Sync() invocations.
type ExpandedResource struct {
	resource.Resource `json:"resource"`

	Dependencies map[string]ResolvedDependency `json:"dependencies"`
}

type ResolvedDependency struct {
	Kind   string          `json:"kind"`
	Output json.RawMessage `json:"output"`
}

type LogChunk struct {
	Data   []byte            `json:"data"`
	Labels map[string]string `json:"labels"`
}
