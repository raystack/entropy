package firehose

import (
	_ "embed"
	"encoding/json"
	"fmt"

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
	ReleaseConfigs helm.ReleaseConfig `json:"release_configs"`
}

func (mc *moduleConfig) sanitiseAndValidate(r resource.Resource) error {
	rc := mc.ReleaseConfigs
	rc.Name = fmt.Sprintf("%s-%s-firehose", r.Project, r.Name)
	rc.Repository = defaultRepositoryString
	rc.Chart = defaultChartString
	rc.Version = defaultVersionString
	rc.Namespace = defaultNamespace
	rc.ForceUpdate = true

	mc.ReleaseConfigs = rc
	return nil
}

func (mc moduleConfig) JSON() []byte {
	b, err := json.Marshal(mc)
	if err != nil {
		panic(err)
	}
	return b
}

func (mc *moduleConfig) merge(overrides moduleConfig) {

}
