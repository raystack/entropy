package volume

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestVolume_GetPodVolume(t *testing.T) {
	type fields struct {
		Kind       Source
		Name       string
		SourceName string
	}
	tests := []struct {
		name   string
		fields fields
		want   v1.Volume
	}{
		{
			name: "Get Volume", fields: fields{
				Kind: ConfigMap, Name: "volume1", SourceName: "confMap1",
			}, want: v1.Volume{
				Name: "volume1",
				VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{Name: "confMap1"},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Volume{
				Kind:       tt.fields.Kind,
				Name:       tt.fields.Name,
				SourceName: tt.fields.SourceName,
			}
			if got := v.GetPodVolume(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPodVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}
