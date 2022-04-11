package module

//go:generate mockery --name=Module -r --case underscore --with-expecter --structname Module --filename=module.go --output=../mocks
//go:generate mockery --name=Loggable -r --case underscore --with-expecter --structname LoggableModule --filename=loggable_module.go --output=../mocks
//go:generate mockery --name=Repository -r --case underscore --with-expecter --structname ModuleRepository --filename=module_repository.go --output=../mocks

import (
	"context"
	"errors"

	"github.com/odpf/entropy/resource"
)

var (
	ErrModuleNotFound      = errors.New("no module(s) found")
	ErrModuleAlreadyExists = errors.New("module already exists")

	ErrModuleConfigParseFailed = errors.New("unable to load and validate config")
	ErrLogStreamingUnsupported = errors.New("log streaming is not supported for this module")
)

type Module interface {
	ID() string
	Act(r *resource.Resource, action string, params map[string]interface{}) (map[string]interface{}, error)
	Apply(r *resource.Resource) (resource.Status, error)
	Validate(r *resource.Resource) error
}

type Loggable interface {
	Module

	Log(ctx context.Context, r *resource.Resource, filter map[string]string) (<-chan LogChunk, error)
}

type LogChunk struct {
	Data   []byte
	Labels map[string]string
}

type Repository interface {
	Get(id string) (Module, error)
	Register(m Module) error
}
