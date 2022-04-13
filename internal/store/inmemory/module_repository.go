package inmemory

import (
	"github.com/odpf/entropy/core/module"
)

type ModuleRepository struct {
	collection map[string]module.Module
}

func NewModuleRepository() *ModuleRepository {
	return &ModuleRepository{
		collection: map[string]module.Module{},
	}
}

func (mr *ModuleRepository) Get(id string) (module.Module, error) {
	if m, exists := mr.collection[id]; exists {
		return m, nil
	}
	return nil, module.ErrModuleNotFound
}

func (mr *ModuleRepository) Register(m module.Module) error {
	if _, exists := mr.collection[m.ID()]; exists {
		return module.ErrModuleAlreadyExists
	}
	mr.collection[m.ID()] = m
	return nil
}
