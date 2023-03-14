package kafka

import (
	"context"

	"github.com/goto/entropy/pkg/kube"
)

const (
	kafkaImage = "bitnami/kafka:2.0.0"
	retries    = 6
)

type ConsumerGroupManager struct {
	brokers   string
	kube      *kube.Client
	namespace string
}

func NewConsumerGroupManager(brokers string, kube *kube.Client, namespace string) *ConsumerGroupManager {
	return &ConsumerGroupManager{
		brokers:   brokers,
		kube:      kube,
		namespace: namespace,
	}
}

func (k ConsumerGroupManager) ResetOffsetToDatetime(ctx context.Context, consumerID string, datetime string) error {
	return k.kube.RunJob(ctx, k.namespace,
		getJobName(consumerID),
		kafkaImage,
		append(k.getDefaultCMD(consumerID), "--to-datetime", datetime),
		retries,
	)
}

func (k ConsumerGroupManager) ResetOffsetToLatest(ctx context.Context, consumerID string) error {
	return k.kube.RunJob(ctx, k.namespace,
		getJobName(consumerID),
		kafkaImage,
		append(k.getDefaultCMD(consumerID), "--to-latest"),
		retries,
	)
}

func (k ConsumerGroupManager) ResetOffsetToEarliest(ctx context.Context, consumerID string) error {
	return k.kube.RunJob(ctx, k.namespace,
		getJobName(consumerID),
		kafkaImage,
		append(k.getDefaultCMD(consumerID), "--to-earliest"),
		retries,
	)
}

func (k ConsumerGroupManager) getDefaultCMD(consumerID string) []string {
	return []string{"kafka-consumer-groups.sh", "--bootstrap-server", k.brokers, "--group", consumerID, "--reset-offsets", "--execute", "--all-topics"}
}

func getJobName(consumerID string) string {
	return consumerID + "-reset"
}
