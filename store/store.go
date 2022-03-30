package store

//go:generate mockery --name=ModuleRepository -r --case underscore --with-expecter --structname ModuleRepository --filename=module_repository.go --output=../mocks
//go:generate mockery --name=ResourceRepository -r --case underscore --with-expecter --structname ResourceRepository --filename=resource_repository.go --output=../mocks

import (
	"errors"

	"github.com/odpf/entropy/domain"
)

// Custom errors which can be used by multiple DB vendors
var (
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrResourceNotFound      = errors.New("no resource(s) found")
	ErrModuleAlreadyExists   = errors.New("module already exists")
	ErrModuleNotFound        = errors.New("no module(s) found")
)

var ResourceRepositoryName = "resources"

type ResourceRepository interface {
	Create(r *domain.Resource) error
	Update(r *domain.Resource) error
	GetByURN(urn string) (*domain.Resource, error)
	Migrate() error
	List(filter map[string]string) ([]*domain.Resource, error)
	Delete(urn string) error
}

type ModuleRepository interface {
	Register(module domain.Module) error
	Get(id string) (domain.Module, error)
}
