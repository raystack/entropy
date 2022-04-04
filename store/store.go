package store

//go:generate mockery --name=ModuleRepository -r --case underscore --with-expecter --structname ModuleRepository --filename=module_repository.go --output=../mocks
//go:generate mockery --name=ResourceRepository -r --case underscore --with-expecter --structname ResourceRepository --filename=resource_repository.go --output=../mocks

import (
	"errors"

	"github.com/odpf/entropy/domain"
)

// Custom errors which can be used by multiple DB vendors
var (
	ResourceAlreadyExistsError = errors.New("resource already exists")
	ResourceNotFoundError      = errors.New("no resource(s) found")
	ModuleAlreadyExistsError   = errors.New("module already exists")
	ModuleNotFoundError        = errors.New("no module(s) found")
	ProviderAlreadyExistsError = errors.New("provider already exists")
	ProviderNotFoundError      = errors.New("no provider(s) found")
)

var ResourceRepositoryName = "resources"
var ProviderRepositoryName = "providers"

type ResourceRepository interface {
	Create(r *domain.Resource) error
	Update(r *domain.Resource) error
	GetByURN(urn string) (*domain.Resource, error)
	Migrate() error
	List(filter map[string]string) ([]*domain.Resource, error)
	Delete(urn string) error
}

type ProviderRepository interface {
	Create(r *domain.Provider) error
	GetByURN(urn string) (*domain.Provider, error)
	List(filter map[string]string) ([]*domain.Provider, error)
	Migrate() error
}

type ModuleRepository interface {
	Register(module domain.Module) error
	Get(id string) (domain.Module, error)
}
