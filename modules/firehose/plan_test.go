package firehose

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func TestFirehoseModule_Plan(t *testing.T) {
	t.Parallel()

	res := resource.Resource{
		URN:     "orn:entropy:firehose:test",
		Kind:    "firehose",
		Name:    "test",
		Project: "demo",
		Spec: resource.Spec{
			Configs: []byte(`{"state":"RUNNING","chart_version":"0.1.1","firehose":{"replicas":1,"kafka_broker_address":"localhost:9092","kafka_topic":"test-topic","kafka_consumer_id":"test-consumer-id","env_variables":{}}}`),
		},
		State: resource.State{},
	}

	table := []struct {
		title   string
		spec    module.ExpandedResource
		act     module.ActionRequest
		want    *module.Plan
		wantErr error
	}{
		{
			title: "InvalidConfiguration",
			spec:  module.ExpandedResource{Resource: res},
			act: module.ActionRequest{
				Name:   module.CreateAction,
				Params: []byte(`{`),
			},
			wantErr: errors.ErrInvalid,
		},
		{
			title: "ValidConfiguration",
			spec:  module.ExpandedResource{Resource: res},
			act: module.ActionRequest{
				Name:   module.CreateAction,
				Params: []byte(`{"state":"RUNNING","firehose":{"replicas":1,"kafka_broker_address":"localhost:9092","kafka_topic":"test-topic","kafka_consumer_id":"test-consumer-id","env_variables":{}}}`),
			},
			want: &module.Plan{
				Resource: resource.Resource{
					URN:     "orn:entropy:firehose:test",
					Kind:    "firehose",
					Name:    "test",
					Project: "demo",
					Spec: resource.Spec{
						Configs: []byte(`{"state":"RUNNING","chart_version":"0.1.1","stop_time":null,"telegraf":null,"firehose":{"replicas":1,"kafka_broker_address":"localhost:9092","kafka_topic":"test-topic","kafka_consumer_id":"test-consumer-id","env_variables":{}}}`),
					},
					State: resource.State{
						Status:     resource.StatusPending,
						ModuleData: []byte(`{"pending_steps":["release_create"]}`),
					},
				},
			},
		},
		{
			title: "InvalidActionParams",
			spec:  module.ExpandedResource{Resource: res},
			act: module.ActionRequest{
				Name:   ScaleAction,
				Params: []byte(`{`),
			},
			wantErr: errors.ErrInvalid,
		},
		{
			title: "ValidScaleRequest",
			spec:  module.ExpandedResource{Resource: res},
			act: module.ActionRequest{
				Name:   ScaleAction,
				Params: []byte(`{"replicas": 5}`),
			},
			want: &module.Plan{
				Resource: resource.Resource{
					URN:     "orn:entropy:firehose:test",
					Kind:    "firehose",
					Name:    "test",
					Project: "demo",
					Spec: resource.Spec{
						Configs: []byte(`{"state":"RUNNING","chart_version":"0.1.1","stop_time":null,"telegraf":null,"firehose":{"replicas":5,"kafka_broker_address":"localhost:9092","kafka_topic":"test-topic","kafka_consumer_id":"test-consumer-id","env_variables":{}}}`),
					},
					State: resource.State{
						Status:     resource.StatusPending,
						ModuleData: []byte(`{"pending_steps":["release_update"]}`),
					},
				},
			},
		},
		{
			title: "ValidResetRequest",
			spec:  module.ExpandedResource{Resource: res},
			act: module.ActionRequest{
				Name:   ResetAction,
				Params: []byte(`{"to":"DATETIME","datetime":"2022-06-22T00:00:00+00:00"}`),
			},
			want: &module.Plan{
				Resource: resource.Resource{
					URN:     "orn:entropy:firehose:test",
					Kind:    "firehose",
					Name:    "test",
					Project: "demo",
					Spec: resource.Spec{
						Configs: []byte(`{"state":"RUNNING","chart_version":"0.1.1","stop_time":null,"telegraf":null,"firehose":{"replicas":1,"kafka_broker_address":"localhost:9092","kafka_topic":"test-topic","kafka_consumer_id":"test-consumer-id","env_variables":{}}}`),
					},
					State: resource.State{
						Status:     resource.StatusPending,
						ModuleData: []byte(`{"pending_steps":["release_update","consumer_reset","release_update"],"reset_to":"2022-06-22T00:00:00+00:00","state_override":"STOPPED"}`),
					},
				},
			},
		},
		{
			title: "WithStopTimeConfiguration",
			spec:  module.Spec{Resource: res},
			act: module.ActionRequest{
				Name:   module.CreateAction,
				Params: []byte(`{"state":"RUNNING","stop_time":"3022-07-13T00:40:14.028016Z","firehose":{"replicas":1,"kafka_broker_address":"localhost:9092","kafka_topic":"test-topic","kafka_consumer_id":"test-consumer-id","env_variables":{}}}`),
			},
			want: &module.Plan{
				Resource: resource.Resource{
					URN:     "orn:entropy:firehose:test",
					Kind:    "firehose",
					Name:    "test",
					Project: "demo",
					Spec: resource.Spec{
						Configs: []byte(`{"state":"RUNNING","chart_version":"0.1.1","stop_time":"3022-07-13T00:40:14.028016Z","telegraf":null,"firehose":{"replicas":1,"kafka_broker_address":"localhost:9092","kafka_topic":"test-topic","kafka_consumer_id":"test-consumer-id","env_variables":{}}}`),
					},
					State: resource.State{
						Status:     resource.StatusPending,
						ModuleData: []byte(`{"pending_steps":["release_create"]}`),
					},
				},
				ScheduleRunAt: parseTime("3022-07-13T00:40:14.028016Z"),
			},
		},
	}

	for _, tt := range table {
		tt := tt
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()
			m := firehoseModule{}

			got, err := m.Plan(context.Background(), tt.spec, tt.act)
			if tt.wantErr != nil || err != nil {
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

func parseTime(timeString string) time.Time {
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		panic(err)
	}
	return t
}
