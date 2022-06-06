package core

import (
	"context"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func (s *Service) GetResource(ctx context.Context, urn string) (*resource.Resource, error) {
	res, err := s.store.GetByURN(ctx, urn)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, errors.ErrNotFound.WithMsgf("resource with urn '%s' not found", urn)
		}
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	return res, nil
}

func (s *Service) ListResources(ctx context.Context, filter resource.Filter) ([]resource.Resource, error) {
	resources, err := s.store.List(ctx, filter)
	if err != nil {
		return nil, errors.ErrInternal.WithCausef(err.Error())
	}
	return filter.Apply(resources), nil
}

func (s *Service) GetLog(ctx context.Context, urn string, filter map[string]string) (<-chan module.LogChunk, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	moduleLogStream, supported := s.rootModule.(module.Loggable)
	if !supported {
		return nil, errors.ErrUnsupported.WithMsgf("log streaming not supported for kind '%s'", res.Kind)
	}

	modSpec, err := s.generateModuleSpec(ctx, *res)
	if err != nil {
		return nil, err
	}

	return moduleLogStream.Log(ctx, *modSpec, filter)
}
