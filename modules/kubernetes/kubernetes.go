package kubernetes

import (
	"context"
	_ "embed"
	"encoding/json"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/pkg/kube"
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

func (*kubeModule) Plan(_ context.Context, spec module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	res := spec.Resource

	conf := kube.DefaultClientConfig()
	if err := json.Unmarshal(act.Params, &conf); err != nil {
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

	res.Spec = resource.Spec{
		Configs:      act.Params,
		Dependencies: nil,
	}
	res.State = resource.State{
		Status: resource.StatusCompleted,
		Output: Output{
			Configs:    conf,
			ServerInfo: *info,
		}.JSON(),
	}
	return &module.Plan{Resource: res}, nil
}

func (*kubeModule) Sync(_ context.Context, spec module.ExpandedResource) (*resource.State, error) {
	return &resource.State{
		Status:     resource.StatusCompleted,
		Output:     spec.Resource.State.Output,
		ModuleData: nil,
	}, nil
}

func (out Output) JSON() []byte {
	b, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	return b
}
