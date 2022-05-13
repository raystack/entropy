package firehose

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

const (
	ResetAction  = "reset"
	StopAction   = "stop"
	StartAction  = "start"
	ScaleAction  = "scale"
	DeleteAction = "delete"
)

var Module = module.Descriptor{
	Kind: "firehose",
	Actions: []module.ActionDesc{
		{
			Name:        module.CreateAction,
			Description: "creates firehose instance",
			ParamSchema: createActionSchema,
		},
		{
			Name:        ScaleAction,
			Description: "creates firehose instance",
			ParamSchema: scaleActionSchema,
		},
	},
	Module: &firehoseModule{},
}

type firehoseModule struct{}

func (m *firehoseModule) Plan(ctx context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	switch act.Name {
	case module.CreateAction:
		return m.planCreate(ctx, spec, act)

	case ScaleAction:
		return m.planScale(ctx, spec, act)
	}

	return &r, nil
}

func (m *firehoseModule) Sync(ctx context.Context, spec module.Spec) (*resource.State, error) {
	return &resource.State{
		Status:     resource.StatusCompleted,
		Output:     map[string]interface{}{"foo": "bar"},
		ModuleData: nil,
	}, nil
}

func (m *firehoseModule) planCreate(ctx context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	var conf moduleConfig
	if err := json.Unmarshal(act.Params, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	} else if err := conf.sanitiseAndValidate(); err != nil {
		return nil, err
	}

	r.Spec.Configs = conf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps: []string{"helm_create"},
		}.JSON(),
	}

	return &r, nil
}

func (m *firehoseModule) planScale(ctx context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	var scaleParams struct {
		Replicas int `json:"replicas"`
	}
	if err := json.Unmarshal(act.Params, &scaleParams); err != nil {
		return nil, err
	}

	r := spec.Resource

	var conf moduleConfig
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	conf.Replicas = scaleParams.Replicas
	if err := conf.sanitiseAndValidate(); err != nil {
		return nil, err
	}

	r.Spec.Configs = conf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps: []string{"helm_update"},
		}.JSON(),
	}

	return &r, nil
}
