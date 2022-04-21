package provider

import (
	"context"
	"time"

	"github.com/odpf/entropy/pkg/errors"
)

func NewService(repository Repository, clock func() time.Time) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{
		clock: clock,
		repo:  repository,
	}
}

type Service struct {
	clock func() time.Time
	repo  Repository
}

func (s *Service) CreateProvider(ctx context.Context, pro Provider) (*Provider, error) {
	if err := pro.Validate(); err != nil {
		return nil, err
	}
	pro.CreatedAt = s.clock()
	pro.UpdatedAt = pro.CreatedAt

	if err := s.repo.Create(ctx, pro); err != nil {
		if errors.Is(err, errors.ErrConflict) {
			return nil, errors.ErrConflict.WithMsgf("provider with urn '%s' already exists", pro.URN)
		}
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	return &pro, nil
}

func (s *Service) ListProviders(ctx context.Context, parent string, kind string) ([]*Provider, error) {
	filter := map[string]string{}
	if kind != "" {
		filter["kind"] = kind
	}
	if parent != "" {
		filter["parent"] = parent
	}

	providers, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	return providers, nil
}
