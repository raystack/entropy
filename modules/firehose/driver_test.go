package firehose

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/helm"
)

func TestFirehoseDriver(t *testing.T) {
	t.Parallel()

	table := []struct {
		title      string
		res        resource.Resource
		kubeOutput kubernetes.Output
		want       *helm.ReleaseConfig
		wantErr    error
	}{
		{
			title: "LOG_Sink",
			res: resource.Resource{
				URN:     "orn:entropy:firehose:project-1:resource-1-firehose",
				Kind:    "firehose",
				Name:    "resource-1",
				Project: "project-1",
				Labels: map[string]string{
					"team": "team-1",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				UpdatedBy: "john.doe@goto.com",
				CreatedBy: "john.doe@goto.com",
				Spec: resource.Spec{
					Configs: []byte(`{
                                     "env_variables": {
                                         "SINK_TYPE": "LOG",
                                         "INPUT_SCHEMA_PROTO_CLASS": "com.foo.Bar",
                                         "SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
                                         "SOURCE_KAFKA_BROKERS": "localhost:9092",
                                         "SOURCE_KAFKA_TOPIC": "foo-log"
                                     },
                                     "replicas": 1
                                 }`),
					Dependencies: map[string]string{},
				},
				State: resource.State{
					Status: resource.StatusPending,
					Output: nil,
				},
			},
			kubeOutput: kubernetes.Output{
				Tolerations: map[string][]kubernetes.Toleration{
					"firehose_LOG": {
						{
							Key:      "key1",
							Operator: "Equal",
							Value:    "value1",
							Effect:   "NoSchedule",
						},
					},
					"firehose_BIGQUERY": {
						{
							Key:      "key2",
							Operator: "Equal",
							Value:    "value2",
							Effect:   "NoSchedule",
						},
					},
				},
			},
			want: &helm.ReleaseConfig{
				Name:        "project-1-resource-1-firehose",
				Repository:  "https://goto.github.io/charts/",
				Chart:       "firehose",
				Version:     "0.1.13",
				Namespace:   "namespace-1",
				Timeout:     300,
				Wait:        true,
				ForceUpdate: true,
				Values: map[string]any{
					"firehose": map[string]any{
						"config": map[string]any{
							"DEFAULT_KEY_IN_FIREHOSE_MODULE_1": "default-key-in-firehose-module-value_1",
							"DEFAULT_KEY_IN_FIREHOSE_MODULE_2": "default-key-in-firehose-module-value_2",
							"DLQ_GCS_CREDENTIAL_PATH":          "/etc/secret/dlq_gcs_auth.json",
							"INPUT_SCHEMA_PROTO_CLASS":         "com.foo.Bar",
							"SINK_BIGQUERY_CREDENTIAL_PATH":    "/etc/secret/bigquery_auth.json",
							"SINK_BIGTABLE_CREDENTIAL_PATH":    "/etc/secret/gcs_auth.json",
							"SINK_BLOB_GCS_CREDENTIAL_PATH":    "/etc/secret/gcs_auth.json",
							"SINK_TYPE":                        "LOG",
							"SOURCE_KAFKA_BROKERS":             "localhost:9092",
							"SOURCE_KAFKA_CONSUMER_GROUP_ID":   "foo-bar-baz",
							"SOURCE_KAFKA_TOPIC":               "foo-log",
						},
						"image": map[string]any{
							"pullPolicy": "IfNotPresent",
							"repository": "gotocompany/firehose",
							"tag":        "0.8.1",
						},
						"resources": map[string]any{
							"limits": map[string]any{
								"cpu":    "6000m",
								"memory": "6000Mi",
							},
							"requests": map[string]any{
								"cpu":    "600m",
								"memory": "2500Mi",
							},
						},
					},
					"init-firehose": map[string]any{
						"enabled": true,
						"image": map[string]any{
							"repository": "busybox",
							"pullPolicy": "IfNotPresent",
							"tag":        "latest",
						},
						"command": []string{"cmd1", "--a"},
						"args":    []string{"arg1", "arg2"},
					},
					"labels": map[string]string{
						"deployment":   "project-1-resource-1-firehose",
						"team":         "team-1",
						"orchestrator": "entropy",
					},
					"mountSecrets": []map[string]string{
						{
							"key":   "gcs_credential",
							"path":  "gcs_auth.json",
							"value": "gcs-credential",
						},
						{
							"key":   "dlq_gcs_credential",
							"path":  "dlq_gcs_auth.json",
							"value": "dlq-gcs-credential",
						},
						{
							"key":   "bigquery_credential",
							"path":  "bigquery_auth.json",
							"value": "big-query-credential",
						},
					},
					"nodeAffinityMatchExpressions": map[string]any{
						"preferredDuringSchedulingIgnoredDuringExecution": []WeightedPreference{
							{
								Weight: 1,
								Preference: []Preference{
									{
										Key:      "another-node-label-key",
										Operator: "In",
										Values:   []string{"another-node-label-value"},
									},
								},
							},
						},
						"requiredDuringSchedulingIgnoredDuringExecution": []Preference{
							{
								Key:      "topology.kubernetes.io/zone",
								Operator: "In",
								Values:   []string{"antarctica-east1", "antarctica-west1"},
							},
						},
					},
					"replicaCount": 1,
					"telegraf": map[string]any{
						"enabled": true,
						"image": map[string]string{
							"pullPolicy": "IfNotPresent",
							"repository": "telegraf",
							"tag":        "1.18.0-alpine",
						},
						"config": map[string]any{
							"output": map[string]any{
								"prometheus_remote_write": map[string]any{
									"enabled": true,
									"url":     "http://goto.com",
								},
							},
							"additional_global_tags": map[string]string{
								"app": "orn:entropy:firehose:project-1:resource-1-firehose",
							},
						},
					},
					"tolerations": []map[string]any{
						{
							"key":      "key1",
							"operator": "Equal",
							"value":    "value1",
							"effect":   "NoSchedule",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			title: "BIGQUERY_Sink",
			res: resource.Resource{
				URN:     "orn:entropy:firehose:project-1:resource-2-firehose",
				Kind:    "firehose",
				Name:    "resource-2",
				Project: "project-1",
				Labels: map[string]string{
					"team": "team-2",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				UpdatedBy: "john.doe2@goto.com",
				CreatedBy: "john.doe2@goto.com",
				Spec: resource.Spec{
					Configs: []byte(`{
                                     "env_variables": {
                                         "SINK_TYPE": "BIGQUERY",
                                         "INPUT_SCHEMA_PROTO_CLASS": "com.foo.Bar-2",
                                         "SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz-2",
                                         "SOURCE_KAFKA_BROKERS": "localhost:9093",
                                         "SOURCE_KAFKA_TOPIC": "foo-log-2"
                                     },
                                     "replicas": 2
                                 }`),
					Dependencies: map[string]string{},
				},
				State: resource.State{
					Status: resource.StatusPending,
					Output: nil,
				},
			},
			kubeOutput: kubernetes.Output{
				Tolerations: map[string][]kubernetes.Toleration{
					"firehose_LOG": {
						{
							Key:      "key1",
							Operator: "Equal",
							Value:    "value1",
							Effect:   "NoSchedule",
						},
					},
					"firehose_BIGQUERY": {
						{
							Key:      "key2",
							Operator: "Equal",
							Value:    "value2",
							Effect:   "NoSchedule",
						},
					},
				},
			},
			want: &helm.ReleaseConfig{
				Name:        "project-1-resource-2-firehose",
				Repository:  "https://goto.github.io/charts/",
				Chart:       "firehose",
				Version:     "0.1.13",
				Namespace:   "namespace-1",
				Timeout:     300,
				Wait:        true,
				ForceUpdate: true,
				Values: map[string]any{
					"firehose": map[string]any{
						"config": map[string]any{
							"DEFAULT_KEY_IN_FIREHOSE_MODULE_1": "default-key-in-firehose-module-value_1",
							"DEFAULT_KEY_IN_FIREHOSE_MODULE_2": "default-key-in-firehose-module-value_2",
							"DLQ_GCS_CREDENTIAL_PATH":          "/etc/secret/dlq_gcs_auth.json",
							"INPUT_SCHEMA_PROTO_CLASS":         "com.foo.Bar-2",
							"SINK_BIGQUERY_CREDENTIAL_PATH":    "/etc/secret/bigquery_auth.json",
							"SINK_BIGTABLE_CREDENTIAL_PATH":    "/etc/secret/gcs_auth.json",
							"SINK_BLOB_GCS_CREDENTIAL_PATH":    "/etc/secret/gcs_auth.json",
							"SINK_TYPE":                        "BIGQUERY",
							"SOURCE_KAFKA_BROKERS":             "localhost:9093",
							"SOURCE_KAFKA_CONSUMER_GROUP_ID":   "foo-bar-baz-2",
							"SOURCE_KAFKA_TOPIC":               "foo-log-2",
						},
						"image": map[string]any{
							"pullPolicy": "IfNotPresent",
							"repository": "gotocompany/firehose",
							"tag":        "0.8.1",
						},
						"resources": map[string]any{
							"limits": map[string]any{
								"cpu":    "6000m",
								"memory": "20000Mi",
							},
							"requests": map[string]any{
								"cpu":    "300m",
								"memory": "2000Mi",
							},
						},
					},
					"init-firehose": map[string]any{
						"enabled": true,
						"image": map[string]any{
							"repository": "busybox",
							"pullPolicy": "IfNotPresent",
							"tag":        "latest",
						},
						"command": []string{"cmd1", "--a"},
						"args":    []string{"arg1", "arg2"},
					},
					"labels": map[string]string{
						"deployment":   "project-1-resource-2-firehose",
						"team":         "team-2",
						"orchestrator": "entropy",
					},
					"mountSecrets": []map[string]string{
						{
							"key":   "gcs_credential",
							"path":  "gcs_auth.json",
							"value": "gcs-credential",
						},
						{
							"key":   "dlq_gcs_credential",
							"path":  "dlq_gcs_auth.json",
							"value": "dlq-gcs-credential",
						},
						{
							"key":   "bigquery_credential",
							"path":  "bigquery_auth.json",
							"value": "big-query-credential",
						},
					},
					"nodeAffinityMatchExpressions": map[string]any{
						"preferredDuringSchedulingIgnoredDuringExecution": []WeightedPreference{
							{
								Weight: 1,
								Preference: []Preference{
									{
										Key:      "another-node-label-key",
										Operator: "In",
										Values:   []string{"another-node-label-value"},
									},
								},
							},
						},
						"requiredDuringSchedulingIgnoredDuringExecution": []Preference{
							{
								Key:      "topology.kubernetes.io/zone",
								Operator: "In",
								Values:   []string{"antarctica-east1", "antarctica-west1"},
							},
						},
					},
					"replicaCount": 2,
					"telegraf": map[string]any{
						"enabled": true,
						"image": map[string]string{
							"pullPolicy": "IfNotPresent",
							"repository": "telegraf",
							"tag":        "1.18.0-alpine",
						},
						"config": map[string]any{
							"output": map[string]any{
								"prometheus_remote_write": map[string]any{
									"enabled": true,
									"url":     "http://goto.com",
								},
							},
							"additional_global_tags": map[string]string{
								"app": "orn:entropy:firehose:project-1:resource-2-firehose",
							},
						},
					},
					"tolerations": []map[string]any{
						{
							"key":      "key2",
							"operator": "Equal",
							"value":    "value2",
							"effect":   "NoSchedule",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			title: "BLOB_Sink",
			res: resource.Resource{
				URN:     "orn:entropy:firehose:project-1:resource-3-firehose",
				Kind:    "firehose",
				Name:    "resource-3",
				Project: "project-1",
				Labels: map[string]string{
					"team": "team-3",
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				UpdatedBy: "john.doe3@goto.com",
				CreatedBy: "john.doe3@goto.com",
				Spec: resource.Spec{
					Configs: []byte(`{
                                     "env_variables": {
                                         "SINK_TYPE": "BLOB",
                                         "INPUT_SCHEMA_PROTO_CLASS": "com.foo.Bar-3",
                                         "SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz-3",
                                         "SOURCE_KAFKA_BROKERS": "localhost:9094",
                                         "SOURCE_KAFKA_TOPIC": "foo-log-3"
                                     },
                                     "replicas": 3
                                 }`),
					Dependencies: map[string]string{},
				},
				State: resource.State{
					Status: resource.StatusPending,
					Output: nil,
				},
			},
			kubeOutput: kubernetes.Output{
				Tolerations: map[string][]kubernetes.Toleration{
					"firehose_LOG": {
						{
							Key:      "key1",
							Operator: "Equal",
							Value:    "value1",
							Effect:   "NoSchedule",
						},
					},
					"firehose_BIGQUERY": {
						{
							Key:      "key2",
							Operator: "Equal",
							Value:    "value2",
							Effect:   "NoSchedule",
						},
					},
					"firehose_BLOB": {
						{
							Key:      "key3",
							Operator: "Equal",
							Value:    "value3",
							Effect:   "NoSchedule",
						},
					},
				},
			},
			want: &helm.ReleaseConfig{
				Name:        "project-1-resource-3-firehose",
				Repository:  "https://goto.github.io/charts/",
				Chart:       "firehose",
				Version:     "0.1.13",
				Namespace:   "namespace-1",
				Timeout:     300,
				Wait:        true,
				ForceUpdate: true,
				Values: map[string]any{
					"firehose": map[string]any{
						"config": map[string]any{
							"DEFAULT_KEY_IN_FIREHOSE_MODULE_1": "default-key-in-firehose-module-value_1",
							"DEFAULT_KEY_IN_FIREHOSE_MODULE_2": "default-key-in-firehose-module-value_2",
							"DLQ_GCS_CREDENTIAL_PATH":          "/etc/secret/dlq_gcs_auth.json",
							"INPUT_SCHEMA_PROTO_CLASS":         "com.foo.Bar-3",
							"SINK_BIGQUERY_CREDENTIAL_PATH":    "/etc/secret/bigquery_auth.json",
							"SINK_BIGTABLE_CREDENTIAL_PATH":    "/etc/secret/gcs_auth.json",
							"SINK_BLOB_GCS_CREDENTIAL_PATH":    "/etc/secret/gcs_auth.json",
							"SINK_TYPE":                        "BLOB",
							"SOURCE_KAFKA_BROKERS":             "localhost:9094",
							"SOURCE_KAFKA_CONSUMER_GROUP_ID":   "foo-bar-baz-3",
							"SOURCE_KAFKA_TOPIC":               "foo-log-3",
						},
						"image": map[string]any{
							"pullPolicy": "IfNotPresent",
							"repository": "gotocompany/firehose",
							"tag":        "0.8.1",
						},
						"resources": map[string]any{
							"limits": map[string]any{
								"cpu":    "6000m",
								"memory": "20000Mi",
							},
							"requests": map[string]any{
								"cpu":    "300m",
								"memory": "2000Mi",
							},
						},
					},
					"init-firehose": map[string]any{
						"enabled": true,
						"image": map[string]any{
							"repository": "busybox",
							"pullPolicy": "IfNotPresent",
							"tag":        "latest",
						},
						"command": []string{"cmd1", "--a"},
						"args":    []string{"arg1", "arg2"},
					},
					"labels": map[string]string{
						"deployment":   "project-1-resource-3-firehose",
						"team":         "team-3",
						"orchestrator": "entropy",
					},
					"mountSecrets": []map[string]string{
						{
							"key":   "gcs_credential",
							"path":  "gcs_auth.json",
							"value": "gcs-credential",
						},
						{
							"key":   "dlq_gcs_credential",
							"path":  "dlq_gcs_auth.json",
							"value": "dlq-gcs-credential",
						},
						{
							"key":   "bigquery_credential",
							"path":  "bigquery_auth.json",
							"value": "big-query-credential",
						},
					},
					"nodeAffinityMatchExpressions": map[string]any{
						"preferredDuringSchedulingIgnoredDuringExecution": []WeightedPreference{
							{
								Weight: 1,
								Preference: []Preference{
									{
										Key:      "another-node-label-key",
										Operator: "In",
										Values:   []string{"another-node-label-value"},
									},
								},
							},
						},
						"requiredDuringSchedulingIgnoredDuringExecution": []Preference{
							{
								Key:      "topology.kubernetes.io/zone",
								Operator: "In",
								Values:   []string{"antarctica-east1", "antarctica-west1"},
							},
						},
					},
					"replicaCount": 3,
					"telegraf": map[string]any{
						"enabled": true,
						"image": map[string]string{
							"pullPolicy": "IfNotPresent",
							"repository": "telegraf",
							"tag":        "1.18.0-alpine",
						},
						"config": map[string]any{
							"output": map[string]any{
								"prometheus_remote_write": map[string]any{
									"enabled": true,
									"url":     "http://goto.com",
								},
							},
							"additional_global_tags": map[string]string{
								"app": "orn:entropy:firehose:project-1:resource-3-firehose",
							},
						},
					},
					"tolerations": []map[string]any{
						{
							"key":      "key3",
							"operator": "Equal",
							"value":    "value3",
							"effect":   "NoSchedule",
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			fd := &firehoseDriver{
				conf:    firehoseDriverConf(),
				timeNow: func() time.Time { return frozenTime },
			}

			conf, _ := readConfig(tt.res, tt.res.Spec.Configs, fd.conf)
			chartVals, _ := mergeChartValues(&fd.conf.ChartValues, conf.ChartValues)

			conf.Telegraf = fd.conf.Telegraf
			conf.Namespace = fd.conf.Namespace
			conf.ChartValues = chartVals

			got, err := fd.getHelmRelease(tt.res, *conf, tt.kubeOutput)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "wantErr=%v\ngotErr=%v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, got)

				wantJSON := string(mustJSON(tt.want))
				gotJSON := string(mustJSON(got))
				assert.JSONEq(t, wantJSON, gotJSON)
			}
		})
	}
}

func firehoseDriverConf() driverConf {
	return driverConf{
		NodeAffinityMatchExpressions: NodeAffinityMatchExpressions{
			RequiredDuringSchedulingIgnoredDuringExecution: []Preference{
				{
					Key:      "topology.kubernetes.io/zone",
					Operator: "In",
					Values:   []string{"antarctica-east1", "antarctica-west1"},
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []WeightedPreference{
				{
					Weight: 1,
					Preference: []Preference{
						{
							Key:      "another-node-label-key",
							Operator: "In",
							Values:   []string{"another-node-label-value"},
						},
					},
				},
			},
		},
		EnvVariables: map[string]string{
			"DEFAULT_KEY_IN_FIREHOSE_MODULE_1": "default-key-in-firehose-module-value_1",
			"DEFAULT_KEY_IN_FIREHOSE_MODULE_2": "default-key-in-firehose-module-value_2",
		},
		ChartValues: ChartValues{
			ChartVersion:    "0.1.13",
			ImageTag:        "0.8.1",
			ImagePullPolicy: "IfNotPresent",
		},
		BigQuerySinkCredential: "big-query-credential",
		GCSSinkCredential:      "gcs-credential",
		DLQGCSSinkCredential:   "dlq-gcs-credential",
		InitContainer: InitContainer{
			Args:       []string{"arg1", "arg2"},
			Command:    []string{"cmd1", "--a"},
			Enabled:    true,
			ImageTag:   "latest",
			PullPolicy: "IfNotPresent",
			Repository: "busybox",
		},
		Labels: map[string]string{
			"team": "{{.team}}",
		},
		Namespace: "namespace-1",
		RequestsAndLimits: map[string]RequestsAndLimits{
			"BIGQUERY": {
				Limits: UsageSpec{
					CPU:    "6000m",
					Memory: "20000Mi",
				},
				Requests: UsageSpec{
					CPU:    "300m",
					Memory: "2000Mi",
				},
			},
			"BLOB": {
				Limits: UsageSpec{
					CPU:    "6000m",
					Memory: "20000Mi",
				},
				Requests: UsageSpec{
					CPU:    "300m",
					Memory: "2000Mi",
				},
			},
			"default": {
				Limits: UsageSpec{
					CPU:    "6000m",
					Memory: "6000Mi",
				},
				Requests: UsageSpec{
					CPU:    "600m",
					Memory: "2500Mi",
				},
			},
		},
		Telegraf: &Telegraf{
			Enabled: true,
			Image: map[string]any{
				"pullPolicy": "IfNotPresent",
				"repository": "telegraf",
				"tag":        "1.18.0-alpine",
			},
			Config: TelegrafConf{
				Output: map[string]any{
					"prometheus_remote_write": map[string]any{
						"enabled": true,
						"url":     "http://goto.com",
					},
				},
				AdditionalGlobalTags: map[string]string{
					"app": "{{ .urn }}",
				},
			},
		},
	}
}
