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
	const (
		backOff      = 2
		pollInterval = 500 * time.Millisecond
	)

	pollTicker := time.NewTimer(pollInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-pollTicker.C:
			err := s.repository.DoPending(ctx, s.syncChange)
			if err != nil {
				if errors.Is(err, errors.ErrNotFound) {
					// backOff to reduce polling pressure.
					pollTicker.Reset(backOff * pollInterval)
				} else {
					e := errors.E(err)
					s.logger.Error("failed to handle pending item",
						zap.Error(e),
						zap.String("cause", e.Cause),
					)
					return e
				}
			} else {
				pollTicker.Reset(pollInterval)
			}
		}
	}
}

func (s *Service) syncChange(ctx context.Context, res resource.Resource) (*resource.Resource, bool, error) {
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

	res.UpdatedAt = s.clock()
	res.State = *newState
	return &res, shouldDelete, nil
}
