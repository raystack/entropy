package firehose

import (
	"context"
	"encoding/json"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kube"
)

func (fd *firehoseDriver) Log(ctx context.Context, res module.ExpandedResource, filter map[string]string) (<-chan module.LogChunk, error) {
	conf, err := readConfig(res.Resource, res.Spec.Configs, fd.conf)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}

	if filter == nil {
		filter = map[string]string{}
	}
	filter["app"] = conf.DeploymentID

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(res.Dependencies[keyKubeDependency].Output, &kubeOut); err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	kubeCl := kube.NewClient(kubeOut.Configs)

	logs, err := kubeCl.StreamLogs(ctx, conf.Namespace, filter)
	if err != nil {
		return nil, err
	}

	mappedLogs := make(chan module.LogChunk)
	go func() {
		defer close(mappedLogs)
		for {
			select {
			case log, ok := <-logs:
				if !ok {
					return
				}
				mappedLogs <- module.LogChunk{Data: log.Data, Labels: log.Labels}
			case <-ctx.Done():
				return
			}
		}
	}()

	return mappedLogs, err
}
