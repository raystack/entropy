package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/odpf/entropy/store"

	"github.com/odpf/entropy/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProviderRepository struct {
	collection *mongo.Collection
}

func NewProviderRepository(collection *mongo.Collection) *ProviderRepository {
	return &ProviderRepository{
		collection: collection,
	}
}

func (rc *ProviderRepository) Migrate() error {
	return createUniqueIndex(rc.collection, "urn", 1)
}

func (rc *ProviderRepository) Create(Provider *domain.Provider) error {
	Provider.Urn = domain.GenerateProviderUrn(Provider)
	Provider.CreatedAt = time.Now()
	Provider.UpdatedAt = time.Now()

	_, err := rc.collection.InsertOne(context.TODO(), Provider)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%w: %s", store.ProviderAlreadyExistsError, err)
		}
		return err
	}
	return nil
}

func (rc *ProviderRepository) GetConfigByURN(urn string) (map[string]interface{}, error) {
	res := &domain.Provider{}
	err := rc.collection.FindOne(context.TODO(), map[string]interface{}{"urn": urn}).Decode(res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: %s", store.ProviderNotFoundError, err)
		}
		return nil, err
	}
	return res.Configs, nil
}
