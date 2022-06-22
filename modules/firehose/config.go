package firehose

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/helm"
)

const (
	defaultNamespace        = "firehose"
	defaultChartString      = "firehose"
	defaultVersionString    = "0.1.1"
	defaultRepositoryString = "https://odpf.github.io/charts/"
	defaultImagePullPolicy  = "IfNotPresent"
	defaultImageRepository  = "odpf/firehose"
	defaultImageTag         = "latest"
	defaultReplicaCount     = 1
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
	State        string `json:"state"`
	ChartVersion string `json:"chart_version"`
	Firehose     struct {
		Replicas           int               `json:"replicas"`
		KafkaBrokerAddress string            `json:"kafka_broker_address"`
		KafkaTopic         string            `json:"kafka_topic"`
		KafkaConsumerID    string            `json:"kafka_consumer_id"`
		EnvVariables       map[string]string `json:"env_variables"`
	} `json:"firehose"`
}

func (mc *moduleConfig) sanitiseAndValidate() {
	if mc.ChartVersion == "" {
		mc.ChartVersion = defaultVersionString
	}
}

func (mc moduleConfig) GetHelmReleaseConfig(r resource.Resource) *helm.ReleaseConfig {
	rc := helm.DefaultReleaseConfig()
	rc.Name = fmt.Sprintf("%s-%s-firehose", r.Project, r.Name)
	rc.Repository = defaultRepositoryString
	rc.Chart = defaultChartString
	rc.Namespace = defaultNamespace
	rc.ForceUpdate = true
	rc.Version = mc.ChartVersion

	fc := mc.Firehose
	fc.EnvVariables["SOURCE_KAFKA_BROKERS"] = fc.KafkaBrokerAddress
	fc.EnvVariables["SOURCE_KAFKA_TOPIC"] = fc.KafkaTopic
	fc.EnvVariables["SOURCE_KAFKA_CONSUMER_GROUP_ID"] = fc.KafkaConsumerID

	hv := map[string]interface{}{
		"replicaCount": defaultReplicaCount,
		"firehose": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": defaultImageRepository,
				"pullPolicy": defaultImagePullPolicy,
				"tag":        defaultImageTag,
			},
			"config": fc.EnvVariables,
		},
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
