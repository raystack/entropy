package core

//go:generate mockery --name=ModuleService -r --case underscore --with-expecter --structname ModuleService  --filename=module_service.go --output=./mocks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
)

type Service struct {
	clock          func() time.Time
	store          resource.Store
	moduleSvc      ModuleService
	syncBackoff    time.Duration
	maxSyncRetries int
}

type ModuleService interface {
	PlanAction(ctx context.Context, res module.ExpandedResource, act module.ActionRequest) (*resource.Resource, error)
	SyncState(ctx context.Context, res module.ExpandedResource) (*resource.State, error)
	StreamLogs(ctx context.Context, res module.ExpandedResource, filter map[string]string) (<-chan module.LogChunk, error)
	GetOutput(ctx context.Context, res module.ExpandedResource) (json.RawMessage, error)
}

func New(repo resource.Store, moduleSvc ModuleService, clockFn func() time.Time) *Service {
	const (
		defaultMaxRetries  = 10
		defaultSyncBackoff = 5 * time.Second
	)

	if clockFn == nil {
		clockFn = time.Now
	}

	return &Service{
		clock:          clockFn,
		store:          repo,
		syncBackoff:    defaultSyncBackoff,
		maxSyncRetries: defaultMaxRetries,
		moduleSvc:      moduleSvc,
	}
}

func (svc *Service) generateModuleSpec(ctx context.Context, res resource.Resource) (*module.ExpandedResource, error) {
	modSpec := module.ExpandedResource{
		Resource:     res,
		Dependencies: map[string]module.ResolvedDependency{},
	}

	for key, resURN := range res.Spec.Dependencies {
		d, err := svc.GetResource(ctx, resURN)
		if err != nil {
			if errors.Is(err, errors.ErrNotFound) {
				return nil, errors.ErrInvalid.
					WithMsgf("dependency '%s' not found", resURN)
			}
			return nil, err
		} else if d.State.Status != resource.StatusCompleted {
			return nil, errors.ErrInvalid.
				WithMsgf("dependency '%s' is in incomplete state (%s)", resURN, d.State.Status)
		} else if d.Project != res.Project {
			return nil, errors.ErrInvalid.
				WithMsgf("dependency '%s' not found", resURN).
				WithCausef("cross-project references not allowed")
		}

		modSpec.Dependencies[key] = module.ResolvedDependency{
			Kind:   d.Kind,
			Output: d.State.Output,
		}
	}

	return &modSpec, nil
}
