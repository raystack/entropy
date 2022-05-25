package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config contains the database configurations.
type Config struct {
	// URI should be valid MongodDB connection string.
	// https://www.mongodb.com/docs/manual/reference/connection-string/
	URI string `mapstructure:"uri" default:"mongodb://localhost:27017"`

	// Name should be the name of the database to use.
	Name string `mapstructure:"name" default:"entropy"`

	// PingTimeout decides the maximum time to wait for ping response.
	PingTimeout time.Duration `mapstructure:"ping_timeout" default:"3s"`
}

// Connect returns the database instance.
func Connect(cfg Config) (*mongo.Database, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	pingCtx, pingCancel := context.WithTimeout(context.TODO(), cfg.PingTimeout)
	defer pingCancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, err
	}

	return client.Database(cfg.Name), nil
}
