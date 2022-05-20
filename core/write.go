package core

import (
	"context"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func (s *Service) CreateResource(ctx context.Context, res resource.Resource) (*resource.Resource, error) {
	if err := res.Validate(); err != nil {
		return nil, err
	}

	act := module.ActionRequest{
		Name:   module.CreateAction,
		Params: res.Spec.Configs,
	}
	res.Spec.Configs = nil

	return s.execAction(ctx, res, act)
}

func (s *Service) UpdateResource(ctx context.Context, urn string, newSpec resource.Spec) (*resource.Resource, error) {
	if len(newSpec.Dependencies) != 0 {
		return nil, errors.ErrUnsupported.WithMsgf("updating dependencies is not supported")
	} else if len(newSpec.Configs) == 0 {
		return nil, errors.ErrInvalid.WithMsgf("no config is being updated, nothing to do")
	}

	return s.ApplyAction(ctx, urn, module.ActionRequest{
		Name:   module.UpdateAction,
		Params: newSpec.Configs,
	})
}

func (s *Service) DeleteResource(ctx context.Context, urn string) error {
	_, actionErr := s.ApplyAction(ctx, urn, module.ActionRequest{
		Name: module.DeleteAction,
	})
	return actionErr
}

func (s *Service) ApplyAction(ctx context.Context, urn string, act module.ActionRequest) (*resource.Resource, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	} else if !res.State.IsTerminal() {
		return nil, errors.ErrInvalid.
			WithMsgf("cannot perform '%s' on resource in '%s'", act.Name, res.State.Status)
	}

	return s.execAction(ctx, *res, act)
}

func (s *Service) execAction(ctx context.Context,
	res resource.Resource, act module.ActionRequest) (*resource.Resource, error) {

	plannedRes, err := s.planChange(ctx, res, act)
	if err != nil {
		return nil, err
	}

	if act.Name == module.CreateAction {
		plannedRes.CreatedAt = s.clock()
		plannedRes.UpdatedAt = plannedRes.CreatedAt
		if err := s.repository.Create(ctx, *plannedRes); err != nil {
			if errors.Is(err, errors.ErrConflict) {
				return nil, errors.ErrConflict.
					WithMsgf("resource with urn '%s' already exists", plannedRes.URN)
			}
			return nil, err
		}
	} else {
		plannedRes.CreatedAt = res.CreatedAt
		plannedRes.UpdatedAt = s.clock()
		if err := s.repository.Update(ctx, *plannedRes); err != nil {
			if errors.Is(err, errors.ErrNotFound) {
				return nil, errors.ErrNotFound.
					WithMsgf("resource with urn '%s' does not exist", plannedRes.URN)
			}
			return nil, errors.ErrInternal.WithCausef(err.Error())
		}
	}

	return plannedRes, nil
}

func (s *Service) planChange(ctx context.Context, res resource.Resource, act module.ActionRequest) (*resource.Resource, error) {
	modSpec, err := s.generateModuleSpec(ctx, res)
	if err != nil {
		return nil, err
	}

	plannedRes, err := s.rootModule.Plan(ctx, *modSpec, act)
	if err != nil {
		if errors.Is(err, errors.ErrInvalid) {
			return nil, err
		}
		return nil, errors.ErrInternal.WithMsgf("plan() failed").WithCausef(err.Error())
	} else if err := plannedRes.Validate(); err != nil {
		return nil, err
	}

	return plannedRes, nil
}
