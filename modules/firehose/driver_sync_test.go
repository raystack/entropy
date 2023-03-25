package firehose

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/helm"
	"github.com/goto/entropy/pkg/kube"
)

func TestFirehoseDriver_Sync(t *testing.T) {
	t.Parallel()

	table := []struct {
		title      string
		kubeDeploy func(t *testing.T) kubeDeployFn
		kubeGetPod func(t *testing.T) kubeGetPodFn

		exr     module.ExpandedResource
		want    *resource.State
		wantErr error
	}{
		{
			title: "InvalidModuleData",
			exr: module.ExpandedResource{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					State:   resource.State{},
				},
			},
			wantErr: errors.ErrInternal,
		},
		{
			title: "InvalidOutput",
			exr: module.ExpandedResource{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					State: resource.State{
						ModuleData: mustJSON(transientData{}),
					},
				},
			},
			wantErr: errors.ErrInternal,
		},
		{
			title: "InvalidConfig",
			exr: module.ExpandedResource{
				Resource: resource.Resource{
					URN:     "urn:goto:entropy:foo:fh1",
					Kind:    "firehose",
					Name:    "fh1",
					Project: "foo",
					State: resource.State{
						Output:     mustJSON(Output{}),
						ModuleData: mustJSON(transientData{}),
					},
				},
			},
			wantErr: errors.ErrInternal,
		},
		{
			title: "NoPendingStep",
			exr: sampleResourceWithState(resource.State{
				Status: resource.StatusPending,
				Output: mustJSON(Output{}),
				ModuleData: mustJSON(transientData{
					PendingSteps: nil,
				}),
			}),
			kubeGetPod: func(t *testing.T) kubeGetPodFn {
				t.Helper()
				return func(ctx context.Context, conf kube.Config, ns string, labels map[string]string) ([]kube.Pod, error) {
					assert.Equal(t, ns, "firehose")
					assert.Equal(t, labels["app"], "firehose-foo-fh1")
					return []kube.Pod{
						{
							Name:       "foo-1",
							Containers: []string{"firehose"},
						},
					}, nil
				}
			},
			want: &resource.State{
				Status: resource.StatusCompleted,
				Output: mustJSON(Output{
					Pods: []kube.Pod{
						{
							Name:       "foo-1",
							Containers: []string{"firehose"},
						},
					},
				}),
				ModuleData: nil,
			},
		},
		{
			title: "Sync_refresh_output_failure",
			exr: sampleResourceWithState(resource.State{
				Status: resource.StatusCompleted,
				Output: mustJSON(Output{}),
			}),
			kubeGetPod: func(t *testing.T) kubeGetPodFn {
				t.Helper()
				return func(ctx context.Context, conf kube.Config, ns string, labels map[string]string) ([]kube.Pod, error) {
					return nil, errors.New("failed")
				}
			},
			wantErr: errors.ErrInternal,
		},
		{
			title: "Sync_release_create_failure",
			exr: sampleResourceWithState(resource.State{
				Status: resource.StatusPending,
				Output: mustJSON(Output{}),
				ModuleData: mustJSON(transientData{
					PendingSteps: []string{stepReleaseCreate},
				}),
			}),
			kubeDeploy: func(t *testing.T) kubeDeployFn {
				t.Helper()
				return func(ctx context.Context, isCreate bool, conf kube.Config, hc helm.ReleaseConfig) error {
					return errors.New("failed")
				}
			},
			kubeGetPod: func(t *testing.T) kubeGetPodFn {
				t.Helper()
				return func(ctx context.Context, conf kube.Config, ns string, labels map[string]string) ([]kube.Pod, error) {
					assert.Equal(t, ns, "firehose")
					assert.Equal(t, labels["app"], "firehose-foo-fh1")
					return []kube.Pod{
						{
							Name:       "foo-1",
							Containers: []string{"firehose"},
						},
					}, nil
				}
			},
			wantErr: errors.ErrInternal,
		},
		{
			title: "Sync_release_create_success",
			exr: sampleResourceWithState(resource.State{
				Status: resource.StatusPending,
				Output: mustJSON(Output{}),
				ModuleData: mustJSON(transientData{
					PendingSteps: []string{stepReleaseCreate},
				}),
			}),
			kubeDeploy: func(t *testing.T) kubeDeployFn {
				t.Helper()
				return func(ctx context.Context, isCreate bool, conf kube.Config, hc helm.ReleaseConfig) error {
					assert.True(t, isCreate)
					assert.Equal(t, hc.Values["replicaCount"], 1)
					return nil
				}
			},
			kubeGetPod: func(t *testing.T) kubeGetPodFn {
				t.Helper()
				return func(ctx context.Context, conf kube.Config, ns string, labels map[string]string) ([]kube.Pod, error) {
					assert.Equal(t, ns, "firehose")
					assert.Equal(t, labels["app"], "firehose-foo-fh1")
					return []kube.Pod{
						{
							Name:       "foo-1",
							Containers: []string{"firehose"},
						},
					}, nil
				}
			},
			want: &resource.State{
				Status: resource.StatusPending,
				Output: mustJSON(Output{}),
				ModuleData: mustJSON(transientData{
					PendingSteps: []string{},
				}),
				NextSyncAt: &frozenTime,
			},
		},
		{
			title: "Sync_release_stop_success",
			exr: sampleResourceWithState(resource.State{
				Status: resource.StatusPending,
				Output: mustJSON(Output{}),
				ModuleData: mustJSON(transientData{
					PendingSteps: []string{stepReleaseStop},
				}),
			}),
			kubeDeploy: func(t *testing.T) kubeDeployFn {
				t.Helper()
				return func(ctx context.Context, isCreate bool, conf kube.Config, hc helm.ReleaseConfig) error {
					assert.False(t, isCreate)
					assert.Equal(t, hc.Values["replicaCount"], 0)
					return nil
				}
			},
			kubeGetPod: func(t *testing.T) kubeGetPodFn {
				t.Helper()
				return func(ctx context.Context, conf kube.Config, ns string, labels map[string]string) ([]kube.Pod, error) {
					assert.Equal(t, ns, "firehose")
					assert.Equal(t, labels["app"], "firehose-foo-fh1")
					return []kube.Pod{
						{
							Name:       "foo-1",
							Containers: []string{"firehose"},
						},
					}, nil
				}
			},
			want: &resource.State{
				Status: resource.StatusPending,
				Output: mustJSON(Output{}),
				ModuleData: mustJSON(transientData{
					PendingSteps: []string{},
				}),
				NextSyncAt: &frozenTime,
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			fd := &firehoseDriver{
				conf:    defaultDriverConf,
				timeNow: func() time.Time { return frozenTime },
			}

			if tt.kubeGetPod != nil {
				fd.kubeGetPod = tt.kubeGetPod(t)
			}

			if tt.kubeDeploy != nil {
				fd.kubeDeploy = tt.kubeDeploy(t)
			}

			got, err := fd.Sync(context.Background(), tt.exr)
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

func sampleResourceWithState(state resource.State) module.ExpandedResource {
	return module.ExpandedResource{
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
			State: state,
		},
		Dependencies: map[string]module.ResolvedDependency{
			"kube_cluster": {
				Kind:   "kubernetes",
				Output: mustJSON(kubernetes.Output{}),
			},
		},
	}
}
