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

	conf, err := readConfig(exr.Resource, exr.Spec.Configs, fd.conf)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(exr.Dependencies[keyKubeDependency].Output, &kubeOut); err != nil {
		return nil, errors.ErrInternal.WithMsgf("invalid kube state").WithCausef(err.Error())
	}

	finalState := resource.State{
		Status: resource.StatusPending,
		Output: exr.Resource.State.Output,
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
			if pendingStep == stepReleaseStop || conf.Stopped {
				conf.Replicas = 0
			}

			isCreate := pendingStep == stepReleaseCreate
			if err := fd.releaseSync(ctx, exr.Resource, isCreate, *conf, kubeOut); err != nil {
				return nil, err
			}

		case stepKafkaReset:
			if err := fd.consumerReset(ctx, *conf, kubeOut, modData.ResetOffsetTo); err != nil {
				return nil, err
			}

		default:
			return nil, errors.ErrInternal.WithMsgf("unknown step: '%s'", pendingStep)
		}

		// we have more pending states, so enqueue resource for another sync
		// as soon as possible.
		immediately := fd.timeNow()
		finalState.NextSyncAt = &immediately
		finalState.ModuleData = mustJSON(modData)

		return &finalState, nil
	}

	// even if the resource is in completed state, we check this time to
	// see if the firehose is expected to be stopped by this time.
	finalState.NextSyncAt = conf.StopTime
	if conf.StopTime != nil && conf.StopTime.Before(fd.timeNow()) {
		conf.Replicas = 0
		if err := fd.releaseSync(ctx, exr.Resource, false, *conf, kubeOut); err != nil {
			return nil, err
		}
		finalState.NextSyncAt = nil
	}

	finalOut, err := fd.refreshOutput(ctx, exr.Resource, *conf, *out, kubeOut)
	if err != nil {
		return nil, err
	}
	finalState.Output = finalOut

	finalState.Status = resource.StatusCompleted
	finalState.ModuleData = nil
	return &finalState, nil
}

func (fd *firehoseDriver) releaseSync(ctx context.Context, r resource.Resource,
	isCreate bool, conf Config, kubeOut kubernetes.Output,
) error {
	rc, err := fd.getHelmRelease(r, conf, kubeOut)
	if err != nil {
		return err
	}

	if err := fd.kubeDeploy(ctx, isCreate, kubeOut.Configs, *rc); err != nil {
		return errors.ErrInternal.WithCausef(err.Error())
	}
	return nil
}
