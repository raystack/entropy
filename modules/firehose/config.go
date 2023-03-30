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

const kubeDeploymentNameLengthLimit = 53

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
	Enabled   bool           `json:"enabled"`
	Image     string         `json:"image,omitempty"`
	ConfigTpl string         `json:"config_tpl,omitempty"`
	TplValues map[string]any `json:"tpl_values,omitempty"`
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
		cfg.DeploymentID = generateSafeReleaseName(r.Project, r.Name)
	} else if len(cfg.DeploymentID) > kubeDeploymentNameLengthLimit {
		return nil, errors.ErrInvalid.WithMsgf("deployment_id must be shorter than 53 chars")
	}

	if consumerID := cfg.EnvVariables[confKeyConsumerID]; consumerID == "" {
		cfg.EnvVariables[confKeyConsumerID] = fmt.Sprintf("%s-0001", cfg.DeploymentID)
	}

	return &cfg, nil
}

func generateSafeReleaseName(project, name string) string {
	const prefix = "firehose-"
	const randomHashLen = 6

	releaseName := fmt.Sprintf("%s%s-%s", prefix, project, name)
	if len(releaseName) >= kubeDeploymentNameLengthLimit {
		releaseName = strings.Trim(releaseName[:kubeDeploymentNameLengthLimit-randomHashLen-1], "-")

		val := sha256.Sum256([]byte(releaseName))
		hash := fmt.Sprintf("%x", val)
		releaseName = releaseName + "-" + hash[:randomHashLen]
	}

	return releaseName
}
