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

const keyKubeDependency = "kube_cluster"

var (
	//go:embed schema/config.json
	completeConfigSchema string

	//go:embed schema/scale.json
	scaleActionSchema string

	//go:embed schema/reset.json
	resetActionSchema string
)

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
		driverCfg := driverConfig{
			ChartRepository: "https://odpf.github.io/charts/",
			ChartName:       "firehose",
			ChartVersion:    "0.1.3",
			ImageRepository: "odpf/firehose",
			ImageName:       "firehose",
			ImageTag:        "latest",
			Namespace:       "firehose",
			ImagePullPolicy: "IfNotPresent",
		}

		if err := json.Unmarshal(conf, &driverCfg); err != nil {
			return nil, err
		}

		return &firehoseDriver{
			Config: driverCfg,
		}, nil
	},
}
