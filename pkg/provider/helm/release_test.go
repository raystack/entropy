//go:build integration
// +build integration

package helm

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var jsonValues = `
{
	"firehose": {
		"image": {
			"tag": "0.1.1"
		}
	}
}
`

var updatedJsonValues = `
{
	"firehose": {
		"image": {
			"tag": "0.1.2"
		}
	}
}
`

var envKubeAPIServer = os.Getenv("TEST_K8S_API_SERVER")
var envKubeSAToken = os.Getenv("TEST_K8S_SA_TOKEN")

func jsonMarshal(data interface{}) string {
	j, _ := json.MarshalIndent(data, "", "  ")
	return string(j)
}

func jsonUnmarshal(data string) map[string]interface{} {
	ret := map[string]interface{}{}
	_ = json.Unmarshal([]byte(data), &ret)
	return ret
}

func TestReleaseCreate(t *testing.T) {
	providerConfig := DefaultProviderConfig()
	providerConfig.Kubernetes.Host = envKubeAPIServer
	providerConfig.Kubernetes.Insecure = true
	providerConfig.Kubernetes.Token = envKubeSAToken
	provider := NewProvider(providerConfig)

	releaseName := fmt.Sprintf("test-entropy-helm-provider-create-%d", rand.Int())

	releaseConfig := DefaultReleaseConfig()
	releaseConfig.Name = releaseName
	releaseConfig.Repository = "https://odpf.github.io/charts/"
	releaseConfig.Chart = "firehose"
	releaseConfig.Version = "0.1.1"
	releaseConfig.Values = jsonUnmarshal(jsonValues)
	releaseConfig.WaitForJobs = true
	rel, err := provider.Release(releaseConfig)

	assert.Nil(t, err)
	assert.NotNil(t, rel)
}

func TestReleaseUpdate(t *testing.T) {
	providerConfig := DefaultProviderConfig()
	providerConfig.Kubernetes.Host = envKubeAPIServer
	providerConfig.Kubernetes.Insecure = true
	providerConfig.Kubernetes.Token = envKubeSAToken
	provider := NewProvider(providerConfig)

	releaseName := fmt.Sprintf("test-entropy-helm-provider-update-%d", rand.Int())

	releaseConfig := DefaultReleaseConfig()
	releaseConfig.Name = releaseName
	releaseConfig.Repository = "https://odpf.github.io/charts/"
	releaseConfig.Chart = "firehose"
	releaseConfig.Version = "0.1.1"
	releaseConfig.Values = jsonUnmarshal(jsonValues)
	releaseConfig.WaitForJobs = true
	rel, err := provider.Release(releaseConfig)

	assert.Nil(t, err)
	assert.NotNil(t, rel)

	releaseConfig.Values = jsonUnmarshal(updatedJsonValues)
	rel2, err := provider.Release(releaseConfig)

	assert.Nil(t, err)
	assert.NotNil(t, rel2)
}
