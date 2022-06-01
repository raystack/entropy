package pgq

import (
	"context"

	"github.com/odpf/entropy/pkg/worker"

	sq "github.com/Masterminds/squirrel"
)

func (q *Queue) Enqueue(ctx context.Context, jobs ...worker.Job) error {
	var insertQuery = sq.Insert(q.table).Columns(
		"id", "kind", "status", "run_at",
		"payload", "created_at", "updated_at",
	)

	for _, job := range jobs {
		insertQuery = insertQuery.Values(
			job.ID, job.Kind, job.Status, job.RunAt,
			job.Payload, job.CreatedAt, job.UpdatedAt,
		)
	}

	_, err := insertQuery.RunWith(q.db).PlaceholderFormat(sq.Dollar).ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}
