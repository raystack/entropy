package firehose

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func TestFirehoseModule_Plan(t *testing.T) {
	t.Parallel()

	res := resource.Resource{
		URN:     "urn:odpf:entropy:firehose:test",
		Kind:    "firehose",
		Name:    "test",
		Project: "demo",
		Spec: resource.Spec{
			Configs: []byte(`{"release_configs": {"values": {"replicaCount": 1, "firehose": {}}}}`),
		},
		State: resource.State{},
	}

	table := []struct {
		title   string
		spec    module.Spec
		act     module.ActionRequest
		want    *resource.Resource
		wantErr error
	}{
		{
			title: "InvalidConfiguration",
			spec:  module.Spec{Resource: res},
			act: module.ActionRequest{
				Name:   module.CreateAction,
				Params: []byte(`{`),
			},
			wantErr: errors.ErrInvalid,
		},
		{
			title: "ValidConfiguration",
			spec:  module.Spec{Resource: res},
			act: module.ActionRequest{
				Name:   module.CreateAction,
				Params: []byte(`{}`),
			},
			want: &resource.Resource{
				URN:     "urn:odpf:entropy:firehose:test",
				Kind:    "firehose",
				Name:    "test",
				Project: "demo",
				Spec: resource.Spec{
					Configs: []byte(`{"state":"","release_configs":{"name":"demo-test-firehose","repository":"https://odpf.github.io/charts/","chart":"firehose","version":"0.1.1","values":null,"namespace":"firehose","timeout":0,"force_update":true,"recreate_pods":false,"wait":false,"wait_for_jobs":false,"replace":false,"description":"","create_namespace":false}}`),
				},
				State: resource.State{
					Status:     resource.StatusPending,
					ModuleData: []byte(`{"pending_steps":["helm_create"]}`),
				},
			},
		},
		{
			title: "InvalidActionParams",
			spec:  module.Spec{Resource: res},
			act: module.ActionRequest{
				Name:   ScaleAction,
				Params: []byte(`{`),
			},
			wantErr: errors.ErrInvalid,
		},
		{
			title: "ValidScaleRequest",
			spec:  module.Spec{Resource: res},
			act: module.ActionRequest{
				Name:   ScaleAction,
				Params: []byte(`{"replicas": 5}`),
			},
			want: &resource.Resource{
				URN:     "urn:odpf:entropy:firehose:test",
				Kind:    "firehose",
				Name:    "test",
				Project: "demo",
				Spec: resource.Spec{
					Configs: []byte(`{"state":"","release_configs":{"name":"demo-test-firehose","repository":"https://odpf.github.io/charts/","chart":"firehose","version":"0.1.1","values":{"firehose":{},"replicaCount":5},"namespace":"firehose","timeout":0,"force_update":true,"recreate_pods":false,"wait":false,"wait_for_jobs":false,"replace":false,"description":"","create_namespace":false}}`),
				},
				State: resource.State{
					Status:     resource.StatusPending,
					ModuleData: []byte(`{"pending_steps":["helm_update"]}`),
				},
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			m := firehoseModule{}

			got, err := m.Plan(context.Background(), tt.spec, tt.act)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got, cmp.Diff(tt.want, got))
			}
		})
	}
}
