package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/pkg/worker"
)

const (
	JobKindSyncResource  = "sync_resource"
	syncJobType          = "sync"
	scheduledSyncJobType = "sched-sync"
)

type syncJobPayload struct {
	ResourceURN string    `json:"resource_urn"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Service) enqueueSyncJob(ctx context.Context, res resource.Resource, runAt time.Time, jobType string) error {
	data := syncJobPayload{
		ResourceURN: res.URN,
		UpdatedAt:   res.UpdatedAt,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return s.worker.Enqueue(ctx, worker.Job{
		ID:      fmt.Sprintf(jobType+"-%s-%d", res.URN, res.UpdatedAt.Unix()),
		Kind:    JobKindSyncResource,
		RunAt:   runAt,
		Payload: payload,
	})
}

// HandleSyncJob is meant to be invoked by asyncWorker when an enqueued job is
// ready.
// TODO: make this private and move the registration of this handler inside New().
func (s *Service) HandleSyncJob(ctx context.Context, job worker.Job) ([]byte, error) {
	const retryBackoff = 5 * time.Second

	var data syncJobPayload
	if err := json.Unmarshal(job.Payload, &data); err != nil {
		return nil, err
	}

	syncedRes, err := s.syncChange(ctx, data.ResourceURN)
	if err != nil {
		if errors.Is(err, errors.ErrInternal) {
			return nil, &worker.RetryableError{
				Cause:      errors.Verbose(err),
				RetryAfter: retryBackoff,
			}
		}

		return nil, errors.Verbose(err)
	}

	return json.Marshal(map[string]interface{}{
		"status": syncedRes.State.Status,
	})
}

func (s *Service) syncChange(ctx context.Context, urn string) (*resource.Resource, error) {
	res, err := s.GetResource(ctx, urn)
	if err != nil {
		return nil, err
	}

	modSpec, err := s.generateModuleSpec(ctx, *res)
	if err != nil {
		return nil, err
	}

	oldState := res.State.Clone()
	newState, err := s.rootModule.Sync(ctx, *modSpec)
	if err != nil {
		if errors.Is(err, errors.ErrInvalid) {
			return nil, err
		}
		return nil, errors.ErrInternal.WithMsgf("sync() failed").WithCausef(err.Error())
	}
	res.UpdatedAt = s.clock()
	res.State = *newState

	// TODO: clarify on behaviour when resource schedule for deletion reaches error.
	shouldDelete := oldState.InDeletion() && newState.IsTerminal()
	if shouldDelete {
		if err := s.DeleteResource(ctx, urn); err != nil {
			return nil, err
		}
	} else {
		if err := s.upsert(ctx, module.Plan{Resource: res}, false); err != nil {
			return nil, err
		}
	}

	return res, nil
}
