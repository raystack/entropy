package driver

import (
	"context"
	"encoding/json"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules/job/config"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
)

type Output struct {
	Namespace string `json:"namespace"`
	JobName   string `json:"jobName"`
}

func (*Driver) refreshOutput(context.Context, resource.Resource, config.Config, Output, kubernetes.Output) (json.RawMessage, error) {
	return json.RawMessage{}, nil
}

func ReadOutputData(exr module.ExpandedResource) (*Output, error) {
	var curOut Output
	if err := json.Unmarshal(exr.Resource.State.Output, &curOut); err != nil {
		return nil, errors.ErrInternal.WithMsgf("corrupted output").WithCausef(err.Error())
	}
	return &curOut, nil
}
