package container

import (
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Container struct {
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
	PreStopCmd      []string
	PostStartCmd    []string
}

type VolumeMount struct {
	Name      string
	MountPath string
}

func (c Container) Template() corev1.Container {
	var env []corev1.EnvVar
	for k, v := range c.EnvMap {
		env = append(env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	var envFrom []corev1.EnvFromSource
	for _, configMap := range c.EnvConfigMaps {
		envFrom = append(envFrom, corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: configMap},
			},
		})
	}
	var mounts []corev1.VolumeMount
	for _, v := range c.VolumeMounts {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      v.Name,
			MountPath: v.MountPath,
		})
	}

	// Shared directory for all the containers
	mounts = append(mounts, corev1.VolumeMount{
		Name:      "shared-data",
		MountPath: "/shared",
	})

	var lifecycle corev1.Lifecycle
	if len(c.PreStopCmd) > 0 {
		lifecycle.PreStop = &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{Command: c.PreStopCmd},
		}
	}
	if len(c.PostStartCmd) > 0 {
		lifecycle.PostStart = &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{Command: c.PostStartCmd},
		}
	}

	return corev1.Container{
		Name:            c.Name,
		Image:           c.Image,
		Command:         c.Command,
		Args:            c.Args,
		EnvFrom:         envFrom,
		Env:             env,
		Resources:       c.parseResources(),
		VolumeMounts:    mounts,
		Lifecycle:       &lifecycle,
		ImagePullPolicy: corev1.PullPolicy(c.ImagePullPolicy),
	}
}

func (c Container) parseResources() corev1.ResourceRequirements {
	cpuLimits, err := resource.ParseQuantity(c.Limits["cpu"])
	if err != nil {
		zap.L().Error(err.Error())
	}
	memLimits, err := resource.ParseQuantity(c.Limits["memory"])
	if err != nil {
		zap.L().Error(err.Error())
	}
	cpuRequests, err := resource.ParseQuantity(c.Requests["cpu"])
	if err != nil {
		zap.L().Error(err.Error())
	}
	memRequests, err := resource.ParseQuantity(c.Requests["memory"])
	if err != nil {
		zap.L().Error(err.Error())
	}
	return corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    cpuLimits,
			corev1.ResourceMemory: memLimits,
		},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    cpuRequests,
			corev1.ResourceMemory: memRequests,
		},
	}
}
