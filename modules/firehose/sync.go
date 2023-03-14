package firehose

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules/firehose/kafka"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/helm"
	"github.com/goto/entropy/pkg/kube"
	"github.com/goto/entropy/pkg/worker"
)

const (
	networkErrorRetryDuration   = 5 * time.Second
	kubeAPIRetryBackoffDuration = 30 * time.Second
)

var (
	ErrNetwork = worker.RetryableError{RetryAfter: networkErrorRetryDuration}
	ErrKubeAPI = worker.RetryableError{RetryAfter: kubeAPIRetryBackoffDuration}
)

func (m *firehoseModule) Sync(ctx context.Context, res module.ExpandedResource) (*resource.State, error) {
	r := res.Resource

	var data moduleData
	var pendingStep string
	if err := json.Unmarshal(r.State.ModuleData, &data); err != nil {
		return nil, err
	}

	if len(data.PendingSteps) != 0 {
		pendingStep = data.PendingSteps[0]
		data.PendingSteps = data.PendingSteps[1:]
	}

	var conf moduleConfig
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(res.Dependencies[keyKubeDependency].Output, &kubeOut); err != nil {
		return nil, err
	}

	switch pendingStep {
	case releaseCreate, releaseUpdate:
		if data.StateOverride != "" {
			conf.State = data.StateOverride
		}
		if err := m.releaseSync(pendingStep == releaseCreate, conf, r, kubeOut); err != nil {
			return nil, err
		}
	case consumerReset:
		if err := m.consumerReset(ctx,
			conf,
			r,
			data.ResetTo,
			kubeOut); err != nil {
			return nil, err
		}
		data.StateOverride = ""
	default:
		if err := m.releaseSync(pendingStep == releaseCreate, conf, r, kubeOut); err != nil {
			return nil, err
		}
	}

	finalStatus := resource.StatusCompleted
	if len(data.PendingSteps) > 0 {
		finalStatus = resource.StatusPending
	}

	output, err := m.Output(ctx, res)
	if err != nil {
		return nil, err
	}

	return &resource.State{
		Status:     finalStatus,
		Output:     output,
		ModuleData: data.JSON(),
	}, nil
}

func (*firehoseModule) releaseSync(isCreate bool, conf moduleConfig, r resource.Resource, kube kubernetes.Output) error {
	helmCl := helm.NewClient(&helm.Config{Kubernetes: kube.Configs})

	if conf.State == stateStopped || (conf.StopTime != nil && conf.StopTime.Before(time.Now())) {
		conf.Firehose.Replicas = 0
	}

	hc, err := conf.GetHelmReleaseConfig(r)
	if err != nil {
		return err
	}

	var helmErr error
	if isCreate {
		_, helmErr = helmCl.Create(hc)
	} else {
		_, helmErr = helmCl.Update(hc)
	}

	return helmErr
}

func (*firehoseModule) consumerReset(ctx context.Context, conf moduleConfig, r resource.Resource, resetTo string, out kubernetes.Output) error {
	releaseConfig, err := conf.GetHelmReleaseConfig(r)
	if err != nil {
		return err
	}

	cgm := kafka.NewConsumerGroupManager(conf.Firehose.KafkaBrokerAddress, kube.NewClient(out.Configs), releaseConfig.Namespace)

	switch resetTo {
	case ResetToEarliest:
		err = cgm.ResetOffsetToEarliest(ctx, conf.Firehose.KafkaConsumerID)
	case ResetToLatest:
		err = cgm.ResetOffsetToLatest(ctx, conf.Firehose.KafkaConsumerID)
	default:
		err = cgm.ResetOffsetToDatetime(ctx, conf.Firehose.KafkaConsumerID, resetTo)
	}

	return handleErr(err)
}

func handleErr(err error) error {
	switch {
	case errors.Is(err, kube.ErrJobCreationFailed):
		return ErrNetwork.WithCause(err)
	case errors.Is(err, kube.ErrJobNotFound):
		return ErrKubeAPI.WithCause(err)
	case errors.Is(err, kube.ErrJobExecutionFailed):
		return ErrKubeAPI.WithCause(err)
	default:
		return err
	}
}
