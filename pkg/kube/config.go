package kube

import (
	"time"

	"k8s.io/client-go/rest"
)

type Config struct {
	// Host - The hostname (in form of URI) of Kubernetes master.
	Host string `json:"host"`

	Timeout time.Duration `json:"timeout" default:"100ms"`

	// Token - Token to authenticate a service account
	Token string `json:"token"`

	// Insecure - Whether server should be accessed without verifying the TLS certificate.
	Insecure bool `json:"insecure" default:"false"`

	// ClientKey - PEM-encoded client key for TLS authentication.
	ClientKey string `json:"client_key"`

	// ClientCertificate - PEM-encoded client certificate for TLS authentication.
	ClientCertificate string `json:"client_certificate"`

	// ClusterCACertificate - PEM-encoded root certificates bundle for TLS authentication.
	ClusterCACertificate string `json:"cluster_ca_certificate"`
}

func (conf Config) RESTConfig() *rest.Config {
	rc := &rest.Config{
		Host:    conf.Host,
		Timeout: conf.Timeout,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   []byte(conf.ClusterCACertificate),
			KeyData:  []byte(conf.ClientKey),
			CertData: []byte(conf.ClientCertificate),
		},
	}

	if conf.Token != "" {
		rc.BearerToken = conf.Token
	}

	return rc
}

func (conf Config) StreamingConfig() *rest.Config {
	rc := conf.RESTConfig()
	rc.Timeout = 0
	return rc
}
