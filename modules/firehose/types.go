package firehose

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kube"
)

const (
	kubeDeploymentNameLengthLimit      = 53
	firehoseConsumerIDStartingSequence = "0001"
)

type Output struct {
	Namespace   string       `json:"namespace,omitempty"`
	ReleaseName string       `json:"release_name,omitempty"`
	Pods        []kube.Pod   `json:"pods,omitempty"`
	Defaults    driverConfig `json:"defaults,omitempty"`
}

type moduleData struct {
	PendingSteps  []string `json:"pending_steps"`
	ResetTo       string   `json:"reset_to,omitempty"`
	StateOverride string   `json:"state_override,omitempty"`
}

type Config struct {
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

func (mc *Config) validateAndSanitize(r resource.Resource) error {
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

func sanitiseDeploymentID(r resource.Resource, mc Config) (string, error) {
	releaseName := mc.Firehose.DeploymentID
	if len(releaseName) == 0 {
		releaseName = generateSafeReleaseName(r.Project, r.Name)
	} else if len(releaseName) >= kubeDeploymentNameLengthLimit {
		return "", errors.ErrInvalid.WithMsgf("deployment_id must be shorter than 53 chars")
	}
	return releaseName, nil
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
