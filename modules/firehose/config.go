package firehose

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/pkg/helm"
)

var (
	//go:embed schema/config.json
	completeConfigSchema string

	//go:embed schema/scale.json
	scaleActionSchema string

	//go:embed schema/reset.json
	resetActionSchema string
)

type moduleConfig struct {
	State    string                 `json:"state"`
	StopTime *time.Time             `json:"stop_time"`
	Telegraf map[string]interface{} `json:"telegraf"`
	Firehose struct {
		Replicas           int               `json:"replicas"`
		KafkaBrokerAddress string            `json:"kafka_broker_address"`
		KafkaTopic         string            `json:"kafka_topic"`
		KafkaConsumerID    string            `json:"kafka_consumer_id"`
		EnvVariables       map[string]string `json:"env_variables"`
	} `json:"firehose"`
}

func (mc *moduleConfig) validate() error {
	if mc.StopTime != nil && mc.StopTime.Before(time.Now()) {
		return errors.ErrInvalid.
			WithMsgf("value for stop_time must be greater than current time")
	}
	return nil
}

func (mc moduleConfig) GetHelmReleaseConfig(r resource.Resource, defaults config) *helm.ReleaseConfig {
	rc := helm.DefaultReleaseConfig()
	rc.Name = fmt.Sprintf("%s-%s-firehose", r.Project, r.Name)
	rc.Repository = defaults.ChartRepository
	rc.Chart = defaults.ChartName
	rc.Namespace = defaults.Namespace
	rc.ForceUpdate = true
	rc.Version = defaults.ChartVersion

	fc := mc.Firehose
	fc.EnvVariables["SOURCE_KAFKA_BROKERS"] = fc.KafkaBrokerAddress
	fc.EnvVariables["SOURCE_KAFKA_TOPIC"] = fc.KafkaTopic
	fc.EnvVariables["SOURCE_KAFKA_CONSUMER_GROUP_ID"] = fc.KafkaConsumerID

	hv := map[string]interface{}{
		"replicaCount": mc.Firehose.Replicas,
		"firehose": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": defaults.ImageRepository,
				"pullPolicy": defaults.ImagePullPolicy,
				"tag":        defaults.ImageTag,
			},
			"config": fc.EnvVariables,
		},
	}
	if len(mc.Telegraf) > 0 {
		hv["telegraf"] = mc.Telegraf
	}
	rc.Values = hv

	return rc
}

func (mc moduleConfig) JSON() []byte {
	b, err := json.Marshal(mc)
	if err != nil {
		panic(err)
	}
	return b
}
