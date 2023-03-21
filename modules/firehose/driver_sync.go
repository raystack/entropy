package firehose

import (
	"context"
	"encoding/json"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
)

func (fd *firehoseDriver) Sync(ctx context.Context, exr module.ExpandedResource) (*resource.State, error) {
	modData, err := readTransientData(exr)
	if err != nil {
		return nil, err
	}

	out, err := readOutputData(exr)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	conf, err := readConfig(exr.Resource, exr.Spec.Configs)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(exr.Dependencies[keyKubeDependency].Output, &kubeOut); err != nil {
		return nil, errors.ErrInternal.WithMsgf("invalid kube state").WithCausef(err.Error())
	}

	// pickup the next pending step if available.
	if len(modData.PendingSteps) > 0 {
		pendingStep := modData.PendingSteps[0]
		modData.PendingSteps = modData.PendingSteps[1:]

		switch pendingStep {
		case stepReleaseCreate, stepReleaseUpdate, stepReleaseStop:
			// we want to stop the current deployment. we do this by setting
			// replicas to 0. But this value will not be persisted to DB since
			// config changes during Sync() are not saved.
			if pendingStep == stepReleaseStop {
				conf.Replicas = 0
			}

			isCreate := pendingStep == stepReleaseCreate
			if err := fd.releaseSync(ctx, isCreate, *conf, kubeOut); err != nil {
				return nil, err
			}

		case stepKafkaReset:
			if err := fd.consumerReset(ctx, *conf, kubeOut, modData.ResetOffsetTo); err != nil {
				return nil, err
			}

		default:
			return nil, errors.ErrInternal.WithMsgf("unknown step: '%s'", pendingStep)
		}
	}

	finalOut, err := fd.refreshOutput(ctx, *conf, *out, kubeOut)
	if err != nil {
		return nil, err
	}

	finalState := resource.State{
		Status:     resource.StatusPending,
		Output:     finalOut,
		ModuleData: mustJSON(modData),
	}

	if len(modData.PendingSteps) == 0 {
		finalState.Status = resource.StatusCompleted
		finalState.ModuleData = nil
	}

	return &finalState, nil
}

func (fd *firehoseDriver) releaseSync(ctx context.Context, isCreate bool, conf Config, kubeOut kubernetes.Output) error {
	rc := fd.getHelmRelease(conf)
	if err := fd.kubeDeploy(ctx, isCreate, kubeOut.Configs, *rc); err != nil {
		return errors.ErrInternal.WithCausef(err.Error())
	}
	return nil
}
