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
	case ResetAction:
		return m.planReset(spec, act)
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
	if err := reqConf.sanitiseAndValidate(); err != nil {
		return nil, err
	}

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
		if err := reqConf.sanitiseAndValidate(); err != nil {
			return nil, err
		}
		conf = reqConf

	case ScaleAction:
		var scaleParams struct {
			Replicas int `json:"replicas"`
		}
		if err := json.Unmarshal(act.Params, &scaleParams); err != nil {
			return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
		}
		conf.Firehose.Replicas = scaleParams.Replicas

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

func (*firehoseModule) planReset(spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	var conf moduleConfig
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	var resetParams struct {
		To       string `json:"to"`
		Datetime string `json:"datetime"`
	}
	if err := json.Unmarshal(act.Params, &resetParams); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid action params json: %v", err)
	}

	var resetTo string
	switch resetParams.To {
	case "DATETIME":
		resetTo = resetParams.Datetime
	default:
		resetTo = resetParams.To
	}

	r.Spec.Configs = conf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps:  []string{releaseUpdate, consumerReset, releaseUpdate},
			ResetTo:       resetTo,
			StateOverride: stateStopped,
		}.JSON(),
	}
	return &r, nil
}
