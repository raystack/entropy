package helm

import (
	"github.com/mcuadros/go-defaults"
	"helm.sh/helm/v3/pkg/release"

	"github.com/goto/entropy/pkg/errors"
)

var (
	typeApplication        = "application"
	ErrChartNotApplication = errors.New("helm chart is not an application chart")
)

type ReleaseConfig struct {
	// Name - Result Name
	Name string `json:"name" mapstructure:"name"`
	// Repository - Repository where to locate the requested chart. If is a URL the chart is installed without installing the repository.
	Repository string `json:"repository" mapstructure:"repository"`
	// Chart - Chart name to be installed. A path may be used.
	Chart string `json:"chart" mapstructure:"chart"`
	// Version - Specify the exact chart version to install. If this is not specified, the latest version is installed.
	Version string `json:"version" mapstructure:"version"`
	// Values - Map of values in to pass to helm.
	Values map[string]interface{} `json:"values" mapstructure:"values"`
	// Namespace - Namespace to install the release into.
	Namespace string `json:"namespace" mapstructure:"namespace" default:"default"`
	// Timeout - Time in seconds to wait for any individual kubernetes operation.
	Timeout int `json:"timeout" mapstructure:"timeout" default:"300"`
	// ForceUpdate - Force resource update through delete/recreate if needed.
	ForceUpdate bool `json:"force_update" mapstructure:"force_update" default:"false"`
	// RecreatePods - Perform pods restart during upgrade/rollback
	RecreatePods bool `json:"recreate_pods" mapstructure:"recreate_pods" default:"false"`
	// Wait - Will wait until all resources are in a ready state before marking the release as successful.
	Wait bool `json:"wait" mapstructure:"wait" default:"true"`
	// WaitForJobs - If wait is enabled, will wait until all Jobs have been completed before marking the release as successful.
	WaitForJobs bool `json:"wait_for_jobs" mapstructure:"wait_for_jobs" default:"false"`
	// Replace - Re-use the given name, even if that name is already used. This is unsafe in production
	Replace bool `json:"replace" mapstructure:"replace" default:"false"`
	// Description - Add a custom description
	Description string `json:"description" mapstructure:"description"`
	// CreateNamespace - Create the namespace if it does not exist
	CreateNamespace bool `json:"create_namespace" mapstructure:"create_namespace" default:"false"`
}

type Result struct {
	Config  *ReleaseConfig
	Release *release.Release
}

func DefaultReleaseConfig() *ReleaseConfig {
	defaultReleaseConfig := &ReleaseConfig{}
	defaults.SetDefaults(defaultReleaseConfig)
	return defaultReleaseConfig
}
