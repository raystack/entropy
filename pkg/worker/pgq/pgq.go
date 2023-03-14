package pgq

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/worker"
)

const (
	pgDriverName         = "postgres"
	tableNamePlaceholder = "__queueTable__"

	extendInterval  = 30 * time.Second
	refreshInterval = 20 * time.Second
)

var (
	//go:embed schema.sql
	sqlSchemaTemplate string

	queueNamePattern    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
	errInvalidQueueName = fmt.Errorf("queue name must match pattern '%s'", queueNamePattern)
)

// Queue implements a JobQueue backed by PostgreSQL. Refer Open() for initialising.
type Queue struct {
	db              *sql.DB
	queueName       string
	tableName       string
	extendInterval  time.Duration
	refreshInterval time.Duration
}

// Open returns a JobQueue implementation backed by the PostgreSQL instance
// discovered by the conString. The table used for the queue will be based
// on the queueName. Necessary migrations will be done automatically.
func Open(conString, queueName string) (*Queue, error) {
	if !queueNamePattern.MatchString(queueName) {
		return nil, errInvalidQueueName
	}
	tableName := fmt.Sprintf("pgq_%s", queueName)

	db, err := sql.Open(pgDriverName, conString)
	if err != nil {
		return nil, err
	}

	q := &Queue{
		db:              db,
		queueName:       queueName,
		tableName:       tableName,
		extendInterval:  extendInterval,
		refreshInterval: refreshInterval,
	}

	if err := q.prepareDB(); err != nil {
		_ = q.Close()
		return nil, err
	}

	return q, nil
}

func (q *Queue) Enqueue(ctx context.Context, jobs ...worker.Job) error {
	insertQuery := sq.Insert(q.tableName).Columns(
		"id", "kind", "status", "run_at",
		"payload", "created_at", "updated_at",
	)

	for _, job := range jobs {
		insertQuery = insertQuery.Values(
			job.ID, job.Kind, job.Status, job.RunAt.UTC(),
			job.Payload, job.CreatedAt.UTC(), job.UpdatedAt.UTC(),
		)
	}

	_, err := insertQuery.RunWith(q.db).PlaceholderFormat(sq.Dollar).ExecContext(ctx)
	if err != nil {
		pqErr := &pq.Error{}
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
			return worker.ErrJobExists
		}
		return err
	}
	return nil
}

func (q *Queue) Dequeue(ctx context.Context, kinds []string, fn worker.DequeueFn) error {
	job, err := q.pickupJob(ctx, kinds)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	resultJob, err := q.handleDequeued(ctx, *job, fn)
	if err != nil {
		return err
	}

	return q.saveJobResult(ctx, *resultJob)
}

func (q *Queue) handleDequeued(baseCtx context.Context, job worker.Job, fn worker.DequeueFn) (*worker.Job, error) {
	jobCtx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	go func() {
		q.runHeartbeat(jobCtx, job.ID)

		// Heartbeat process stopped for some reason. job should be
		// released as soon as possible. so cancel context.
		cancel()
	}()

	return fn(jobCtx, job)
}

func (q *Queue) Close() error { return q.db.Close() }

func (q *Queue) prepareDB() error {
	sqlSchema := strings.ReplaceAll(sqlSchemaTemplate, tableNamePlaceholder, q.tableName)
	_, execErr := q.db.Exec(sqlSchema)
	return execErr
}
