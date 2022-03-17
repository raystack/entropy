package log

import (
	"errors"
	"fmt"
	"github.com/odpf/entropy/domain"
	gjs "github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
	"strings"
)

type LogLevel string

const (
	Error LogLevel = "ERROR"
	Warn  LogLevel = "WARN"
	Info  LogLevel = "INFO"
	Debug LogLevel = "DEBUG"
)

const configSchemaString = `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "properties": {
    "log_level": {
      "type": "string",
      "enum": ["ERROR", "WARN", "INFO", "DEBUG"]
    }
  },
  "required": [
    "log_level"
  ]
}
`

type Module struct {
	schema *gjs.Schema
	logger *zap.Logger
}

func (m *Module) ID() string {
	return "log"
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

func (m *Module) Apply(r *domain.Resource) (domain.ResourceStatus, error) {
	switch r.Configs["log_level"].(LogLevel) {
	case Error:
		m.logger.Sugar().Error(r)
	case Warn:
		m.logger.Sugar().Warn(r)
	case Info:
		m.logger.Sugar().Info(r)
	case Debug:
		m.logger.Sugar().Debug(r)
	default:
		return domain.ResourceStatusError, errors.New("unknown log level")
	}
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
	switch action {
	case "escalate":
		r.Configs["log_level"] = increaseLogLevel(r.Configs["log_level"].(LogLevel))
	}
	return r.Configs, nil
}

func increaseLogLevel(currentLevel LogLevel) LogLevel {
	switch currentLevel {
	case Error:
		return Error
	case Warn:
		return Error
	case Info:
		return Warn
	case Debug:
		return Info
	default:
		return Info
	}
}
