package resource

import (
	"context"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/store"
	"strings"
)

type ServiceInterface interface {
	CreateResource(ctx context.Context, res *domain.Resource) (*domain.Resource, error)
	UpdateResource(ctx context.Context, urn string, configs map[string]interface{}) (*domain.Resource, error)
}

type Service struct {
	resourceRepository store.ResourceRepository
}

func NewService(repository store.ResourceRepository) *Service {
	return &Service{
		resourceRepository: repository,
	}
}

func (s *Service) CreateResource(ctx context.Context, res *domain.Resource) (*domain.Resource, error) {
	res.Urn = generateResourceUrn(res)
	res.Status = "PENDING"
	err := s.resourceRepository.Create(res)
	if err != nil {
		return nil, err
	}
	createdResource, err := s.resourceRepository.GetByURN(res.Urn)
	if err != nil {
		return nil, err
	}
	return createdResource, nil
}

func (s *Service) UpdateResource(ctx context.Context, urn string, configs map[string]interface{}) (*domain.Resource, error) {
	res, err := s.resourceRepository.GetByURN(urn)
	if err != nil {
		return nil, err
	}
	res.Configs = configs
	err = s.resourceRepository.Update(res)
	if err != nil {
		return nil, err
	}
	updatedRes, err := s.resourceRepository.GetByURN(urn)
	if err != nil {
		return nil, err
	}
	return updatedRes, nil
}

func generateResourceUrn(res *domain.Resource) string {
	return strings.Join([]string{
		sanitizeString(res.Parent),
		sanitizeString(res.Name),
		sanitizeString(res.Kind),
	}, "-")
}

func sanitizeString(s string) string {
	return strings.Replace(s, " ", "_", -1)
}
