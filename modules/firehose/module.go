package firehose

import (
	"errors"
	"fmt"
	"strings"

	"github.com/odpf/entropy/domain"
	gjs "github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
)

const configSchemaString = `
{
	"$schema": "http://json-schema.org/draft-04/schema#",
	"type": "object",
	"properties": {
		"image": { "type": "string" },
		"configuration": {
			"type": "object",
			"properties": {
				"SOURCE_KAFKA_BROKERS": { "type": "string" },
				"SOURCE_KAFKA_TOPIC": { "type": "string" },
				"SOURCE_KAFKA_CONSUMER_GROUP_ID": { "type": "string" },
				"INPUT_SCHEMA_PROTO_CLASS": { "type": "string" },
			}
		},
		"replicas": { "type": "number" },
		"namespace": { "type": "string" },
		"cluster": { "type": "string" },
		"sink_type": { "type": "string",},
		"state": { "type": "string", "enum": [ "RUNNING", "STOPPED" ] },
		"stop_date": { "type": "string", "format": "date-time" },
		"description": { "type": "string" },
	},
	"oneOf": [
		{
		  "if": {
			"properties": { "sink_type": { "const": "LOG" } }
		  },
		  "then": {
			"properties": { "configuration": { 
				"KAFKA_RECORD_PARSER_MODE": { "type": "string" },
				"SOURCE_KAFKA_BROKERS": { "type": "string" },
				"SOURCE_KAFKA_TOPIC": { "type": "string" },
				"SOURCE_KAFKA_CONSUMER_GROUP_ID": { "type": "string" },
				"INPUT_SCHEMA_PROTO_CLASS": { "type": "string" },
			 } }
		  }
		},
		{
		  "if": {
			"properties": { "country": { "const": "HTTP" } },
		  },
		  "then": {
			"properties": { "configuration": {
				"SOURCE_KAFKA_BROKERS": { "type": "string" },
				"SOURCE_KAFKA_TOPIC": { "type": "string" },
				"SOURCE_KAFKA_CONSUMER_GROUP_ID": { "type": "string" },
				"INPUT_SCHEMA_PROTO_CLASS": { "type": "string" },
				"SINK_HTTP_RETRY_STATUS_CODE_RANGES": { "type": "string" },
				"SINK_HTTP_REQUEST_LOG_STATUS_CODE_RANGES": { "type": "string" },
				"SINK_HTTP_REQUEST_TIMEOUT_MS": { "type": "number" },
				"SINK_HTTP_REQUEST_METHOD": { "type": "string", "enum": [ "put", "post" ] },
				"SINK_HTTP_MAX_CONNECTIONS": { "type": "number" },
				"SINK_HTTP_SERVICE_URL": { "type": "string" },
				"SINK_HTTP_HEADERS": { "type": "string" },
				"SINK_HTTP_PARAMETER_SOURCE": { "type": "string", "enum": [ "key", "message", "disabled" ] },
				"SINK_HTTP_DATA_FORMAT": { "type": "string", "enum": [ "proto", "json" ] },
				"SINK_HTTP_OAUTH2_ENABLE": { "type": "boolean" },
				"SINK_HTTP_OAUTH2_ACCESS_TOKEN_URL": { "type": "string" },
				"SINK_HTTP_OAUTH2_CLIENT_NAME": { "type": "string" },
				"SINK_HTTP_OAUTH2_CLIENT_SECRET": { "type": "string" },
				"SINK_HTTP_OAUTH2_SCOPE": { "type": "string" },
				"SINK_HTTP_JSON_BODY_TEMPLATE": { "type": "string" },
				"SINK_HTTP_PARAMETER_PLACEMENT": { "type": "string", "enum": [ "query", "header" ] },
				"SINK_HTTP_PARAMETER_SCHEMA_PROTO_CLASS": { "type": "string" },
			 } }
		  }
		},
	],
	"required": [ "image", "replicas", "namespace", "state" ]
}
`

type Module struct {
	schema *gjs.Schema
	logger *zap.Logger
}

func (m *Module) ID() string {
	return "firehose"
}

func New(logger *zap.Logger) *Module {
	schemaLoader := gjs.NewStringLoader(configSchemaString)
	schema, err := gjs.NewSchema(schemaLoader)
	if err != nil {
		return nil
	}
	return &Module{
		schema: schema,
		logger: logger,
	}
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
