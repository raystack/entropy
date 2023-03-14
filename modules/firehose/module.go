package firehose

import (
	_ "embed"
	"encoding/json"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/modules/kubernetes"
)

const (
	StopAction    = "stop"
	StartAction   = "start"
	ScaleAction   = "scale"
	ResetAction   = "reset"
	UpgradeAction = "upgrade"
)

const (
	releaseCreate = "release_create"
	releaseUpdate = "release_update"
	consumerReset = "consumer_reset"
)

const (
	stateRunning = "RUNNING"
	stateStopped = "STOPPED"
)

const (
	ResetToDateTime = "DATETIME"
	ResetToEarliest = "EARLIEST"
	ResetToLatest   = "LATEST"
)

const keyKubeDependency = "kube_cluster"

var Module = module.Descriptor{
	Kind: "firehose",
	Dependencies: map[string]string{
		keyKubeDependency: kubernetes.Module.Kind,
	},
	Actions: []module.ActionDesc{
		{
			Name:        module.CreateAction,
			Description: "Creates firehose instance.",
			ParamSchema: completeConfigSchema,
		},
		{
			Name:        module.UpdateAction,
			Description: "Updates an existing firehose instance.",
			ParamSchema: completeConfigSchema,
		},
		{
			Name:        ScaleAction,
			Description: "Scale-up or scale-down an existing firehose instance.",
			ParamSchema: scaleActionSchema,
		},
		{
			Name:        StopAction,
			Description: "Stop firehose and all its components.",
		},
		{
			Name:        StartAction,
			Description: "Start firehose and all its components.",
		},
		{
			Name:        ResetAction,
			Description: "Reset firehose kafka consumer group to given timestamp",
			ParamSchema: resetActionSchema,
		},
		{
			Name:        UpgradeAction,
			Description: "Upgrade firehose to current stable version",
		},
	},
	DriverFactory: func(conf json.RawMessage) (module.Driver, error) {
		fm := firehoseModuleWithDefaultConfigs()
		err := json.Unmarshal(conf, fm)
		if err != nil {
			return nil, err
		}
		return fm, nil
	},
}

type firehoseModule struct {
	Config config `json:"config"`
}

type config struct {
	ChartRepository string `json:"chart_repository,omitempty"`
	ChartName       string `json:"chart_name,omitempty"`
	ChartVersion    string `json:"chart_version,omitempty"`
	ImageRepository string `json:"image_repository,omitempty"`
	ImageName       string `json:"image_name,omitempty"`
	ImageTag        string `json:"image_tag,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	ImagePullPolicy string `json:"image_pull_policy,omitempty"`
}

func firehoseModuleWithDefaultConfigs() *firehoseModule {
	return &firehoseModule{
		config{
			ChartRepository: "https://odpf.github.io/charts/",
			ChartName:       "firehose",
			ChartVersion:    "0.1.3",
			ImageRepository: "odpf/firehose",
			ImageName:       "firehose",
			ImageTag:        "latest",
			Namespace:       "firehose",
			ImagePullPolicy: "IfNotPresent",
		},
	}
}
