package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/goto/entropy/pkg/errors"
)

const listResourceByFilterQuery = `SELECT r.id, r.urn, r.kind, r.name, r.project, r.created_at, r.updated_at, r.spec_configs, r.state_status, r.state_output, r.state_module_data, r.state_next_sync, r.state_sync_result, r.created_by, r.updated_by,
       array_agg(rt.tag)::text[] AS tags,
       jsonb_object_agg(COALESCE(rd.dependency_key, ''), d.urn) AS dependencies
FROM resources r
         LEFT JOIN resource_dependencies rd ON r.id = rd.resource_id
         LEFT JOIN resources d ON rd.depends_on = d.id
         LEFT JOIN resource_tags rt ON r.id = rt.resource_id
WHERE ($1 = '' OR r.project = $1)
  AND ($2 = '' OR r.kind = $2)
GROUP BY r.id
`

type resourceModel struct {
	ID              int64           `db:"id"`
	URN             string          `db:"urn"`
	Kind            string          `db:"kind"`
	Name            string          `db:"name"`
	Project         string          `db:"project"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
	CreatedBy       string          `db:"created_by"`
	UpdatedBy       string          `db:"updated_by"`
	SpecConfigs     []byte          `db:"spec_configs"`
	StateStatus     string          `db:"state_status"`
	StateOutput     []byte          `db:"state_output"`
	StateModuleData []byte          `db:"state_module_data"`
	StateNextSync   *time.Time      `db:"state_next_sync"`
	StateSyncResult json.RawMessage `db:"state_sync_result"`
}

type ListResourceByFilterRow struct {
	ID              int64
	Urn             string
	Kind            string
	Name            string
	Project         string
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
	SpecConfigs     []byte
	StateStatus     string
	StateOutput     []byte
	StateModuleData []byte
	StateNextSync   *time.Time
	StateSyncResult []byte
	CreatedBy       string
	UpdatedBy       string
	Tags            []byte
	Dependencies    []byte
}

func listResourceByFilter(ctx context.Context, db *sqlx.DB, project, kind string) ([]ListResourceByFilterRow, error) {
	rows, err := db.QueryContext(ctx, listResourceByFilterQuery, project, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListResourceByFilterRow
	for rows.Next() {
		var i ListResourceByFilterRow
		if err := rows.Scan(
			&i.ID,
			&i.Urn,
			&i.Kind,
			&i.Name,
			&i.Project,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.SpecConfigs,
			&i.StateStatus,
			&i.StateOutput,
			&i.StateModuleData,
			&i.StateNextSync,
			&i.StateSyncResult,
			&i.CreatedBy,
			&i.UpdatedBy,
			&i.Tags,
			&i.Dependencies,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func readResourceRecord(ctx context.Context, r sqlx.QueryerContext, urn string, into *resourceModel) error {
	cols := []string{
		"id", "urn", "kind", "project", "name", "created_at", "updated_at", "created_by", "updated_by",
		"spec_configs", "state_status", "state_output", "state_module_data",
		"state_next_sync", "state_sync_result",
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
