package resource

import (
	"context"
	"time"

	"github.com/odpf/entropy/pkg/errors"
)

const (
	ActionCreate = "create"
	ActionUpdate = "update"
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
	act := Action{
		Name:   ActionCreate,
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

func (s *Service) UpdateResource(ctx context.Context, urn string, newSpec Spec) (*Resource, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	res.UpdatedAt = s.clock()
	res.Spec = newSpec
	res.State = State{
		Status:     StatusPending,
		Output:     res.State.Output,
		ModuleData: res.State.ModuleData,
	}

	plannedRes, err := s.planChange(ctx, *res, Action{Name: ActionUpdate})
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

	res.State.Status = StatusDeleted
	res.UpdatedAt = s.clock()
	if err := s.resourceRepository.Update(ctx, *res); err != nil {
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

func (s *Service) planChange(ctx context.Context, res Resource, act Action) (*Resource, error) {
	modSpec := ModuleSpec{
		Resource:     res,
		Dependencies: map[string]Output{},
	}

	m, err := s.moduleRegistry.Get(res.Kind)
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

		return nil, errors.ErrInternal.
			WithMsgf("plan() failed").
			WithCausef(err.Error())
	}
	return plannedRes, nil
}
