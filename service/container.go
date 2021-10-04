package service

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Container struct{}

func Init(db *mongo.Database) (*Container, error) {
	return &Container{}, nil
}

func (container *Container) MigrateAll(db *mongo.Database) error {
	return nil
}
