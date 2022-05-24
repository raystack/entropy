package firehose

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/modules/kubernetes"
	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/pkg/helm"
)

const (
	StopAction  = "stop"
	StartAction = "start"
	ScaleAction = "scale"
)

const (
	helmCreate = "helm_create"
	helmUpdate = "helm_update"

	stateRunning = "RUNNING"
	stateStopped = "STOPPED"
)

const (
	keyReplicaCount   = "replicaCount"
	keyKubeDependency = "kube_cluster"
)

var Module = module.Descriptor{
	Kind: "firehose",
	Dependencies: map[string]string{
		keyKubeDependency: kubernetes.Module.Kind,
	},
	Actions: []module.ActionDesc{
		{
			Name:        module.CreateAction,
			Description: "Creates firehose instance.",
			ParamSchema: completeConfigSchema,
		},
		{
			Name:        module.UpdateAction,
			Description: "Updates an existing firehose instance.",
			ParamSchema: completeConfigSchema,
		},
		{
			Name:        ScaleAction,
			Description: "Scale-up or scale-down an existing firehose instance.",
			ParamSchema: scaleActionSchema,
		},
		{
			Name:        StopAction,
			Description: "Stop firehose and all its components.",
		},
		{
			Name:        StartAction,
			Description: "Start firehose and all its components.",
		},
	},
	Module: &firehoseModule{},
}

type firehoseModule struct{}

func (m *firehoseModule) Plan(_ context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	if act.Name == module.CreateAction {
		return m.planCreate(spec, act)
	}
	return m.planChange(spec, act)
}

func (m *firehoseModule) Sync(_ context.Context, spec module.Spec) (*resource.State, error) {
	r := spec.Resource

	var data moduleData
	if err := json.Unmarshal(r.State.ModuleData, &data); err != nil {
		return nil, err
	}

	if len(data.PendingSteps) == 0 {
		return &resource.State{
			Status:     resource.StatusCompleted,
			Output:     r.State.Output,
			ModuleData: r.State.ModuleData,
		}, nil
	}

	pendingStep := data.PendingSteps[0]
	data.PendingSteps = data.PendingSteps[1:]

	var conf moduleConfig
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(spec.Dependencies[keyKubeDependency].Output, &kubeOut); err != nil {
		return nil, err
	}

	if err := m.helmSync(pendingStep == helmCreate, conf, kubeOut); err != nil {
		return nil, err
	}

	return &resource.State{
		Status: resource.StatusCompleted,
		Output: Output{
			Namespace:   conf.ReleaseConfigs.Namespace,
			ReleaseName: conf.ReleaseConfigs.Name,
		}.JSON(),
		ModuleData: data.JSON(),
	}, nil
}

func (*firehoseModule) planCreate(spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	var reqConf moduleConfig
	if err := json.Unmarshal(act.Params, &reqConf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	} else {
		reqConf.sanitiseAndValidate(r)
	}

	r.Spec.Configs = reqConf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps: []string{helmCreate},
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
		conf = reqConf

	case ScaleAction:
		var scaleParams struct {
			Replicas int `json:"replicas"`
		}
		if err := json.Unmarshal(act.Params, &scaleParams); err != nil {
			return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
		}
		conf.ReleaseConfigs.Values[keyReplicaCount] = scaleParams.Replicas

	case StartAction:
		conf.State = stateRunning

	case StopAction:
		conf.State = stateStopped
	}

	conf.sanitiseAndValidate(r)
	r.Spec.Configs = conf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps: []string{helmUpdate},
		}.JSON(),
	}
	return &r, nil
}

func (*firehoseModule) helmSync(isCreate bool, conf moduleConfig, kube kubernetes.Output) error {
	helmCl := helm.NewClient(&helm.Config{Kubernetes: kube.Configs})

	if conf.State == stateStopped {
		conf.ReleaseConfigs.Values[keyReplicaCount] = 0
	}

	var helmErr error
	if isCreate {
		_, helmErr = helmCl.Create(&conf.ReleaseConfigs)
	} else {
		_, helmErr = helmCl.Update(&conf.ReleaseConfigs)
	}

	return helmErr
}
