package store

import (
	"context"
	"fmt"

	"github.com/odpf/entropy/domain"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// New returns the database instance
func New(config *domain.DBConfig) (*mongo.Database, error) {
	uri := fmt.Sprintf(
		"mongodb://%s:%s/%s",
		config.Host,
		config.Port,
		config.Name,
	)

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	return client.Database(config.Name), err
}
