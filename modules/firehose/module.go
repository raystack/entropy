package firehose

import (
	_ "embed"
	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/modules/kubernetes"
)

const (
	StopAction  = "stop"
	StartAction = "start"
	ScaleAction = "scale"
	ResetAction = "reset"
)

const (
	releaseCreate = "release_create"
	releaseUpdate = "release_update"

	consumerReset = "consumer_reset"
)

const (
	stateRunning = "RUNNING"
	stateStopped = "STOPPED"
)

const (
	keyReplicaCount   = "replicaCount"
	keyKubeDependency = "kube_cluster"
)

var Module = module.Descriptor{
	Kind: "firehose",
	Dependencies: map[string]string{
		keyKubeDependency: kubernetes.Module.Kind,
	},
	Actions: []module.ActionDesc{
		{
			Name:        module.CreateAction,
			Description: "Creates firehose instance.",
			ParamSchema: completeConfigSchema,
		},
		{
			Name:        module.UpdateAction,
			Description: "Updates an existing firehose instance.",
			ParamSchema: completeConfigSchema,
		},
		{
			Name:        ScaleAction,
			Description: "Scale-up or scale-down an existing firehose instance.",
			ParamSchema: scaleActionSchema,
		},
		{
			Name:        StopAction,
			Description: "Stop firehose and all its components.",
		},
		{
			Name:        StartAction,
			Description: "Start firehose and all its components.",
		},
		{
			Name:        ResetAction,
			Description: "Reset firehose kafka consumer group to given timestamp",
			ParamSchema: resetActionSchema,
		},
	},
	Module: &firehoseModule{},
}

type firehoseModule struct{}
