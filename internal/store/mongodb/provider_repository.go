package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/odpf/entropy/core/provider"
)

const providerRepoName = "providers"

type ProviderRepository struct {
	collection *mongo.Collection
}

func NewProviderRepository(db *mongo.Database) *ProviderRepository {
	return &ProviderRepository{
		collection: db.Collection(providerRepoName),
	}
}

func (rc *ProviderRepository) Migrate(ctx context.Context) error {
	return createUniqueIndex(ctx, rc.collection, "urn", 1)
}

func (rc *ProviderRepository) Create(ctx context.Context, pro provider.Provider) error {
	pro.URN = provider.GenerateURN(pro)
	pro.CreatedAt = time.Now()
	pro.UpdatedAt = time.Now()

	_, err := rc.collection.InsertOne(ctx, pro)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: %s", provider.ErrProviderAlreadyExists, err)
		}
		return err
	}
	return nil
}

func (rc *ProviderRepository) GetByURN(ctx context.Context, urn string) (*provider.Provider, error) {
	pro := &provider.Provider{}
	err := rc.collection.FindOne(ctx, map[string]interface{}{"urn": urn}).Decode(pro)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: %s", provider.ErrProviderNotFound, err)
		}
		return nil, err
	}
	return pro, nil
}

func (rc *ProviderRepository) List(ctx context.Context, filter map[string]string) ([]*provider.Provider, error) {
	var pro []*provider.Provider
	cur, err := rc.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	err = cur.All(ctx, &pro)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return pro, nil
		}
		return nil, err
	}
	return pro, nil
}
