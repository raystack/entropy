package firehose

import (
	"context"
	"encoding/json"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/modules/kubernetes"
	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/pkg/kube"
)

type Output struct {
	Namespace   string     `json:"namespace"`
	ReleaseName string     `json:"release_name"`
	Pods        []kube.Pod `json:"pods"`
}

func (out Output) JSON() []byte {
	b, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	return b
}

func (m *firehoseModule) Output(ctx context.Context, res module.ExpandedResource) (json.RawMessage, error) {
	var conf moduleConfig
	if err := json.Unmarshal(res.Resource.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	pods, err := m.podDetails(ctx, res)
	if err != nil {
		return nil, err
	}

	return Output{
		Namespace:   conf.GetHelmReleaseConfig(res.Resource).Namespace,
		ReleaseName: conf.GetHelmReleaseConfig(res.Resource).Name,
		Pods:        pods,
	}.JSON(), nil
}

func (*firehoseModule) podDetails(ctx context.Context, res module.ExpandedResource) ([]kube.Pod, error) {
	r := res.Resource

	var conf moduleConfig
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(res.Dependencies[keyKubeDependency].Output, &kubeOut); err != nil {
		return nil, err
	}

	kubeCl := kube.NewClient(kubeOut.Configs)
	return kubeCl.GetPodDetails(ctx, defaultNamespace, map[string]string{"app": conf.GetHelmReleaseConfig(r).Name})
}
