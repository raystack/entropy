package inmemory

import (
	"github.com/odpf/entropy/core/resource"
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
	return nil, resource.ErrModuleNotFound
}

func (mr *ModuleRepository) Register(m resource.Module) error {
	if _, exists := mr.collection[m.ID()]; exists {
		return resource.ErrModuleAlreadyExists
	}
	mr.collection[m.ID()] = m
	return nil
}
