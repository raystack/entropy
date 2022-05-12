package firehose

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/helm"
	gjs "github.com/xeipuuv/gojsonschema"
)

const (
	releaseConfigString = "release_configs"
	stateString         = "state"
	replicaCountString  = "replicaCount"
	releaseStateRunning = "RUNNING"
	releaseStateStopped = "STOPPED"

	providerKindKubernetes = "kubernetes"

	defaultRepositoryString = "https://odpf.github.io/charts/"
	defaultChartString      = "firehose"
	defaultVersionString    = "0.1.1"
	defaultNamespace        = "firehose"

	ResetAction  = "reset"
	StopAction   = "stop"
	StartAction  = "start"
	ScaleAction  = "scale"
	DeleteAction = "delete"
)

//go:embed create_schema.json
var createActionSchema string

//go:embed create_schema.json
var updateActionSchema string

//go:embed scale_schema.json
var scaleActionSchema string

type Module struct {
	schema *gjs.Schema
}

type moduleData struct {
	PendingList []Pending
}

type Pending struct {
	Name   string
	Params map[string]interface{}
}

func (md moduleData) JSON() (json.RawMessage, error) {
	bytes, err := json.Marshal(md)
	return bytes, err
}

func (m *Module) Describe() module.Desc {
	return module.Desc{
		Kind: "firehose",
		Actions: []module.ActionDesc{
			{
				Name:        "create",
				Description: "creates firehose instance",
				ParamSchema: createActionSchema,
			},
			{
				Name:        "update",
				Description: "updates firehose instance",
				ParamSchema: updateActionSchema,
			},
			{
				Name:        "scale",
				Description: "scales firehose instance to given replicas",
				ParamSchema: scaleActionSchema,
			},
			{
				Name:        "start",
				Description: "starts firehose instance",
			},
			{
				Name:        "stop",
				Description: "stops firehose instance",
			},
		},
	}
}

func (m *Module) Plan(ctx context.Context, spec module.Spec, act module.ActionRequest) (*resource.Resource, error) {
	r := spec.Resource

	switch act.Name {
	case module.CreateAction:
		r.Spec.Configs = act.Params

		moduleDataJson, err := moduleData{
			PendingList: []Pending{
				{Name: "helm_create"},
			},
		}.JSON()
		if err != nil {
			return nil, err
		}

		r.State = resource.State{
			Status:     resource.StatusPending,
			ModuleData: moduleDataJson,
		}
	case module.UpdateAction:
		r.Spec.Configs = act.Params
		moduleDataJson, err := moduleData{
			PendingList: []Pending{
				{Name: "helm_update"},
			},
		}.JSON()
		if err != nil {
			return nil, err
		}

		r.State = resource.State{
			Status:     resource.StatusPending,
			ModuleData: moduleDataJson,
		}
	case StopAction:
		r.Spec.Configs[stateString] = releaseStateStopped
		moduleDataJson, err := moduleData{
			PendingList: []Pending{
				{Name: "helm_update"},
			},
		}.JSON()
		if err != nil {
			return nil, err
		}

		r.State = resource.State{
			Status:     resource.StatusPending,
			ModuleData: moduleDataJson,
		}
	case StartAction:
		r.Spec.Configs[stateString] = releaseStateRunning
		moduleDataJson, err := moduleData{
			PendingList: []Pending{
				{Name: "helm_update"},
			},
		}.JSON()
		if err != nil {
			return nil, err
		}

		r.State = resource.State{
			Status:     resource.StatusPending,
			ModuleData: moduleDataJson,
		}
	case ScaleAction:
		releaseConfig, err := getReleaseConfig(r)
		if err != nil {
			return nil, err
		}
		releaseConfig.Values[replicaCountString] = act.Params[replicaCountString]
		moduleDataJson, err := moduleData{
			PendingList: []Pending{
				{Name: "helm_update"},
			},
		}.JSON()
		if err != nil {
			return nil, err
		}

		r.State = resource.State{
			Status:     resource.StatusPending,
			ModuleData: moduleDataJson,
		}
	case DeleteAction:
		moduleDataJson, err := moduleData{
			PendingList: []Pending{
				{Name: "helm_delete"},
			},
		}.JSON()
		if err != nil {
			return nil, err
		}

		r.State = resource.State{
			Status:     resource.StatusDeleted,
			ModuleData: moduleDataJson,
		}
	}

	return &r, nil
}

func (m *Module) Sync(ctx context.Context, spec module.Spec) (*resource.Output, error) {
	return &resource.Output{"foo": "bar"}, nil
}

func getReleaseConfig(r resource.Resource) (*helm.ReleaseConfig, error) {
	releaseConfig := helm.DefaultReleaseConfig()
	releaseConfig.Repository = defaultRepositoryString
	releaseConfig.Chart = defaultChartString
	releaseConfig.Version = defaultVersionString
	releaseConfig.Namespace = defaultNamespace
	releaseConfig.Name = r.URN
	err := mapstructure.Decode(r.Spec.Configs[releaseConfigString], &releaseConfig)
	if err != nil {
		return releaseConfig, err
	}

	return releaseConfig, nil
}
