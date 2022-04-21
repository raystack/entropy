package resource

//go:generate mockery --name=Module -r --case underscore --with-expecter --structname Module --filename=module.go --output=./mocks
//go:generate mockery --name=LoggableModule -r --case underscore --with-expecter --structname LoggableModule --filename=loggable_module.go --output=./mocks
//go:generate mockery --name=ModuleRegistry -r --case underscore --with-expecter --structname ModuleRegistry --filename=module_registry.go --output=./mocks

import (
	"context"
)

type Module interface {
	ID() string
	Act(r Resource, action string, params map[string]interface{}) (map[string]interface{}, error)
	Apply(r Resource) (Status, error)
	Validate(r Resource) error
}

type LoggableModule interface {
	Module

	Log(ctx context.Context, r Resource, filter map[string]string) (<-chan LogChunk, error)
}

type LogChunk struct {
	Data   []byte
	Labels map[string]string
}

type ModuleRegistry interface {
	Get(id string) (Module, error)
	Register(m Module) error
}
