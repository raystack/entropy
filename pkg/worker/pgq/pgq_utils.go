package pgq

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/goto/entropy/pkg/worker"
)

type txnFn func(ctx context.Context, tx *sql.Tx) error

func (q *Queue) withTx(ctx context.Context, readOnly bool, fn txnFn) error {
	opts := &sql.TxOptions{ReadOnly: readOnly}

	tx, err := q.db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	if fnErr := fn(ctx, tx); fnErr != nil {
		_ = tx.Rollback()
		return fnErr
	}

	return tx.Commit()
}

func (q *Queue) runHeartbeat(ctx context.Context, id string) {
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

	txErr := q.withTx(ctx, false, func(ctx context.Context, tx *sql.Tx) error {
		j, err := fetchReadyJob(ctx, tx, q.tableName, kinds)
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
	updateQuery := sq.Update(q.tableName).
		Where(sq.Eq{
			"id":     job.ID,
			"status": worker.StatusPending,
		}).
		Set("updated_at", job.UpdatedAt.UTC()).
		Set("status", job.Status).
		Set("result", job.Result).
		Set("attempts_done", sq.Expr("attempts_done + 1")).
		Set("last_error", job.LastError).
		Set("last_attempt_at", job.LastAttemptAt.UTC())

	_, err := updateQuery.PlaceholderFormat(sq.Dollar).RunWith(q.db).ExecContext(ctx)
	return err
}

func fetchReadyJob(ctx context.Context, r sq.BaseRunner, tableName string, kinds []string) (*worker.Job, error) {
	selectQuery := sq.Select().From(tableName).
		Columns("id", "kind", "status", "run_at",
			"payload", "created_at", "updated_at",
			"result", "attempts_done", "last_attempt_at", "last_error").
		Where(sq.Eq{
			"kind":   kinds,
			"status": worker.StatusPending,
		}).
		Where(sq.Expr("run_at < current_timestamp")).
		Limit(1).
		Suffix("FOR UPDATE SKIP LOCKED")

	row := selectQuery.PlaceholderFormat(sq.Dollar).RunWith(r).QueryRowContext(ctx)
	return rowIntoJob(row)
}

func (q *Queue) extendWaitTime(ctx context.Context, r sq.BaseRunner, id string) error {
	extendTo := sq.Expr("current_timestamp + (? ||' seconds')::interval ", q.extendInterval.Seconds())
	extendQuery := sq.Update(q.tableName).
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
