package module

import (
	"context"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/store"
)

type ServiceInterface interface {
	TriggerSync(ctx context.Context, urn string) error
}

type Service struct {
	resourceRepository store.ResourceRepository
	moduleRepository   store.ModuleRepository
}

func NewService(resourceRepository store.ResourceRepository, moduleRepository store.ModuleRepository) *Service {
	return &Service{
		resourceRepository: resourceRepository,
		moduleRepository:   moduleRepository,
	}
}

func (s *Service) TriggerSync(ctx context.Context, urn string) error {
	r, err := s.resourceRepository.GetByURN(urn)
	if err != nil {
		return err
	}

	module, err := s.moduleRepository.Get(r.Kind)
	if err != nil {
		r.Status = domain.ResourceStatusError
		updateErr := s.resourceRepository.Update(r)
		if updateErr != nil {
			return updateErr
		}
		return err
	}

	status, err := module.Apply(r)
	if err != nil {
		r.Status = domain.ResourceStatusError
		updateErr := s.resourceRepository.Update(r)
		if updateErr != nil {
			return updateErr
		}
		return err
	}

	r.Status = status
	err = s.resourceRepository.Update(r)
	if err != nil {
		return err
	}

	return nil
}
