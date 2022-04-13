package provider

import "context"

type Service struct {
	repo Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repo: repository,
	}
}

func (s *Service) CreateProvider(ctx context.Context, pro Provider) (*Provider, error) {
	err := s.repo.Create(pro)
	if err != nil {
		return nil, err
	}
	createdProvider, err := s.repo.GetByURN(pro.URN)
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
	return s.repo.List(filter)
}
