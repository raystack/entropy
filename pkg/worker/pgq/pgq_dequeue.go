package pgq

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/pkg/worker"
)

func (q *Queue) Dequeue(baseCtx context.Context, kinds []string, fn worker.JobFn) error {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	job, err := q.pickupJob(ctx, kinds)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	// Heartbeat goroutine: Keeps extending the ready_at timestamp
	// until the job-handler is running to make sure no other worker
	// picks up the same job.
	go q.runHeartbeat(ctx, cancel, job.ID)
	job.Attempt(ctx, time.Now(), fn)
	return q.saveJobResult(ctx, *job)
}

func (q *Queue) runHeartbeat(ctx context.Context, cancel context.CancelFunc, id string) {
	defer cancel()

	tick := time.NewTicker(q.refreshInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-tick.C:
			if err := q.extendWaitTime(ctx, q.db, id); err != nil {
				return
			}
		}
	}
}

func (q *Queue) pickupJob(ctx context.Context, kinds []string) (*worker.Job, error) {
	var job worker.Job

	txErr := q.withTxn(ctx, false, func(ctx context.Context, tx *sql.Tx) error {
		j, err := q.fetchReadyJob(ctx, tx, kinds)
		if err != nil {
			return err
		} else if extendErr := q.extendWaitTime(ctx, tx, j.ID); extendErr != nil {
			return extendErr
		}
		job = *j

		return nil
	})

	return &job, txErr
}

func (q *Queue) saveJobResult(ctx context.Context, job worker.Job) error {
	updateQuery := sq.Update(q.table).
		Where(sq.Eq{
			"id":     job.ID,
			"status": worker.StatusPending,
		}).
		Set("updated_at", job.UpdatedAt).
		Set("status", job.Status).
		Set("result", job.Result).
		Set("attempts_done", sq.Expr("attempts_done + 1")).
		Set("last_error", job.LastError).
		Set("last_attempt_at", job.LastAttemptAt)

	_, err := updateQuery.PlaceholderFormat(sq.Dollar).RunWith(q.db).ExecContext(ctx)
	return err
}

func (q *Queue) fetchReadyJob(ctx context.Context, r sq.BaseRunner, kinds []string) (*worker.Job, error) {
	selectQuery := sq.Select().From(q.table).
		Columns(
			"id", "kind", "status", "run_at",
			"payload", "created_at", "updated_at",
			"result", "attempts_done", "last_attempt_at", "last_error",
		).
		Where(sq.Eq{
			"kind":   kinds,
			"status": worker.StatusPending,
		}).
		Where(sq.Lt{"run_at": time.Now()}).
		Limit(1).
		Suffix("FOR UPDATE SKIP LOCKED")

	row := selectQuery.PlaceholderFormat(sq.Dollar).RunWith(r).QueryRowContext(ctx)
	return rowIntoJob(row)
}

func (q *Queue) extendWaitTime(ctx context.Context, r sq.BaseRunner, id string) error {
	extendTo := sq.Expr("current_timestamp + (? ||' seconds')::interval ", q.extendInterval.Seconds())
	extendQuery := sq.Update(q.table).
		Set("run_at", extendTo).
		Where(sq.Eq{"id": id})

	_, err := extendQuery.PlaceholderFormat(sq.Dollar).RunWith(r).ExecContext(ctx)
	return err
}

func rowIntoJob(row sq.RowScanner) (*worker.Job, error) {
	var job worker.Job
	var lastErr sql.NullString
	var lastAttemptAt sql.NullTime

	fieldPtrs := []interface{}{
		&job.ID, &job.Kind, &job.Status, &job.RunAt,
		&job.Payload, &job.CreatedAt, &job.UpdatedAt,

		// execution results.
		&job.Result, &job.AttemptsDone, &lastAttemptAt, &lastErr,
	}

	if err := row.Scan(fieldPtrs...); err != nil {
		return nil, err
	}

	job.LastAttemptAt = lastAttemptAt.Time
	job.LastError = lastErr.String

	return &job, nil
}
