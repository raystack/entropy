package helm

import (
	"github.com/mcuadros/go-defaults"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type providerConfig struct {
	// HelmDriver - The backend storage driver. Values are - configmap, secret, memory, sql
	HelmDriver string `default:"secret"`
	// Kubernetes configuration.
	Kubernetes KubernetesConfig
}

func DefaultProviderConfig() *providerConfig {
	defaultProviderConfig := new(providerConfig)
	defaults.SetDefaults(defaultProviderConfig)
	return defaultProviderConfig
}

const ProviderID = "helm"

type KubernetesConfig struct {
	// Host - The hostname (in form of URI) of Kubernetes master.
	Host string
	// Insecure - Whether server should be accessed without verifying the TLS certificate.
	Insecure bool `default:"false"`
	// ClientCertificate - PEM-encoded client certificate for TLS authentication.
	ClientCertificate string
	// ClientKey - PEM-encoded client key for TLS authentication.
	ClientKey string
	// ClusterCACertificate - PEM-encoded root certificates bundle for TLS authentication.
	ClusterCACertificate string
	// Token - Token to authenticate a service account
	Token string
}

type Provider struct {
	config      *providerConfig
	cliSettings *cli.EnvSettings
}

func NewProvider(config *providerConfig) *Provider {
	return &Provider{config: config, cliSettings: cli.New()}
}

func (p *Provider) ID() string {
	return ProviderID
}

func (p *Provider) getActionConfiguration(namespace string) (*action.Configuration, error) {
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
