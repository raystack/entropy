package container

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestContainer_Template(t *testing.T) {
	quantity, _ := resource.ParseQuantity("100")
	type fields struct {
		Image           string
		Name            string
		EnvConfigMaps   []string
		Command         []string
		EnvMap          map[string]string
		Args            []string
		ImagePullPolicy string
		VolumeMounts    []VolumeMount
		Requests        map[string]string
		Limits          map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   v1.Container
	}{
		{name: "Container Create", fields: fields{
			Image:           "image:v1.0",
			Name:            "container-name",
			EnvConfigMaps:   []string{"cm1"},
			Command:         []string{"cmd1", "cmd2"},
			EnvMap:          map[string]string{"a": "b"},
			Args:            nil,
			ImagePullPolicy: "Never",
			VolumeMounts: []VolumeMount{{
				Name:      "v1",
				MountPath: "/tmp/v1",
			}},
			Requests: map[string]string{"cpu": "100", "memory": "100"},
			Limits:   map[string]string{"cpu": "100", "memory": "100"},
		}, want: v1.Container{
			Name:    "container-name",
			Image:   "image:v1.0",
			Command: []string{"cmd1", "cmd2"},
			EnvFrom: []v1.EnvFromSource{{
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{Name: "cm1"},
				},
			}},
			Env: []v1.EnvVar{{Name: "a", Value: "b"}},
			Resources: v1.ResourceRequirements{
				Limits: map[v1.ResourceName]resource.Quantity{
					v1.ResourceCPU:    quantity,
					v1.ResourceMemory: quantity,
				},
				Requests: map[v1.ResourceName]resource.Quantity{
					v1.ResourceCPU:    quantity,
					v1.ResourceMemory: quantity,
				},
			},
			Lifecycle: &v1.Lifecycle{},
			VolumeMounts: []v1.VolumeMount{
				{Name: "v1", MountPath: "/tmp/v1"},
				{Name: "shared-data", MountPath: "/shared"},
			},
			ImagePullPolicy: "Never",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Container{
				Image:           tt.fields.Image,
				Name:            tt.fields.Name,
				EnvConfigMaps:   tt.fields.EnvConfigMaps,
				Command:         tt.fields.Command,
				EnvMap:          tt.fields.EnvMap,
				Args:            tt.fields.Args,
				ImagePullPolicy: tt.fields.ImagePullPolicy,
				VolumeMounts:    tt.fields.VolumeMounts,
				Requests:        tt.fields.Requests,
				Limits:          tt.fields.Limits,
			}
			if got := c.Template(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Template() = %v\n, want %v\n", got, tt.want)
			}
		})
	}
}
