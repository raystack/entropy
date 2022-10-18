package firehose

import (
	"context"
	"encoding/json"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/modules/kubernetes"
	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/pkg/kube"
)

func (*firehoseModule) Log(ctx context.Context, res module.ExpandedResource, filter map[string]string) (<-chan module.LogChunk, error) {
	r := res.Resource

	var conf moduleConfig
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(res.Dependencies[keyKubeDependency].Output, &kubeOut); err != nil {
		return nil, err
	}

	if filter == nil {
		filter = make(map[string]string)
	}
	filter["app"] = conf.GetHelmReleaseConfig(r).Name

	kubeCl := kube.NewClient(kubeOut.Configs)
	logs, err := kubeCl.StreamLogs(ctx, defaultNamespace, filter)
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
