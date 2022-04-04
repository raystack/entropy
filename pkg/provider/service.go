package provider

import (
	"context"

	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/store"
)

type ServiceInterface interface {
	CreateProvider(ctx context.Context, res *domain.Provider) (*domain.Provider, error)
	ListProviders(ctx context.Context, parent string, kind string) ([]*domain.Provider, error)
}

type Service struct {
	providerRepository store.ProviderRepository
}

func NewService(repository store.ProviderRepository) *Service {
	return &Service{
		providerRepository: repository,
	}
}

func (s *Service) CreateProvider(ctx context.Context, pro *domain.Provider) (*domain.Provider, error) {
	err := s.providerRepository.Create(pro)
	if err != nil {
		return nil, err
	}
	createdProvider, err := s.providerRepository.GetByURN(pro.Urn)
	if err != nil {
		return nil, err
	}
	return createdProvider, nil
}

func (s *Service) ListProviders(ctx context.Context, parent string, kind string) ([]*domain.Provider, error) {
	filter := map[string]string{}
	if kind != "" {
		filter["kind"] = kind
	}
	if parent != "" {
		filter["parent"] = parent
	}
	return s.providerRepository.List(filter)
}
