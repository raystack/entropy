package core

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
)

// RunSyncer runs the syncer thread that keeps performing resource-sync at
// regular intervals.
func (svc *Service) RunSyncer(ctx context.Context, interval time.Duration) error {
	tick := time.NewTimer(interval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-tick.C:
			tick.Reset(interval)

			err := svc.store.SyncOne(ctx, svc.handleSync)
			if err != nil {
				svc.logger.Warn("SyncOne() failed", zap.Error(err))
			}
		}
	}
}

func (svc *Service) handleSync(ctx context.Context, res resource.Resource) (*resource.Resource, error) {
	logEntry := svc.logger.With(
		zap.String("resource_urn", res.URN),
		zap.String("resource_status", res.State.Status),
		zap.Int("retries", res.State.SyncResult.Retries),
		zap.String("last_err", res.State.SyncResult.LastError),
	)

	modSpec, err := svc.generateModuleSpec(ctx, res)
	if err != nil {
		logEntry.Error("SyncOne() failed", zap.Error(err))
		return nil, err
	}

	newState, err := svc.moduleSvc.SyncState(ctx, *modSpec)
	if err != nil {
		logEntry.Error("SyncOne() failed", zap.Error(err))

		res.State.SyncResult.LastError = err.Error()
		res.State.SyncResult.Retries++
		if errors.Is(err, errors.ErrInvalid) {
			// ErrInvalid is expected to be returned when config is invalid.
			// There is no point in retrying in this case.
			res.State.Status = resource.StatusError
			res.State.NextSyncAt = nil
		} else if svc.maxSyncRetries > 0 && res.State.SyncResult.Retries >= svc.maxSyncRetries {
			// Some other error occurred and no more retries remaining.
			// move the resource to failure state.
			res.State.Status = resource.StatusError
			res.State.NextSyncAt = nil
		} else {
			// Some other error occurred and we still have remaining retries.
			// need to backoff and retry in some time.
			tryAgainAt := svc.clock().Add(svc.syncBackoff)
			res.State.NextSyncAt = &tryAgainAt
		}
	} else {
		res.State.SyncResult.Retries = 0
		res.State.SyncResult.LastError = ""
		res.UpdatedAt = svc.clock()
		res.State = *newState

		logEntry.Info("SyncOne() finished",
			zap.String("final_status", res.State.Status),
			zap.Timep("next_sync", res.State.NextSyncAt),
		)
	}

	return &res, nil
}
