package domain

import "errors"

var (
	ModuleConfigParseFailed = errors.New("unable to load and validate config")
)

type Module interface {
	ID() string
	Apply(r *Resource) (ResourceStatus, error)
	Validate(r *Resource) error
}
