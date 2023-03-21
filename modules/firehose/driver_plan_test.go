package firehose

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
)

func TestFirehoseDriver_Plan(t *testing.T) {
	t.Parallel()

	table := []struct {
		title   string
		exr     module.ExpandedResource
		act     module.ActionRequest
		want    *module.Plan
		wantErr error
	}{
		// create action tests
		{
			title: "Create_InvalidParamsJSON",
			exr:   module.ExpandedResource{},
			act: module.ActionRequest{
				Name:   module.CreateAction,
				Params: []byte("{"),
			},
			wantErr: errors.ErrInvalid,
		},
		{
			title: "Create_InvalidParamsValue",
			exr:   module.ExpandedResource{},
			act: module.ActionRequest{
				Name:   module.CreateAction,
				Params: []byte("{}"),
			},
			wantErr: errors.ErrInvalid,
		},
		{
			title: "Create_LongName",
			exr: module.ExpandedResource{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:ABCDEFGHIJKLMNOPQRSTUVWXYZ:abcdefghijklmnopqrstuvwxyz",
					Kind:    "firehose",
					Name:    "abcdefghijklmnopqrstuvwxyz",
					Project: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
				},
			},
			act: module.ActionRequest{
				Name: module.CreateAction,
				Params: mustJSON(map[string]any{
					"replicas": 1,
					"env_variables": map[string]string{
						"SINK_TYPE":                "LOG",
						"INPUT_SCHEMA_PROTO_CLASS": "com.foo.Bar",
						"SOURCE_KAFKA_BROKERS":     "localhost:9092",
						"SOURCE_KAFKA_TOPIC":       "foo-log",
					},
				}),
			},
			want: &module.Plan{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:ABCDEFGHIJKLMNOPQRSTUVWXYZ:abcdefghijklmnopqrstuvwxyz",
					Kind:    "firehose",
					Name:    "abcdefghijklmnopqrstuvwxyz",
					Project: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
					Spec: resource.Spec{
						Configs: mustJSON(map[string]any{
							"replicas":      1,
							"namespace":     "firehose",
							"deployment_id": "firehose-ABCDEFGHIJKLMNOPQRSTUVWXYZ-abcdefghij-9bf099",
							"telegraf": map[string]any{
								"enabled": false,
							},
							"chart_values": map[string]string{
								"chart_version":     "0.1.3",
								"image_pull_policy": "IfNotPresent",
								"image_tag":         "latest",
							},
							"env_variables": map[string]string{
								"SINK_TYPE":                      "LOG",
								"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
								"SOURCE_KAFKA_CONSUMER_GROUP_ID": "firehose-ABCDEFGHIJKLMNOPQRSTUVWXYZ-abcdefghij-9bf099-0001",
								"SOURCE_KAFKA_BROKERS":           "localhost:9092",
								"SOURCE_KAFKA_TOPIC":             "foo-log",
							},
						}),
					},
					State: resource.State{
						Status: resource.StatusPending,
						Output: mustJSON(Output{
							Namespace:   "firehose",
							ReleaseName: "firehose-ABCDEFGHIJKLMNOPQRSTUVWXYZ-abcdefghij-9bf099",
						}),
						ModuleData: mustJSON(transientData{
							PendingSteps: []string{stepReleaseCreate},
						}),
					},
				},
				Reason: "firehose_create",
			},
			wantErr: nil,
		},
		{
			title: "Create_ValidRequest",
			exr: module.ExpandedResource{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
				},
			},
			act: module.ActionRequest{
				Name: module.CreateAction,
				Params: mustJSON(map[string]any{
					"replicas": 1,
					"env_variables": map[string]string{
						"SINK_TYPE":                      "LOG",
						"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
						"SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
						"SOURCE_KAFKA_BROKERS":           "localhost:9092",
						"SOURCE_KAFKA_TOPIC":             "foo-log",
					},
				}),
			},
			want: &module.Plan{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					Spec: resource.Spec{
						Configs: mustJSON(map[string]any{
							"replicas":      1,
							"namespace":     "firehose",
							"deployment_id": "firehose-foo-fh1",
							"telegraf": map[string]any{
								"enabled": false,
							},
							"chart_values": map[string]string{
								"chart_version":     "0.1.3",
								"image_pull_policy": "IfNotPresent",
								"image_tag":         "latest",
							},
							"env_variables": map[string]string{
								"SINK_TYPE":                      "LOG",
								"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
								"SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
								"SOURCE_KAFKA_BROKERS":           "localhost:9092",
								"SOURCE_KAFKA_TOPIC":             "foo-log",
							},
						}),
					},
					State: resource.State{
						Status: resource.StatusPending,
						Output: mustJSON(Output{
							Namespace:   "firehose",
							ReleaseName: "firehose-foo-fh1",
						}),
						ModuleData: mustJSON(transientData{
							PendingSteps: []string{stepReleaseCreate},
						}),
					},
				},
				Reason: "firehose_create",
			},
			wantErr: nil,
		},

		// update action tests
		{
			title: "Update_Valid",
			exr: module.ExpandedResource{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					Spec: resource.Spec{
						Configs: mustJSON(map[string]any{
							"replicas":      1,
							"deployment_id": "firehose-deployment-x",
							"env_variables": map[string]string{
								"SINK_TYPE":                "LOG",
								"INPUT_SCHEMA_PROTO_CLASS": "com.foo.Bar",
								"SOURCE_KAFKA_BROKERS":     "localhost:9092",
								"SOURCE_KAFKA_TOPIC":       "foo-log",
							},
						}),
					},
					State: resource.State{
						Status: resource.StatusCompleted,
						Output: mustJSON(Output{
							Namespace:   "foo",
							ReleaseName: "bar",
						}),
					},
				},
			},
			act: module.ActionRequest{
				Name: module.UpdateAction,
				Params: mustJSON(map[string]any{
					"replicas": 10,
					"env_variables": map[string]string{
						"SINK_TYPE":                      "HTTP", // the change being applied
						"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
						"SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
						"SOURCE_KAFKA_BROKERS":           "localhost:9092",
						"SOURCE_KAFKA_TOPIC":             "foo-log",
					},
				}),
			},
			want: &module.Plan{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					Spec: resource.Spec{
						Configs: mustJSON(map[string]any{
							"replicas":      10,
							"deployment_id": "firehose-deployment-x",
							"env_variables": map[string]string{
								"SINK_TYPE":                      "HTTP",
								"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
								"SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
								"SOURCE_KAFKA_BROKERS":           "localhost:9092",
								"SOURCE_KAFKA_TOPIC":             "foo-log",
							},
						}),
					},
					State: resource.State{
						Status: resource.StatusPending,
						Output: mustJSON(Output{
							Namespace:   "foo",
							ReleaseName: "bar",
						}),
						ModuleData: mustJSON(transientData{
							PendingSteps: []string{stepReleaseUpdate},
						}),
					},
				},
				Reason: "firehose_update",
			},
			wantErr: nil,
		},

		// reset action tests
		{
			title: "Reset_InValid",
			exr: module.ExpandedResource{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					Spec: resource.Spec{
						Configs: mustJSON(map[string]any{
							"replicas":      1,
							"deployment_id": "firehose-deployment-x",
							"env_variables": map[string]string{
								"SINK_TYPE":                      "LOG",
								"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
								"SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
								"SOURCE_KAFKA_BROKERS":           "localhost:9092",
								"SOURCE_KAFKA_TOPIC":             "foo-log",
							},
						}),
					},
					State: resource.State{
						Status: resource.StatusCompleted,
						Output: mustJSON(Output{
							Namespace:   "foo",
							ReleaseName: "bar",
						}),
					},
				},
			},
			act: module.ActionRequest{
				Name: ResetAction,
				Params: mustJSON(map[string]any{
					"reset_to": "some_random",
				}),
			},
			wantErr: errors.ErrInvalid,
		},
		{
			title: "Reset_Valid",
			exr: module.ExpandedResource{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					Spec: resource.Spec{
						Configs: mustJSON(map[string]any{
							"replicas":      1,
							"deployment_id": "firehose-deployment-x",
							"env_variables": map[string]string{
								"SINK_TYPE":                      "LOG",
								"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
								"SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
								"SOURCE_KAFKA_BROKERS":           "localhost:9092",
								"SOURCE_KAFKA_TOPIC":             "foo-log",
							},
						}),
					},
					State: resource.State{
						Status: resource.StatusCompleted,
						Output: mustJSON(Output{
							Namespace:   "foo",
							ReleaseName: "bar",
						}),
					},
				},
			},
			act: module.ActionRequest{
				Name: ResetAction,
				Params: mustJSON(map[string]any{
					"to": "latest",
				}),
			},
			want: &module.Plan{
				Reason: "firehose_reset",
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					Spec: resource.Spec{
						Configs: mustJSON(map[string]any{
							"replicas":      1,
							"deployment_id": "firehose-deployment-x",
							"env_variables": map[string]string{
								"SINK_TYPE":                      "LOG",
								"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
								"SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
								"SOURCE_KAFKA_BROKERS":           "localhost:9092",
								"SOURCE_KAFKA_TOPIC":             "foo-log",
							},
						}),
					},
					State: resource.State{
						Status: resource.StatusPending,
						Output: mustJSON(Output{
							Namespace:   "foo",
							ReleaseName: "bar",
						}),
						ModuleData: mustJSON(transientData{
							ResetOffsetTo: "latest",
							PendingSteps: []string{
								stepReleaseStop,
								stepKafkaReset,
								stepReleaseUpdate,
							},
						}),
					},
				},
			},
		},

		// upgrade action tests
		{
			title: "Upgrade_Valid",
			exr: module.ExpandedResource{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					Spec: resource.Spec{
						Configs: mustJSON(map[string]any{
							"replicas":      1,
							"deployment_id": "firehose-deployment-x",
							"chart_values": map[string]string{
								"chart_version":     "0.1.0",
								"image_pull_policy": "IfNotPresent",
								"image_tag":         "latest",
							},
							"env_variables": map[string]string{
								"SINK_TYPE":                      "LOG",
								"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
								"SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
								"SOURCE_KAFKA_BROKERS":           "localhost:9092",
								"SOURCE_KAFKA_TOPIC":             "foo-log",
							},
						}),
					},
					State: resource.State{
						Status: resource.StatusCompleted,
						Output: mustJSON(Output{
							Namespace:   "foo",
							ReleaseName: "bar",
						}),
					},
				},
			},
			act: module.ActionRequest{
				Name: UpgradeAction,
			},
			want: &module.Plan{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					Spec: resource.Spec{
						Configs: mustJSON(map[string]any{
							"replicas":      1,
							"deployment_id": "firehose-deployment-x",
							"chart_values": map[string]string{
								"chart_version":     "0.1.3",
								"image_pull_policy": "IfNotPresent",
								"image_tag":         "latest",
							},
							"env_variables": map[string]string{
								"SINK_TYPE":                      "LOG",
								"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
								"SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
								"SOURCE_KAFKA_BROKERS":           "localhost:9092",
								"SOURCE_KAFKA_TOPIC":             "foo-log",
							},
						}),
					},
					State: resource.State{
						Status: resource.StatusPending,
						Output: mustJSON(Output{
							Namespace:   "foo",
							ReleaseName: "bar",
						}),
						ModuleData: mustJSON(transientData{
							PendingSteps: []string{stepReleaseUpdate},
						}),
					},
				},
				Reason: "firehose_upgrade",
			},
			wantErr: nil,
		},

		// scale action tests
		{
			title: "Scale_Invalid_params",
			exr: module.ExpandedResource{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					Spec: resource.Spec{
						Configs: mustJSON(map[string]any{
							"replicas":      1,
							"deployment_id": "firehose-deployment-x",
							"chart_values": map[string]string{
								"chart_version":     "0.1.0",
								"image_pull_policy": "IfNotPresent",
								"image_tag":         "latest",
							},
							"env_variables": map[string]string{
								"SINK_TYPE":                      "LOG",
								"INPUT_SCHEMA_PROTO_CLASS":       "com.foo.Bar",
								"SOURCE_KAFKA_CONSUMER_GROUP_ID": "foo-bar-baz",
								"SOURCE_KAFKA_BROKERS":           "localhost:9092",
								"SOURCE_KAFKA_TOPIC":             "foo-log",
							},
						}),
					},
					State: resource.State{
						Status: resource.StatusCompleted,
						Output: mustJSON(Output{
							Namespace:   "foo",
							ReleaseName: "bar",
						}),
					},
				},
			},
			act: module.ActionRequest{
				Name:   ScaleAction,
				Params: []byte("{}"),
			},
			wantErr: errors.ErrInvalid,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			dr := &firehoseDriver{
				conf: defaultDriverConf,
			}

			got, err := dr.Plan(context.Background(), tt.exr, tt.act)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, got)
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
