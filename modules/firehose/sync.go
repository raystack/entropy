package firehose

import (
	"context"
	"encoding/json"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/modules/kubernetes"
	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/pkg/helm"
)

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

	if err := m.releaseSync(pendingStep == releaseCreate, conf, r, kubeOut); err != nil {
		return nil, err
	}

	return &resource.State{
		Status: resource.StatusCompleted,
		Output: Output{
			Namespace:   conf.GetHelmReleaseConfig(r).Namespace,
			ReleaseName: conf.GetHelmReleaseConfig(r).Name,
		}.JSON(),
		ModuleData: data.JSON(),
	}, nil
}

func (*firehoseModule) releaseSync(isCreate bool, conf moduleConfig, r resource.Resource, kube kubernetes.Output) error {
	helmCl := helm.NewClient(&helm.Config{Kubernetes: kube.Configs})

	if conf.State == stateStopped {
		conf.FirehoseConfigs.Replicas = 0
	}

	var helmErr error
	if isCreate {
		_, helmErr = helmCl.Create(conf.GetHelmReleaseConfig(r))
	} else {
		_, helmErr = helmCl.Update(conf.GetHelmReleaseConfig(r))
	}

	return helmErr
}
