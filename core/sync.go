package core

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

// RunSync runs a loop for processing resources in pending state. It runs until
// context is cancelled.
func (s *Service) RunSync(ctx context.Context) error {
	const pollInterval = 500 * time.Millisecond

	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-pollTicker.C:
			// TODO: handle repeated failure scenarios?
			if err := s.repository.DoPending(ctx, s.syncChange); err != nil {
				s.logger.Error("failed to handle pending item", zap.Error(err))
			}
		}
	}
}

func (s *Service) syncChange(ctx context.Context,
	res resource.Resource) (updated *resource.Resource, delete bool, err error) {

	modSpec, err := s.generateModuleSpec(ctx, res)
	if err != nil {
		return nil, false, err
	}

	oldState := res.State.Clone()
	newState, err := s.rootModule.Sync(ctx, *modSpec)
	if err != nil {
		if errors.Is(err, errors.ErrInvalid) {
			return nil, false, err
		}
		return nil, false, errors.ErrInternal.WithMsgf("sync() failed").WithCausef(err.Error())
	}

	// TODO: clarify on behaviour when resource schedule for deletion reaches error.
	shouldDelete := oldState.InDeletion() && newState.IsTerminal()

	res.UpdatedAt = time.Now()
	res.State = *newState
	return &res, shouldDelete, nil
}
