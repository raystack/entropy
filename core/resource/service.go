package resource

import (
	"context"
)

func NewService(repository Repository, moduleReg ModuleRegistry) *Service {
	return &Service{
		moduleRegistry:     moduleReg,
		resourceRepository: repository,
	}
}

type Service struct {
	moduleRegistry     ModuleRegistry
	resourceRepository Repository
}

func (s *Service) GetResource(ctx context.Context, urn string) (*Resource, error) {
	return s.resourceRepository.GetByURN(ctx, urn)
}

func (s *Service) ListResources(ctx context.Context, parent string, kind string) ([]Resource, error) {
	filter := map[string]string{}
	if kind != "" {
		filter["kind"] = kind
	}
	if parent != "" {
		filter["parent"] = parent
	}

	resources, err := s.resourceRepository.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	var result []Resource
	for _, res := range resources {
		result = append(result, *res)
	}
	return result, nil
}

func (s *Service) CreateResource(ctx context.Context, res Resource) (*Resource, error) {
	res.Status = StatusPending
	res.URN = GenerateURN(res)

	if err := s.validate(ctx, res); err != nil {
		return nil, err
	}

	err := s.resourceRepository.Create(ctx, res)
	if err != nil {
		return nil, err
	}

	syncedRes, err := s.sync(ctx, res)
	if err != nil {
		return nil, err
	}

	if err := s.resourceRepository.Update(ctx, *syncedRes); err != nil {
		return nil, err
	}

	return syncedRes, nil
}

func (s *Service) UpdateResource(ctx context.Context, urn string, updates Updates) (*Resource, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	res.Status = StatusPending
	res.Configs = updates.Configs
	if err := s.validate(ctx, *res); err != nil {
		return nil, err
	}

	if err := s.resourceRepository.Update(ctx, *res); err != nil {
		return nil, err
	}

	syncedRes, err := s.sync(ctx, *res)
	if err != nil {
		return nil, err
	}

	if err := s.resourceRepository.Update(ctx, *syncedRes); err != nil {
		return nil, err
	}
	return syncedRes, nil
}

func (s *Service) DeleteResource(ctx context.Context, urn string) error {
	// TODO: notify the module about deletion.
	return s.resourceRepository.Delete(ctx, urn)
}

func (s *Service) ApplyAction(ctx context.Context, urn string, action Action) (*Resource, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	m, err := s.moduleRegistry.Get(res.Kind)
	if err != nil {
		return nil, err
	}

	configs, err := m.Act(*res, action.Name, action.Params)
	if err != nil {
		return nil, err
	}

	res.Configs = configs
	syncedRes, err := s.sync(ctx, *res)
	if err != nil {
		return nil, err
	}

	if err := s.resourceRepository.Update(ctx, *syncedRes); err != nil {
		return nil, err
	}

	return syncedRes, nil
}

func (s *Service) GetLog(ctx context.Context, urn string, filter map[string]string) (<-chan LogChunk, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	m, err := s.moduleRegistry.Get(res.Kind)
	if err != nil {
		return nil, err
	}

	moduleLogStream, supported := m.(LoggableModule)
	if !supported {
		return nil, ErrLogStreamingUnsupported
	}

	return moduleLogStream.Log(ctx, *res, filter)
}

func (s *Service) sync(ctx context.Context, r Resource) (*Resource, error) {
	m, err := s.moduleRegistry.Get(r.Kind)
	if err != nil {
		r.Status = StatusError
		return &r, err
	}
	status, err := m.Apply(r)
	r.Status = status
	return &r, err
}

func (s *Service) validate(ctx context.Context, r Resource) error {
	m, err := s.moduleRegistry.Get(r.Kind)
	if err != nil {
		return err
	}
	return m.Validate(r)
}
