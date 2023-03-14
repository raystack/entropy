package firehose

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/helm"
)

const (
	firehoseConsumerIDStartingSequence = "0001"
	kubeDeploymentNameLengthLimit      = 63
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
		DeploymentID       string            `json:"deployment_id,omitempty"`
	} `json:"firehose"`
}

func (mc *moduleConfig) validateAndSanitize(r resource.Resource) error {
	if mc.StopTime != nil && mc.StopTime.Before(time.Now()) {
		return errors.ErrInvalid.
			WithMsgf("value for stop_time must be greater than current time")
	}

	if mc.Firehose.KafkaConsumerID == "" {
		mc.Firehose.KafkaConsumerID = fmt.Sprintf("%s-%s-%s-%s",
			r.Project, r.Name, "-firehose-", firehoseConsumerIDStartingSequence)
	}

	return nil
}

func (mc *moduleConfig) GetHelmReleaseConfig(r resource.Resource) (*helm.ReleaseConfig, error) {
	var output Output
	err := json.Unmarshal(r.State.Output, &output)
	if err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid output json: %v", err)
	}
	defaults := output.Defaults

	relName, err := sanitiseDeploymentID(r, *mc)
	if err != nil {
		return nil, err
	}

	rc := helm.DefaultReleaseConfig()
	rc.Name = relName
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

	return rc, nil
}

func (mc *moduleConfig) JSON() []byte {
	b, err := json.Marshal(mc)
	if err != nil {
		panic(err)
	}
	return b
}

func sanitiseDeploymentID(r resource.Resource, mc moduleConfig) (string, error) {
	releaseName := mc.Firehose.DeploymentID
	if len(releaseName) == 0 {
		releaseName = generateSafeReleaseName(r.Project, r.Name)
	} else if len(releaseName) >= kubeDeploymentNameLengthLimit {
		return "", errors.ErrInvalid.WithMsgf("deployment_id must be shorter than 63 chars")
	}
	return releaseName, nil
}

func generateSafeReleaseName(project, name string) string {
	const suffix = "-firehose"

	releaseName := fmt.Sprintf("%s-%s", project, name)
	if len(releaseName)+len(suffix) >= kubeDeploymentNameLengthLimit {
		val := sha256.Sum256([]byte(releaseName))
		hash := fmt.Sprintf("%x", val)
		releaseName = fmt.Sprintf("%s-%s", releaseName[:48], hash[:5])
	}

	return releaseName + suffix
}
