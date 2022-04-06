package domain

//go:generate mockery --name=Module -r --case underscore --with-expecter --structname Module --filename=module.go --output=../mocks

import (
	"context"
	"errors"
)

var (
	ErrModuleConfigParseFailed = errors.New("unable to load and validate config")
)

type Module interface {
	ID() string
	Apply(r *Resource) (ResourceStatus, error)
	Validate(r *Resource) error
	Act(r *Resource, action string, params map[string]interface{}) (map[string]interface{}, error)
	Log(ctx context.Context, r *Resource, filter map[string]string) (<-chan LogChunk, error)
}

type LogChunk struct {
	Data   []byte
	Labels map[string]string
}
