package firehose

import (
	"errors"
	"fmt"
	"strings"

	"github.com/odpf/entropy/domain"
	gjs "github.com/xeipuuv/gojsonschema"
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
	schema *gjs.Schema
}

func (m *Module) ID() string {
	return "firehose"
}

func New() *Module {
	schemaLoader := gjs.NewStringLoader(configSchemaString)
	schema, err := gjs.NewSchema(schemaLoader)
	if err != nil {
		return nil
	}
	return &Module{
		schema: schema,
	}
}

func (m *Module) Apply(r *domain.Resource) (domain.ResourceStatus, error) {
	return domain.ResourceStatusCompleted, nil
}

func (m *Module) Validate(r *domain.Resource) error {
	resourceLoader := gjs.NewGoLoader(r.Configs)
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
