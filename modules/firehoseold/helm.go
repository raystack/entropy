package firehoseold

import (
	"encoding/json"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/helm"
)

func getHelmReleaseConf(r resource.Resource, mc Config) (*helm.ReleaseConfig, error) {
	var output Output
	if err := json.Unmarshal(r.State.Output, &output); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid output json: %v", err)
	}
	defaults := output.Defaults

	relName, err := sanitiseDeploymentID(r, mc)
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
