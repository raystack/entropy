package store

import (
	"errors"
	"github.com/odpf/entropy/domain"
)

// Custom errors which can be used by multiple DB vendors
var (
	ResourceAlreadyExistsError = errors.New("resource already exists")
	ResourceNotFoundError      = errors.New("no resource(s) found")
)

var ResourceRepositoryName = "resources"

type ResourceRepository interface {
	Create(r *domain.Resource) error
	Update(r *domain.Resource) error
	GetByURN(urn string) (*domain.Resource, error)
	Migrate() error
}
