package module

//go:generate mockery --name=Loggable -r --case underscore --with-expecter --structname LoggableModule --filename=loggable_module.go --output=../mocks

import "context"

// Loggable extension of module allows streaming log data for a resource.
type Loggable interface {
	Module

	Log(ctx context.Context, spec Spec, filter map[string]string) (<-chan LogChunk, error)
}

type LogChunk struct {
	Data   []byte
	Labels map[string]string
}
