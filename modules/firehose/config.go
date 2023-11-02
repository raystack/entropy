package firehose

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/validator"
)

const (
	confSinkType        = "SINK_TYPE"
	confKeyConsumerID   = "SOURCE_KAFKA_CONSUMER_GROUP_ID"
	confKeyKafkaBrokers = "SOURCE_KAFKA_BROKERS"
)

const helmReleaseNameMaxLength = 53

var (
	//go:embed schema/config.json
	configSchemaRaw []byte

	validateConfig = validator.FromJSONSchema(configSchemaRaw)
)

type ScaleParams struct {
	Replicas int `json:"replicas"`
}

type StartParams struct {
	StopTime *time.Time `json:"stop_time"`
}

type Config struct {
	// Stopped flag when set forces the firehose to be stopped on next sync.
	Stopped bool `json:"stopped"`

	// StopTime can be set to schedule the firehose to be stopped at given time.
	StopTime *time.Time `json:"stop_time,omitempty"`

	// Replicas is the number of firehose instances to run.
	Replicas int `json:"replicas"`

	// Namespace is the target namespace where firehose should be deployed.
	// Inherits from driver config.
	Namespace string `json:"namespace,omitempty"`

	// DeploymentID will be used as the release-name for the deployment.
	// Must be shorter than 53 chars if set. If not set, one will be generated
	// automatically.
	DeploymentID string `json:"deployment_id,omitempty"`

	// EnvVariables contains all the firehose environment config values.
	EnvVariables map[string]string `json:"env_variables,omitempty"`

	// ResetOffset represents the value to which kafka consumer offset was set to
	ResetOffset string `json:"reset_offset,omitempty"`

	Limits        UsageSpec     `json:"limits,omitempty"`
	Requests      UsageSpec     `json:"requests,omitempty"`
	Telegraf      *Telegraf     `json:"telegraf,omitempty"`
	ChartValues   *ChartValues  `json:"chart_values,omitempty"`
	InitContainer InitContainer `json:"init_container,omitempty"`
}

type Telegraf struct {
	Enabled bool           `json:"enabled,omitempty"`
	Image   map[string]any `json:"image,omitempty"`
	Config  TelegrafConf   `json:"config,omitempty"`
}

type TelegrafConf struct {
	Output               map[string]any    `json:"output"`
	AdditionalGlobalTags map[string]string `json:"additional_global_tags"`
}

type ChartValues struct {
	ImageTag        string `json:"image_tag" validate:"required"`
	ChartVersion    string `json:"chart_version" validate:"required"`
	ImagePullPolicy string `json:"image_pull_policy" validate:"required"`
}

func readConfig(r resource.Resource, confJSON json.RawMessage, dc driverConf) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(confJSON, &cfg); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json").WithCausef(err.Error())
	}

	cfg.EnvVariables = modules.CloneAndMergeMaps(dc.EnvVariables, cfg.EnvVariables)
	cfg.InitContainer = dc.InitContainer

	if cfg.Replicas <= 0 {
		cfg.Replicas = 1
	}

	if err := validateConfig(confJSON); err != nil {
		return nil, err
	}

	// note: enforce the kubernetes deployment name length limit.
	if len(cfg.DeploymentID) == 0 {
		cfg.DeploymentID = modules.SafeName(fmt.Sprintf("%s-%s", r.Project, r.Name), "-firehose", helmReleaseNameMaxLength)
	} else if len(cfg.DeploymentID) > helmReleaseNameMaxLength {
		return nil, errors.ErrInvalid.WithMsgf("deployment_id must not have more than 53 chars")
	}

	// we name a consumer group by adding a sequence suffix to the deployment name
	// this sequence will later be incremented to name new consumer group while resetting offset
	if consumerID := cfg.EnvVariables[confKeyConsumerID]; consumerID == "" {
		cfg.EnvVariables[confKeyConsumerID] = fmt.Sprintf("%s-1", cfg.DeploymentID)
	}

	rl := dc.RequestsAndLimits[defaultKey]
	if overrides, ok := dc.RequestsAndLimits[cfg.EnvVariables[confSinkType]]; ok {
		rl.Limits = rl.Limits.merge(overrides.Limits)
		rl.Requests = rl.Requests.merge(overrides.Requests)
	}
	cfg.Limits = rl.Limits.merge(cfg.Limits)
	cfg.Requests = rl.Requests.merge(cfg.Requests)

	if cfg.Namespace == "" {
		ns := dc.Namespace[defaultKey]
		if override, ok := dc.Namespace[cfg.EnvVariables[confSinkType]]; ok {
			ns = override
		}
		cfg.Namespace = ns
	}

	return &cfg, nil
}
