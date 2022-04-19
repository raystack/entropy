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

const resourceRepoName = "resources"

type ResourceRepository struct {
	collection *mongo.Collection
}

func NewResourceRepository(db *mongo.Database) *ResourceRepository {
	return &ResourceRepository{
		collection: db.Collection(resourceRepoName),
	}
}

func (rc *ResourceRepository) Migrate(ctx context.Context) error {
	return createUniqueIndex(ctx, rc.collection, "urn", 1)
}

func (rc *ResourceRepository) Create(ctx context.Context, res resource.Resource) error {
	res.CreatedAt = time.Now()
	res.UpdatedAt = time.Now()

	_, err := rc.collection.InsertOne(ctx, res)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: %s", resource.ErrResourceAlreadyExists, err)
		}
		return err
	}
	return nil
}

func (rc *ResourceRepository) Update(ctx context.Context, r resource.Resource) error {
	r.UpdatedAt = time.Now()
	singleResult := rc.collection.FindOneAndUpdate(ctx,
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

func (rc *ResourceRepository) GetByURN(ctx context.Context, urn string) (*resource.Resource, error) {
	res := &resource.Resource{}
	err := rc.collection.FindOne(ctx, map[string]interface{}{"urn": urn}).Decode(res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: %s", resource.ErrResourceNotFound, err)
		}
		return nil, err
	}
	return res, nil
}

func (rc *ResourceRepository) List(ctx context.Context, filter map[string]string) ([]*resource.Resource, error) {
	var res []*resource.Resource
	cur, err := rc.collection.Find(ctx, filter)
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

func (rc *ResourceRepository) Delete(ctx context.Context, urn string) error {
	_, err := rc.collection.DeleteOne(ctx, map[string]interface{}{"urn": urn})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("%w: %s", resource.ErrResourceNotFound, err)
		}
		return err
	}
	return nil
}

func createUniqueIndex(ctx context.Context, collection *mongo.Collection, key string, order int) error {
	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: key, Value: order}},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}
