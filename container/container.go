package container

import (
	"github.com/odpf/entropy/pkg/resource"
	"github.com/odpf/entropy/store"
	"github.com/odpf/entropy/store/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct {
	ResourceService    resource.ServiceInterface
	resourceRepository store.ResourceRepository
}

func NewContainer(db *mongo.Database) (*Container, error) {
	resourceRepository := mongodb.NewRepository(db)
	resourceService := resource.NewService(resourceRepository)
	return &Container{
		ResourceService:    resourceService,
		resourceRepository: resourceRepository,
	}, nil
}

func (container *Container) MigrateAll(db *mongo.Database) error {
	return container.resourceRepository.Migrate()
}
