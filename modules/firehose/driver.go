package firehose

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/helm"
	"github.com/goto/entropy/pkg/kube"
)

const (
	stepReleaseCreate = "release_create"
	stepReleaseUpdate = "release_update"
	stepReleaseStop   = "release_stop"
	stepKafkaReset    = "consumer_reset"
)

const (
	chartRepo = "https://goto.github.io/charts/"
	chartName = "firehose"
	imageRepo = "gotocompany/firehose"
)

const (
	labelsConfKey = "labels"

	labelDeployment   = "deployment"
	labelOrchestrator = "orchestrator"
	labelURN          = "urn"
	labelName         = "name"
	labelNamespace    = "namespace"

	orchestratorLabelValue = "entropy"
)

const defaultKey = "default"

var defaultDriverConf = driverConf{
	Namespace: map[string]string{
		defaultKey: "firehose",
	},
	ChartValues: ChartValues{
		ImageTag:        "latest",
		ChartVersion:    "0.1.3",
		ImagePullPolicy: "IfNotPresent",
	},
	RequestsAndLimits: map[string]RequestsAndLimits{
		defaultKey: {
			Limits: UsageSpec{
				CPU:    "200m",
				Memory: "512Mi",
			},
			Requests: UsageSpec{
				CPU:    "200m",
				Memory: "512Mi",
			},
		},
	},
}

type firehoseDriver struct {
	timeNow       func() time.Time
	conf          driverConf
	kubeDeploy    kubeDeployFn
	kubeGetPod    kubeGetPodFn
	consumerReset consumerResetFn
}

type (
	kubeDeployFn    func(ctx context.Context, isCreate bool, conf kube.Config, hc helm.ReleaseConfig) error
	kubeGetPodFn    func(ctx context.Context, conf kube.Config, ns string, labels map[string]string) ([]kube.Pod, error)
	consumerResetFn func(ctx context.Context, conf Config, out kubernetes.Output, resetTo string, offsetResetDelaySeconds int) error
)

type driverConf struct {
	// Labels to be injected to the chart during deployment. Values can be Go templates.
	Labels map[string]string `json:"labels,omitempty"`

	// Telegraf is the telegraf configuration for the deployment.
	Telegraf *Telegraf `json:"telegraf"`

	// Namespace is the kubernetes namespace where firehoses will be deployed.
	Namespace map[string]string `json:"namespace" validate:"required"`

	// ChartValues is the chart and image version information.
	ChartValues ChartValues `json:"chart_values" validate:"required"`

	// Tolerations represents the tolerations to be set for the deployment.
	// The key in the map is the sink-type in upper case.
	Tolerations map[string]kubernetes.Toleration `json:"tolerations"`

	EnvVariables map[string]string `json:"env_variables,omitempty"`

	// InitContainer can be set to have a container that is used as init_container on the
	// deployment.
	InitContainer InitContainer `json:"init_container"`

	// GCSSinkCredential can be set to the name of kubernetes secret containing GCS credential.
	// The secret must already exist on the target kube cluster in the same namespace.
	// The secret will be mounted as a volume and the appropriate credential path will be set.
	GCSSinkCredential string `json:"gcs_sink_credential,omitempty"`

	// DLQGCSSinkCredential is same as GCSSinkCredential but for DLQ.
	DLQGCSSinkCredential string `json:"dlq_gcs_sink_credential,omitempty"`

	// BigQuerySinkCredential is same as GCSSinkCredential but for BigQuery credential.
	BigQuerySinkCredential string `json:"big_query_sink_credential,omitempty"`

	// RequestsAndLimits can be set to configure the container cpu/memory requests & limits.
	// 'default' key will be used as base and any sink-type will be used as the override.
	RequestsAndLimits map[string]RequestsAndLimits `json:"requests_and_limits" validate:"required"`

	// NodeAffinityMatchExpressions can be used to set node-affinity for the deployment.
	NodeAffinityMatchExpressions NodeAffinityMatchExpressions `json:"node_affinity_match_expressions"`

	// delay between stopping a firehose and making an offset reset request
	OffsetResetDelaySeconds int `json:"offset_reset_delay_seconds"`
}

