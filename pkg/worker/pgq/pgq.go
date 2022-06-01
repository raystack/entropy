package pgq

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq" // postgres driver.
)

const (
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

type Queue struct {
	db              *sql.DB
	name            string
	table           string
	extendInterval  time.Duration
	refreshInterval time.Duration
}

type txnFn func(ctx context.Context, tx *sql.Tx) error

func Open(conString, queueName string) (*Queue, error) {
	if !queueNamePattern.MatchString(queueName) {
		return nil, errInvalidQueueName
	}
	tableName := fmt.Sprintf("pgq_%s", queueName)

	db, err := sql.Open("postgres", conString)
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

func (q *Queue) Close() error { return q.db.Close() }

func (q *Queue) prepareDB() error {
	sqlSchema := strings.ReplaceAll(sqlSchemaTemplate, tableNamePlaceholder, q.table)
	_, execErr := q.db.Exec(sqlSchema)
	return execErr
}

func (q *Queue) withTxn(ctx context.Context, readOnly bool, fn txnFn) error {
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
