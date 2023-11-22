package driver

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules"
	"github.com/goto/entropy/modules/job/config"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kube"
	"github.com/goto/entropy/pkg/kube/job"
)

type Driver struct {
	Conf       config.DriverConf
	CreateJob  func(ctx context.Context, conf kube.Config, j *job.Job) error
	SuspendJob func(ctx context.Context, conf kube.Config, j *job.Job) error
	DeleteJob  func(ctx context.Context, conf kube.Config, j *job.Job) error
	StartJob   func(ctx context.Context, conf kube.Config, j *job.Job) error
	GetJobPods func(ctx context.Context, conf kube.Config, j *job.Job, labels map[string]string) ([]kube.Pod, error)
	StreamLogs func(ctx context.Context, kubeConf kube.Config, j *job.Job, filter map[string]string) (<-chan module.LogChunk, error)
}

func (driver *Driver) Plan(_ context.Context, res module.ExpandedResource, act module.ActionRequest) (*resource.Resource, error) {
	switch act.Name {
	case module.CreateAction:
		return driver.planCreate(res, act)
	case SuspendAction:
		return driver.planSuspend(res)
	case module.DeleteAction:
		return driver.planDelete(res)
	case StartAction:
		return driver.planStart(res)
	default:
		return &resource.Resource{}, nil
	}
}

func (driver *Driver) Sync(ctx context.Context, exr module.ExpandedResource) (*resource.State, error) {
	modData, err := ReadTransientData(exr)
	if err != nil {
		return nil, err
	}

	out, err := ReadOutputData(exr)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	conf, err := config.ReadConfig(exr.Resource, exr.Spec.Configs, driver.Conf)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(exr.Dependencies[KeyKubeDependency].Output, &kubeOut); err != nil {
		return nil, errors.ErrInternal.WithMsgf("invalid kube state").WithCausef(err.Error())
	}

	finalState := resource.State{
		Status: resource.StatusPending,
		Output: exr.Resource.State.Output,
	}

	if len(modData.PendingSteps) > 0 {
		pendingStep := modData.PendingSteps[0]
		modData.PendingSteps = modData.PendingSteps[1:]
		switch pendingStep {
		case Create:
			if err := driver.create(ctx, exr.Resource, conf, kubeOut); err != nil {
				return nil, err
			}
		case Suspend:
			if err := driver.suspend(ctx, conf, kubeOut); err != nil {
				return nil, err
			}
		case Delete:
			if err := driver.delete(ctx, conf, kubeOut); err != nil {
				return nil, err
			}
		case Start:
			if err := driver.start(ctx, conf, kubeOut); err != nil {
				return nil, err
			}
		default:
			return nil, errors.ErrInternal.WithMsgf("unknown step: '%s'", pendingStep)
		}

		immediately := time.Now()
		finalState.NextSyncAt = &immediately
		finalState.ModuleData = modules.MustJSON(modData)

		return &finalState, nil
	}

	finalOut, err := driver.refreshOutput(ctx, *conf, *out, kubeOut)
	if err != nil {
		return nil, err
	}
	finalState.Output = finalOut

	finalState.Status = resource.StatusCompleted
	finalState.ModuleData = nil
	return &finalState, nil
}

func (driver *Driver) Output(ctx context.Context, exr module.ExpandedResource) (json.RawMessage, error) {
	output, err := ReadOutputData(exr)
	if err != nil {
		return nil, err
	}

	conf, err := config.ReadConfig(exr.Resource, exr.Spec.Configs, driver.Conf)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(exr.Dependencies[KeyKubeDependency].Output, &kubeOut); err != nil {
		return nil, errors.ErrInternal.WithMsgf("invalid kube state").WithCausef(err.Error())
	}

	return driver.refreshOutput(ctx, *conf, *output, kubeOut)
}
