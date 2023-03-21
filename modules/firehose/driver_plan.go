package firehose

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kafka"
)

func (fd *firehoseDriver) Plan(_ context.Context, exr module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	switch act.Name {
	case module.CreateAction:
		return fd.planCreate(exr, act)

	case ResetAction:
		return fd.planReset(exr, act)

	default:
		return fd.planChange(exr, act)
	}
}

func (fd *firehoseDriver) planChange(exr module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	curConf, err := readConfig(exr.Resource, exr.Resource.Spec.Configs)
	if err != nil {
		return nil, err
	}

	enqueueSteps := []string{stepReleaseUpdate}
	switch act.Name {
	case module.UpdateAction:
		newConf, err := readConfig(exr.Resource, act.Params)
		if err != nil {
			return nil, err
		}

		// restore configs that are not user-controlled.
		newConf.DeploymentID = curConf.DeploymentID
		newConf.ChartValues = curConf.ChartValues
		newConf.Namespace = curConf.Namespace
		newConf.Telegraf = curConf.Telegraf

		curConf = newConf

	case ScaleAction:
		var scaleParams struct {
			Replicas int `json:"replicas"`
		}
		if err := json.Unmarshal(act.Params, &scaleParams); err != nil {
			return nil, errors.ErrInvalid.WithMsgf("invalid params for scale action").WithCausef(err.Error())
		} else if scaleParams.Replicas < 1 {
			return nil, errors.ErrInvalid.WithMsgf("replicas must be >= 1")
		}

		curConf.Replicas = scaleParams.Replicas

	case StartAction:
		// nothing to do here since stepReleaseUpdate will automatically
		// start the firehose with last known value of 'replicas'.

	case StopAction:
		enqueueSteps = []string{stepReleaseStop}

	case UpgradeAction:
		// upgrade the chart values to the latest project-level config.
		curConf.ChartValues = &fd.conf.ChartValues
	}

	exr.Resource.Spec.Configs = mustJSON(curConf)
	exr.Resource.State = resource.State{
		Status: resource.StatusPending,
		Output: exr.Resource.State.Output,
		ModuleData: mustJSON(transientData{
			PendingSteps: enqueueSteps,
		}),
	}

	return &module.Plan{
		Reason:   fmt.Sprintf("firehose_%s", act.Name),
		Resource: exr.Resource,
	}, nil
}

func (fd *firehoseDriver) planCreate(exr module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	conf, err := readConfig(exr.Resource, act.Params)
	if err != nil {
		return nil, err
	}

	// set project defaults.
	conf.Telegraf = fd.conf.Telegraf
	conf.Namespace = fd.conf.Namespace
	conf.ChartValues = &fd.conf.ChartValues

	exr.Resource.Spec.Configs = mustJSON(conf)
	exr.Resource.State = resource.State{
		Status: resource.StatusPending,
		Output: mustJSON(Output{
			Namespace:   conf.Namespace,
			ReleaseName: conf.DeploymentID,
		}),
		ModuleData: mustJSON(transientData{
			PendingSteps: []string{stepReleaseCreate},
		}),
	}

	return &module.Plan{
		Reason:   "firehose_create",
		Resource: exr.Resource,
	}, nil
}

func (*firehoseDriver) planReset(exr module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	resetValue, err := kafka.ParseResetParams(act.Params)
	if err != nil {
		return nil, err
	}

	exr.Resource.State = resource.State{
		Status: resource.StatusPending,
		Output: exr.Resource.State.Output,
		ModuleData: mustJSON(transientData{
			ResetOffsetTo: resetValue,
			PendingSteps: []string{
				stepReleaseStop,
				stepKafkaReset,    // reset the consumer group offset value.
				stepReleaseUpdate, // restart the deployment.
			},
		}),
	}
	return &module.Plan{
		Reason:   "firehose_reset",
		Resource: exr.Resource,
	}, nil
}
