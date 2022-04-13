package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/odpf/entropy/core/resource"
)

type ResourceRepository struct {
	collection *mongo.Collection
}

func NewResourceRepository(collection *mongo.Collection) *ResourceRepository {
	return &ResourceRepository{
		collection: collection,
	}
}

func (rc *ResourceRepository) Migrate() error {
	return createUniqueIndex(rc.collection, "urn", 1)
}

func createUniqueIndex(collection *mongo.Collection, key string, order int) error {
	_, err := collection.Indexes().CreateOne(
		context.TODO(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: key, Value: order}},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

func (rc *ResourceRepository) Create(res *resource.Resource) error {
	res.CreatedAt = time.Now()
	res.UpdatedAt = time.Now()

	_, err := rc.collection.InsertOne(context.TODO(), res)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: %s", resource.ErrResourceAlreadyExists, err)
		}
		return err
	}
	return nil
}

func (rc *ResourceRepository) Update(r *resource.Resource) error {
	r.UpdatedAt = time.Now()
	singleResult := rc.collection.FindOneAndUpdate(context.TODO(),
		map[string]interface{}{"urn": r.URN},
		map[string]interface{}{"$set": r},
	)
	err := singleResult.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("%w: urn = %s", resource.ErrResourceNotFound, r.URN)
		}
		return err
	}
	return nil
}

func (rc *ResourceRepository) GetByURN(urn string) (*resource.Resource, error) {
	res := &resource.Resource{}
	err := rc.collection.FindOne(context.TODO(), map[string]interface{}{"urn": urn}).Decode(res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: %s", resource.ErrResourceNotFound, err)
		}
		return nil, err
	}
	return res, nil
}

func (rc *ResourceRepository) List(filter map[string]string) ([]*resource.Resource, error) {
	var res []*resource.Resource
	cur, err := rc.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	err = cur.All(context.TODO(), &res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return res, nil
		}
		return nil, err
	}
	return res, nil
}

func (rc *ResourceRepository) Delete(urn string) error {
	_, err := rc.collection.DeleteOne(context.TODO(), map[string]interface{}{"urn": urn})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("%w: %s", resource.ErrResourceNotFound, err)
		}
		return err
	}
	return nil
}
