package helm

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/goto/entropy/pkg/errors"
)

type Client struct {
	config      *Config
	cliSettings *cli.EnvSettings
}

func (p *Client) Upsert(config *ReleaseConfig, canUpdateCheck func(rel *release.Release) bool) (*Result, error) {
	actionConfig, err := p.getActionConfiguration(config.Namespace)
	if err != nil {
		return nil, errors.ErrInternal.WithMsgf("error while getting action configuration  : %s", err)
	}

	rel, err := fetchRelease(actionConfig, config.Name)
	if err != nil && !errors.Is(err, errors.ErrNotFound) {
		return nil, errors.ErrInternal.WithMsgf("failed to find release").WithCausef(err.Error())
	}

	isCreate := rel == nil                        // release doesn't exist already.
	canUpdate := !isCreate && canUpdateCheck(rel) // exists already and we are allowed to update.
	if !isCreate && !canUpdate {
		return nil, errors.ErrConflict.WithMsgf("release with same name exists, but update not possible")
	}

	if isCreate {
		// release does not exist.
		return p.doCreate(actionConfig, config)
	}

	// already exists and is updatable.
	return p.doUpdate(actionConfig, config)
}

func (p *Client) Delete(config *ReleaseConfig) error {
	actionConfig, err := p.getActionConfiguration(config.Namespace)
	if err != nil {
		return errors.ErrInternal.WithMsgf("error while getting action configuration  : %s", err)
	}

	act := action.NewUninstall(actionConfig)
	if _, err := act.Run(config.Name); err != nil {
		return errors.ErrInternal.WithMsgf("unable to uninstall release %s", err)
	}
	return nil
}

func (p *Client) doCreate(actionConfig *action.Configuration, config *ReleaseConfig) (*Result, error) {
	fetchedChart, chartPathOpts, err := p.getChart(config)
	if err != nil {
		return nil, errors.ErrInternal.WithMsgf("error while getting chart").WithCausef(err.Error())
	}

	act := action.NewInstall(actionConfig)
	act.Wait = config.Wait
	act.DryRun = false
	act.Timeout = time.Duration(config.Timeout) * time.Second
	act.Replace = config.Replace
	act.OutputDir = ""
	act.Namespace = config.Namespace
	act.ClientOnly = false
	act.Description = config.Description
	act.WaitForJobs = config.WaitForJobs
	act.ReleaseName = config.Name
	act.GenerateName = false
	act.NameTemplate = ""
	act.CreateNamespace = config.CreateNamespace
	act.ChartPathOptions = *chartPathOpts

	rel, err := act.Run(fetchedChart, config.Values)
	if err != nil {
		return nil, errors.ErrInternal.WithMsgf("create-release failed").WithCausef(err.Error())
	}

	return &Result{
		Config:  config,
		Release: rel,
	}, nil
}

func (p *Client) doUpdate(actionConfig *action.Configuration, config *ReleaseConfig) (*Result, error) {
	fetchedChart, chartPathOpts, err := p.getChart(config)
	if err != nil {
		return nil, errors.ErrInternal.WithMsgf("error while getting chart").WithCausef(err.Error())
	}

	act := action.NewUpgrade(actionConfig)
	act.ChartPathOptions = *chartPathOpts
	act.DryRun = false
	act.Wait = config.Wait
	act.WaitForJobs = config.WaitForJobs
	act.Timeout = time.Duration(config.Timeout) * time.Second
	act.Namespace = config.Namespace
	act.Description = config.Description

	rel, err := act.Run(config.Name, fetchedChart, config.Values)
	if err != nil {
		if isReleaseNotFoundErr(err) {
			return nil, errors.ErrNotFound.
				WithMsgf("update-release failed").
				WithCausef("release with given name not found")
		}
		return nil, errors.ErrInternal.WithMsgf("update-release failed").WithCausef(err.Error())
	}

	return &Result{
		Config:  config,
		Release: rel,
	}, nil
}

func (p *Client) getChart(config *ReleaseConfig) (*chart.Chart, *action.ChartPathOptions, error) {
	repositoryURL, chartName := resolveChartName(config.Repository, strings.TrimSpace(config.Chart))

	chartPathOpts := &action.ChartPathOptions{
		RepoURL: repositoryURL,
		Version: getVersion(config.Version),
	}

	// TODO: Add a lock as Load function blows up if accessed concurrently
	path, err := chartPathOpts.LocateChart(chartName, p.cliSettings)
	if err != nil {
		return nil, nil, err
	}

	fetchedChart, err := loader.Load(path)
	if err != nil {
		return nil, nil, err
	}

	// TODO: check if chart has dependencies and load those dependencies
	if fetchedChart.Metadata.Type != typeApplication {
		return nil, nil, ErrChartNotApplication
	}

	return fetchedChart, chartPathOpts, nil
}

func (p *Client) getActionConfiguration(namespace string) (*action.Configuration, error) {
	hasCA := len(p.config.Kubernetes.ClusterCACertificate) != 0
	hasCert := len(p.config.Kubernetes.ClientCertificate) != 0
	defaultTLS := hasCA || hasCert || p.config.Kubernetes.Insecure
	host, _, err := rest.DefaultServerURL(p.config.Kubernetes.Host, "", apimachineryschema.GroupVersion{}, defaultTLS)
	if err != nil {
		return nil, err
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*api.NewConfig(), &clientcmd.ConfigOverrides{
		AuthInfo: api.AuthInfo{
			Token:                 p.config.Kubernetes.Token,
			ClientKeyData:         []byte(p.config.Kubernetes.ClientKey),
			ClientCertificateData: []byte(p.config.Kubernetes.ClientCertificate),
		},
		ClusterInfo: api.Cluster{
			Server:                   host.String(),
			InsecureSkipTLSVerify:    p.config.Kubernetes.Insecure,
			CertificateAuthorityData: []byte(p.config.Kubernetes.ClusterCACertificate),
		},
	})
	kubeConf := &kubeClientGetter{ClientConfig: clientConfig}

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(kubeConf, namespace, p.config.HelmDriver, noOpLog); err != nil {
		return nil, err
	}
	return actionConfig, nil
}

func NewClient(config *Config) *Client {
	return &Client{config: config, cliSettings: cli.New()}
}

func fetchRelease(cfg *action.Configuration, name string) (*release.Release, error) {
	get := action.NewGet(cfg)
	res, err := get.Run(name)
	if err != nil {
		if isReleaseNotFoundErr(err) {
			return nil, errors.ErrNotFound.WithCausef(err.Error())
		}
		return nil, err
	}
	return res, nil
}

func getVersion(version string) string {
	if version == "" {
		return ">0.0.0-0"
	}
	return strings.TrimSpace(version)
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

func isReleaseNotFoundErr(err error) bool {
	return strings.Contains(err.Error(), "release: not found")
}

func noOpLog(_ string, _ ...interface{}) {}
