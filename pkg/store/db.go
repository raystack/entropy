package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBConfig contains the database configuration
type DBConfig struct {
	Host string `mapstructure:"host" default:"localhost"`
	Port string `mapstructure:"port" default:"27017"`
	Name string `mapstructure:"name" default:"entropy"`
}

type DB struct {
	db *mongo.Database
}

func (d *DB) GetCollection(name string) *Collection {
	return &Collection{
		collection: d.db.Collection(name),
	}
}

// New returns the database instance
func New(config *DBConfig) (*DB, error) {
	uri := fmt.Sprintf(
		"mongodb://%s:%s/%s",
		config.Host,
		config.Port,
		config.Name,
	)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer pingCancel()

	err = client.Ping(pingCtx, nil)

	return &DB{
		db: client.Database(config.Name),
	}, err
}
