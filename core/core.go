package core

//go:generate mockery --name=AsyncWorker -r --case underscore --with-expecter --structname AsyncWorker  --filename=async_worker.go --output=./mocks

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/pkg/worker"
)

type Service struct {
	logger     *zap.Logger
	clock      func() time.Time
	store      resource.Store
	worker     AsyncWorker
	rootModule module.Driver
}

type AsyncWorker interface {
	Enqueue(ctx context.Context, jobs ...worker.Job) error
}

func New(repo resource.Store, rootModule module.Driver, asyncWorker AsyncWorker, clockFn func() time.Time, lg *zap.Logger) *Service {
	if clockFn == nil {
		clockFn = time.Now
	}

	return &Service{
		logger:     lg,
		clock:      clockFn,
		store:      repo,
		worker:     asyncWorker,
		rootModule: rootModule,
	}
}

func (s *Service) generateModuleSpec(ctx context.Context, res resource.Resource) (*module.ExpandedResource, error) {
	modSpec := module.ExpandedResource{
		Resource:     res,
		Dependencies: map[string]module.ResolvedDependency{},
	}

	for key, resURN := range res.Spec.Dependencies {
		d, err := s.GetResource(ctx, resURN)
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
