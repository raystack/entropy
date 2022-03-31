package firehose

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/pkg/provider/helm"
	"github.com/odpf/entropy/store/mongodb"
	gjs "github.com/xeipuuv/gojsonschema"
)

const (
	nameConfigString            = "name"
	repositoryConfigString      = "repository"
	chartConfigString           = "chart"
	versionConfigString         = "version"
	valuesConfigString          = "values"
	namespaceConfigString       = "namespace"
	timeoutConfigString         = "timeout"
	forceUpdateConfigString     = "force_update"
	recreatePodsConfigString    = "recreate_pods"
	waitConfigString            = "wait"
	waitForJobsConfigString     = "wait_for_jobs"
	replaceConfigString         = "replace"
	descriptionConfigString     = "description"
	CreateNamespaceConfigString = "create_namespace"
)

const configSchemaString = `
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"$id": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"properties": {
	  "image": {
		"type": "string"
	  },
	  "replicas": {
		"type": "number"
	  },
	  "namespace": {
		"type": "string"
	  },
	  "cluster": {
		"type": "string"
	  },
	  "sink_type": {
		"type": "string",
		"enum": [
		  "LOG",
		  "HTTP"
		]
	  },
	  "state": {
		"type": "string",
		"enum": [
		  "RUNNING",
		  "STOPPED"
		]
	  },
	  "stop_date": {
		"type": "string",
		"format": "date-time"
	  },
	  "description": {
		"type": "string"
	  }
	},
	"allOf": [
	  {
		"if": {
		  "properties": {
			"sink_type": {
			  "const": "LOG"
			}
		  },
		  "required": [
			"sink_type"
		  ]
		},
		"then": {
		  "properties": {
			"configuration": {
			  "type": "object",
			  "properties": {
				"KAFKA_RECORD_PARSER_MODE": {
				  "type": "string"
				},
				"SOURCE_KAFKA_BROKERS": {
				  "type": "string"
				},
				"SOURCE_KAFKA_TOPIC": {
				  "type": "string"
				},
				"SOURCE_KAFKA_CONSUMER_GROUP_ID": {
				  "type": "string"
				},
				"INPUT_SCHEMA_PROTO_CLASS": {
				  "type": "string"
				}
			  },
			  "required": [
				"KAFKA_RECORD_PARSER_MODE",
				"SOURCE_KAFKA_BROKERS",
				"SOURCE_KAFKA_TOPIC",
				"SOURCE_KAFKA_CONSUMER_GROUP_ID",
				"INPUT_SCHEMA_PROTO_CLASS"
			  ]
			}
		  }
		}
	  },
	  {
		"if": {
		  "properties": {
			"sink_type": {
			  "const": "HTTP"
			}
		  },
		  "required": [
			"sink_type"
		  ]
		},
		"then": {
		  "properties": {
			"configuration": {
			  "type": "object",
			  "properties": {
				"SOURCE_KAFKA_BROKERS": {
				  "type": "string"
				},
				"SOURCE_KAFKA_TOPIC": {
				  "type": "string"
				},
				"SOURCE_KAFKA_CONSUMER_GROUP_ID": {
				  "type": "string"
				},
				"INPUT_SCHEMA_PROTO_CLASS": {
				  "type": "string"
				},
				"SINK_HTTP_RETRY_STATUS_CODE_RANGES": {
				  "type": "string"
				},
				"SINK_HTTP_REQUEST_LOG_STATUS_CODE_RANGES": {
				  "type": "string"
				},
				"SINK_HTTP_REQUEST_TIMEOUT_MS": {
				  "type": "number"
				},
				"SINK_HTTP_REQUEST_METHOD": {
				  "type": "string",
				  "enum": [
					"put",
					"post"
				  ]
				},
				"SINK_HTTP_MAX_CONNECTIONS": {
				  "type": "number"
				},
				"SINK_HTTP_SERVICE_URL": {
				  "type": "string"
				},
				"SINK_HTTP_HEADERS": {
				  "type": "string"
				},
				"SINK_HTTP_PARAMETER_SOURCE": {
				  "type": "string",
				  "enum": [
					"key",
					"message",
					"disabled"
				  ]
				},
				"SINK_HTTP_DATA_FORMAT": {
				  "type": "string",
				  "enum": [
					"proto",
					"json"
				  ]
				},
				"SINK_HTTP_OAUTH2_ENABLE": {
				  "type": "boolean"
				},
				"SINK_HTTP_OAUTH2_ACCESS_TOKEN_URL": {
				  "type": "string"
				},
				"SINK_HTTP_OAUTH2_CLIENT_NAME": {
				  "type": "string"
				},
				"SINK_HTTP_OAUTH2_CLIENT_SECRET": {
				  "type": "string"
				},
				"SINK_HTTP_OAUTH2_SCOPE": {
				  "type": "string"
				},
				"SINK_HTTP_JSON_BODY_TEMPLATE": {
				  "type": "string"
				},
				"SINK_HTTP_PARAMETER_PLACEMENT": {
				  "type": "string",
				  "enum": [
					"query",
					"header"
				  ]
				},
				"SINK_HTTP_PARAMETER_SCHEMA_PROTO_CLASS": {
				  "type": "string"
				}
			  },
			  "required": [
				"SOURCE_KAFKA_BROKERS",
				"SOURCE_KAFKA_TOPIC",
				"SOURCE_KAFKA_CONSUMER_GROUP_ID",
				"INPUT_SCHEMA_PROTO_CLASS",
				"SINK_HTTP_PARAMETER_SCHEMA_PROTO_CLASS",
				"SINK_HTTP_PARAMETER_PLACEMENT",
				"SINK_HTTP_JSON_BODY_TEMPLATE",
				"SINK_HTTP_OAUTH2_SCOPE",
				"SINK_HTTP_OAUTH2_CLIENT_SECRET",
				"SINK_HTTP_OAUTH2_CLIENT_NAME",
				"SINK_HTTP_OAUTH2_ACCESS_TOKEN_URL",
				"SINK_HTTP_OAUTH2_ENABLE",
				"SINK_HTTP_DATA_FORMAT",
				"SINK_HTTP_PARAMETER_SOURCE",
				"SINK_HTTP_HEADERS",
				"SINK_HTTP_SERVICE_URL",
				"SINK_HTTP_MAX_CONNECTIONS",
				"SINK_HTTP_REQUEST_METHOD",
				"SINK_HTTP_REQUEST_TIMEOUT_MS",
				"SINK_HTTP_REQUEST_LOG_STATUS_CODE_RANGES",
				"SINK_HTTP_RETRY_STATUS_CODE_RANGES"
			  ]
			}
		  }
		}
	  }
	],
	"required": [
	  "image",
	  "replicas",
	  "namespace",
	  "state"
	]
  }
`

