package driver

import (
	"context"
	"encoding/json"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/modules"
	"github.com/goto/entropy/modules/job/config"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kube"
)

type Output struct {
	Namespace string     `json:"namespace"`
	JobName   string     `json:"jobName"`
	Pods      []kube.Pod `json:"pods"`
}

func (driver *Driver) refreshOutput(ctx context.Context, conf config.Config, output Output, kubeOut kubernetes.Output) (json.RawMessage, error) {
	pods, err := driver.GetJobPods(ctx, kubeOut.Configs, map[string]string{"job-name": conf.Name})
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	output.Pods = pods

	return modules.MustJSON(output), nil
}

func ReadOutputData(exr module.ExpandedResource) (*Output, error) {
	var curOut Output
	if len(exr.Resource.State.Output) == 0 {
		return &curOut, nil
	}
	if err := json.Unmarshal(exr.Resource.State.Output, &curOut); err != nil {
		return nil, errors.ErrInternal.WithMsgf("corrupted output").WithCausef(err.Error())
	}
	return &curOut, nil
}
