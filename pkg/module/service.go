package module

import (
	"context"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/store"
)

type ServiceInterface interface {
	Sync(ctx context.Context, res *domain.Resource) (*domain.Resource, error)
}

type Service struct {
	moduleRepository store.ModuleRepository
}

func NewService(moduleRepository store.ModuleRepository) *Service {
	return &Service{
		moduleRepository: moduleRepository,
	}
}

func (s *Service) Sync(ctx context.Context, r *domain.Resource) (*domain.Resource, error) {
	module, err := s.moduleRepository.Get(r.Kind)
	if err != nil {
		r.Status = domain.ResourceStatusError
		return r, err
	}
	status, err := module.Apply(r)
	r.Status = status
	return r, err
}
