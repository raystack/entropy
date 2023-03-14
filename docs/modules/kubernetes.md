# Kubernetes

[Kubernetes](https://kubernetes.io/) is an open-source container orchestration system for automating software deployment, scaling, and management. Google originally designed Kubernetes, but the Cloud Native Computing Foundation now maintains the project.

## What happens in Plan?

Entropy creates a kubernetes client and stores the config and client info in the resource output.

## What happens in Sync?

Sync in Kubernetes module is a passive one i.e. it has no side-effects. It just gives the resource information with "STATUS_COMPLETED".

## Kubernetes Module Configuration

The configuration struct for Kubernetes module looks like:

```
type Config struct {
	Host string `json:"host"`

	Timeout time.Duration `json:"timeout" default:"100ms"`

	Token string `json:"token"`

	Insecure bool `json:"insecure" default:"false"`

	ClientKey string `json:"client_key"`

	ClientCertificate string `json:"client_certificate"`

	ClusterCACertificate string `json:"cluster_ca_certificate"`
}
```

| Fields | |
| :--- | :--- |
| `Host` | `string` The hostname (in form of URI) of Kubernetes master. |
| `Timeout` | `number` Connection timeout time. default: 100 |
| `Token` | `string` Token to authenticate a service account. |
| `Insecure` | `bool` Whether server should be accessed without verifying the TLS certificate. Default: false |
| `ClusterCACertificate` | `string` PEM-encoded root certificates bundle for TLS authentication. |
| `ClientKey` | `string` PEM-encoded client key for TLS authentication. |
| `ClientCertificate` | `string` PEM-encoded client certificate for TLS authentication. |

Note: User shall either enable Insecure or set ClusterCACertificate. Also, user can either use Token to aunthenate a service account or they can use ClientKey & ClientCertificate for TLS authentication.
Detailed JSONSchema for config can be referenced [here](https://github.com/goto/entropy/blob/main/modules/kubernetes/config_schema.json).

## Supported actions

| Fields | |
| :--- | :--- |
| `Create` | To create a Kubernetes resource. |
| `Update` | To update a Kubernetes resource. |