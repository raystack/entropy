//go:build integration
// +build integration

package helm

import (
	"encoding/base64"
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

func jsonMarshal(data interface{}) string {
	j, _ := json.MarshalIndent(data, "", "  ")
	return string(j)
}

func jsonUnmarshal(data string) map[string]interface{} {
	ret := map[string]interface{}{}
	_ = json.Unmarshal([]byte(data), &ret)
	return ret
}

func getClient() *Client {
	envKubeAPIServer := os.Getenv("TEST_K8S_API_SERVER")
	envKubeSAToken := os.Getenv("TEST_K8S_SA_TOKEN")
	clientConfig := DefaultClientConfig()
	clientConfig.Kubernetes.Host = envKubeAPIServer
	clientConfig.Kubernetes.Insecure = true
	tokenBytes, _ := base64.StdEncoding.DecodeString(envKubeSAToken)
	clientConfig.Kubernetes.Token = string(tokenBytes)
	return NewClient(clientConfig)
}

func TestReleaseCreate(t *testing.T) {
	client := getClient()

	releaseName := fmt.Sprintf("test-entropy-helm-client-create-%d", rand.Int())

	releaseConfig := DefaultReleaseConfig()
	releaseConfig.Name = releaseName
	releaseConfig.Repository = "https://odpf.github.io/charts/"
	releaseConfig.Chart = "firehose"
	releaseConfig.Version = "0.1.1"
	releaseConfig.Values = jsonUnmarshal(jsonValues)
	releaseConfig.WaitForJobs = true
	rel, err := client.Create(releaseConfig)

	assert.Nil(t, err)
	assert.NotNil(t, rel)

	_ = client.Delete(releaseConfig)
}

func TestReleaseUpdate(t *testing.T) {
	client := getClient()

	releaseName := fmt.Sprintf("test-entropy-helm-client-update-%d", rand.Int())

	releaseConfig := DefaultReleaseConfig()
	releaseConfig.Name = releaseName
	releaseConfig.Repository = "https://odpf.github.io/charts/"
	releaseConfig.Chart = "firehose"
	releaseConfig.Version = "0.1.1"
	releaseConfig.Values = jsonUnmarshal(jsonValues)
	releaseConfig.WaitForJobs = true
	rel, err := client.Create(releaseConfig)

	assert.Nil(t, err)
	assert.NotNil(t, rel)

	releaseConfig.Values = jsonUnmarshal(updatedJsonValues)
	rel2, err := client.Update(releaseConfig)

	assert.Nil(t, err)
	assert.NotNil(t, rel2)

	_ = client.Delete(releaseConfig)
}

func TestReleaseDelete(t *testing.T) {
	client := getClient()

	releaseName := fmt.Sprintf("test-entropy-helm-client-delete-%d", rand.Int())

	releaseConfig := DefaultReleaseConfig()
	releaseConfig.Name = releaseName
	releaseConfig.Repository = "https://odpf.github.io/charts/"
	releaseConfig.Chart = "firehose"
	releaseConfig.Version = "0.1.1"
	releaseConfig.Values = jsonUnmarshal(jsonValues)
	releaseConfig.WaitForJobs = true
	rel, err := client.Create(releaseConfig)

	assert.Nil(t, err)
	assert.NotNil(t, rel)

	err = client.Delete(releaseConfig)

	assert.Nil(t, err)
}
