package kafka

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kube"
)

const (
	kafkaImage = "bitnami/kafka:2.0.0"
	retries    = 6
)

const (
	resetLatest   = "latest"
	resetEarliest = "earliest"
	resetDatetime = "datetime"
)

type ResetParams struct {
	To       string `json:"to"`
	Datetime string `json:"datetime"`
}

// DoReset executes a kubernetes job with kafka-consumer-group.sh installed to
// reset offset policy for the given consumer id on all topics.
func DoReset(ctx context.Context, jobCluster *kube.Client, kubeNamespace, kafkaBrokers, kafkaConsumerID, kafkaResetValue, resetJobName string) error {
	jobName := resetJobName + "-reset"

	return jobCluster.RunJob(ctx, kubeNamespace,
		jobName,
		kafkaImage,
		prepCommand(kafkaBrokers, kafkaConsumerID, kafkaResetValue),
		retries,
	)
}

// ParseResetParams parses the given JSON data as reset parameters value and
// returns the actual reset value to be used with DoReset().
func ParseResetParams(bytes json.RawMessage) (string, error) {
	var params ResetParams
	if err := json.Unmarshal(bytes, &params); err != nil {
		return "", errors.ErrInvalid.
			WithMsgf("invalid reset params").
			WithCausef(err.Error())
	}

	resetValue := strings.ToLower(params.To)
	if params.To == resetDatetime {
		resetValue = params.Datetime
	} else if resetValue != resetLatest && resetValue != resetEarliest {
		return "", errors.ErrInvalid.
			WithMsgf("reset_value must be one of %v", []string{resetEarliest, resetLatest, resetDatetime})
	}

	return resetValue, nil
}

func prepCommand(brokers, consumerID, kafkaResetValue string) []string {
	args := []string{
		"kafka-consumer-groups.sh",
		"--bootstrap-server", brokers,
		"--group", consumerID,
		"--reset-offsets",
		"--execute",
		"--all-topics",
	}

	switch kafkaResetValue {
	case resetLatest:
		args = append(args, "--to-latest")

	case resetEarliest:
		args = append(args, "--to-earliest")

	default:
		args = append(args, "--to-datetime", kafkaResetValue)
	}

	return args
}