type RequestsAndLimits struct {
	Limits   UsageSpec `json:"limits,omitempty"`
	Requests UsageSpec `json:"requests,omitempty"`
}

type NodeAffinityMatchExpressions struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []Preference         `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPreference `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

type WeightedPreference struct {
	Weight     int          `json:"weight" validate:"required"`
	Preference []Preference `json:"preference" validate:"required"`
}

type Preference struct {
	Key      string   `json:"key" validate:"required"`
	Operator string   `json:"operator" validate:"required"`
	Values   []string `json:"values"`
}

type InitContainer struct {
	Enabled bool `json:"enabled"`

	Args    []string `json:"args"`
	Command []string `json:"command"`

	Repository string `json:"repository"`
	ImageTag   string `json:"image_tag"`
	PullPolicy string `json:"pull_policy"`
}

type UsageSpec struct {
	CPU    string `json:"cpu,omitempty" validate:"required"`
	Memory string `json:"memory,omitempty" validate:"required"`
}

type Output struct {
	Pods        []kube.Pod `json:"pods,omitempty"`
	Namespace   string     `json:"namespace,omitempty"`
	ReleaseName string     `json:"release_name,omitempty"`
}

type transientData struct {
	PendingSteps  []string `json:"pending_steps"`
	ResetOffsetTo string   `json:"reset_offset_to,omitempty"`
}

