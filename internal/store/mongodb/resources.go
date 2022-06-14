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

var _ resource.Store = (*ResourceStore)(nil)

type ResourceStore struct{ coll *mongo.Collection }

func NewResourceStore(db *mongo.Database) *ResourceStore {
	return &ResourceStore{
		coll: db.Collection(resourceRepoName),
	}
}

func (rc *ResourceStore) GetByURN(ctx context.Context, urn string) (*resource.Resource, error) {
	var rm resourceModel

	filter := map[string]interface{}{"urn": urn}
	if err := rc.coll.FindOne(ctx, filter).Decode(&rm); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}

	return modelToResource(rm), nil
}

func (rc *ResourceStore) List(ctx context.Context, filter resource.Filter) ([]resource.Resource, error) {
	fq := map[string]string{}

	if filter.Kind != "" {
		fq["kind"] = filter.Kind
	}

	if filter.Project != "" {
		fq["project"] = filter.Project
	}

	cur, err := rc.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var records []resourceModel
	if err = cur.All(ctx, &records); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	var res []resource.Resource
	for _, rec := range records {
		r := modelToResource(rec)
		res = append(res, *r)
	}
	return res, nil
}

func (rc *ResourceStore) Create(ctx context.Context, res resource.Resource, _ ...resource.MutationHook) error {
	res.CreatedAt = time.Now()
	res.UpdatedAt = time.Now()

	_, err := rc.coll.InsertOne(ctx, modelFromResource(res))
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.ErrConflict
		}
		return err
	}

	return nil
}

func (rc *ResourceStore) Update(ctx context.Context, res resource.Resource, _ ...resource.MutationHook) error {
	res.UpdatedAt = time.Now()

	filter := map[string]interface{}{"urn": res.URN}
	updates := map[string]interface{}{"$set": modelFromResource(res)}

	_, err := rc.coll.UpdateOne(ctx, filter, updates)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.ErrNotFound
		}
		return err
	}

	return nil
}

func (rc *ResourceStore) Delete(ctx context.Context, urn string, _ ...resource.MutationHook) error {
	_, err := rc.coll.DeleteOne(ctx, map[string]interface{}{"urn": urn})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.ErrNotFound
		}
		return err
	}
	return nil
}

func (rc *ResourceStore) Migrate(ctx context.Context) error {
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
