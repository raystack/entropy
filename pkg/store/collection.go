package store

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection struct {
	collection *mongo.Collection
}

var (
	AlreadyExistsError = errors.New("failed to insert into db: already exists")
	UpdateFailedError  = errors.New("failed to update in db")
	FindFailedError    = errors.New("failed to query db")
	NotFoundError      = errors.New("no such record(s) in db")
)

func (c *Collection) CreateUniqueIndex(key string, order int) error {
	_, err := c.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: key, Value: order}},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

func (c *Collection) InsertOne(document interface{}) error {
	_, err := c.collection.InsertOne(context.TODO(), document)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: %s", AlreadyExistsError, err)
		}
		return err
	}
	return nil
}

func (c *Collection) UpdateOne(filter interface{}, document interface{}) error {
	singleResult := c.collection.FindOneAndUpdate(context.TODO(),
		filter,
		map[string]interface{}{"$set": document},
	)
	err := singleResult.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("%w: filter = %s", NotFoundError, filter)
		}
		return fmt.Errorf("%w: %s", UpdateFailedError, err)
	}
	return nil
}

func (c *Collection) FindOne(filter interface{}, result interface{}) error {
	err := c.collection.FindOne(context.TODO(), filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("%w: %s", NotFoundError, err)
		}
		return fmt.Errorf("%w: %s", FindFailedError, err)
	}
	return nil
}

func (c *Collection) Find(filter map[string]interface{}, result interface{}) error {
	cursor, err := c.collection.Find(context.TODO(), filter)
	if err != nil {
		return fmt.Errorf("%w: %s", FindFailedError, err)
	}
	err = cursor.All(context.TODO(), result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("%w: %s", NotFoundError, err)
		}
		return fmt.Errorf("%w: %s", FindFailedError, err)
	}
	return nil
}
