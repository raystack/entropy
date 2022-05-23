package kubernetes

import (
	_ "embed"
	"time"

	"k8s.io/client-go/rest"
)

type moduleConf struct {
	Host          string        `json:"host"`
	CAData        string        `json:"ca_data"`
	KeyData       string        `json:"key_data"`
	CertData      string        `json:"cert_data"`
	ClientTimeout time.Duration `json:"client_timeout"`
}

func (mc moduleConf) toRESTConfig() *rest.Config {
	return &rest.Config{
		Host: mc.Host,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   []byte(mc.CAData),
			KeyData:  []byte(mc.KeyData),
			CertData: []byte(mc.CertData),
		},
		Timeout: mc.ClientTimeout,
	}
}
