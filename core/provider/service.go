package provider

import (
	"context"

	"github.com/odpf/entropy/pkg/errors"
)

type Service struct {
	repo Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repo: repository,
	}
}

func (s *Service) CreateProvider(ctx context.Context, pro Provider) (*Provider, error) {
	if err := pro.Validate(); err != nil {
		return nil, err
	}

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
