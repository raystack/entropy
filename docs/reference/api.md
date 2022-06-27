import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Entropy CLI and API references

## Entropy Migration

Run DB migrations

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
EXAMPLE
  $ entropy migrate
```

  </TabItem>
</Tabs>

## Entropy Serve

Start gRPC & HTTP servers and optionally workers

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
EXAMPLE
  $ entropy serve
```

  </TabItem>
</Tabs>

## Managing Resources

### Creating Resources

1. Using `entropy resource create` CLI command
2. Calling to `POST /api/v1beta1/resources` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```
FLAGS
  -f, --file string          path to body of resource
  -o, --out -o json | yaml   output format, -o json | yaml

EXAMPLE
  $ entropy resource create --file=<file-path> --out=json
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```console
curl --location --request POST '{{HOST}}/api/v1beta1/resources' \
--header 'Content-Type: application/json' \
--data-raw '{
	"kind": "firehose",
	"name": "p-godata-firehose-001",
	"project": "fb728613-f866-4c1f-8a25-d5032464716d",
	"spec": {
		"dependencies": [
			{
				"key": "kube_cluster",
				"value": "urn:odpf:entropy:kubernetes:test:p-godata-pilot"
			}
		],
		"configs": {
			"release_configs": {
				"namespace": "firehose",
				"values": {
					"firehose": {
						"image": {
							"tag": "latest"
						},
						"config": {
							"KAFKA_RECORD_PARSER_MODE": "foo",
							"SOURCE_KAFKA_BROKERS": "plaintext://localhost:9092",
							"SOURCE_KAFKA_TOPIC": "foo",
							"SOURCE_KAFKA_CONSUMER_GROUP_ID": "consuemr-foo",
							"INPUT_SCHEMA_PROTO_CLASS": "com.gojek.foo.ProtoMessage",
							"SINK_TYPE": "LOG"
						}
					}
				}
			}
		}
	}
}'
```

  </TabItem>
</Tabs>

### Listing Resources

1. Using `entropy resource list` CLI command
2. Calling to `GET /api/v1beta1/resources/` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
FLAGS
  -k, --kind string          kind of resources
  -o, --out -o json | yaml   output format, -o json | yaml
  -p, --project string       project of resources

EXAMPLE
  $ entropy resource list --kind=<resource-kind> --project=<project-name> --out=json
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```console
curl --location --request GET '{{HOST}}/api/v1beta1/resources'
```

  </TabItem>
</Tabs>

### Update Resource

1. Using `entropy resource edit` CLI command
2. Calling to `PATCH /api/v1beta1/resources/:urn` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
FLAGS
  -f, --file string   path to the updated spec of resource

EXAMPLE
  $ entropy resource edit <resource-urn> --file=<file-path>
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```console
curl --location --request PATCH '{{HOST}}/api/v1beta1/resources/{{resource_urn}}' \
--header 'Content-Type: application/json' \
--data-raw '{
		"configs": {
			"release_configs": {
				"values": {
					"firehose": {
						"config": {
							"KAFKA_RECORD_PARSER_MODE": "bar",
							"SOURCE_KAFKA_BROKERS": "http://localhost:9092",
						}
					}
				}
			}
		}
	}'
```

  </TabItem>
</Tabs>

### Viewing Resource

1. Using `entropy resource view` CLI command
2. Calling to `GET /api/v1beta1/resources/:resource` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
FLAGS
  -o, --out -o json | yaml   output format, -o json | yaml

EXAMPLE
  $ entropy resource view <resource-urn> --out=json
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```console
curl --location --request GET '{{HOST}}/api/v1beta1/resources/{{resource_urn}}'
```

  </TabItem>
</Tabs>

### Delete Resource

1. Using `entropy resource delete` CLI command
2. Calling to `DELETE /api/v1beta1/resources/:resource` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
EXAMPLE
  $ entropy resource delete <resource-urn>
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```console
curl --location --request DELETE '{{HOST}}/api/v1beta1/resources/{{resource_urn}}'
```

  </TabItem>
</Tabs>

## Entropy actions

1. Using `entropy action` CLI command
2. Calling to `POST /api/v1beta1/resources/:urn/actions/:action` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
FLAGS
  -f, --file string          path to the params file
  -o, --out -o json | yaml   output format, -o json | yaml
  -u, --urn string           urn of the resource

EXAMPLE
  $ entropy action start --urn=<resource-urn> --file=<file-path> --out=json
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```console
curl --location --request POST '{{HOST}}/api/v1beta1/resources/{{resource_urn}}/actions/{{resource_action}}'
```

  </TabItem>
</Tabs>

## Entropy Logs

1. Using `entropy logs` CLI command
2. Calling to `GET /api/v1beta1/resources/:resource/logs` API

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
FLAGS
  -f, --filter stringArray   Use filters. Example: --filter="key=value"

EXAMPLE
  $ entropy logs <resource-urn> --filter="key1=value1" --filter="key2=value2"
```

  </TabItem>
  <TabItem value="http" label="HTTP">

```console
curl --location --request GET '{{HOST}}/api/v1beta1/resources/{{resource_urn}}/logs'
```

  </TabItem>
</Tabs>

## Entropy Configs

Display configurations currently loaded

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
EXAMPLE
  $ entropy configs
```

  </TabItem>
</Tabs>

## Entropy Version

Shows version information

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
EXAMPLE
  $ entropy version
```

  </TabItem>
</Tabs>

## Entropy Completion

Generate the autocompletion script for entropy for the specified shell. See each sub-command's help for details on how to use the generated script.

### Bash

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

        source <(entropy completion bash)

To load completions for every new session, execute once:

#### Linux:

        entropy completion bash > /etc/bash_completion.d/entropy

#### macOS:

        entropy completion bash > /usr/local/etc/bash_completion.d/entropy

You will need to start a new shell for this setup to take effect.

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
FLAGS
  --no-descriptions   disable completion descriptions

EXAMPLE
  $ entropy completion bash --no-descriptions
```

  </TabItem>
</Tabs>

### Fish

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

        entropy completion fish | source

To load completions for every new session, execute once:

        entropy completion fish > ~/.config/fish/completions/entropy.fish

You will need to start a new shell for this setup to take effect.

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
FLAGS
  --no-descriptions   disable completion descriptions

EXAMPLE
  $ entropy completion fish --no-descriptions
```

  </TabItem>
</Tabs>

### Powershell

Generate the autocompletion script for powershell.

To load completions in your current shell session:

        entropy completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
FLAGS
  --no-descriptions   disable completion descriptions

EXAMPLE
  $ entropy completion powershell --no-descriptions
```

  </TabItem>
</Tabs>

### ZSH

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

        echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for every new session, execute once:

#### Linux:

        entropy completion zsh > "${fpath[1]}/_entropy"

#### macOS:

        entropy completion zsh > /usr/local/share/zsh/site-functions/_entropy

You will need to start a new shell for this setup to take effect.

<Tabs groupId="api">
  <TabItem value="cli" label="CLI" default>

```console
FLAGS
  --no-descriptions   disable completion descriptions

EXAMPLE
  $ entropy completion zsh --no-descriptions
```

  </TabItem>
</Tabs>

