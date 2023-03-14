package postgres

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/pkg/errors"
)

const tableModules = "modules"

type moduleModel struct {
	URN       string    `db:"urn"`
	Name      string    `db:"name"`
	Project   string    `db:"project"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Configs   []byte    `db:"configs"`
}

func (mm moduleModel) toModule() module.Module {
	return module.Module{
		URN:       mm.URN,
		Name:      mm.Name,
		Project:   mm.Project,
		Configs:   mm.Configs,
		CreatedAt: mm.CreatedAt,
		UpdatedAt: mm.UpdatedAt,
	}
}

func readModuleRecord(ctx context.Context, r sqlx.QueryerContext, urn string, into *moduleModel) error {
	cols := []string{"urn", "project", "name", "created_at", "updated_at", "configs"}
	builder := sq.Select(cols...).From(tableModules).Where(sq.Eq{"urn": urn})

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

func insertModuleRecord(ctx context.Context, runner sq.BaseRunner, mod module.Module) error {
	q := sq.Insert(tableModules).
		Columns("urn", "project", "name", "created_at", "updated_at", "configs").
		Values(mod.URN, mod.Project, mod.Name, mod.CreatedAt, mod.UpdatedAt, mod.Configs).
		PlaceholderFormat(sq.Dollar)

	_, err := q.RunWith(runner).ExecContext(ctx)
	return err
}
