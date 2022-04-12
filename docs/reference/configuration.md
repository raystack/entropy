# Configuration

| Go struct                     | YAML path        | ENV var          | default     | Valid values                             |
|-------------------------------|------------------|------------------|-------------|------------------------------------------|
| Config.Service.Port           | service.port     | SERVICE_PORT     | 8080        | 0-65535                                  |
| Config.Service.Host           | service.host     | SERVICE_HOST     | ""          | valid hostname or IP address             |
| Config.DBConfig.Name          | db.name          | DB_NAME          | entropy     |                                          |
| Config.DBConfig.Port          | db.port          | DB_PORT          | 27017       | 0-65535                                  |
| Config.DBConfig.Host          | db.host          | DB_HOST          | localhost   | valid hostname name or IP address        |
| Config.NewRelicConfig.Enabled | newrelic.enabled | NEWRELIC_ENABLED | false       | bool                                     |
| Config.NewRelicConfig.License | newrelic.license | NEWRELIC_LICENSE |             | 40 char NewRelic license key             |
| Config.NewRelicConfig.AppName | newrelic.appname | NEWRELIC_APPNAME | entropy-dev | string                                   |
| Config.LogConfig.Level        | log.level        | LOG_LEVEL        | info        | debug,info,warn,error,dpanic,panic,fatal |

## How to configure

There are 3 ways to configure app:

- Using env variables
- Using a yaml file
- or using a combination of both

### Using env variables

Example:

```sh
$ export PORT=9999
$ entropy serve
```

This will run the service on port 9999 instead of the default 8080

### Using a yaml file

By default `entropy` looks for a configuration file at `./entropy.yaml`. 

This behaviour can be overriden by using `--config <config-file>` flag.

Usage example:

```sh
$ entropy serve --config entropy.yaml
```

### Using a combination of both

If any key that is set via both env vars and yaml the value set in env vars will take effect.
