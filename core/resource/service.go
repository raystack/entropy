package resource

import (
	"context"
	"log"
	"time"

	"github.com/odpf/entropy/pkg/errors"
)

func NewService(repository Repository, moduleReg ModuleRegistry, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{
		clock:              now,
		moduleRegistry:     moduleReg,
		resourceRepository: repository,
	}
}

type Service struct {
	clock              func() time.Time
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
	if err := res.Validate(); err != nil {
		return nil, err
	}
	res.Status = StatusPending
	res.CreatedAt = s.clock()
	res.UpdatedAt = res.CreatedAt

	if err := s.validateByModule(ctx, res); err != nil {
		return nil, err
	}

	if err := s.resourceRepository.Create(ctx, res); err != nil {
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
	res.UpdatedAt = s.clock()
	if err := s.validateByModule(ctx, *res); err != nil {
		return nil, err
	}

	if err := s.resourceRepository.Update(ctx, *res); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	return s.sync(ctx, *res)
}

func (s *Service) DeleteResource(ctx context.Context, urn string) error {
	// TODO: notify the module about deletion.

	if err := s.resourceRepository.Delete(ctx, urn); err != nil {
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
		r.Status, err = m.Apply(r)
		if err != nil {
			log.Printf("apply failed: %v", err)
		}
	}

	r.UpdatedAt = s.clock()
	if err := s.resourceRepository.Update(ctx, r); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	return &r, nil
}

func (s *Service) validateByModule(ctx context.Context, r Resource) error {
	m, err := s.moduleRegistry.Get(r.Kind)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return errors.ErrInvalid.WithMsgf("invalid resource kind '%s'", r.Kind)
		}
		return errors.ErrInternal.WithCausef(err.Error())
	}
	return m.Validate(r)
}
