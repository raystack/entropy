package postgres

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/odpf/entropy/core/module"
	"github.com/odpf/entropy/pkg/errors"
)

const tableModules = "modules"

type moduleModel struct {
	URN         string    `db:"urn"`
	Name        string    `db:"name"`
	Project     string    `db:"project"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	SpecLoader  string    `db:"spec_loader"`
	SpecPath    string    `db:"spec_path"`
	SpecConfigs []byte    `db:"spec_configs"`
}

func (mm moduleModel) toModule() module.Module {
	return module.Module{
		URN:     mm.URN,
		Name:    mm.Name,
		Project: mm.Project,
		Spec: module.Spec{
			Path:    mm.SpecPath,
			Loader:  mm.SpecLoader,
			Configs: mm.SpecConfigs,
		},
		CreatedAt: mm.CreatedAt,
		UpdatedAt: mm.UpdatedAt,
	}
}

func readModuleRecord(ctx context.Context, r sqlx.QueryerContext, urn string, into *moduleModel) error {
	cols := []string{
		"urn", "project", "name", "created_at", "updated_at",
		"spec_configs", "spec_loader", "spec_path",
	}
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
		Columns("urn", "project", "name", "created_at", "updated_at",
			"spec_configs", "spec_loader", "spec_path").
		Values(mod.URN, mod.Project, mod.Name, mod.CreatedAt, mod.UpdatedAt,
			mod.Spec.Configs, mod.Spec.Loader, mod.Spec.Path).
		PlaceholderFormat(sq.Dollar)

	_, err := q.RunWith(runner).ExecContext(ctx)
	return err
}
