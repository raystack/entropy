package service

import (
	"github.com/odpf/entropy/domain/resource"
	"github.com/odpf/entropy/pkg/store"
)

type Container struct {
	ResourceRepository *resource.Repository
}

func Init(db *store.DB) (*Container, error) {
	resourceRepository := &resource.Repository{DB: db}
	return &Container{
		ResourceRepository: resourceRepository,
	}, nil
}

func (container *Container) MigrateAll(db *store.DB) error {
	return container.ResourceRepository.Migrate()
}
