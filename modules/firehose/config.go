package firehose

import (
	_ "embed"
	"encoding/json"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/helm"
)

const (
	defaultNamespace        = "firehose"
	defaultChartString      = "firehose"
	defaultVersionString    = "0.1.1"
	defaultRepositoryString = "https://odpf.github.io/charts/"
)

const (
	stateString         = "state"
	releaseStateRunning = "RUNNING"
	releaseStateStopped = "STOPPED"
)

var (
	//go:embed create_schema.json
	createActionSchema string

	//go:embed scale_schema.json
	scaleActionSchema string
)

type moduleConfig struct {
	Version         string `json:"version"`
	Replicas        int    `json:"replicas"`
	Namespace       string `json:"namespace"`
	CreateNamespace bool   `json:"create_namespace"`
}

func (mc *moduleConfig) sanitiseAndValidate() error {
	return nil
}

func (mc moduleConfig) helmReleaseConfig(r resource.Resource) *helm.ReleaseConfig {
	rc := helm.DefaultReleaseConfig()
	rc.Name = r.URN
	rc.Repository = defaultRepositoryString
	rc.Chart = defaultChartString
	rc.Version = defaultVersionString
	rc.Namespace = defaultNamespace

	if mc.Version != "" {
		rc.Version = mc.Version
	}

	if mc.Namespace != "" {
		rc.Namespace = mc.Namespace
	}
	rc.CreateNamespace = mc.CreateNamespace

	return rc
}

func (mc moduleConfig) JSON() []byte {
	b, err := json.Marshal(mc)
	if err != nil {
		panic(err)
	}
	return b
}
