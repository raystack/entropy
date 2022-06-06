package core

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

type Service struct {
	logger     *zap.Logger
	clock      func() time.Time
	store      resource.Store
	rootModule module.Module
}

func New(repo resource.Store, rootModule module.Module, clockFn func() time.Time, lg *zap.Logger) *Service {
	if clockFn == nil {
		clockFn = time.Now
	}
	return &Service{
		logger:     lg,
		clock:      clockFn,
		store:      repo,
		rootModule: rootModule,
	}
}

func (s *Service) generateModuleSpec(ctx context.Context, res resource.Resource) (*module.Spec, error) {
	modSpec := module.Spec{
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
		}

		modSpec.Dependencies[key] = module.ResolvedDependency{
			Kind:   d.Kind,
			Output: d.State.Output,
		}
	}

	return &modSpec, nil
}
