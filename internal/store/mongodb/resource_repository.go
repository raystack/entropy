package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

const resourceRepoName = "resources"

func NewResourceRepository(db *mongo.Database) *ResourceRepository {
	return &ResourceRepository{
		coll: db.Collection(resourceRepoName),
	}
}

type ResourceRepository struct{ coll *mongo.Collection }

func (rc *ResourceRepository) GetByURN(ctx context.Context, urn string) (*resource.Resource, error) {
	res := &resource.Resource{}
	err := rc.coll.FindOne(ctx, map[string]interface{}{"urn": urn}).Decode(res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return res, nil
}

func (rc *ResourceRepository) List(ctx context.Context, filter map[string]string) ([]*resource.Resource, error) {
	var res []*resource.Resource
	cur, err := rc.coll.Find(ctx, filter)
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

func (rc *ResourceRepository) Create(ctx context.Context, res resource.Resource) error {
	res.CreatedAt = time.Now()
	res.UpdatedAt = time.Now()

	_, err := rc.coll.InsertOne(ctx, res)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.ErrConflict
		}
		return err
	}

	return nil
}

func (rc *ResourceRepository) Update(ctx context.Context, r resource.Resource) error {
	r.UpdatedAt = time.Now()

	filter := map[string]interface{}{"urn": r.URN}
	updates := map[string]interface{}{"$set": r}

	_, err := rc.coll.UpdateOne(ctx, filter, updates)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrNotFound
		}
		return err
	}

	return nil
}

func (rc *ResourceRepository) Delete(ctx context.Context, urn string) error {
	_, err := rc.coll.DeleteOne(ctx, map[string]interface{}{"urn": urn})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrNotFound
		}
		return err
	}
	return nil
}

func (rc *ResourceRepository) DoPending(ctx context.Context, fn resource.PendingHandler) error {
	var res resource.Resource

	filter := map[string]interface{}{"state.status": resource.StatusPending}
	if err := rc.coll.FindOne(ctx, filter).Decode(res); err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrNotFound // no pending item is available.
		}
		return err
	}

	modified, shouldDelete, err := fn(ctx, res)
	if err != nil {
		return err
	}

	if shouldDelete {
		_, err := rc.coll.DeleteOne(ctx, map[string]interface{}{"urn": res.URN})
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}
		return nil
	}

	return rc.Update(ctx, *modified)
}

func (rc *ResourceRepository) Migrate(ctx context.Context) error {
	return createUniqueIndex(ctx, rc.coll, "urn", 1)
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
