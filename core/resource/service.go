package resource

import (
	"context"

	"github.com/odpf/entropy/pkg/errors"
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
	res, err := s.resourceRepository.GetByURN(ctx, urn)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, errors.ErrNotFound.WithMsgf("resource with urn '%s' not found", urn)
		}
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	return res, nil
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
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	var result []Resource
	for _, res := range resources {
		result = append(result, *res)
	}
	return result, nil
}

func (s *Service) CreateResource(ctx context.Context, res Resource) (*Resource, error) {
	res.Status = StatusPending
	res.URN = generateURN(res)

	if err := s.validate(ctx, res); err != nil {
		return nil, err
	}

	err := s.resourceRepository.Create(ctx, res)
	if err != nil {
		if errors.Is(err, errors.ErrConflict) {
			return nil, errors.ErrConflict.WithMsgf("resource with urn '%s' already exists", res.URN)
		}
		return nil, err
	}

	return s.sync(ctx, res)
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
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	return s.sync(ctx, *res)
}

func (s *Service) DeleteResource(ctx context.Context, urn string) error {
	// TODO: notify the module about deletion.

	err := s.resourceRepository.Delete(ctx, urn)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil
		}
		return errors.ErrInternal.WithCausef(err.Error())
	}
	return nil
}

func (s *Service) ApplyAction(ctx context.Context, urn string, action Action) (*Resource, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	m, err := s.moduleRegistry.Get(res.Kind)
	if err != nil {
		return nil, errors.ErrInternal.
			WithMsgf("failed to resolve module for kind '%s'", res.Kind).
			WithCausef(err.Error())
	}

	configs, err := m.Act(*res, action.Name, action.Params)
	if err != nil {
		return nil, errors.ErrInternal.
			WithMsgf("executing module action failed").
			WithCausef(err.Error())
	}
	res.Configs = configs

	return s.sync(ctx, *res)
}

func (s *Service) GetLog(ctx context.Context, urn string, filter map[string]string) (<-chan LogChunk, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	m, err := s.moduleRegistry.Get(res.Kind)
	if err != nil {
		return nil, errors.ErrInternal.
			WithMsgf("failed to resolve module for kind '%s'", res.Kind).
			WithCausef(err.Error())
	}

	moduleLogStream, supported := m.(LoggableModule)
	if !supported {
		return nil, errors.ErrUnsupported.WithMsgf("log streaming not supported for kind '%s'", res.Kind)
	}

	return moduleLogStream.Log(ctx, *res, filter)
}

func (s *Service) sync(ctx context.Context, r Resource) (*Resource, error) {
	// TODO: clarify and fix the expected behaviour here.
	m, err := s.moduleRegistry.Get(r.Kind)
	if err != nil {
		r.Status = StatusError
	} else {
		r.Status, _ = m.Apply(r)
	}

	if err := s.resourceRepository.Update(ctx, r); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	return &r, nil
}

func (s *Service) validate(ctx context.Context, r Resource) error {
	m, err := s.moduleRegistry.Get(r.Kind)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return errors.ErrInvalid.WithMsgf("invalid kind '%s'", r.Kind)
		}
		return errors.ErrInternal.WithCausef(err.Error())
	}
	return m.Validate(r)
}
