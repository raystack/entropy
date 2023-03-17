package firehose

import (
	"context"
	"encoding/json"
	"time"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/helm"
	"github.com/goto/entropy/pkg/kafka"
	"github.com/goto/entropy/pkg/kube"
	"github.com/goto/entropy/pkg/worker"
)

const (
	networkErrorRetryDuration   = 5 * time.Second
	kubeAPIRetryBackoffDuration = 30 * time.Second
)

const (
	releaseCreate = "release_create"
	releaseUpdate = "release_update"
	consumerReset = "consumer_reset"
)

const (
	stateRunning = "RUNNING"
	stateStopped = "STOPPED"
)

const (
	ResetToDateTime = "DATETIME"
	ResetToEarliest = "EARLIEST"
	ResetToLatest   = "LATEST"
)

var (
	ErrNetwork = worker.RetryableError{RetryAfter: networkErrorRetryDuration}
	ErrKubeAPI = worker.RetryableError{RetryAfter: kubeAPIRetryBackoffDuration}
)

type driverConfig struct {
	ChartRepository string `json:"chart_repository,omitempty"`
	ChartName       string `json:"chart_name,omitempty"`
	ChartVersion    string `json:"chart_version,omitempty"`
	ImageRepository string `json:"image_repository,omitempty"`
	ImageName       string `json:"image_name,omitempty"`
	ImageTag        string `json:"image_tag,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	ImagePullPolicy string `json:"image_pull_policy,omitempty"`
}

type firehoseDriver struct {
	Config driverConfig `json:"config"`
}

func (m *firehoseDriver) Plan(_ context.Context, res module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	switch act.Name {
	case module.CreateAction:
		return m.planCreate(res, act)

	case ResetAction:
		return m.planReset(res, act)

	default:
		return m.planChange(res, act)
	}
}

func (m *firehoseDriver) Sync(ctx context.Context, res module.ExpandedResource) (*resource.State, error) {
	r := res.Resource

	var data moduleData
	if err := json.Unmarshal(r.State.ModuleData, &data); err != nil {
		return nil, errors.ErrInternal.WithMsgf("invalid module data").WithCausef(err.Error())
	}

	var conf Config
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json").WithCausef(err.Error())
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(res.Dependencies[keyKubeDependency].Output, &kubeOut); err != nil {
		return nil, errors.ErrInternal.WithMsgf("invalid kube state").WithCausef(err.Error())
	}

	if len(data.PendingSteps) > 0 {
		pendingStep := data.PendingSteps[0]
		data.PendingSteps = data.PendingSteps[1:]

		switch pendingStep {
		case releaseCreate, releaseUpdate:
			if data.StateOverride != "" {
				conf.State = data.StateOverride
			}
			if err := m.releaseSync(pendingStep == releaseCreate, conf, r, kubeOut); err != nil {
				return nil, err
			}

		case consumerReset:
			if err := m.consumerReset(ctx, conf, r, data.ResetTo, kubeOut); err != nil {
				return nil, err
			}
			data.StateOverride = ""

		default:
			if err := m.releaseSync(pendingStep == releaseCreate, conf, r, kubeOut); err != nil {
				return nil, err
			}
		}
	}

	output, err := m.Output(ctx, res)
	if err != nil {
		return nil, err
	}

	finalStatus := resource.StatusCompleted
	if len(data.PendingSteps) > 0 {
		finalStatus = resource.StatusPending
	}

	return &resource.State{
		Status:     finalStatus,
		Output:     output,
		ModuleData: mustJSON(data),
	}, nil
}

func (*firehoseDriver) Log(ctx context.Context, res module.ExpandedResource, filter map[string]string) (<-chan module.LogChunk, error) {
	r := res.Resource

	var conf Config
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

	hc, err := getHelmReleaseConf(r, conf)
	if err != nil {
		return nil, err
	}

	filter["app"] = hc.Name

	kubeCl := kube.NewClient(kubeOut.Configs)
	logs, err := kubeCl.StreamLogs(ctx, hc.Namespace, filter)
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

func (m *firehoseDriver) Output(ctx context.Context, res module.ExpandedResource) (json.RawMessage, error) {
	var conf Config
	if err := json.Unmarshal(res.Resource.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	var output Output
	if err := json.Unmarshal(res.Resource.State.Output, &output); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid output json: %v", err)
	}

	pods, err := m.podDetails(ctx, res)
	if err != nil {
		return nil, err
	}

	hc, err := getHelmReleaseConf(res.Resource, conf)
	if err != nil {
		return nil, err
	}

	return mustJSON(Output{
		Namespace:   hc.Namespace,
		ReleaseName: hc.Name,
		Pods:        pods,
		Defaults:    output.Defaults,
	}), nil
}

func (*firehoseDriver) podDetails(ctx context.Context, res module.ExpandedResource) ([]kube.Pod, error) {
	r := res.Resource

	var conf Config
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	var kubeOut kubernetes.Output
	if err := json.Unmarshal(res.Dependencies[keyKubeDependency].Output, &kubeOut); err != nil {
		return nil, err
	}

	hc, err := getHelmReleaseConf(r, conf)
	if err != nil {
		return nil, err
	}

	kubeCl := kube.NewClient(kubeOut.Configs)
	return kubeCl.GetPodDetails(ctx, hc.Namespace, map[string]string{"app": hc.Name})
}

func (m *firehoseDriver) planCreate(res module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	var plan module.Plan
	r := res.Resource

	var reqConf Config
	if err := json.Unmarshal(act.Params, &reqConf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}
	if err := reqConf.validateAndSanitize(res.Resource); err != nil {
		return nil, err
	}

	output := mustJSON(Output{
		Defaults: m.Config,
	})

	r.Spec.Configs = mustJSON(reqConf)
	r.State = resource.State{
		Status: resource.StatusPending,
		ModuleData: mustJSON(moduleData{
			PendingSteps: []string{releaseCreate},
		}),
		Output: output,
	}

	plan.Resource = r
	if reqConf.StopTime != nil {
		plan.ScheduleRunAt = *reqConf.StopTime
	}
	plan.Reason = "firehose created"
	return &plan, nil
}

func (m *firehoseDriver) planChange(res module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	var plan module.Plan
	r := res.Resource

	var conf Config
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	switch act.Name {
	case module.UpdateAction:
		var reqConf Config
		if err := json.Unmarshal(act.Params, &reqConf); err != nil {
			return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
		}
		if err := reqConf.validateAndSanitize(r); err != nil {
			return nil, err
		}
		conf = reqConf

		if conf.StopTime != nil {
			plan.ScheduleRunAt = *conf.StopTime
		}
		plan.Reason = "firehose config updated"

	case ScaleAction:
		var scaleParams struct {
			Replicas int `json:"replicas"`
		}
		if err := json.Unmarshal(act.Params, &scaleParams); err != nil {
			return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
		}
		conf.Firehose.Replicas = scaleParams.Replicas
		plan.Reason = "firehose scaled"

	case StartAction:
		conf.State = stateRunning
		plan.Reason = "firehose started"

	case StopAction:
		conf.State = stateStopped
		plan.Reason = "firehose stopped"

	case UpgradeAction:
		var output Output
		err := json.Unmarshal(res.State.Output, &output)
		if err != nil {
			return nil, errors.ErrInvalid.WithMsgf("invalid output json: %v", err)
		}

		output.Defaults = m.Config
		res.State.Output = mustJSON(output)

		plan.Reason = "firehose upgraded"
	}

	r.Spec.Configs = mustJSON(conf)
	r.State = resource.State{
		Status: resource.StatusPending,
		Output: res.State.Output,
		ModuleData: mustJSON(moduleData{
			PendingSteps: []string{releaseUpdate},
		}),
	}
	plan.Resource = r
	return &plan, nil
}

func (*firehoseDriver) planReset(res module.ExpandedResource, act module.ActionRequest) (*module.Plan, error) {
	r := res.Resource

	var conf Config
	if err := json.Unmarshal(r.Spec.Configs, &conf); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid config json: %v", err)
	}

	var resetParams struct {
		To       string `json:"to"`
		Datetime string `json:"datetime"`
	}
	if err := json.Unmarshal(act.Params, &resetParams); err != nil {
		return nil, errors.ErrInvalid.WithMsgf("invalid action params json: %v", err)
	}

	resetValue := resetParams.To
	if resetParams.To == "DATETIME" {
		resetValue = resetParams.Datetime
	}

	r.Spec.Configs = mustJSON(conf)
	r.State = resource.State{
		Status: resource.StatusPending,
		Output: res.State.Output,
		ModuleData: mustJSON(moduleData{
			ResetTo:       resetValue,
			PendingSteps:  []string{releaseUpdate, consumerReset, releaseUpdate},
			StateOverride: stateStopped,
		}),
	}

	return &module.Plan{Resource: r, Reason: "firehose consumer reset"}, nil
}

func (*firehoseDriver) releaseSync(isCreate bool, conf Config, r resource.Resource, kube kubernetes.Output) error {
	helmCl := helm.NewClient(&helm.Config{Kubernetes: kube.Configs})

	if conf.State == stateStopped || (conf.StopTime != nil && conf.StopTime.Before(time.Now())) {
		conf.Firehose.Replicas = 0
	}

	hc, err := getHelmReleaseConf(r, conf)
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

func (*firehoseDriver) consumerReset(ctx context.Context, conf Config, r resource.Resource, resetTo string, out kubernetes.Output) error {
	releaseConfig, err := getHelmReleaseConf(r, conf)
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

func mustJSON(v any) json.RawMessage {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bytes
}
