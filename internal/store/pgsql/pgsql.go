package pgsql

import (
	"context"
	"database/sql"
	_ "embed"
)

//go:embed schema.sql
var schema string

func Open(conStr string) (*Store, error) {
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

type Store struct{ db *sql.DB }

func (st *Store) Migrate(ctx context.Context) error {
	_, err := st.db.ExecContext(ctx, schema)

	return err
}

func (st *Store) Close() error {
	return st.db.Close()
}
