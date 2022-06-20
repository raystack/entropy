package firehose

import (
	"context"
	"encoding/json"
	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func (m *firehoseModule) Plan(_ context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	switch act.Name {
	case module.CreateAction:
		return m.planCreate(spec, act)
	default:
		return m.planChange(spec, act)
	}
}

func (*firehoseModule) planCreate(spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	var reqConf moduleConfig
	if err := json.Unmarshal(act.Params, &reqConf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}
	reqConf.sanitiseAndValidate()

	r.Spec.Configs = reqConf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps: []string{releaseCreate},
		}.JSON(),
	}
	return &r, nil
}

func (*firehoseModule) planChange(spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	var conf moduleConfig
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	switch act.Name {
	case module.UpdateAction:
		var reqConf moduleConfig
		if err := json.Unmarshal(act.Params, &reqConf); err != nil {
			return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
		}
		reqConf.sanitiseAndValidate()
		conf = reqConf

	case ScaleAction:
		var scaleParams struct {
			Replicas int `json:"replicas"`
		}
		if err := json.Unmarshal(act.Params, &scaleParams); err != nil {
			return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
		}
		conf.FirehoseConfigs.Replicas = scaleParams.Replicas

	case StartAction:
		conf.State = stateRunning

	case StopAction:
		conf.State = stateStopped
	}

	r.Spec.Configs = conf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps: []string{releaseUpdate},
		}.JSON(),
	}
	return &r, nil
}
