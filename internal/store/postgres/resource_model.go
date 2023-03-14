package postgres

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/goto/entropy/pkg/errors"
)

type resourceModel struct {
	ID              int64     `db:"id"`
	URN             string    `db:"urn"`
	Kind            string    `db:"kind"`
	Name            string    `db:"name"`
	Project         string    `db:"project"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
	SpecConfigs     []byte    `db:"spec_configs"`
	StateStatus     string    `db:"state_status"`
	StateOutput     []byte    `db:"state_output"`
	StateModuleData []byte    `db:"state_module_data"`
}

func readResourceRecord(ctx context.Context, r sqlx.QueryerContext, urn string, into *resourceModel) error {
	cols := []string{
		"id", "urn", "kind", "project", "name", "created_at", "updated_at",
		"spec_configs", "state_status", "state_output", "state_module_data",
	}
	builder := sq.Select(cols...).From(tableResources).Where(sq.Eq{"urn": urn})

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

func readResourceTags(ctx context.Context, r sq.BaseRunner, id int64, into *[]string) error {
	return readTags(ctx, r, tableResourceTags, columnResourceID, id, into)
}

func readResourceDeps(ctx context.Context, r sq.BaseRunner, id int64, into map[string]string) error {
	q := sq.Select("rd.dependency_key as dep_key", "r.urn as dep_urn").
		From("resource_dependencies rd").
		Join("resources r ON r.id=rd.depends_on").
		Where(sq.Eq{"resource_id": id})

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(r).QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var key, val string
		if err := rows.Scan(&key, &val); err != nil {
			return err
		}
		into[key] = val
	}
	return rows.Err()
}

func translateURNToID(ctx context.Context, r sq.BaseRunner, urn string) (int64, error) {
	row := sq.Select("id").
		From(tableResources).
		Where(sq.Eq{"urn": urn}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRowContext(ctx)

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
