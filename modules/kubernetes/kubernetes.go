package kubernetes

import (
	"context"
	"encoding/json"

	"k8s.io/client-go/kubernetes"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

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
	Module: &kubeModule{},
}

type kubeModule struct{}

func (k *kubeModule) Plan(_ context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	res := spec.Resource

	var conf moduleConf
	if err := json.Unmarshal(act.Params, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid json config value").WithCausef(err.Error())
	}

	clientSet, err := kubernetes.NewForConfig(conf.toRESTConfig())
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
		Output: map[string]interface{}{
			"host":           conf.Host,
			"ca_data":        conf.CertData,
			"key_data":       conf.KeyData,
			"cert_data":      conf.CertData,
			"client_timeout": conf.ClientTimeout,
			"server_info": map[string]interface{}{
				"platform":    info.Platform,
				"major":       info.Major,
				"minor":       info.Minor,
				"git_version": info.GitVersion,
				"git_commit":  info.GitCommit,
			},
		},
	}
	return &res, nil
}

func (k *kubeModule) Sync(_ context.Context, spec module.Spec) (*resource.State, error) {
	return &resource.State{
		Status:     resource.StatusCompleted,
		Output:     spec.Resource.State.Output,
		ModuleData: nil,
	}, nil
}
