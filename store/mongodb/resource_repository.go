package mongodb

import (
	"context"
	"fmt"
	"github.com/odpf/entropy/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"

	"github.com/odpf/entropy/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

var ResourceRepositoryName = "resource"

type ResourceRepository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) *ResourceRepository {
	return &ResourceRepository{
		db: db,
	}
}

func (rc *ResourceRepository) Migrate() error {
	coll := rc.db.Collection(ResourceRepositoryName)
	return createUniqueIndex(coll, "urn", 1)
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

func (rc *ResourceRepository) Create(resource *domain.Resource) error {
	resource.CreatedAt = time.Now()
	resource.UpdatedAt = time.Now()
	coll := rc.db.Collection(ResourceRepositoryName)

	_, err := coll.InsertOne(context.TODO(), resource)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: %s", store.ResourceAlreadyExistsError, err)
		}
		return err
	}
	return nil
}

func (rc *ResourceRepository) Update(r *domain.Resource) error {
	r.UpdatedAt = time.Now()
	coll := rc.db.Collection(ResourceRepositoryName)
	singleResult := coll.FindOneAndUpdate(context.TODO(),
		map[string]interface{}{"urn": r.Urn},
		map[string]interface{}{"$set": r},
	)
	err := singleResult.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("%w: urn = %s", store.ResourceNotFoundError, r.Urn)
		}
		return err
	}
	return nil
}

func (rc *ResourceRepository) GetByURN(urn string) (*domain.Resource, error) {
	coll := rc.db.Collection(ResourceRepositoryName)
	res := &domain.Resource{}
	err := coll.FindOne(context.TODO(), map[string]interface{}{"urn": urn}).Decode(res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: %s", store.ResourceNotFoundError, err)
		}
		return nil, err
	}
	return res, nil
}
