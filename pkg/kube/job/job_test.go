package job

import (
	"reflect"
	"testing"

	v12 "k8s.io/api/batch/v1"
	v13 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/goto/entropy/pkg/kube/pod"
)

func TestJob_Template(t *testing.T) {
	var constant int32 = 1
	type fields struct {
		Pod         *pod.Pod
		Name        string
		Namespace   string
		Labels      map[string]string
		Parallelism *int32
		BackOffList *int32
	}
	tests := []struct {
		name   string
		fields fields
		want   *v12.Job
	}{
		{name: "Job template", fields: fields{
			Pod: &pod.Pod{
				Name: "pod-name",
			},
			Name:        "job-name",
			Namespace:   "default",
			Labels:      nil,
			Parallelism: &constant,
			BackOffList: &constant,
		}, want: &v12.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      "job-name",
				Namespace: "default",
			},
			Spec: v12.JobSpec{
				BackoffLimit: &constant,
				Parallelism:  &constant,
				Template: v13.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{Name: "pod-name"},
					Spec: v13.PodSpec{
						Containers: nil,
						Volumes: []v13.Volume{{
							Name:         "shared-data",
							VolumeSource: v13.VolumeSource{EmptyDir: &v13.EmptyDirVolumeSource{}},
						}},
						RestartPolicy: v13.RestartPolicyNever,
					},
				},
			},
			Status: v12.JobStatus{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &Job{
				Pod:         tt.fields.Pod,
				Name:        tt.fields.Name,
				Namespace:   tt.fields.Namespace,
				Labels:      tt.fields.Labels,
				Parallelism: tt.fields.Parallelism,
				BackOffList: tt.fields.BackOffList,
			}
			if got := j.Template(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Template() = %v, want %v", got, tt.want)
			}
		})
	}
}
