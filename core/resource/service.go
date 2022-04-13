package resource

import "context"

type Service struct {
	resourceRepository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		resourceRepository: repository,
	}
}

func (s *Service) GetResource(ctx context.Context, urn string) (*Resource, error) {
	return s.resourceRepository.GetByURN(urn)
}

func (s *Service) ListResources(ctx context.Context, parent string, kind string) ([]*Resource, error) {
	filter := map[string]string{}
	if kind != "" {
		filter["kind"] = kind
	}
	if parent != "" {
		filter["parent"] = parent
	}
	return s.resourceRepository.List(filter)
}

func (s *Service) CreateResource(ctx context.Context, res *Resource) (*Resource, error) {
	res.Status = StatusPending

	err := s.resourceRepository.Create(res)
	if err != nil {
		return nil, err
	}

	createdResource, err := s.resourceRepository.GetByURN(res.URN)
	if err != nil {
		return nil, err
	}
	return createdResource, nil
}

func (s *Service) UpdateResource(ctx context.Context, res *Resource) (*Resource, error) {
	err := s.resourceRepository.Update(res)
	if err != nil {
		return nil, err
	}
	updatedRes, err := s.resourceRepository.GetByURN(res.URN)
	if err != nil {
		return nil, err
	}
	return updatedRes, nil
}

func (s *Service) DeleteResource(ctx context.Context, urn string) error {
	err := s.resourceRepository.Delete(urn)
	if err != nil {
		return err
	}
	return nil
}
