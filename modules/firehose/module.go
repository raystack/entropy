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
	  "replicaCount": {
		"type": "integer"
	  },
	  "nameOverride": {
		"type": "string"
	  },
	  "fullnameOverride": {
		"type": "string"
	  },
	  "labels": {
		"type": "object",
		"properties": {
		  "application": {
			"const": "firehose"
		  }
		}
	  },
	  "firehose": {
		"type": "object",
		"properties": {
		  "image": {
			"type": "object",
			"properties": {
			  "repository": {
				"const": "odpf/firehose"
			  },
			  "pullPolicy": {
				"type": "string"
			  },
			  "tag": {
				"type": "string"
			  }
			}
		  },
		  "config": {
			"type": "object",
			"properties": {
			  "SOURCE_KAFKA_BROKERS": {
				"type": "string"
			  },
			  "SOURCE_KAFKA_CONSUMER_GROUP_ID": {
				"type": "string"
			  },
			  "SOURCE_KAFKA_TOPIC": {
				"type": "string"
			  },
			  "SINK_TYPE": {
				"type": "string"
			  },
			  "SOURCE_KAFKA_CONSUMER_CONFIG_AUTO_OFFSET_RESET": {
				"type": "string"
			  },
			  "INPUT_SCHEMA_PROTO_CLASS": {
				"type": "string"
			  },
			  "JAVA_TOOL_OPTIONS": {
				"type": "string"
			  }
			}
		  },
		  "args": {
			"type": "array",
			"items": {
			  "type": "string"
			},
			"additionalItems": true
		  },
		  "resources": {
			"type": "object",
			"properties": {
			  "limits": {
				"type": "object",
				"properties": {
				  "cpu": {
					"type": "string"
				  },
				  "memory": {
					"type": "string"
				  }
				}
			  },
			  "requests": {
				"type": "object",
				"properties": {
				  "cpu": {
					"type": "string"
				  },
				  "memory": {
					"type": "string"
				  }
				}
			  }
			}
		  }
		}
	  },
	  "init-firehose": {
		"type": "object",
		"properties": {
		  "enabled": {
			"type": "boolean"
		  },
		  "image": {
			"type": "object",
			"properties": {
			  "repository": {
				"const": "busybox"
			  },
			  "pullPolicy": {
				"type": "string"
			  },
			  "tag": {
				"type": "string"
			  }
			}
		  },
		  "command": {
			"type": "array",
			"items": {
			  "type": "string"
			},
			"additionalItems": true
		  },
		  "args": {
			"type": "array",
			"items": {
			  "type": "string"
			},
			"additionalItems": true
		  }
		}
	  },
	  "telegraf": {
		"type": "object",
		"properties": {
		  "enabled": {
			"type": "boolean"
		  },
		  "image": {
			"type": "object",
			"properties": {
			  "repository": {
				"const": "telegraf"
			  },
			  "pullPolicy": {
				"type": "string"
			  },
			  "tag": {
				"type": "string"
			  }
			}
		  },
		  "config": {
			"type": "object",
			"properties": {
			  "output": {
				"type": "object",
				"properties": {
				  "influxdb": {
					"type": "object",
					"properties": {
					  "enabled": {
						"type": "boolean"
					  },
					  "urls": {
						"type": "array",
						"items": {
						  "type": "string"
						},
						"additionalItems": true
					  },
					  "database": {
						"type": "string"
					  },
					  "retention_policy": {
						"type": "string"
					  }
					}
				  }
				}
			  }
			}
		  },
		  "resources": {
			"type": "object",
			"properties": {
			  "limits": {
				"type": "object",
				"properties": {
				  "cpu": {
					"type": "string"
				  },
				  "memory": {
					"type": "string"
				  }
				}
			  },
			  "requests": {
				"type": "object",
				"properties": {
				  "cpu": {
					"type": "string"
				  },
				  "memory": {
					"type": "string"
				  }
				}
			  }
			}
		  }
		}
	  }
	}
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