type Module struct {
	schema             *gjs.Schema
	providerRepository *mongodb.ProviderRepository
}

func (m *Module) ID() string {
	return "firehose"
}

func New(providerRepository *mongodb.ProviderRepository) *Module {
	schemaLoader := gjs.NewStringLoader(configSchemaString)
	schema, err := gjs.NewSchema(schemaLoader)
	if err != nil {
		return nil
	}
	return &Module{
		schema:             schema,
		providerRepository: providerRepository,
	}
}

func (m *Module) Apply(r *domain.Resource) (domain.ResourceStatus, error) {
	kubeProviderConfig, err := m.providerRepository.GetConfigByURN(r.Providers["kubernetes"].Urn)
	if err != nil {
		return domain.ResourceStatusError, err
	}
	kubeConfig := helm.ToKubeConfig(kubeProviderConfig)
	helmConfig := &helm.ProviderConfig{
		Kubernetes: kubeConfig,
	}
	helmProvider := helm.NewProvider(helmConfig)

	v := reflect.ValueOf(r.Configs[valuesConfigString])
	var values = make(map[string]interface{})
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			strct := v.MapIndex(key)
			values[key.String()] = strct.Interface()
		}
	}

	var rc helm.ReleaseConfig
	if err := mapstructure.Decode(r.Configs, &rc); err != nil {
		return domain.ResourceStatusError, err
	}

	releaseConfig := helm.DefaultReleaseConfig()
	releaseConfig.Values = values

	if r.Configs[nameConfigString] != nil {
		releaseConfig.Name = r.Configs[nameConfigString].(string)
	}

	if r.Configs[repositoryConfigString] != nil {
		releaseConfig.Repository = r.Configs[repositoryConfigString].(string)
	}

	if r.Configs[chartConfigString] != nil {
		releaseConfig.Chart = r.Configs[chartConfigString].(string)
	}

	if r.Configs[versionConfigString] != nil {
		releaseConfig.Version = r.Configs[versionConfigString].(string)
	}

	if r.Configs[namespaceConfigString] != nil {
		releaseConfig.Namespace = r.Configs[namespaceConfigString].(string)
	}

	if r.Configs[timeoutConfigString] != nil {
		releaseConfig.Timeout = r.Configs[timeoutConfigString].(int)
	}

	if r.Configs[descriptionConfigString] != nil {
		releaseConfig.Description = r.Configs[descriptionConfigString].(string)
	}

	if r.Configs[forceUpdateConfigString] != nil {
		releaseConfig.ForceUpdate = r.Configs[forceUpdateConfigString].(bool)
	}

	if r.Configs[recreatePodsConfigString] != nil {
		releaseConfig.RecreatePods = r.Configs[recreatePodsConfigString].(bool)
	}

	if r.Configs[waitConfigString] != nil {
		releaseConfig.Wait = r.Configs[waitConfigString].(bool)
	}

	if r.Configs[waitForJobsConfigString] != nil {
		releaseConfig.WaitForJobs = r.Configs[waitForJobsConfigString].(bool)
	}

	if r.Configs[replaceConfigString] != nil {
		releaseConfig.Replace = r.Configs[replaceConfigString].(bool)
	}

	if r.Configs[CreateNamespaceConfigString] != nil {
		releaseConfig.CreateNamespace = r.Configs[CreateNamespaceConfigString].(bool)
	}

	_, err = helmProvider.Release(releaseConfig)
	if err != nil {
		return domain.ResourceStatusError, nil
	}

	return domain.ResourceStatusCompleted, nil
}

func (m *Module) Validate(r *domain.Resource) error {
	resourceLoader := gjs.NewGoLoader(r.Configs["values"])
	result, err := m.schema.Validate(resourceLoader)
	if err != nil {
		return fmt.Errorf("%w: %s", domain.ModuleConfigParseFailed, err)
	}
	if !result.Valid() {
		var errorStrings []string
		for _, resultErr := range result.Errors() {
			errorStrings = append(errorStrings, resultErr.String())
		}
		errorString := strings.Join(errorStrings, "\n")
		return errors.New(errorString)
	}
	return nil
}

func (m *Module) Act(r *domain.Resource, action string, params map[string]interface{}) (map[string]interface{}, error) {
	return r.Configs, nil
}
