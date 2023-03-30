package firehose

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/validator"
)

const (
	confKeyConsumerID   = "SOURCE_KAFKA_CONSUMER_GROUP_ID"
	confKeyKafkaBrokers = "SOURCE_KAFKA_BROKERS"
)

const helmReleaseNameMaxLength = 53

var (
	//go:embed schema/config.json
	configSchemaRaw []byte

	validateConfig = validator.FromJSONSchema(configSchemaRaw)
)

type Config struct {
	Stopped      bool              `json:"stopped"`
	StopTime     *time.Time        `json:"stop_time,omitempty"`
	Telegraf     *Telegraf         `json:"telegraf,omitempty"`
	Replicas     int               `json:"replicas"`
	Namespace    string            `json:"namespace,omitempty"`
	DeploymentID string            `json:"deployment_id,omitempty"`
	ChartValues  *ChartValues      `json:"chart_values,omitempty"`
	EnvVariables map[string]string `json:"env_variables,omitempty"`
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

func readConfig(r resource.Resource, confJSON json.RawMessage) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(confJSON, &cfg); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json").WithCausef(err.Error())
	}

	if cfg.Replicas <= 0 {
		cfg.Replicas = 1
	}

	if err := validateConfig(confJSON); err != nil {
		return nil, err
	}

	// note: enforce the kubernetes deployment name length limit.
	if len(cfg.DeploymentID) == 0 {
		cfg.DeploymentID = safeReleaseName(fmt.Sprintf("%s-%s", r.Project, r.Name))
	} else if len(cfg.DeploymentID) > helmReleaseNameMaxLength {
		return nil, errors.ErrInvalid.WithMsgf("deployment_id must not have more than 53 chars")
	}

	if consumerID := cfg.EnvVariables[confKeyConsumerID]; consumerID == "" {
		cfg.EnvVariables[confKeyConsumerID] = fmt.Sprintf("%s-0001", cfg.DeploymentID)
	}

	return &cfg, nil
}

func safeReleaseName(concatName string) string {
	const randomHashLen = 6
	suffix := "-firehose"

	// remove suffix if already there.
	concatName = strings.TrimSuffix(concatName, suffix)

	if len(concatName) <= helmReleaseNameMaxLength-len(suffix) {
		return concatName + suffix
	}

	val := sha256.Sum256([]byte(concatName))
	hash := fmt.Sprintf("%x", val)
	suffix = fmt.Sprintf("-%s%s", hash[:randomHashLen], suffix)

	// truncate and make room for the suffix. also trim any leading, trailing
	// hyphens to prevent '--' (not allowed in deployment names).
	truncLen := helmReleaseNameMaxLength - len(suffix)
	truncated := concatName[0:truncLen]
	truncated = strings.Trim(truncated, "-")
	return truncated + suffix
}
