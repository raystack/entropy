package modules

import (
	"context"
	"reflect"
	"sync"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/pkg/errors"
)

// Registry maintains a list of supported/enabled modules.
type Registry struct {
	mu      sync.RWMutex
	modules map[string]module.Descriptor
}

func (mr *Registry) GetDriver(_ context.Context, mod module.Module) (module.Driver, module.Descriptor, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	desc, found := mr.modules[mod.Name]
	if !found {
		return nil, module.Descriptor{}, errors.ErrNotFound
	}

	driver, err := desc.DriverFactory(mod.Configs)
	if err != nil {
		return nil, module.Descriptor{}, errors.ErrInvalid.
			WithMsgf("failed to initialise module").
			WithCausef(err.Error())
	}

	return driver, desc, nil
}

// Register adds a module to the registry.
func (mr *Registry) Register(desc module.Descriptor) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	if mr.modules == nil {
		mr.modules = map[string]module.Descriptor{}
	}

	if v, exists := mr.modules[desc.Kind]; exists {
		return errors.ErrConflict.
			WithMsgf("module '%s' is already registered for kind '%s'", reflect.TypeOf(v), desc.Kind)
	}

	for i, action := range desc.Actions {
		if err := action.Sanitise(); err != nil {
			return err
		}
		desc.Actions[i] = action
	}
	mr.modules[desc.Kind] = desc
	return nil
}
