package postgres

import (
	"context"
	_ "embed"

	"github.com/jmoiron/sqlx"
)

const (
	tableResources            = "resources"
	tableResourceTags         = "resource_tags"
	tableResourceDependencies = "resource_dependencies"
	columnResourceID          = "resource_id"

	tableRevisions    = "revisions"
	tableRevisionTags = "revision_tags"
	columnRevisionID  = "revision_id"
)

// schema represents the storage schema.
// Note: Update the constants above if the table name is changed.
//
//go:embed schema.sql
var schema string

type Store struct {
	db *sqlx.DB
}

func (st *Store) Migrate(ctx context.Context) error {
	_, err := st.db.ExecContext(ctx, schema)
	return err
}

func (st *Store) Close() error { return st.db.Close() }

// Open returns store instance backed by PostgreSQL.
func Open(conStr string) (*Store, error) {
	db, err := sqlx.Open("postgres", conStr)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}
