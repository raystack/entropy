package firehose

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	gjs "github.com/xeipuuv/gojsonschema"

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
	defaultNamespaceString  = "firehose"
)

//go:embed config_schema.json
var configSchemaString string

func New(providerRepository provider.Repository) *Module {
	schemaLoader := gjs.NewStringLoader(configSchemaString)
	schema, err := gjs.NewSchema(schemaLoader)
	if err != nil {
		return nil
	}
	return &Module{
		schema:             schema,
		providerRepository: providerRepository,
	}
}

type Module struct {
	schema             *gjs.Schema
	providerRepository provider.Repository
}

func (m *Module) ID() string { return "firehose" }

func (m *Module) Apply(r resource.Resource) (resource.Status, error) {
	for _, p := range r.Providers {
		p, err := m.providerRepository.GetByURN(context.TODO(), p.URN)
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
				return resource.StatusError, nil
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

func (m *Module) Log(ctx context.Context, r *resource.Resource, filter map[string]string) (<-chan resource.LogChunk, error) {
	var releaseConfig *helm.ReleaseConfig
	if err := mapstructure.Decode(r.Configs[releaseConfigString], &releaseConfig); err != nil {
		return nil, errors.New("unable to parse configs")
	}

	var selectors []string
	for k, v := range filter {
		s := fmt.Sprintf("%s=%s", k, v)
		selectors = append(selectors, s)
	}
	selector := strings.Join(selectors, ",")

	logs, err := kubelogger.GetStreamingLogs(ctx, releaseConfig.Namespace, selector)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func getReleaseConfig(r resource.Resource) (*helm.ReleaseConfig, error) {
	releaseConfig := helm.DefaultReleaseConfig()
	releaseConfig.Repository = defaultRepositoryString
	releaseConfig.Chart = defaultChartString
	releaseConfig.Version = defaultVersionString
	releaseConfig.Namespace = defaultNamespaceString
	releaseConfig.Name = r.URN
	err := mapstructure.Decode(r.Configs[releaseConfigString], &releaseConfig)
	if err != nil {
		return releaseConfig, err
	}

	return releaseConfig, nil
}
