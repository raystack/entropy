package core

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func New(repo resource.Repository, rootModule module.Module, clockFn func() time.Time, lg *zap.Logger) *Service {
	if clockFn == nil {
		clockFn = time.Now
	}
	return &Service{
		clock:      clockFn,
		repository: repo,
		rootModule: rootModule,
	}
}

type Service struct {
	logger     *zap.Logger
	clock      func() time.Time
	repository resource.Repository
	rootModule module.Module
}

func (s *Service) generateModuleSpec(ctx context.Context, res resource.Resource) (*module.Spec, error) {
	modSpec := module.Spec{
		Resource:     res,
		Dependencies: map[string]resource.Output{},
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

		modSpec.Dependencies[key] = d.State.Output
	}

	return &modSpec, nil
}
