package core

import (
	"context"
	"fmt"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
)

func (svc *Service) CreateResource(ctx context.Context, res resource.Resource) (*resource.Resource, error) {
	if err := res.Validate(true); err != nil {
		return nil, err
	}

	act := module.ActionRequest{
		Name:   module.CreateAction,
		Params: res.Spec.Configs,
		Labels: res.Labels,
	}
	res.Spec.Configs = nil

	return svc.execAction(ctx, res, act)
}

func (svc *Service) UpdateResource(ctx context.Context, urn string, req resource.UpdateRequest) (*resource.Resource, error) {
	if len(req.Spec.Dependencies) != 0 {
		return nil, errors.ErrUnsupported.WithMsgf("updating dependencies is not supported")
	} else if len(req.Spec.Configs) == 0 {
		return nil, errors.ErrInvalid.WithMsgf("no config is being updated, nothing to do")
	}

	return svc.ApplyAction(ctx, urn, module.ActionRequest{
		Name:   module.UpdateAction,
		Params: req.Spec.Configs,
		Labels: req.Labels,
	})
}

func (svc *Service) DeleteResource(ctx context.Context, urn string) error {
	_, actionErr := svc.ApplyAction(ctx, urn, module.ActionRequest{
		Name: module.DeleteAction,
	})
	return actionErr
}

func (svc *Service) ApplyAction(ctx context.Context, urn string, act module.ActionRequest) (*resource.Resource, error) {
	res, err := svc.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	} else if !res.State.IsTerminal() {
		return nil, errors.ErrInvalid.
			WithMsgf("cannot perform '%s' on resource in '%s'", act.Name, res.State.Status)
	}

	return svc.execAction(ctx, *res, act)
}

func (svc *Service) execAction(ctx context.Context, res resource.Resource, act module.ActionRequest) (*resource.Resource, error) {
	planned, err := svc.planChange(ctx, res, act)
	if err != nil {
		return nil, err
	}

	if isCreate(act.Name) {
		planned.CreatedAt = svc.clock()
		planned.UpdatedAt = planned.CreatedAt
	} else {
		planned.CreatedAt = res.CreatedAt
		planned.UpdatedAt = svc.clock()
	}

	reason := fmt.Sprintf("action:%s", act.Name)
	if err := svc.upsert(ctx, *planned, isCreate(act.Name), true, reason); err != nil {
		return nil, err
	}
	return planned, nil
}

func (svc *Service) planChange(ctx context.Context, res resource.Resource, act module.ActionRequest) (*resource.Resource, error) {
	modSpec, err := svc.generateModuleSpec(ctx, res)
	if err != nil {
		return nil, err
	}

	planned, err := svc.moduleSvc.PlanAction(ctx, *modSpec, act)
	if err != nil {
		if errors.Is(err, errors.ErrInvalid) {
			return nil, err
		}
		return nil, errors.ErrInternal.WithMsgf("plan() failed").WithCausef(err.Error())
	}

	planned.Labels = act.Labels
	if err := planned.Validate(isCreate(act.Name)); err != nil {
		return nil, err
	}

	return planned, nil
}

func (svc *Service) upsert(ctx context.Context, res resource.Resource, isCreate bool, saveRevision bool, reason string) error {
	var err error
	if isCreate {
		err = svc.store.Create(ctx, res)
	} else {
		err = svc.store.Update(ctx, res, saveRevision, reason)
	}

	if err != nil {
		if isCreate && errors.Is(err, errors.ErrConflict) {
			return errors.ErrConflict.WithMsgf("resource with urn '%s' already exists", res.URN)
		} else if !isCreate && errors.Is(err, errors.ErrNotFound) {
			return errors.ErrNotFound.WithMsgf("resource with urn '%s' does not exist", res.URN)
		}
		return errors.ErrInternal.WithCausef(err.Error())
	}

	return nil
}

func isCreate(actionName string) bool {
	return actionName == module.CreateAction
}
