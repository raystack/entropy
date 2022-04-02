package module

//go:generate mockery --name=ServiceInterface -r --case underscore --with-expecter --structname ModuleService  --filename=module_service.go --output=../../mocks

import (
	"context"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/store"
)

type ServiceInterface interface {
	Sync(ctx context.Context, r *domain.Resource) (*domain.Resource, error)
	Validate(ctx context.Context, r *domain.Resource) error
	Act(ctx context.Context, r *domain.Resource, action string, params map[string]interface{}) (map[string]interface{}, error)
	Log(ctx context.Context, r *domain.Resource, filter map[string]string) (chan domain.LogChunk, error)
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

func (s *Service) Validate(ctx context.Context, r *domain.Resource) error {
	module, err := s.moduleRepository.Get(r.Kind)
	if err != nil {
		return err
	}
	err = module.Validate(r)
	return err
}

func (s *Service) Act(ctx context.Context, r *domain.Resource, action string, params map[string]interface{}) (map[string]interface{}, error) {
	module, err := s.moduleRepository.Get(r.Kind)
	if err != nil {
		return nil, err
	}
	output, err := module.Act(r, action, params)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *Service) Log(ctx context.Context, r *domain.Resource, filter map[string]string) (chan domain.LogChunk, error) {
	module, err := s.moduleRepository.Get(r.Kind)
	if err != nil {
		return nil, err
	}
	logOutput, err := module.Log(ctx, r, filter)
	if err != nil {
		return nil, err
	}
	return logOutput, nil
}
