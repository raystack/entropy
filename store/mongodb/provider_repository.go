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
			return fmt.Errorf("%w: %s", store.ErrProviderAlreadyExists, err)
		}
		return err
	}
	return nil
}

func (rc *ProviderRepository) GetByURN(urn string) (*domain.Provider, error) {
	pro := &domain.Provider{}
	err := rc.collection.FindOne(context.TODO(), map[string]interface{}{"urn": urn}).Decode(pro)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("%w: %s", store.ErrProviderNotFound, err)
		}
		return nil, err
	}
	return pro, nil
}

func (rc *ProviderRepository) List(filter map[string]string) ([]*domain.Provider, error) {
	var pro []*domain.Provider
	cur, err := rc.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	err = cur.All(context.TODO(), &pro)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return pro, nil
		}
		return nil, err
	}
	return pro, nil
}
