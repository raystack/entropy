package volume

import (
	v1 "k8s.io/api/core/v1"
)

const (
	Secret Source = iota
	ConfigMap
)

type Source int

type Volume struct {
	Kind       Source
	Name       string
	SourceName string
}

func (v Volume) GetPodVolume() v1.Volume {
	var vSource v1.VolumeSource
	switch v.Kind {
	case Secret:
		vSource.Secret = &v1.SecretVolumeSource{
			SecretName: v.SourceName,
		}
	case ConfigMap:
		vSource.ConfigMap = &v1.ConfigMapVolumeSource{
			LocalObjectReference: v1.LocalObjectReference{Name: v.SourceName},
		}
	}
	return v1.Volume{
		Name:         v.Name,
		VolumeSource: vSource,
	}
}
