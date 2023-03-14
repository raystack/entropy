package kubernetes

import (
	"context"
	_ "embed"
	"encoding/json"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kube"
)

//go:embed config_schema.json
var configSchema string

var Module = module.Descriptor{
	Kind: "kubernetes",
	Actions: []module.ActionDesc{
		{
			Name:        module.CreateAction,
			ParamSchema: configSchema,
		},
		{
			Name:        module.UpdateAction,
			ParamSchema: configSchema,
		},
	},
	DriverFactory: func(conf json.RawMessage) (module.Driver, error) {
		return &kubeModule{}, nil
	},
}

type kubeModule struct{}

type Output struct {
	Configs    kube.Config  `json:"configs"`
	ServerInfo version.Info `json:"server_info"`
}

func (m *kubeModule) Plan(ctx context.Context, res module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
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
	return &module.Plan{Resource: res.Resource, Reason: "kubernetes cluster details updated"}, nil
}

func (*kubeModule) Sync(_ context.Context, res module.ExpandedResource) (*resource.State, error) {
	return &resource.State{
		Status:     resource.StatusCompleted,
		Output:     res.Resource.State.Output,
		ModuleData: nil,
	}, nil
}

func (*kubeModule) Output(_ context.Context, res module.ExpandedResource) (json.RawMessage, error) {
	conf := kube.DefaultClientConfig()
	if err := json.Unmarshal(res.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid json config value").WithCausef(err.Error())
	}

	clientSet, err := kubernetes.NewForConfig(conf.RESTConfig())
	if err != nil {
		return nil, errors.ErrInvalid.WithMsgf("failed to create client: %v", err)
	}

	info, err := clientSet.ServerVersion()
	if err != nil {
		return nil, errors.ErrInvalid.WithMsgf("failed to fetch server info: %v", err)
	}

	return Output{
		Configs:    conf,
		ServerInfo: *info,
	}.JSON(), nil
}

func (out Output) JSON() []byte {
	b, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	return b
}
