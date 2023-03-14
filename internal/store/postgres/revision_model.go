package postgres

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/goto/entropy/pkg/errors"
)

type revisionModel struct {
	ID          int64     `db:"id"`
	URN         string    `db:"urn"`
	Reason      string    `db:"reason"`
	CreatedAt   time.Time `db:"created_at"`
	SpecConfigs []byte    `db:"spec_configs"`
}

func readRevisionRecord(ctx context.Context, r sqlx.QueryerContext, id int64, into *revisionModel) error {
	cols := []string{"id", "urn", "reason", "created_at", "spec_configs"}
	builder := sq.Select(cols...).From(tableRevisions).Where(sq.Eq{"id": id})

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(ctx, query, args...).StructScan(into); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.ErrNotFound
		}
		return err
	}
	return nil
}

func readRevisionTags(ctx context.Context, r sq.BaseRunner, revisionID int64, into *[]string) error {
	return readTags(ctx, r, tableRevisionTags, columnRevisionID, revisionID, into)
}
