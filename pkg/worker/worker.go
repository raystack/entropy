package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/odpf/entropy/pkg/errors"
)

// Worker provides asynchronous job processing using a job-queue.
type Worker struct {
	workers int
	pollInt time.Duration

	queue    JobQueue
	logger   *zap.Logger
	handlers map[string]JobFn
}

type Option func(w *Worker) error

func New(queue JobQueue, opts ...Option) (*Worker, error) {
	w := &Worker{queue: queue}
	for _, opt := range withDefaults(opts) {
		if err := opt(w); err != nil {
			return nil, err
		}
	}

	if len(w.handlers) == 0 {
		return nil, errors.New("at-least one job handler must be registered")
	}
	return w, nil
}

// Enqueue enqueues all jobs for processing.
func (w *Worker) Enqueue(ctx context.Context, jobs ...Job) error {
	for i, job := range jobs {
		if err := job.sanitise(); err != nil {
			return err
		} else if _, knownKind := w.handlers[job.Kind]; !knownKind {
			return fmt.Errorf("%w: kind '%s'", ErrUnknownKind, job.Kind)
		}
		jobs[i] = job
	}

	return w.queue.Enqueue(ctx, jobs...)
}

// Run starts the worker threads that dequeue and process ready jobs. Run blocks
// until all workers exit or context is cancelled. Context cancellation will do
// graceful shutdown of the worker threads.
func (w *Worker) Run(baseCtx context.Context) error {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	wg := &sync.WaitGroup{}
	for i := 0; i < w.workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			if err := w.runWorker(ctx); err != nil && !errors.Is(err, context.Canceled) {
				w.logger.Error("worker exited with error",
					zap.Error(err),
					zap.Int("worker_id", id),
				)
			}
		}(i)
	}
	wg.Wait()

	return cleanupCtxErr(ctx.Err())
}

func (w *Worker) runWorker(ctx context.Context) error {
	timer := time.NewTimer(w.pollInt)
	defer timer.Stop()

	var kinds []string
	for kind := range w.handlers {
		kinds = append(kinds, kind)
	}

	for {
		select {
		case <-ctx.Done():
			return cleanupCtxErr(ctx.Err())

		case <-timer.C:
			timer.Reset(w.pollInt)

			w.logger.Info("looking for a job")
			if err := w.queue.Dequeue(ctx, kinds, w.handleJob); err != nil {
				w.logger.Error("dequeue failed", zap.Error(err))
			}
		}
	}
}

func (w *Worker) handleJob(ctx context.Context, job Job) ([]byte, error) {
	const invalidKindBackoff = 5 * time.Minute

	fn, exists := w.handlers[job.Kind]
	if !exists {
		// Note: This should never happen since Dequeue() has `kinds` filter.
		//       It is only kept as a safety net to prevent nil-dereferences.
		return nil, &RunError{
			JobID:      job.ID,
			Cause:      errors.New("job kind is invalid"),
			RetryAfter: invalidKindBackoff,
		}
	}

	return fn(ctx, job)
}

func cleanupCtxErr(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
