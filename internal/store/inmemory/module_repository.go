package inmemory

import (
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

type ModuleRepository struct {
	collection map[string]resource.Module
}

func NewModuleRepository() *ModuleRepository {
	return &ModuleRepository{
		collection: map[string]resource.Module{},
	}
}

func (mr *ModuleRepository) Get(id string) (resource.Module, error) {
	if m, exists := mr.collection[id]; exists {
		return m, nil
	}
	return nil, errors.ErrNotFound
}

func (mr *ModuleRepository) Register(m resource.Module) error {
	id := m.ID()
	if _, exists := mr.collection[id]; exists {
		return errors.ErrConflict.WithMsgf("module with id '%s' already exists", id)
	}
	mr.collection[m.ID()] = m
	return nil
}
