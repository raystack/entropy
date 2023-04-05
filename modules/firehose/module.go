package firehose

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"

	"helm.sh/helm/v3/pkg/release"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/helm"
	"github.com/goto/entropy/pkg/kafka"
	"github.com/goto/entropy/pkg/kube"
	"github.com/goto/entropy/pkg/validator"
	"github.com/goto/entropy/pkg/worker"
)

const (
	keyKubeDependency = "kube_cluster"

	ScaleAction   = "scale"
	StartAction   = "start"
	StopAction    = "stop"
	ResetAction   = "reset"
	UpgradeAction = "upgrade"
)

var Module = module.Descriptor{
	Kind: "firehose",
	Dependencies: map[string]string{
		keyKubeDependency: kubernetes.Module.Kind,
	},
	Actions: []module.ActionDesc{
		{
			Name:        module.CreateAction,
			Description: "Creates a new firehose",
		},
		{
			Name:        module.UpdateAction,
			Description: "Update all configurations of firehose",
		},
		{
			Name:        ResetAction,
			Description: "Stop firehose, reset consumer group, restart",
		},
		{
			Name:        StopAction,
			Description: "Stop all replicas of this firehose.",
		},
		{
			Name:        StartAction,
			Description: "Start the firehose if it is currently stopped.",
		},
		{
			Name:        ScaleAction,
			Description: "Scale the number of replicas to given number.",
		},
		{
			Name:        UpgradeAction,
			Description: "Upgrade firehose version",
		},
	},
	DriverFactory: func(confJSON json.RawMessage) (module.Driver, error) {
		conf := defaultDriverConf // clone the default value
		if err := json.Unmarshal(confJSON, &conf); err != nil {
			return nil, err
		} else if err := validator.TaggedStruct(conf); err != nil {
			return nil, err
		}

		return &firehoseDriver{
			conf:    conf,
			timeNow: time.Now,
			kubeDeploy: func(_ context.Context, isCreate bool, kubeConf kube.Config, hc helm.ReleaseConfig) error {
				canUpdate := func(rel *release.Release) bool {
					curLabels, ok := rel.Config[labelsConfKey].(map[string]any)
					if !ok {
						return false
					}
					newLabels, ok := hc.Values[labelsConfKey].(map[string]string)
					if !ok {
						return false
					}

					isManagedByEntropy := curLabels[labelOrchestrator] == orchestratorLabelValue
					isSameProject := curLabels[labelProject] == newLabels[labelProject]
					isSameName := curLabels[labelName] == newLabels[labelName]

					return isManagedByEntropy && isSameProject && isSameName
				}

				helmCl := helm.NewClient(&helm.Config{Kubernetes: kubeConf})
				_, errHelm := helmCl.Upsert(&hc, canUpdate)
				return errHelm
			},
			kubeGetPod: func(ctx context.Context, conf kube.Config, ns string, labels map[string]string) ([]kube.Pod, error) {
				kubeCl := kube.NewClient(conf)
				return kubeCl.GetPodDetails(ctx, ns, labels)
			},
			consumerReset: consumerReset,
		}, nil
	},
}

func consumerReset(ctx context.Context, conf Config, out kubernetes.Output, resetTo string) error {
	const (
		networkErrorRetryDuration   = 5 * time.Second
		kubeAPIRetryBackoffDuration = 30 * time.Second
	)

	var (
		errNetwork = worker.RetryableError{RetryAfter: networkErrorRetryDuration}
		errKubeAPI = worker.RetryableError{RetryAfter: kubeAPIRetryBackoffDuration}
	)

	brokerAddr := conf.EnvVariables[confKeyKafkaBrokers]
	consumerID := conf.EnvVariables[confKeyConsumerID]

	err := kafka.DoReset(ctx, kube.NewClient(out.Configs), conf.Namespace, brokerAddr, consumerID, resetTo)
	if err != nil {
		switch {
		case errors.Is(err, kube.ErrJobCreationFailed):
			return errNetwork.WithCause(err)

		case errors.Is(err, kube.ErrJobNotFound):
			return errKubeAPI.WithCause(err)

		case errors.Is(err, kube.ErrJobExecutionFailed):
			return errKubeAPI.WithCause(err)

		default:
			return err
		}
	}

	return nil
}
