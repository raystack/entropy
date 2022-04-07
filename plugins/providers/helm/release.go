package helm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mcuadros/go-defaults"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
)

var ErrReleaseNotFound = errors.New("release not found")
var ErrChartNotApplication = errors.New("helm chart is not an application chart")

type ReleaseConfig struct {
	// Name - Release Name
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
	CreateNamespace bool   `json:"create_namespace" mapstructure:"create_namespace" default:"false"`
	State           string `json:"state" mapstructure:"state"`
}

func DefaultReleaseConfig() *ReleaseConfig {
	defaultReleaseConfig := new(ReleaseConfig)
	defaults.SetDefaults(defaultReleaseConfig)
	return defaultReleaseConfig
}

type Release struct {
	Config *ReleaseConfig
	Output ReleaseOutput
}

type ReleaseOutput struct {
	// Status - Status of the release.
	Status Status
	// Revision - Revision of the release.
	Release string
}

// Release - creates or updates a helm release with its configs
func (p *Provider) Release(config *ReleaseConfig) (*Release, error) {
	releaseExists, _ := p.resourceReleaseExists(config.Name, config.Namespace)
	if releaseExists {
		return p.update(config)
	}
	return p.create(config)
}

func (p *Provider) create(config *ReleaseConfig) (*Release, error) {
	actionConfig, err := p.getActionConfiguration(config.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error while getting action configuration  : %w", err)
	}

	chartPathOptions, chartName := p.chartPathOptions(config)

	chart, _, err := p.getChart(config, chartName, chartPathOptions)
	if err != nil {
		return nil, fmt.Errorf("error while getting chart : %w", err)
	}

	// TODO: check if chart has dependencies and load those dependencies

	if chart.Metadata.Type != "application" {
		return nil, ErrChartNotApplication
	}

	client := action.NewInstall(actionConfig)
	client.ChartPathOptions = *chartPathOptions
	client.ClientOnly = false
	client.DryRun = false
	client.Wait = config.Wait
	client.WaitForJobs = config.WaitForJobs
	client.Timeout = time.Second * time.Duration(config.Timeout)
	client.Namespace = config.Namespace
	client.ReleaseName = config.Name
	client.GenerateName = false
	client.NameTemplate = ""
	client.OutputDir = ""
	client.Replace = config.Replace
	client.Description = config.Description
	client.CreateNamespace = config.CreateNamespace

	rel, err := client.Run(chart, config.Values)
	if err != nil && rel == nil {
		return nil, fmt.Errorf("error while installing release: %w", err)
	}

	if err != nil && rel != nil {
		releaseExists, releaseErr := p.resourceReleaseExists(config.Name, config.Namespace)

		if releaseErr != nil {
			return nil, fmt.Errorf("release exists: %w", releaseErr)
		}

		if !releaseExists {
			return nil, fmt.Errorf("release doesn't exists: %w", err)
		}

		releaseJson, err := json.Marshal(rel)
		if err != nil {
			return nil, fmt.Errorf("error while json marshalling release: %w", err)
		}

		return &Release{
			Config: config,
			Output: ReleaseOutput{
				Status:  mapReleaseStatus(rel.Info.Status),
				Release: string(releaseJson),
			},
		}, fmt.Errorf("helm release created with failure: %w", err)
	}

	releaseJson, err := json.Marshal(rel)
	if err != nil {
		return nil, err
	}

	return &Release{
		Config: config,
		Output: ReleaseOutput{
			Status:  mapReleaseStatus(rel.Info.Status),
			Release: string(releaseJson),
		},
	}, nil
}

