package module

import (
	"context"

	"github.com/odpf/entropy/core/resource"
)

type Service struct {
	moduleRepository Repository
}

func NewService(moduleRepository Repository) *Service {
	return &Service{
		moduleRepository: moduleRepository,
	}
}

func (s *Service) Sync(ctx context.Context, r resource.Resource) (*resource.Resource, error) {
	module, err := s.moduleRepository.Get(r.Kind)
	if err != nil {
		r.Status = resource.StatusError
		return &r, err
	}
	status, err := module.Apply(r)
	r.Status = status
	return &r, err
}

func (s *Service) Validate(ctx context.Context, r resource.Resource) error {
	module, err := s.moduleRepository.Get(r.Kind)
	if err != nil {
		return err
	}
	err = module.Validate(r)
	return err
}

func (s *Service) Act(ctx context.Context, r resource.Resource, action string, params map[string]interface{}) (map[string]interface{}, error) {
	module, err := s.moduleRepository.Get(r.Kind)
	if err != nil {
		return nil, err
	}
	output, err := module.Act(r, action, params)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *Service) Log(ctx context.Context, r resource.Resource, filter map[string]string) (<-chan LogChunk, error) {
	module, err := s.moduleRepository.Get(r.Kind)
	if err != nil {
		return nil, err
	}

	moduleLogStream, supported := module.(Loggable)
	if !supported {
		return nil, ErrLogStreamingUnsupported
	}

	logOutput, err := moduleLogStream.Log(ctx, r, filter)
	if err != nil {
		return nil, err
	}

	return logOutput, nil
}
