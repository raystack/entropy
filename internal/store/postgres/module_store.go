package postgres

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"github.com/goto/entropy/core/module"
	"github.com/goto/entropy/pkg/errors"
)

func (st *Store) GetModule(ctx context.Context, urn string) (*module.Module, error) {
	var rec moduleModel
	if err := readModuleRecord(ctx, st.db, urn, &rec); err != nil {
		return nil, err
	}
	return &module.Module{
		URN:       rec.URN,
		Name:      rec.Name,
		Project:   rec.Project,
		Configs:   rec.Configs,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
	}, nil
}

func (st *Store) ListModules(ctx context.Context, project string) ([]module.Module, error) {
	q := sq.Select("urn").From(tableModules)
	if project != "" {
		q = q.Where(sq.Eq{"project": project})
	}

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(st.db).QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var mods []module.Module
	for rows.Next() {
		var urn string
		if err := rows.Scan(&urn); err != nil {
			return nil, err
		}

		var mod moduleModel
		if err := readModuleRecord(ctx, st.db, urn, &mod); err != nil {
			return nil, err
		}
		mods = append(mods, mod.toModule())
	}

	return mods, rows.Err()
}

func (st *Store) CreateModule(ctx context.Context, m module.Module) error {
	err := insertModuleRecord(ctx, st.db, m)
	if err != nil {
		return translateErr(err)
	}
	return nil
}

func (st *Store) UpdateModule(ctx context.Context, m module.Module) error {
	updateSpec := sq.Update(tableModules).
		Where(sq.Eq{"urn": m.URN}).
		SetMap(map[string]interface{}{
			"configs":    m.Configs,
			"updated_at": sq.Expr("current_timestamp"),
		}).
		PlaceholderFormat(sq.Dollar)

	_, err := updateSpec.RunWith(st.db).ExecContext(ctx)
	return translateErr(err)
}

func (st *Store) DeleteModule(ctx context.Context, urn string) error {
	_, err := sq.Delete(tableModules).
		Where(sq.Eq{"urn": urn}).
		PlaceholderFormat(sq.Dollar).
		RunWith(st.db).
		ExecContext(ctx)
	return translateErr(err)
}
