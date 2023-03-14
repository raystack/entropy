package helm

import (
	"github.com/mcuadros/go-defaults"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/goto/entropy/pkg/kube"
)

type Config struct {
	// HelmDriver - The backend storage driver. Values are - configmap, secret, memory, sql
	HelmDriver string `default:"secret"`
	// Kubernetes configuration.
	Kubernetes kube.Config
}

type Client struct {
	config      *Config
	cliSettings *cli.EnvSettings
}

func DefaultClientConfig() *Config {
	defaultProviderConfig := new(Config)
	defaults.SetDefaults(defaultProviderConfig)
	return defaultProviderConfig
}

func NewClient(config *Config) *Client {
	return &Client{config: config, cliSettings: cli.New()}
}

func (p *Client) getActionConfiguration(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	overrides := &clientcmd.ConfigOverrides{}

	overrides.AuthInfo.ClientCertificateData = []byte(p.config.Kubernetes.ClientCertificate)
	overrides.AuthInfo.ClientKeyData = []byte(p.config.Kubernetes.ClientKey)
	overrides.AuthInfo.Token = p.config.Kubernetes.Token
	overrides.ClusterInfo.CertificateAuthorityData = []byte(p.config.Kubernetes.ClusterCACertificate)
	overrides.ClusterInfo.InsecureSkipTLSVerify = p.config.Kubernetes.Insecure

	hasCA := len(overrides.ClusterInfo.CertificateAuthorityData) != 0
	hasCert := len(overrides.AuthInfo.ClientCertificateData) != 0
	defaultTLS := hasCA || hasCert || overrides.ClusterInfo.InsecureSkipTLSVerify
	host, _, err := rest.DefaultServerURL(p.config.Kubernetes.Host, "", apimachineryschema.GroupVersion{}, defaultTLS)
	if err != nil {
		return nil, err
	}
	overrides.ClusterInfo.Server = host.String()

	clientConfig := clientcmd.NewDefaultClientConfig(*api.NewConfig(), overrides)

	if err := actionConfig.Init(&KubeConfig{ClientConfig: clientConfig}, namespace, p.config.HelmDriver, func(format string, v ...interface{}) {}); err != nil {
		return nil, err
	}
	return actionConfig, nil
}
