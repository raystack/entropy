package driver

import (
	"context"
	"encoding/json"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/modules/job/config"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
)

func (driver *Driver) Log(ctx context.Context, res module.ExpandedResource, filter map[string]string) (<-chan module.LogChunk, error) {
	conf, err := config.ReadConfig(res.Resource, res.Spec.Configs, driver.Conf)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	if filter == nil {
		filter = map[string]string{}
	}
	filter["app"] = conf.Name

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(res.Dependencies[KeyKubeDependency].Output, &kubeOut); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	return driver.StreamLogs(ctx, kubeOut.Configs, filter)
}
