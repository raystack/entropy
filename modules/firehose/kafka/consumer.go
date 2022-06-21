package kafka

import "github.com/odpf/entropy/modules/kubernetes"

type KafkaConsumerGroups struct {
	brokers string
	topics  []string
	kube    kubernetes.Output
}

func (k KafkaConsumerGroups) ResetOffsetByTimestamp() {

}

func (k KafkaConsumerGroups) ResetOffsetByPeriod() {

}

func (k KafkaConsumerGroups) ResetOffsetToLatest() {

}

func (k KafkaConsumerGroups) ResetOffsetToEarliest() {

}
