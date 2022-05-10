//go:build integration
// +build integration

package kubelogger

import (
	"context"
	"os"
	"testing"

	"github.com/odpf/entropy/pkg/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

var envClusterCACert = os.Getenv("TEST_K8S_CLUSTER_CA_CERT")
var envClientKey = os.Getenv("TEST_K8S_CLIENT_KEY")
var envClientCert = os.Getenv("TEST_K8S_CLIENT_CERT")
var envHost = os.Getenv("TEST_K8S_HOST")
var envNamespace = os.Getenv("TEST_K8S_NAMESPACE")
var envPod = os.Getenv("TEST_K8S_POD")
var envContainer = os.Getenv("TEST_K8S_CONTAINER")

func TestGetStreamingLogs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name          string
		ClusterCACert string
		ClientKey     string
		ClientCert    string
		Host          string
		Namespace     string
		Pod           string
		Container     string
		wantErr       error
	}{
		{
			Name:          "InvalidCredentials",
			ClusterCACert: "invalid cluster certificate",
			ClientKey:     "invalid client key",
			ClientCert:    "invalid client certificate",
			Host:          "invalid host",
			Namespace:     "invalid namespace",
			Pod:           "invalid pod",
			Container:     "invalid container",
			wantErr:       errors.New("invalid credentials"),
		},
		{
			Name:          "StreamFromOnePod",
			ClusterCACert: envClusterCACert,
			ClientKey:     envClientKey,
			ClientCert:    envClientCert,
			Host:          envHost,
			Namespace:     envNamespace,
			Pod:           envPod,
			Container:     envContainer,
			wantErr:       nil,
		},
		{
			Name:          "StreamFromAllPods",
			ClusterCACert: envClusterCACert,
			ClientKey:     envClientKey,
			ClientCert:    envClientCert,
			Host:          envHost,
			Namespace:     envNamespace,
			Pod:           "",
			Container:     "",
			wantErr:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			filter := make(map[string]string)
			filter["pod"] = tt.Pod
			filter["container"] = tt.Container

			cfg := rest.Config{
				Host: tt.Host,
				TLSClientConfig: rest.TLSClientConfig{
					CAData:   []byte(tt.ClusterCACert),
					KeyData:  []byte(tt.ClientKey),
					CertData: []byte(tt.ClientCert),
				},
			}

			ctx := new(context.Context)
			_, err := GetStreamingLogs(*ctx, tt.Namespace, filter, cfg)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
