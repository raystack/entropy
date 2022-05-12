package firehose

import (
	"context"
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/mitchellh/mapstructure"
	gjs "github.com/xeipuuv/gojsonschema"
	"k8s.io/client-go/rest"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/plugins/helm"
	"github.com/odpf/entropy/plugins/kubelogger"
)

const (
	releaseConfigString     = "release_configs"
	replicaCountString      = "replicaCount"
	releaseStateRunning     = "RUNNING"
	releaseStateStopped     = "STOPPED"
	providerKindKubernetes  = "kubernetes"
	defaultRepositoryString = "https://odpf.github.io/charts/"
	defaultChartString      = "firehose"
	defaultVersionString    = "0.1.1"
	defaultNamespace        = "firehose"
)

//go:embed config_schema.json
var configSchemaString string

type Module struct {
	schema *gjs.Schema
}

func (m *Module) Describe() module.Desc {
	return module.Desc{
		Kind: "firehose",
	}
}

func (m *Module) Plan(r resource.Resource, act module.ActionRequest) (*resource.Resource, error) {
	if act.Name == module.CreateAction {
		err := m.validate(r)
		if err != nil {
			return nil, err
		}
		r.State = resource.State{Status: "status_pending", ModuleData: json.RawMessage{"pending": "helm_release"}}
		return &r, nil
	} else if act.Name == "reset" {
		err := m.validate(r)
		if err != nil {
			return nil, err
		}
		r.State = resource.State{Status: "status_pending", ModuleData: json.RawMessage{
			"pending": [
				{"name": "stop_firehose", "params": {}},
				{"name": "reset_offset", "params": {"reset_to": action.params["reset_to"]}},
				{"name": "start_firehose", "params": {}},
			],
		}
} else if act.Name == "stop_firehose" {
	err := m.validate(r)
	if err != nil {
		return nil, err
	}
	releaseConfig, err := getReleaseConfig(r)
	if err != nil {
		return nil, err
	}
	releaseConfig.State = releaseStateStopped
	r.Spec.Configs[releaseConfigString] = releaseConfig
	r.State = resource.State{Status: "status_pending", ModuleData: json.RawMessage{"pending": ["helm_update"]}

} else if act.Name == "start_firehose" {
	err := m.validate(r)
	if err != nil {
		return nil, err
	}
	releaseConfig, err := getReleaseConfig(r)
	if err != nil {
		return nil, err
	}
	releaseConfig.State = releaseStateRunning
	r.Spec.Configs[releaseConfigString] = releaseConfig
	r.State = resource.State{Status: "status_pending", ModuleData: json.RawMessage{"pending": ["helm_update"]}
}

	return &r, nil
}

func (m *Module) Sync(r resource.Resource) (*resource.State, error) {

	return &resource.State{}, nil
}

/*func (m *Module) Apply(r resource.Resource) (resource.Status, error) {
	for _, p := range r.Providers {
		p, err := m.providerSvc.GetByURN(context.TODO(), p.URN)
		if err != nil {
			return resource.StatusError, err
		}

		if p.Kind == providerKindKubernetes {
			releaseConfig, err := getReleaseConfig(r)
			if err != nil {
				return resource.StatusError, err
			}

			if releaseConfig.State == releaseStateStopped {
				releaseConfig.Values[replicaCountString] = 0
			}

			kubeConfig := helm.ToKubeConfig(p.Configs)
			helmConfig := &helm.ProviderConfig{
				Kubernetes: kubeConfig,
			}
			helmProvider := helm.NewProvider(helmConfig)
			_, err = helmProvider.Release(releaseConfig)
			if err != nil {
				return resource.StatusError, err
			}
		}
	}

	return resource.StatusCompleted, nil
}*/

func (m *Module) validate(r resource.Resource) error {
	resourceLoader := gjs.NewGoLoader(r.Spec.Configs)
	result, err := m.schema.Validate(resourceLoader)
	if err != nil {
		return errors.ErrInvalid.WithCausef(err.Error())
	}

	if !result.Valid() {
		var errorStrings []string
		for _, resultErr := range result.Errors() {
			errorStrings = append(errorStrings, resultErr.String())
		}
		errorString := strings.Join(errorStrings, "\n")
		return errors.New(errorString)
	}
	return nil
}
/*
func (m *Module) Act(r resource.Resource, action string, params map[string]interface{}) (map[string]interface{}, error) {
	releaseConfig, err := getReleaseConfig(r)
	if err != nil {
		return nil, err
	}

	switch action {
	case "start":
		releaseConfig.State = releaseStateRunning

	case "stop":
		releaseConfig.State = releaseStateStopped

	case "scale":
		releaseConfig.Values[replicaCountString] = params[replicaCountString]
	}
	r.Configs[releaseConfigString] = releaseConfig
	return r.Configs, nil
}
*/
func (m *Module) Log(ctx context.Context, r resource.Resource, filter map[string]string) (<-chan resource.LogChunk, error) {
	var releaseConfig helm.ReleaseConfig
	if err := mapstructure.Decode(r.Configs[releaseConfigString], &releaseConfig); err != nil {
		return nil, errors.New("unable to parse configs")
	}

	cfg, err := m.loadKubeConfig(ctx, r.Providers)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	if filter == nil {
		filter = make(map[string]string)
	}
	filter["app"] = r.URN

	return kubelogger.GetStreamingLogs(ctx, defaultNamespace, filter, *cfg)
}

func (m *Module) loadKubeConfig(ctx context.Context, providers []resource.ProviderSelector) (*rest.Config, error) {
	for _, providerSelector := range providers {
		p, err := m.providerSvc.GetByURN(ctx, providerSelector.URN)
		if err != nil {
			return nil, err
		}

		if p.Kind == providerKindKubernetes {
			var kubeCfg kubeConfig
			if err := mapstructure.Decode(p.Configs, &kubeCfg); err != nil {
				return nil, err
			}
			return kubeCfg.ToRESTConfig(), nil
		}
	}

	return nil, errors.ErrInternal.WithCausef("kubernetes provider not found in resource")
}

func getReleaseConfig(r resource.Resource) (*helm.ReleaseConfig, error) {
	releaseConfig := helm.DefaultReleaseConfig()
	releaseConfig.Repository = defaultRepositoryString
	releaseConfig.Chart = defaultChartString
	releaseConfig.Version = defaultVersionString
	releaseConfig.Namespace = defaultNamespace
	releaseConfig.Name = r.URN
	err := mapstructure.Decode(r.Configs[releaseConfigString], &releaseConfig)
	if err != nil {
		return releaseConfig, err
	}

	return releaseConfig, nil
}

type kubeConfig struct {
	Host          string `mapstructure:"host"`
	ClientKey     string `mapstructure:"clientKey"`
	ClientCert    string `mapstructure:"clientCertificate"`
	ClusterCACert string `mapstructure:"clusterCACertificate"`
}

func (kc kubeConfig) ToRESTConfig() *rest.Config {
	return &rest.Config{
		Host: kc.Host,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   []byte(kc.ClusterCACert),
			KeyData:  []byte(kc.ClientKey),
			CertData: []byte(kc.ClientCert),
		},
	}
}
