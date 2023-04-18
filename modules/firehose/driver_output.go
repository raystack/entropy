package firehose

import (
	"context"
	"encoding/json"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
)

func (fd *firehoseDriver) Output(ctx context.Context, exr module.ExpandedResource) (json.RawMessage, error) {
	output, err := readOutputData(exr)
	if err != nil {
		return nil, err
	}

	conf, err := readConfig(exr.Resource, exr.Spec.Configs, fd.conf)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(exr.Dependencies[keyKubeDependency].Output, &kubeOut); err != nil {
		return nil, errors.ErrInternal.WithMsgf("invalid kube state").WithCausef(err.Error())
	}

	return fd.refreshOutput(ctx, exr.Resource, *conf, *output, kubeOut)
}

func (fd *firehoseDriver) refreshOutput(ctx context.Context, r resource.Resource,
	conf Config, output Output, kubeOut kubernetes.Output,
) (json.RawMessage, error) {
	rc, err := fd.getHelmRelease(r, conf, kubeOut)
	if err != nil {
		return nil, err
	}

	pods, err := fd.kubeGetPod(ctx, kubeOut.Configs, rc.Namespace, map[string]string{"app": rc.Name})
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	output.Pods = pods

	return mustJSON(output), nil
}
