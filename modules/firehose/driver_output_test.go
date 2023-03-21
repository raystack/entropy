package firehose

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/core/resource"
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
			title: "GetPod_Failure",
			exr: sampleResourceWithState(resource.State{
				Status: resource.StatusCompleted,
				Output: mustJSON(Output{}),
			}),
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
				Output: mustJSON(Output{
					Pods:        nil,
					Namespace:   "firehose",
					ReleaseName: "foo-bar",
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
			want: mustJSON(Output{
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
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			fd := &firehoseDriver{
				conf: defaultDriverConf,
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
