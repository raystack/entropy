package module

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
)

type Service struct {
	store    Store
	registry Registry
}

func NewService(registry Registry, store Store) *Service {
	return &Service{
		store:    store,
		registry: registry,
	}
}

func (mr *Service) PlanAction(ctx context.Context, res ExpandedResource, act ActionRequest) (*Plan, error) {
	mod, err := mr.discoverModule(ctx, res.Kind, res.Project)
	if err != nil {
		return nil, err
	}

	driver, desc, err := mr.initDriver(ctx, *mod)
	if err != nil {
		return nil, err
	} else if err := desc.validateDependencies(res.Dependencies); err != nil {
		return nil, err
	} else if err := desc.validateActionReq(res, act); err != nil {
		return nil, err
	}

	return driver.Plan(ctx, res, act)
}

func (mr *Service) SyncState(ctx context.Context, res ExpandedResource) (*resource.State, error) {
	mod, err := mr.discoverModule(ctx, res.Kind, res.Project)
	if err != nil {
		return nil, err
	}

	driver, desc, err := mr.initDriver(ctx, *mod)
	if err != nil {
		return nil, err
	} else if err := desc.validateDependencies(res.Dependencies); err != nil {
		return nil, err
	}

	return driver.Sync(ctx, res)
}

func (mr *Service) StreamLogs(ctx context.Context, res ExpandedResource, filter map[string]string) (<-chan LogChunk, error) {
	mod, err := mr.discoverModule(ctx, res.Kind, res.Project)
	if err != nil {
		return nil, err
	}

	driver, _, err := mr.initDriver(ctx, *mod)
	if err != nil {
		return nil, err
	}

	lg, supported := driver.(Loggable)
	if !supported {
		return nil, errors.ErrUnsupported.WithMsgf("log streaming not supported for kind '%s'", res.Kind)
	}

	return lg.Log(ctx, res, filter)
}

func (mr *Service) GetOutput(ctx context.Context, res ExpandedResource) (json.RawMessage, error) {
	mod, err := mr.discoverModule(ctx, res.Kind, res.Project)
	if err != nil {
		return nil, err
	}

	driver, _, err := mr.initDriver(ctx, *mod)
	if err != nil {
		return nil, err
	}

	return driver.Output(ctx, res)
}

func (mr *Service) GetModule(ctx context.Context, urn string) (*Module, error) {
	return mr.store.GetModule(ctx, urn)
}

func (mr *Service) ListModules(ctx context.Context, project string) ([]Module, error) {
	return mr.store.ListModules(ctx, project)
}

func (mr *Service) CreateModule(ctx context.Context, mod Module) (*Module, error) {
	if err := mod.sanitise(true); err != nil {
		return nil, err
	}

	if _, _, err := mr.registry.GetDriver(ctx, mod); err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, errors.ErrInvalid.WithMsgf("driver not found for kind '%s'", mod.Name)
		} else if errors.Is(err, errors.ErrInvalid) {
			return nil, errors.ErrInvalid.
				WithMsgf("failed to init driver with given configs").
				WithCausef(err.Error())
		}
		return nil, err
	}

	if err := mr.store.CreateModule(ctx, mod); err != nil {
		if errors.Is(err, errors.ErrConflict) {
			return nil, errors.ErrConflict.
				WithMsgf("module with given name and project already exists").
				WithCausef(err.Error())
		}
		return nil, err
	}
	return &mod, nil
}

func (mr *Service) UpdateModule(ctx context.Context, urn string, newConfigs json.RawMessage) (*Module, error) {
	mod, err := mr.store.GetModule(ctx, urn)
	if err != nil {
		return nil, err
	}
	mod.Configs = newConfigs

	if err := mod.sanitise(false); err != nil {
		return nil, err
	}

	if err := mr.store.UpdateModule(ctx, *mod); err != nil {
		return nil, err
	}
	return mod, nil
}

func (mr *Service) DeleteModule(ctx context.Context, urn string) error {
	return mr.store.DeleteModule(ctx, urn)
}

func (mr *Service) discoverModule(ctx context.Context, kind, project string) (*Module, error) {
	urn := generateURN(kind, project)

	m, err := mr.store.GetModule(ctx, urn)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, errors.ErrInvalid.
				WithMsgf("kind '%s' is not valid in project '%s'", kind, project).
				WithCausef("failed to find module with urn '%s'", urn)
		}
		return nil, err
	}
	return m, nil
}

func (mr *Service) initDriver(ctx context.Context, mod Module) (Driver, Descriptor, error) {
	driver, desc, err := mr.registry.GetDriver(ctx, mod)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, Descriptor{}, errors.ErrInvalid.WithMsgf("driver not found for kind '%s'", mod.Name)
		} else if errors.Is(err, errors.ErrInvalid) {
			return nil, Descriptor{}, errors.ErrInvalid.
				WithMsgf("failed to init driver with given configs").
				WithCausef(err.Error())
		}
		return nil, Descriptor{}, err
	}
	return driver, desc, nil
}

func generateURN(name, project string) string {
	return fmt.Sprintf("orn:entropy:module:%s:%s", project, name)
}
