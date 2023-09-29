package config

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/validator"
)

const (
	maxJobNameLength = 53
	Default          = "default"
)

var (
	//go:embed schema/config.json
	configSchemaRaw []byte
	validateConfig  = validator.FromJSONSchema(configSchemaRaw)
)

type DriverConf struct {
	Namespace         string                       `json:"namespace"`         // maybe we shouldn't restrict namespace?
	RequestsAndLimits map[string]RequestsAndLimits `json:"requestsAndLimits"` // to use when not provided
	EnvVariables      map[string]string            `json:"env_variables"`
}

type RequestsAndLimits struct {
	Limits   UsageSpec `json:"limits,omitempty"`
	Requests UsageSpec `json:"requests,omitempty"`
}
type UsageSpec struct {
	CPU    string `json:"cpu,omitempty" validate:"required"`
	Memory string `json:"memory,omitempty" validate:"required"`
}

type Config struct {
	Replicas   int32             `json:"replicas"`
	Namespace  string            `json:"namespace"`
	Name       string            `json:"name,omitempty"`
	Containers []Container       `json:"containers,omitempty"`
	JobLabels  map[string]string `json:"job_labels,omitempty"`
	Volumes    []Volume          `json:"volumes,omitempty"`
	TTLSeconds *int32            `json:"ttl_seconds,omitempty"`
}

type Volume struct {
	Name string
	Kind string
}

type Container struct {
	Name              string            `json:"name"`
	Image             string            `json:"image"`
	ImagePullPolicy   string            `json:"image_pull_policy,omitempty"`
	Command           []string          `json:"command,omitempty"`
	SecretsVolumes    []Secret          `json:"secrets_volumes,omitempty"`
	ConfigMapsVolumes []ConfigMap       `json:"config_maps_volumes,omitempty"`
	Limits            UsageSpec         `json:"limits,omitempty"`
	Requests          UsageSpec         `json:"requests,omitempty"`
	EnvConfigMaps     []string          `json:"env_config_maps,omitempty"`
	EnvVariables      map[string]string `json:"env_variables,omitempty"`
}

type Secret struct {
	Name  string `json:"name"`
	Mount string `json:"mount"`
}

type ConfigMap struct {
	Name  string `json:"name"`
	Mount string `json:"mount"`
}

func (dc DriverConf) getDefaultResources() RequestsAndLimits {
	return dc.RequestsAndLimits[Default]
}

func ReadConfig(r resource.Resource, confJSON json.RawMessage, dc DriverConf) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(confJSON, &cfg); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json").WithCausef(err.Error())
	}
	// for each container
	rl := dc.getDefaultResources()
	for i := range cfg.Containers {
		c := &cfg.Containers[i]
		c.EnvVariables = modules.CloneAndMergeMaps(dc.EnvVariables, c.EnvVariables)
		if c.Requests.CPU == "" {
			c.Requests.CPU = rl.Requests.CPU
		}
		if c.Requests.Memory == "" {
			c.Requests.Memory = rl.Requests.Memory
		}
		if c.Limits.CPU == "" {
			c.Limits.CPU = rl.Limits.CPU
		}
		if c.Limits.Memory == "" {
			c.Limits.Memory = rl.Limits.Memory
		}
	}
	if err := validateConfig(confJSON); err != nil {
		return nil, err
	}

	if len(cfg.Name) == 0 {
		cfg.Name = modules.SafeName(fmt.Sprintf("%s-%s", r.Project, r.Name), "-job", maxJobNameLength)
	} else if len(cfg.Name) > maxJobNameLength {
		return nil, errors.ErrInvalid.WithMsgf("Job name must not have more than %d chars", maxJobNameLength)
	}
	cfg.Namespace = dc.Namespace
	if cfg.Replicas < 1 {
		cfg.Replicas = 1
	}
	return &cfg, nil
}
