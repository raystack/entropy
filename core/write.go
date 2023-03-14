package core

import (
	"context"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
)

func (s *Service) CreateResource(ctx context.Context, res resource.Resource) (*resource.Resource, error) {
	if err := res.Validate(true); err != nil {
		return nil, err
	}

	act := module.ActionRequest{
		Name:   module.CreateAction,
		Params: res.Spec.Configs,
	}
	res.Spec.Configs = nil

	return s.execAction(ctx, res, act)
}

func (s *Service) UpdateResource(ctx context.Context, urn string, req resource.UpdateRequest) (*resource.Resource, error) {
	if len(req.Spec.Dependencies) != 0 {
		return nil, errors.ErrUnsupported.WithMsgf("updating dependencies is not supported")
	} else if len(req.Spec.Configs) == 0 {
		return nil, errors.ErrInvalid.WithMsgf("no config is being updated, nothing to do")
	}

	return s.ApplyAction(ctx, urn, module.ActionRequest{
		Name:   module.UpdateAction,
		Params: req.Spec.Configs,
		Labels: req.Labels,
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

func (s *Service) execAction(ctx context.Context, res resource.Resource, act module.ActionRequest) (*resource.Resource, error) {
	planned, err := s.planChange(ctx, res, act)
	if err != nil {
		return nil, err
	}

	if isCreate(act.Name) {
		planned.Resource.CreatedAt = s.clock()
		planned.Resource.UpdatedAt = planned.Resource.CreatedAt
	} else {
		planned.Resource.CreatedAt = res.CreatedAt
		planned.Resource.UpdatedAt = s.clock()
	}

	if err := s.upsert(ctx, *planned, isCreate(act.Name), true, planned.Reason); err != nil {
		return nil, err
	}
	return &planned.Resource, nil
}

func isCreate(actionName string) bool {
	return actionName == module.CreateAction
}

func (s *Service) planChange(ctx context.Context, res resource.Resource, act module.ActionRequest) (*module.Plan, error) {
	modSpec, err := s.generateModuleSpec(ctx, res)
	if err != nil {
		return nil, err
	}

	planned, err := s.moduleSvc.PlanAction(ctx, *modSpec, act)
	if err != nil {
		if errors.Is(err, errors.ErrInvalid) {
			return nil, err
		}
		return nil, errors.ErrInternal.WithMsgf("plan() failed").WithCausef(err.Error())
	}

	planned.Resource.Labels = act.Labels
	if err := planned.Resource.Validate(isCreate(act.Name)); err != nil {
		return nil, err
	}

	return planned, nil
}

func (s *Service) upsert(ctx context.Context, plan module.Plan, isCreate bool, saveRevision bool, reason string) error {
	var hooks []resource.MutationHook
	hooks = append(hooks, func(ctx context.Context) error {
		if plan.Resource.State.IsTerminal() {
			// no need to enqueue if resource has reached terminal state.
			return nil
		}

		return s.enqueueSyncJob(ctx, plan.Resource, s.clock(), JobKindSyncResource)
	})

	if !plan.ScheduleRunAt.IsZero() {
		hooks = append(hooks, func(ctx context.Context) error {
			return s.enqueueSyncJob(ctx, plan.Resource, plan.ScheduleRunAt, JobKindScheduledSyncResource)
		})
	}

	var err error
	if isCreate {
		err = s.store.Create(ctx, plan.Resource, hooks...)
	} else {
		err = s.store.Update(ctx, plan.Resource, saveRevision, reason, hooks...)
	}

	if err != nil {
		if isCreate && errors.Is(err, errors.ErrConflict) {
			return errors.ErrConflict.WithMsgf("resource with urn '%s' already exists", plan.Resource.URN)
		} else if !isCreate && errors.Is(err, errors.ErrNotFound) {
			return errors.ErrNotFound.WithMsgf("resource with urn '%s' does not exist", plan.Resource.URN)
		}
		return errors.ErrInternal.WithCausef(err.Error())
	}

	return nil
}
