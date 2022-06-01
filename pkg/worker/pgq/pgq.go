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
	_ "github.com/lib/pq" // postgres driver.

	"github.com/odpf/entropy/pkg/errors"
	"github.com/odpf/entropy/pkg/worker"
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

// Queue implements a JobQueue backed by PostgreSQL.
// Refer Open() for initialising.
type Queue struct {
	db              *sql.DB
	name            string
	table           string
	extendInterval  time.Duration
	refreshInterval time.Duration
}

type txnFn func(ctx context.Context, tx *sql.Tx) error

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
		name:            queueName,
		table:           tableName,
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
	insertQuery := sq.Insert(q.table).Columns(
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

	go func() {
		job.Attempt(ctx, time.Now(), fn)
		cancel() // Attempt finished. Stop the heartbeat.
	}()

	// Keep extending the run_at timestamp until the job-handler
	// is running to make sure no other worker picks up the same job.
	q.runHeartbeat(ctx, cancel, job.ID)

	return q.saveJobResult(ctx, *job)
}

func (q *Queue) Close() error { return q.db.Close() }

func (q *Queue) prepareDB() error {
	sqlSchema := strings.ReplaceAll(sqlSchemaTemplate, tableNamePlaceholder, q.table)
	_, execErr := q.db.Exec(sqlSchema)
	return execErr
}

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
