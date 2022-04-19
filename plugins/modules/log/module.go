package log

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	gjs "github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"

	"github.com/odpf/entropy/core/resource"
)

type Level string

const (
	LevelError Level = "ERROR"
	LevelWarn  Level = "WARN"
	LevelInfo  Level = "INFO"
	LevelDebug Level = "DEBUG"

	levelConfigString = "log_level"
)

const configSchemaString = `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "properties": {
    "log_level": {
      "type": "string",
      "enum": ["ERROR", "WARN", "INFO", "DEBUG"]
    },
    "delay_ms": {
      "type": "integer"
    }
  },
  "required": [
    "log_level",
    "delay_ms"
  ]
}
`

type config struct {
	LogLevel Level `mapstructure:"log_level"`
	DelayMs  int   `mapstructure:"delay_ms"`
}

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

func (m *Module) Apply(r resource.Resource) (resource.Status, error) {
	var cfg config
	if err := mapstructure.Decode(r.Configs, &cfg); err != nil {
		return resource.StatusError, errors.New("unable to parse configs")
	}
	switch cfg.LogLevel {
	case LevelError:
		m.logger.Sugar().Error(r)
	case LevelWarn:
		m.logger.Sugar().Warn(r)
	case LevelInfo:
		m.logger.Sugar().Info(r)
	case LevelDebug:
		m.logger.Sugar().Debug(r)
	default:
		return resource.StatusError, errors.New("unknown log level")
	}

	return resource.StatusCompleted, nil
}

func (m *Module) Validate(r resource.Resource) error {
	resourceLoader := gjs.NewGoLoader(r.Configs)
	result, err := m.schema.Validate(resourceLoader)
	if err != nil {
		return fmt.Errorf("%w: %s", resource.ErrModuleConfigParseFailed, err)
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

func (m *Module) Act(r resource.Resource, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "escalate":
		r.Configs[levelConfigString] = increaseLogLevel(r.Configs[levelConfigString].(Level))
	}
	return r.Configs, nil
}

func (m *Module) Log(ctx context.Context, r resource.Resource, filter map[string]string) (<-chan resource.LogChunk, error) {
	var cfg config
	if err := mapstructure.Decode(r.Configs, &cfg); err != nil {
		return nil, errors.New("unable to parse configs")
	}
	logs := make(chan resource.LogChunk)
	go func() {
		defer close(logs)
		for {
			select {
			case logs <- resource.LogChunk{
				Data:   []byte(fmt.Sprintf("%v", r)),
				Labels: map[string]string{"resource": r.URN},
			}:
				time.Sleep(time.Millisecond * time.Duration(cfg.DelayMs))
			case <-ctx.Done():
				return
			}
		}
	}()
	return logs, nil
}

func increaseLogLevel(currentLevel Level) Level {
	switch currentLevel {
	case LevelError:
		return LevelError
	case LevelWarn:
		return LevelError
	case LevelInfo:
		return LevelWarn
	case LevelDebug:
		return LevelInfo
	default:
		return LevelInfo
	}
}
