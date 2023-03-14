# Firehose

[Firehose](https://odpf.github.io/firehose/) is an extensible, no-code, and cloud-native service to load real-time streaming data from Kafka to data stores, data lakes, and analytical storage systems.

## What happens in Plan?

Plan handles two actions in the firehose module. It creates a new reosurce or updates(change) the exisiting ones. 
While creating a new firehose, simply a ***release_create*** step is added to the ***moduleData***. Updation in firehose adds ***release_create*** step to the ***moduleData***. It can either be `scale`, `start`, `stop` action or it can just be an `update` of other firehose configs. Firehose configs are adjusted here.

## What happens in Sync?

Sync in Firehose would receive pending step which will be either a "release_create" or "release_update", and it uses a helm client to implementation it.

## Firehose Module Configuration

The configuration struct for Firehose module looks like:

```
type moduleConfig struct {
	State        string `json:"state"`
	Firehose     struct {
		Replicas           int               `json:"replicas"`
		KafkaBrokerAddress string            `json:"kafka_broker_address"`
		KafkaTopic         string            `json:"kafka_topic"`
		KafkaConsumerID    string            `json:"kafka_consumer_id"`
		EnvVariables       map[string]string `json:"env_variables"`
	} `json:"firehose"`
}
```

| Fields | |
| :--- | :--- |
| `State` | `string` State of the firehose, "RUNNING" or "STOPPED". |
| `ChartVersion` | `string` Chart version you want to use. |
| `Firehose` | `struct` Holds firehose configuration. |

Detailed JSONSchema for config can be referenced [here](https://github.com/goto/entropy/blob/main/modules/firehose/schema/config.json).