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
	ResetAction  = "reset"
	StopAction   = "stop"
	StartAction  = "start"
	ScaleAction  = "scale"
	DeleteAction = "delete"
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

func (m *firehoseModule) Plan(ctx context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	switch act.Name {
	case module.CreateAction:
		return m.planCreate(ctx, spec, act)

	case module.UpdateAction:
		return m.planUpdate(ctx, spec, act)

	case ScaleAction:
		return m.planScale(ctx, spec, act)

	case StartAction:
		return m.planStart(ctx, spec, act)

	case StopAction:
		return m.planStop(ctx, spec, act)
	}

	return &r, nil
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

	helmCl := helm.NewClient(&helm.Config{
		Kubernetes: kubeOut.Configs,
	})

	if conf.State == stateStopped {
		conf.ReleaseConfigs.Values[keyReplicaCount] = 0
	}

	var helmErr error
	if pendingStep == helmCreate {
		_, helmErr = helmCl.Create(&conf.ReleaseConfigs)
	} else if pendingStep == helmUpdate {
		_, helmErr = helmCl.Update(&conf.ReleaseConfigs)
	}

	if helmErr != nil {
		return nil, helmErr
	}

	return &resource.State{
		Status: resource.StatusCompleted,
		Output: Output{
			// TODO: populate the outputs as required.
		}.JSON(),
		ModuleData: data.JSON(),
	}, nil
}

func (m *firehoseModule) planCreate(ctx context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	var reqConf moduleConfig
	if err := json.Unmarshal(act.Params, &reqConf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	} else if err := reqConf.sanitiseAndValidate(r); err != nil {
		return nil, err
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

func (m *firehoseModule) planUpdate(ctx context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	var reqConf moduleConfig
	if err := json.Unmarshal(act.Params, &reqConf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	if err := reqConf.sanitiseAndValidate(r); err != nil {
		return nil, err
	}

	r.Spec.Configs = reqConf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps: []string{helmUpdate},
		}.JSON(),
	}
	return &r, nil
}

func (m *firehoseModule) planScale(ctx context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	var scaleParams struct {
		Replicas int `json:"replicas"`
	}
	if err := json.Unmarshal(act.Params, &scaleParams); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	r := spec.Resource

	var conf moduleConfig
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	conf.ReleaseConfigs.Values[keyReplicaCount] = scaleParams.Replicas
	if err := conf.sanitiseAndValidate(r); err != nil {
		return nil, err
	}

	r.Spec.Configs = conf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps: []string{helmUpdate},
		}.JSON(),
	}
	return &r, nil
}

func (m *firehoseModule) planStart(ctx context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	var data moduleData
	if err := json.Unmarshal(r.State.ModuleData, &data); err != nil {
		return nil, err
	}

	var curConf moduleConfig
	if err := json.Unmarshal(r.Spec.Configs, &curConf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	curConf.State = stateRunning

	r.Spec.Configs = curConf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps: []string{helmUpdate},
		}.JSON(),
	}
	return &r, nil
}

func (m *firehoseModule) planStop(_ context.Context, spec module.Spec, _ module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	var data moduleData
	if err := json.Unmarshal(r.State.ModuleData, &data); err != nil {
		return nil, err
	}

	var curConf moduleConfig
	if err := json.Unmarshal(r.Spec.Configs, &curConf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	curConf.State = stateStopped

	r.Spec.Configs = curConf.JSON()
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: moduleData{
			PendingSteps: []string{helmUpdate},
		}.JSON(),
	}
	return &r, nil
}
