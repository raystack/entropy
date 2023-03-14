package firehose

import (
	"context"
	"encoding/json"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kube"
)

type Output struct {
	Namespace   string     `json:"namespace,omitempty"`
	ReleaseName string     `json:"release_name,omitempty"`
	Pods        []kube.Pod `json:"pods,omitempty"`
	Defaults    config     `json:"defaults,omitempty"`
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

	var output Output
	if err := json.Unmarshal(res.Resource.State.Output, &output); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid output json: %v", err)
	}

	pods, err := m.podDetails(ctx, res)
	if err != nil {
		return nil, err
	}

	hc, err := conf.GetHelmReleaseConfig(res.Resource)
	if err != nil {
		return nil, err
	}

	return Output{
		Namespace:   hc.Namespace,
		ReleaseName: hc.Name,
		Pods:        pods,
		Defaults:    output.Defaults,
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

	hc, err := conf.GetHelmReleaseConfig(r)
	if err != nil {
		return nil, err
	}

	kubeCl := kube.NewClient(kubeOut.Configs)
	return kubeCl.GetPodDetails(ctx, hc.Namespace, map[string]string{"app": hc.Name})
}
