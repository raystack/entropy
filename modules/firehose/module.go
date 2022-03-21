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
		"name": { "type": "string" },
		"title": { "type": "string" },
		"stream_name": { "type": "string" },
		"topic_name": { "type": "string" },
		"image": { "type": "string" },
		"configuration": {},
		"replicas": { "type": "number" },
		"namespace": { "type": "string" },
		"cluster": { "type": "string" },
		"organization": { "type": "string" },
		"entity": { "type": "string" },
		"consumer_group_id": { "type": "string" },
		"team": { "type": "string" },
		"landscape": { "type": "string" },
		"environment": { "type": "string", "enum": [ "integration", "production" ] },
		"projectID": { "type": "string" },
		"sink_type": { "type": "string" },
		"state": { "type": "string", "enum": [ "running", "stopped" ] },
		"stop_date": { "type": "string", "format": "date-time" },
		"description": { "type": "string" },
		"created_by": { "type": "string" },
		"updated": { "type": "string", "format": "date-time" },
		"created": { "type": "string", "format": "date-time" }
	  },
	  "required": [ "title", "stream_name", "image", "replicas", "namespace", "state" ]
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