func (fd *firehoseDriver) getHelmRelease(res resource.Resource, conf Config,
	kubeOut kubernetes.Output,
) (*helm.ReleaseConfig, error) {
	var telegrafConf Telegraf

	entropyLabels := map[string]string{
		labelDeployment:   conf.DeploymentID,
		labelOrchestrator: orchestratorLabelValue,
	}

	otherLabels := map[string]string{
		labelURN:       res.URN,
		labelName:      res.Name,
		labelNamespace: conf.Namespace,
	}

	deploymentLabels, err := renderTpl(fd.conf.Labels, modules.CloneAndMergeMaps(res.Labels, modules.CloneAndMergeMaps(entropyLabels, otherLabels)))
	if err != nil {
		return nil, err
	}

	if conf.Telegraf != nil && conf.Telegraf.Enabled {
		mergedLabelsAndEnvVariablesMap := modules.CloneAndMergeMaps(modules.CloneAndMergeMaps(conf.EnvVariables, modules.CloneAndMergeMaps(deploymentLabels, modules.CloneAndMergeMaps(res.Labels, entropyLabels))), otherLabels)

		conf.EnvVariables, err = renderTpl(conf.EnvVariables, mergedLabelsAndEnvVariablesMap)
		if err != nil {
			return nil, err
		}

		telegrafTags, err := renderTpl(conf.Telegraf.Config.AdditionalGlobalTags, mergedLabelsAndEnvVariablesMap)
		if err != nil {
			return nil, err
		}

		for key, val := range conf.Telegraf.Config.Output {
			valAsMap, ok := val.(map[string]interface{})
			if !ok {
				continue
			}

			valAsMap, err = renderTplOfMapStringAny(valAsMap, mergedLabelsAndEnvVariablesMap)
			if err != nil {
				return nil, err
			}

			conf.Telegraf.Config.Output[key] = valAsMap
		}

		telegrafConf = Telegraf{
			Enabled: true,
			Image:   conf.Telegraf.Image,
			Config: TelegrafConf{
				Output:               conf.Telegraf.Config.Output,
				AdditionalGlobalTags: telegrafTags,
			},
		}
	}

	tolerationKey := fmt.Sprintf("firehose_%s", conf.EnvVariables["SINK_TYPE"])
	tolerations := []map[string]any{}
	for _, t := range kubeOut.Tolerations[tolerationKey] {
		tolerations = append(tolerations, map[string]any{
			"key":      t.Key,
			"value":    t.Value,
			"effect":   t.Effect,
			"operator": t.Operator,
		})
	}

	mountSecrets := []map[string]any{}
	requiredDuringSchedulingIgnoredDuringExecution := []Preference{}
	preferredDuringSchedulingIgnoredDuringExecution := []WeightedPreference{}

	if fd.conf.NodeAffinityMatchExpressions.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		requiredDuringSchedulingIgnoredDuringExecution = fd.conf.NodeAffinityMatchExpressions.RequiredDuringSchedulingIgnoredDuringExecution
	}
	if fd.conf.NodeAffinityMatchExpressions.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		preferredDuringSchedulingIgnoredDuringExecution = fd.conf.NodeAffinityMatchExpressions.PreferredDuringSchedulingIgnoredDuringExecution
	}

	if fd.conf.GCSSinkCredential != "" {
		const mountFile = "gcs_auth.json"
		credPath := fmt.Sprintf("/etc/secret/%s", mountFile)

		mountSecrets = append(mountSecrets, map[string]any{
			"value": fd.conf.GCSSinkCredential,
			"key":   "gcs_credential",
			"path":  mountFile,
		})
		conf.EnvVariables["SINK_BLOB_GCS_CREDENTIAL_PATH"] = credPath
		conf.EnvVariables["SINK_BIGTABLE_CREDENTIAL_PATH"] = credPath
	}

	if fd.conf.DLQGCSSinkCredential != "" {
		const mountFile = "dlq_gcs_auth.json"
		credPath := fmt.Sprintf("/etc/secret/%s", mountFile)

		mountSecrets = append(mountSecrets, map[string]any{
			"value": fd.conf.DLQGCSSinkCredential,
			"key":   "dlq_gcs_credential",
			"path":  mountFile,
		})
		conf.EnvVariables["DLQ_GCS_CREDENTIAL_PATH"] = credPath
	}

	if fd.conf.BigQuerySinkCredential != "" {
		const mountFile = "bigquery_auth.json"
		credPath := fmt.Sprintf("/etc/secret/%s", mountFile)

		mountSecrets = append(mountSecrets, map[string]any{
			"value": fd.conf.BigQuerySinkCredential,
			"key":   "bigquery_credential",
			"path":  mountFile,
		})
		conf.EnvVariables["SINK_BIGQUERY_CREDENTIAL_PATH"] = credPath
	}

	rc := helm.DefaultReleaseConfig()
	rc.Name = conf.DeploymentID
	rc.Repository = chartRepo
	rc.Chart = chartName
	rc.Namespace = conf.Namespace
	rc.ForceUpdate = true
	rc.Version = conf.ChartValues.ChartVersion
	rc.Values = map[string]any{
		labelsConfKey:  modules.CloneAndMergeMaps(deploymentLabels, entropyLabels),
		"replicaCount": conf.Replicas,
		"firehose": map[string]any{
			"image": map[string]any{
				"repository": imageRepo,
				"pullPolicy": conf.ChartValues.ImagePullPolicy,
				"tag":        conf.ChartValues.ImageTag,
			},
			"config": conf.EnvVariables,
			"resources": map[string]any{
				"limits": map[string]any{
					"cpu":    conf.Limits.CPU,
					"memory": conf.Limits.Memory,
				},
				"requests": map[string]any{
					"cpu":    conf.Requests.CPU,
					"memory": conf.Requests.Memory,
				},
			},
		},
		"tolerations": tolerations,
		"nodeAffinityMatchExpressions": map[string]any{
			"requiredDuringSchedulingIgnoredDuringExecution":  requiredDuringSchedulingIgnoredDuringExecution,
			"preferredDuringSchedulingIgnoredDuringExecution": preferredDuringSchedulingIgnoredDuringExecution,
		},
		"init-firehose": map[string]any{
			"enabled": fd.conf.InitContainer.Enabled,
			"image": map[string]any{
				"repository": fd.conf.InitContainer.Repository,
				"pullPolicy": fd.conf.InitContainer.PullPolicy,
				"tag":        fd.conf.InitContainer.ImageTag,
			},
			"command": fd.conf.InitContainer.Command,
			"args":    fd.conf.InitContainer.Args,
		},
		"telegraf": map[string]any{
			"enabled": telegrafConf.Enabled,
			"image":   telegrafConf.Image,
			"config": map[string]any{
				"output":                 telegrafConf.Config.Output,
				"additional_global_tags": telegrafConf.Config.AdditionalGlobalTags,
			},
		},
		"mountSecrets": mountSecrets,
	}

	return rc, nil
}

