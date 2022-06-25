# CLI

## `entropy migrate`

Run DB migrations

```
EXAMPLE
  $ entropy migrate
```

## `entropy serve`

Start gRPC & HTTP servers and optionally workers

```
EXAMPLE
  $ entropy serve
```

## `entropy resource`

Manage resources

### `entropy resource create [flags]`

Create a resource

```
FLAGS
  -f, --file string          path to body of resource
  -o, --out -o json | yaml   output format, -o json | yaml
```

```
EXAMPLE
  $ entropy resource create --file=<file-path> --out=json
```

### `entropy resource list [flags]`

```
FLAGS
  -k, --kind string          kind of resources
  -o, --out -o json | yaml   output format, -o json | yaml
  -p, --project string       project of resources
```

```
EXAMPLE
  $ entropy resource list --kind=<resource-kind> --project=<project-name> --out=json
```

### `entropy resource view <resource-urn> [flags]`

View a resource

```
FLAGS
  -o, --out -o json | yaml   output format, -o json | yaml
```

```
EXAMPLE
  $ entropy resource view <resource-urn> --out=json
```

### `entropy resource edit <resource-urn> [flags]`

Edit a resource

```
FLAGS
  -f, --file string   path to the updated spec of resource
```

```
EXAMPLE
  $ entropy resource edit <resource-urn> --file=<file-path>
```

### `entropy resource delete <resource-urn>`

Delete a resource

```
EXAMPLE
  $ entropy resource delete <resource-urn>
```

## `entropy action <action-name> [flags]`

Manage actions

```
FLAGS
  -f, --file string          path to the params file
  -o, --out -o json | yaml   output format, -o json | yaml
  -u, --urn string           urn of the resource
```

```
EXAMPLE
  $ entropy action start --urn=<resource-urn> --file=<file-path> --out=json
```

## `entropy logs <resource-urn> [flags]`

Gets logs

```
FLAGS
  -f, --filter stringArray   Use filters. Example: --filter="key=value"
```

```
EXAMPLE
  $ entropy logs <resource-urn> --filter="key1=value1" --filter="key2=value2"
```

## `entropy configs`

Display configurations currently loaded

```
EXAMPLE
  $ entropy configs
```

## `entropy version`

Show version information

```
EXAMPLE
  $ entropy version
```

## `entropy completion`

Generate the autocompletion script for entropy for the specified shell.
See each sub-command's help for details on how to use the generated script.

### `entropy completion bash [flags]`

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

```
FLAGS
  --no-descriptions   disable completion descriptions
```

```
EXAMPLE
  $ entropy completion bash --no-descriptions
```

### `entropy completion fish [flags]`

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

        entropy completion fish | source

To load completions for every new session, execute once:

        entropy completion fish > ~/.config/fish/completions/entropy.fish

You will need to start a new shell for this setup to take effect.

```
FLAGS
  --no-descriptions   disable completion descriptions
```

```
EXAMPLE
  $ entropy completion fish --no-descriptions
```

### `entropy completion powershell [flags]`

Generate the autocompletion script for powershell.

To load completions in your current shell session:

        entropy completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.

```
FLAGS
  --no-descriptions   disable completion descriptions
```

```
EXAMPLE
  $ entropy completion powershell --no-descriptions
```

### `entropy completion zsh [flags]`

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

```
FLAGS
  --no-descriptions   disable completion descriptions
```

```
EXAMPLE
  $ entropy completion zsh --no-descriptions
```