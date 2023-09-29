package driver

import (
	"time"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules"
	"github.com/goto/entropy/modules/job/config"
)

const (
	KeyKubeDependency = "kube_cluster"
	SuspendAction     = "suspend"
	StartAction       = "start"
)

const (
	Create  PendingStep = "create"
	Suspend PendingStep = "suspend"
	Delete  PendingStep = "delete"
	Start   PendingStep = "start"
)

type (
	PendingStep   string
	IgnoreError   bool
	TransientData struct {
		PendingSteps []PendingStep `json:"pending_steps"`
	}
)

func (driver *Driver) planCreate(exr module.ExpandedResource, act module.ActionRequest) (*resource.Resource, error) {
	conf, err := config.ReadConfig(exr.Resource, act.Params, driver.Conf)
	if err != nil {
		return nil, err
	}
	return driver.planPendingWithConf(conf, exr, []PendingStep{Create})
}

func (driver *Driver) planPendingWithConf(conf *config.Config, exr module.ExpandedResource, steps []PendingStep) (*resource.Resource, error) {
	conf.Namespace = driver.Conf.Namespace
	immediately := time.Now()
	exr.Resource.Spec.Configs = modules.MustJSON(conf)
	exr.Resource.State = resource.State{
		Status: resource.StatusPending,
		Output: modules.MustJSON(Output{
			Namespace: conf.Namespace,
			JobName:   conf.Name,
		}),
		NextSyncAt: &immediately,
		ModuleData: modules.MustJSON(TransientData{
			PendingSteps: steps,
		}),
	}
	return &exr.Resource, nil
}

func (driver *Driver) planPendingWithExistingResource(exr module.ExpandedResource, step []PendingStep) (*resource.Resource, error) {
	conf, err := config.ReadConfig(exr.Resource, exr.Resource.Spec.Configs, driver.Conf)
	if err != nil {
		return nil, err
	}
	return driver.planPendingWithConf(conf, exr, step)
}

func (driver *Driver) planDelete(exr module.ExpandedResource) (*resource.Resource, error) {
	return driver.planPendingWithExistingResource(exr, []PendingStep{Delete})
}

func (driver *Driver) planSuspend(exr module.ExpandedResource) (*resource.Resource, error) {
	return driver.planPendingWithExistingResource(exr, []PendingStep{Suspend})
}

func (driver *Driver) planStart(exr module.ExpandedResource) (*resource.Resource, error) {
	return driver.planPendingWithExistingResource(exr, []PendingStep{Start})
}
