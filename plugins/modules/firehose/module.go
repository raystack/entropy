package firehose

import (
	"context"
	_ "embed"
	"strings"

	"github.com/mitchellh/mapstructure"
	gjs "github.com/xeipuuv/gojsonschema"
	"k8s.io/client-go/rest"

	"github.com/odpf/entropy/core/provider"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/plugins/providers/helm"
	"github.com/odpf/entropy/plugins/providers/kubelogger"
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

func New(providerSvc providerService) *Module {
	schemaLoader := gjs.NewStringLoader(configSchemaString)
	schema, err := gjs.NewSchema(schemaLoader)
	if err != nil {
		return nil
	}
	return &Module{
		schema:      schema,
		providerSvc: providerSvc,
	}
}

type providerService interface {
	GetByURN(ctx context.Context, urn string) (*provider.Provider, error)
}

type Module struct {
	schema      *gjs.Schema
	providerSvc providerService
}

func (m *Module) ID() string { return "firehose" }

func (m *Module) Apply(r resource.Resource) (resource.Status, error) {
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
}

func (m *Module) Validate(r resource.Resource) error {
	resourceLoader := gjs.NewGoLoader(r.Configs)
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
