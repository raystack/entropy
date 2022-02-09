package inmemory

import (
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/store"
)

type ModuleRepository struct {
	collection map[string]domain.Module
}

func NewModuleRepository() *ModuleRepository {
	return &ModuleRepository{
		collection: map[string]domain.Module{},
	}
}

func (mr *ModuleRepository) Register(module domain.Module) error {
	if _, exists := mr.collection[module.ID()]; exists {
		return store.ModuleAlreadyExistsError
	}
	mr.collection[module.ID()] = module
	return nil
}

func (mr *ModuleRepository) Get(id string) (domain.Module, error) {
	if module, exists := mr.collection[id]; exists {
		return module, nil
	}
	return nil, store.ModuleNotFoundError
}
