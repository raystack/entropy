package firehose

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/modules"
	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/kube"
)

func TestFirehoseDriver_Output(t *testing.T) {
	t.Parallel()

	table := []struct {
		title      string
		kubeGetPod func(t *testing.T) kubeGetPodFn
		exr        module.ExpandedResource
		want       json.RawMessage
		wantErr    error
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
						ModuleData: modules.MustJSON(transientData{}),
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
						Output:     modules.MustJSON(Output{}),
						ModuleData: modules.MustJSON(transientData{}),
					},
				},
			},
			wantErr: errors.ErrInternal,
		},
		{
			title: "GetPod_Failure",
			exr: sampleResourceWithState(resource.State{
				Status: resource.StatusCompleted,
				Output: modules.MustJSON(Output{}),
			}, "LOG", "firehose"),
			kubeGetPod: func(t *testing.T) kubeGetPodFn {
				t.Helper()
				return func(ctx context.Context, conf kube.Config, ns string, labels map[string]string) ([]kube.Pod, error) {
					assert.Equal(t, ns, "firehose")
					assert.Equal(t, labels["app"], "firehose-foo-fh1")
					return nil, errors.New("failed")
				}
			},
			wantErr: errors.ErrInternal,
		},
		{
			title: "GetPod_Success",
			exr: sampleResourceWithState(resource.State{
				Status: resource.StatusCompleted,
				Output: modules.MustJSON(Output{
					Pods:        nil,
					Namespace:   "firehose",
					ReleaseName: "foo-bar",
				}),
			}, "LOG", "firehose"),
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
			want: modules.MustJSON(Output{
				Pods: []kube.Pod{
					{
						Name:       "foo-1",
						Containers: []string{"firehose"},
					},
				},
				Namespace:   "firehose",
				ReleaseName: "foo-bar",
			}),
		},
		{
			title: "Update_Namespace",
			exr: sampleResourceWithState(resource.State{
				Status: resource.StatusCompleted,
				Output: modules.MustJSON(Output{
					Pods:        nil,
					Namespace:   "firehose",
					ReleaseName: "foo-bar",
				}),
			}, "BIGQUERY", "bigquery-firehose"),
			kubeGetPod: func(t *testing.T) kubeGetPodFn {
				t.Helper()
				return func(ctx context.Context, conf kube.Config, ns string, labels map[string]string) ([]kube.Pod, error) {
					assert.Equal(t, ns, "bigquery-firehose")
					assert.Equal(t, labels["app"], "firehose-foo-fh1")
					return []kube.Pod{
						{
							Name:       "foo-1",
							Containers: []string{"firehose"},
						},
					}, nil
				}
			},
			want: modules.MustJSON(Output{
				Pods: []kube.Pod{
					{
						Name:       "foo-1",
						Containers: []string{"firehose"},
					},
				},
				Namespace:   "bigquery-firehose",
				ReleaseName: "foo-bar",
			}),
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			fd := &firehoseDriver{
				conf:    defaultDriverConf,
				timeNow: func() time.Time { return frozenTime },
			}

			fd.conf.Namespace = map[string]string{
				defaultKey: "firehose",
				"BIGQUERY": "bigquery-firehose",
			}

			if tt.kubeGetPod != nil {
				fd.kubeGetPod = tt.kubeGetPod(t)
			}

			got, err := fd.Output(context.Background(), tt.exr)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "wantErr=%v\ngotErr=%v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, got)
				assert.JSONEq(t, string(tt.want), string(got))
			}
		})
	}
}
