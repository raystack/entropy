package kubernetes

import (
	"context"
	_ "embed"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

//go:embed config_schema.json
var configSchema string

var KubeModule = module.Descriptor{
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

	cfg := &rest.Config{
		Host: act.Params["host"].(string),
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   []byte(act.Params["ca_cert"].(string)),
			KeyData:  []byte(act.Params["client_key"].(string)),
			CertData: []byte(act.Params["client_cert"].(string)),
		},
		Timeout: time.Duration(act.Params["client_timeout"].(float64)) * time.Millisecond,
	}

	clientSet, err := kubernetes.NewForConfig(cfg)
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
		Output: mergeMap(
			act.Params,
			map[string]interface{}{
				"server_info": map[string]interface{}{
					"platform":    info.Platform,
					"major":       info.Major,
					"minor":       info.Minor,
					"git_version": info.GitVersion,
					"git_commit":  info.GitCommit,
				},
			},
		),
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

func mergeMap(m1, m2 map[string]interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range m1 {
		m[k] = v
	}
	for k, v := range m2 {
		m[k] = v
	}
	return m
}
