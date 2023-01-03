package firehose

import (
	"context"
	"encoding/json"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func (m *firehoseModule) Plan(_ context.Context, res module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	switch act.Name {
	case module.CreateAction:
		return m.planCreate(res, act)
	case ResetAction:
		return m.planReset(res, act)
	default:
		return m.planChange(res, act)
	}
}

func (*firehoseModule) planCreate(res module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	var plan module.Plan
	r := res.Resource

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

	plan.Resource = r
	if reqConf.StopTime != nil {
		plan.ScheduleRunAt = *reqConf.StopTime
	}
	return &plan, nil
}

func (*firehoseModule) planChange(res module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	var plan module.Plan
	r := res.Resource

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

		if conf.StopTime != nil {
			plan.ScheduleRunAt = *conf.StopTime
		}

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
		Output: res.State.Output,
		ModuleData: moduleData{
			PendingSteps: []string{releaseUpdate},
		}.JSON(),
	}
	plan.Resource = r
	return &plan, nil
}

func (*firehoseModule) planReset(res module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	r := res.Resource

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
		Output: res.State.Output,
		ModuleData: moduleData{
			PendingSteps:  []string{releaseUpdate, consumerReset, releaseUpdate},
			ResetTo:       resetTo,
			StateOverride: stateStopped,
		}.JSON(),
	}

	return &module.Plan{Resource: r}, nil
}
