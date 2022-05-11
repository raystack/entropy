package core

import (
	"context"
	"time"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func New(resourceRepo resource.Repository, resolver moduleResolverFn, clockFn func() time.Time) *Service {
	if clockFn == nil {
		clockFn = time.Now
	}
	return &Service{
		clock:              clockFn,
		resolveModule:      resolver,
		resourceRepository: resourceRepo,
	}
}

type moduleResolverFn func(kind string) (module.Module, error)

type Service struct {
	clock              func() time.Time
	resolveModule      moduleResolverFn
	resourceRepository resource.Repository
}

func (s *Service) GetResource(ctx context.Context, urn string) (*resource.Resource, error) {
	res, err := s.resourceRepository.GetByURN(ctx, urn)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, errors.ErrNotFound.WithMsgf("resource with urn '%s' not found", urn)
		}
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	return res, nil
}

func (s *Service) ListResources(ctx context.Context, project string, kind string) ([]resource.Resource, error) {
	filter := map[string]string{}
	if kind != "" {
		filter["kind"] = kind
	}
	if project != "" {
		filter["project"] = project
	}

	resources, err := s.resourceRepository.List(ctx, filter)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	var result []resource.Resource
	for _, res := range resources {
		result = append(result, *res)
	}
	return result, nil
}

func (s *Service) CreateResource(ctx context.Context, res resource.Resource) (*resource.Resource, error) {
	act := module.ActionRequest{
		Name:   module.CreateAction,
		Params: res.Spec.Configs,
	}
	res.Spec.Configs = nil

	plannedRes, err := s.planChange(ctx, res, act)
	if err != nil {
		return nil, err
	}

	plannedRes.CreatedAt = s.clock()
	plannedRes.UpdatedAt = plannedRes.CreatedAt
	if err := plannedRes.Validate(); err != nil {
		return nil, err
	}

	if err := s.resourceRepository.Create(ctx, *plannedRes); err != nil {
		if errors.Is(err, errors.ErrConflict) {
			return nil, errors.ErrConflict.WithMsgf("resource with urn '%s' already exists", res.URN)
		}
		return nil, err
	}
	return plannedRes, nil
}

func (s *Service) UpdateResource(ctx context.Context, urn string, newSpec resource.Spec) (*resource.Resource, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	res.UpdatedAt = s.clock()
	res.Spec = newSpec
	res.State = resource.State{
		Status:     resource.StatusPending,
		Output:     res.State.Output,
		ModuleData: res.State.ModuleData,
	}

	plannedRes, err := s.planChange(ctx, *res, module.ActionRequest{Name: module.UpdateAction})
	if err != nil {
		return nil, err
	}

	if err := s.resourceRepository.Update(ctx, *plannedRes); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	return plannedRes, nil
}

func (s *Service) DeleteResource(ctx context.Context, urn string) error {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return err
	}

	res.State.Status = resource.StatusDeleted
	res.UpdatedAt = s.clock()
	if err := s.resourceRepository.Update(ctx, *res); err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil
		}
		return errors.ErrInternal.WithCausef(err.Error())
	}
	return nil
}

func (s *Service) ApplyAction(ctx context.Context, urn string, action module.ActionRequest) (*resource.Resource, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	plannedRes, err := s.planChange(ctx, *res, action)
	if err != nil {
		return nil, err
	}

	plannedRes.CreatedAt = res.CreatedAt
	plannedRes.UpdatedAt = s.clock()
	if err := s.resourceRepository.Update(ctx, *plannedRes); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	return plannedRes, nil
}

func (s *Service) GetLog(ctx context.Context, urn string, filter map[string]string) (<-chan module.LogChunk, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	m, err := s.resolveModule(res.Kind)
	if err != nil {
		return nil, errors.ErrInternal.
			WithMsgf("failed to resolve module for kind '%s'", res.Kind).
			WithCausef(err.Error())
	}

	moduleLogStream, supported := m.(module.Loggable)
	if !supported {
		return nil, errors.ErrUnsupported.WithMsgf("log streaming not supported for kind '%s'", res.Kind)
	}

	modSpec := module.Spec{
		Resource:     *res,
		Dependencies: map[string]resource.Output{},
	}

	return moduleLogStream.Log(ctx, modSpec, filter)
}

func (s *Service) planChange(ctx context.Context, res resource.Resource, act module.ActionRequest) (*resource.Resource, error) {
	modSpec := module.Spec{
		Resource:     res,
		Dependencies: map[string]resource.Output{},
	}

	m, err := s.resolveModule(res.Kind)
	if err != nil {
		return nil, errors.ErrInvalid.
			WithMsgf("failed to resolve module for kind '%s'", res.Kind).
			WithCausef(err.Error())
	}

	plannedRes, err := m.Plan(ctx, modSpec, act)
	if err != nil {
		if errors.Is(err, errors.ErrInvalid) {
			return nil, err
		}
		return nil, errors.ErrInternal.WithMsgf("plan() failed").WithCausef(err.Error())
	}

	return plannedRes, nil
}
