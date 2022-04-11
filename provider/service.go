package provider

import "context"

type Service struct {
	providerRepository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		providerRepository: repository,
	}
}

func (s *Service) CreateProvider(ctx context.Context, pro *Provider) (*Provider, error) {
	err := s.providerRepository.Create(pro)
	if err != nil {
		return nil, err
	}
	createdProvider, err := s.providerRepository.GetByURN(pro.URN)
	if err != nil {
		return nil, err
	}
	return createdProvider, nil
}

func (s *Service) ListProviders(ctx context.Context, parent string, kind string) ([]*Provider, error) {
	filter := map[string]string{}
	if kind != "" {
		filter["kind"] = kind
	}
	if parent != "" {
		filter["parent"] = parent
	}
	return s.providerRepository.List(filter)
}
