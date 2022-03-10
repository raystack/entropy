package domain

//go:generate mockery --name=Module -r --case underscore --with-expecter --structname Module --filename=module.go --output=../mocks

import "errors"

var (
	ModuleConfigParseFailed = errors.New("unable to load and validate config")
)

type Module interface {
	ID() string
	Apply(r *Resource) (ResourceStatus, error)
	Validate(r *Resource) error
}
