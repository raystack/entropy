package core

import (
	"context"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

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