func (p *Provider) update(config *ReleaseConfig) (*Release, error) {
	var rel *release.Release
	var err error

	actionConfig, err := p.getActionConfiguration(config.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error while getting action configuration  : %w", err)
	}

	chartPathOptions, chartName := p.chartPathOptions(config)

	chart, _, err := p.getChart(config, chartName, chartPathOptions)
	if err != nil {
		return nil, fmt.Errorf("error while getting chart : %w", err)
	}

	// TODO: check if chart has dependencies and load those dependencies

	if chart.Metadata.Type != "application" {
		return nil, ErrChartNotApplication
	}

	client := action.NewUpgrade(actionConfig)
	client.ChartPathOptions = *chartPathOptions
	client.DryRun = false
	client.Wait = config.Wait
	client.WaitForJobs = config.WaitForJobs
	client.Timeout = time.Second * time.Duration(config.Timeout)
	client.Namespace = config.Namespace
	client.Description = config.Description

	rel, err = client.Run(config.Name, chart, config.Values)
	if err != nil && rel == nil {
		return nil, fmt.Errorf("error while updating release: %w", err)
	}

	if err != nil && rel != nil {
		releaseExists, releaseErr := p.resourceReleaseExists(config.Name, config.Namespace)

		if releaseErr != nil {
			return nil, fmt.Errorf("release exists: %w", releaseErr)
		}

		if !releaseExists {
			return nil, fmt.Errorf("release doesn't exists: %w", err)
		}

		releaseJson, err := json.Marshal(rel)
		if err != nil {
			return nil, fmt.Errorf("error while json marshalling release: %w", err)
		}

		return &Release{
			Config: config,
			Output: ReleaseOutput{
				Status:  mapReleaseStatus(rel.Info.Status),
				Release: string(releaseJson),
			},
		}, fmt.Errorf("helm release created with failure: %w", err)
	}

	releaseJson, err := json.Marshal(rel)
	if err != nil {
		return nil, err
	}

	return &Release{
		Config: config,
		Output: ReleaseOutput{
			Status:  mapReleaseStatus(rel.Info.Status),
			Release: string(releaseJson),
		},
	}, nil
}

func (p *Provider) chartPathOptions(config *ReleaseConfig) (*action.ChartPathOptions, string) {
	repositoryURL, chartName := resolveChartName(config.Repository, strings.TrimSpace(config.Chart))

	version := getVersion(config.Version)

	return &action.ChartPathOptions{
		RepoURL: repositoryURL,
		Version: version,
	}, chartName
}

func resolveChartName(repository, name string) (string, string) {
	_, err := url.ParseRequestURI(repository)
	if err == nil {
		return repository, name
	}

	if !strings.Contains(name, "/") && repository != "" {
		name = fmt.Sprintf("%s/%s", repository, name)
	}

	return "", name
}

func getVersion(version string) string {
	if version == "" {
		return ">0.0.0-0"
	}
	return strings.TrimSpace(version)
}

func (p *Provider) getChart(config *ReleaseConfig, name string, cpo *action.ChartPathOptions) (*chart.Chart, string, error) {
	// TODO: Add a lock as Load function blows up if accessed concurrently

	path, err := cpo.LocateChart(name, p.cliSettings)
	if err != nil {
		return nil, "", err
	}

	c, err := loader.Load(path)
	if err != nil {
		return nil, "", err
	}

	return c, path, nil
}

func (p *Provider) resourceReleaseExists(name string, namespace string) (bool, error) {
	c, err := p.getActionConfiguration(namespace)
	if err != nil {
		return false, err
	}

	_, err = p.getRelease(c, name)

	if err == nil {
		return true, nil
	}

	if err == ErrReleaseNotFound {
		return false, nil
	}

	return false, err
}

func (p *Provider) getRelease(cfg *action.Configuration, name string) (*release.Release, error) {
	//TODO: Add provider level lock to make sure no other operation is changing this release

	get := action.NewGet(cfg)
	res, err := get.Run(name)
	if err != nil {
		if strings.Contains(err.Error(), "release: not found") {
			return nil, ErrReleaseNotFound
		}
		return nil, err
	}
	return res, nil
}

func mapReleaseStatus(status release.Status) Status {
	switch status {
	case "unknown":
		return StatusUnknown
	case "deployed":
		return StatusSuccess
	default:
		return StatusFailed
	}
}
