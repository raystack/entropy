package resource

import (
	"errors"
	"fmt"
	"time"

	"github.com/odpf/entropy/domain/model"
	"github.com/odpf/entropy/pkg/store"
)

const RepositoryName = "resources"

type Repository struct {
	DB *store.DB
}

var (
	ResourceAlreadyExistsError = errors.New("resource already exists")
	NoResourceFoundError       = errors.New("no resource(s) found")
)

func (rc *Repository) Migrate() error {
	coll := rc.DB.GetCollection(RepositoryName)
	return coll.CreateUniqueIndex("urn", 1)
}

func (rc *Repository) Create(r *model.Resource) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	coll := rc.DB.GetCollection(RepositoryName)
	err := coll.InsertOne(r)
	if err != nil && errors.Is(err, store.AlreadyExistsError) {
		return fmt.Errorf("%w: URN = %s", ResourceAlreadyExistsError, r.Urn)
	}
	return err
}

func (rc *Repository) Update(r *model.Resource) error {
	r.UpdatedAt = time.Now()
	coll := rc.DB.GetCollection(RepositoryName)
	err := coll.UpdateOne(map[string]interface{}{"urn": r.Urn}, r)
	if err != nil {
		return err
	}
	return nil
}

func (rc *Repository) GetByURN(urn string) (*model.Resource, error) {
	coll := rc.DB.GetCollection(RepositoryName)
	res := model.Resource{}
	err := coll.FindOne(map[string]interface{}{"urn": urn}, &res)
	if err != nil {
		if errors.Is(err, store.NotFoundError) {
			return nil, fmt.Errorf("%w: URN = %s", NoResourceFoundError, urn)
		}
		return nil, err
	}
	return &res, nil
}

func (rc *Repository) Get(kind string, parent string) ([]*model.Resource, error) {
	var res []*model.Resource
	coll := rc.DB.GetCollection(RepositoryName)
	err := coll.Find(map[string]interface{}{"kind": kind, "parent": parent}, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (rc *Repository) Delete(urn string) error { return nil }