func renderTpl(labelsTpl map[string]string, labelsValues map[string]string) (map[string]string, error) {
	const useZeroValueForMissingKey = "missingkey=zero"

	finalLabels := map[string]string{}
	for k, v := range labelsTpl {
		var buf bytes.Buffer
		t, err := template.New("").Option(useZeroValueForMissingKey).Parse(v)
		if err != nil {
			return nil, errors.ErrInvalid.
				WithMsgf("label template for '%s' is invalid", k).WithCausef(err.Error())
		} else if err := t.Execute(&buf, labelsValues); err != nil {
			return nil, errors.ErrInvalid.
				WithMsgf("failed to render label template").WithCausef(err.Error())
		}

		// allow empty values
		//		labelVal := strings.TrimSpace(buf.String())
		//		if labelVal == "" {
		//			continue
		//		}

		finalLabels[k] = buf.String()
	}
	return finalLabels, nil
}

func mergeChartValues(cur, newVal *ChartValues) (*ChartValues, error) {
	if newVal == nil {
		return cur, nil
	}

	merged := ChartValues{
		ImageTag:        cur.ImageTag,
		ChartVersion:    cur.ChartVersion,
		ImagePullPolicy: cur.ImagePullPolicy,
	}

	newTag := strings.TrimSpace(newVal.ImageTag)
	if newTag != "" {
		if strings.Contains(newTag, ":") && !strings.HasPrefix(newTag, imageRepo) {
			return nil, errors.ErrInvalid.
				WithMsgf("unknown image repo: '%s', must start with '%s'", newTag, imageRepo)
		}
		merged.ImageTag = strings.TrimPrefix(newTag, imageRepo+":")
	}

	return &merged, nil
}

func readOutputData(exr module.ExpandedResource) (*Output, error) {
	var curOut Output
	if len(exr.Resource.State.Output) == 0 {
		return &curOut, nil
	}
	if err := json.Unmarshal(exr.Resource.State.Output, &curOut); err != nil {
		return nil, errors.ErrInternal.WithMsgf("corrupted output").WithCausef(err.Error())
	}
	return &curOut, nil
}

func readTransientData(exr module.ExpandedResource) (*transientData, error) {
	if len(exr.Resource.State.ModuleData) == 0 {
		return &transientData{}, nil
	}

	var modData transientData
	if err := json.Unmarshal(exr.Resource.State.ModuleData, &modData); err != nil {
		return nil, errors.ErrInternal.WithMsgf("corrupted transient data").WithCausef(err.Error())
	}
	return &modData, nil
}

func (us UsageSpec) merge(overide UsageSpec) UsageSpec {
	clone := us

	if overide.CPU != "" {
		clone.CPU = overide.CPU
	}

	if overide.Memory != "" {
		clone.Memory = overide.Memory
	}

	return clone
}

func renderTplOfMapStringAny(labelsTpl map[string]any, labelsValues map[string]string) (map[string]any, error) {
	outputMap := make(map[string]string)

	for key, value := range labelsTpl {
		if strValue, ok := value.(string); ok {
			outputMap[key] = strValue
		}
	}

	outputMap, err := renderTpl(outputMap, labelsValues)
	if err != nil {
		return nil, err
	}

	for key, val := range outputMap {
		labelsTpl[key] = val
	}

	return labelsTpl, nil
}
