package kubernetes

import (
	"context"
	"encoding/json"

	"k8s.io/client-go/kubernetes"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kube"
)

type kubeDriver struct {
	Tolerations map[string][]Toleration `json:"tolerations"`
}

func (m *kubeDriver) Plan(ctx context.Context, res module.ExpandedResource,
	act module.ActionRequest,
) (*resource.Resource, error) {
	res.Resource.Spec = resource.Spec{
		Configs:      act.Params,
		Dependencies: nil,
	}

	output, err := m.Output(ctx, res)
	if err != nil {
		return nil, err
	}

	res.Resource.State = resource.State{
		Status: resource.StatusCompleted,
		Output: output,
	}

	return &res.Resource, nil
}

func (*kubeDriver) Sync(_ context.Context, res module.ExpandedResource) (*resource.State, error) {
	return &resource.State{
		Status:     resource.StatusCompleted,
		Output:     res.Resource.State.Output,
		ModuleData: nil,
	}, nil
}

func (m *kubeDriver) Output(ctx context.Context, res module.ExpandedResource) (json.RawMessage, error) {
	conf := kube.DefaultClientConfig()
	if err := json.Unmarshal(res.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid json config value").WithCausef(err.Error())
	} else if err := conf.Sanitise(); err != nil {
		return nil, err
	}

	restConfig, err := conf.RESTConfig(ctx)
	if err != nil {
		return nil, errors.ErrInternal.WithMsgf("failed to create new kube client on kube driver output").WithCausef(err.Error())
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.ErrInvalid.WithMsgf("failed to create client: %v", err)
	}

	info, err := clientSet.ServerVersion()
	if err != nil {
		return nil, errors.ErrInvalid.WithMsgf("failed to fetch server info: %v", err)
	}

	return Output{
		Configs:     conf,
		ServerInfo:  *info,
		Tolerations: m.Tolerations,
	}.JSON(), nil
}
