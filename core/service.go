package core

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func New(resourceRepo resource.Repository, rootModule module.Module, clockFn func() time.Time, lg *zap.Logger) *Service {
	if clockFn == nil {
		clockFn = time.Now
	}
	return &Service{
		clock:      clockFn,
		repository: resourceRepo,
		rootModule: rootModule,
	}
}

type Service struct {
	logger     *zap.Logger
	clock      func() time.Time
	repository resource.Repository
	rootModule module.Module
}

func (s *Service) GetResource(ctx context.Context, urn string) (*resource.Resource, error) {
	res, err := s.repository.GetByURN(ctx, urn)
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

	resources, err := s.repository.List(ctx, filter)
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

	plannedRes, err := s.planChange(ctx, res, act)
	if err != nil {
		return nil, err
	}

	plannedRes.CreatedAt = s.clock()
	plannedRes.UpdatedAt = plannedRes.CreatedAt
	if err := plannedRes.Validate(); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, *plannedRes); err != nil {
		if errors.Is(err, errors.ErrConflict) {
			return nil, errors.ErrConflict.WithMsgf("resource with urn '%s' already exists", res.URN)
		}
		return nil, err
	}
	return plannedRes, nil
}

func (s *Service) UpdateResource(ctx context.Context, urn string, newSpec resource.Spec) (*resource.Resource, error) {
	if len(newSpec.Dependencies) != 0 {
		return nil, errors.ErrUnsupported.WithMsgf("updating dependencies is not supported")
	} else if len(newSpec.Configs) == 0 {
		return nil, errors.ErrInvalid.WithMsgf("no config is being updated, nothing to do")
	}

	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	} else if !res.State.IsTerminal() {
		return nil, errors.ErrInvalid.WithMsgf("resource must be in a terminal state to be updatable")
	}

	act := module.ActionRequest{
		Name:   module.UpdateAction,
		Params: newSpec.Configs,
	}

	plannedRes, err := s.planChange(ctx, *res, act)
	if err != nil {
		return nil, err
	}

	plannedRes.UpdatedAt = s.clock()
	if err := s.repository.Update(ctx, *plannedRes); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	return plannedRes, nil
}

func (s *Service) DeleteResource(ctx context.Context, urn string) error {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return err
	} else if !res.State.IsTerminal() {
		return errors.ErrInvalid.
			WithMsgf("resource state '%s' is inappropriate for scheduling deletion", res.State.Status)
	}

	res.State.Status = resource.StatusDeleted
	res.UpdatedAt = s.clock()
	if err := s.repository.Update(ctx, *res); err != nil {
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
	} else if !res.State.IsTerminal() {
		return nil, errors.ErrInvalid.WithMsgf("resource must be in terminal state for applying actions")
	}

	plannedRes, err := s.planChange(ctx, *res, action)
	if err != nil {
		return nil, err
	}

	plannedRes.CreatedAt = res.CreatedAt
	plannedRes.UpdatedAt = s.clock()
	if err := s.repository.Update(ctx, *plannedRes); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	return plannedRes, nil
}

func (s *Service) GetLog(ctx context.Context, urn string, filter map[string]string) (<-chan module.LogChunk, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	moduleLogStream, supported := s.rootModule.(module.Loggable)
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
	modSpec, err := s.generateModuleSpec(ctx, res)
	if err != nil {
		return nil, err
	}
	res.Spec.Configs = map[string]interface{}{}

	plannedRes, err := s.rootModule.Plan(ctx, *modSpec, act)
	if err != nil {
		if errors.Is(err, errors.ErrInvalid) {
			return nil, err
		}
		return nil, errors.ErrInternal.WithMsgf("plan() failed").WithCausef(err.Error())
	}

	return plannedRes, nil
}

func (s *Service) generateModuleSpec(ctx context.Context, res resource.Resource) (*module.Spec, error) {
	modSpec := module.Spec{
		Resource:     res,
		Dependencies: map[string]resource.Output{},
	}

	for key, resURN := range res.Spec.Dependencies {
		d, err := s.GetResource(ctx, resURN)
		if err != nil {
			if errors.Is(err, errors.ErrNotFound) {
				return nil, errors.ErrInvalid.
					WithMsgf("dependency '%s' not found", resURN)
			}
			return nil, err
		} else if d.State.Status != resource.StatusCompleted {
			return nil, errors.ErrInvalid.
				WithMsgf("dependency '%s' is in incomplete state (%s)", resURN, d.State.Status)
		}

		modSpec.Dependencies[key] = d.State.Output
	}

	return &modSpec, nil
}
