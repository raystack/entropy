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

const tableNamePlaceholder = "__queueTable__"

var (
	//go:embed schema.sql
	sqlSchemaTemplate string

	queueNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
)

type Queue struct {
	db              *sql.DB
	name            string
	table           string
	extendInterval  time.Duration
	refreshInterval time.Duration
}

type txnFn func(ctx context.Context, tx *sql.Tx) error

func (q *Queue) Close() error { return q.db.Close() }

func (q *Queue) init() error {
	table, schema, err := generateSchema(q.name)
	if err != nil {
		return err
	}
	q.table = table

	_, execErr := q.db.Exec(schema)
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

func generateSchema(qName string) (table, schema string, err error) {
	if !queueNamePattern.MatchString(qName) {
		return "", "", fmt.Errorf("queue name must match pattern '%s'", queueNamePattern)
	}

	tableName := fmt.Sprintf("pgq_%s", qName)
	sqlSchema := strings.Replace(sqlSchemaTemplate, tableNamePlaceholder, tableName, -1)

	return tableName, sqlSchema, nil
}

func Open(conString, queueName string) (*Queue, error) {
	db, err := sql.Open("postgres", conString)
	if err != nil {
		return nil, err
	}

	q := &Queue{
		db:              db,
		name:            queueName,
		extendInterval:  30 * time.Second,
		refreshInterval: 20 * time.Second,
	}
	if err := q.init(); err != nil {
		_ = q.Close()
		return nil, err
	}

	return q, nil
}
