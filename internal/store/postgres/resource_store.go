package postgres

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

func (st *Store) GetByURN(ctx context.Context, urn string) (*resource.Resource, error) {
	var rec resourceModel
	var tags []string
	deps := map[string]string{}

	readResourceParts := func(ctx context.Context, tx *sqlx.Tx) error {
		if err := readResourceRecord(ctx, tx, urn, &rec); err != nil {
			return err
		}

		if err := readResourceTags(ctx, tx, rec.ID, &tags); err != nil {
			return err
		}

		if err := readResourceDeps(ctx, tx, rec.ID, deps); err != nil {
			return err
		}

		return nil
	}

	if txErr := withinTx(ctx, st.db, true, readResourceParts); txErr != nil {
		return nil, txErr
	}

	return &resource.Resource{
		URN:       rec.URN,
		Kind:      rec.Kind,
		Name:      rec.Name,
		Project:   rec.Project,
		Labels:    tagsToLabelMap(tags),
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
		Spec: resource.Spec{
			Configs:      rec.SpecConfigs,
			Dependencies: deps,
		},
		State: resource.State{
			Status:     rec.StateStatus,
			Output:     rec.StateOutput,
			ModuleData: rec.StateModuleData,
		},
	}, nil
}

func (st *Store) List(ctx context.Context, filter resource.Filter) ([]resource.Resource, error) {
	q := sq.Select("urn").From("resources")
	if filter.Kind != "" {
		q = q.Where(sq.Eq{"kind": filter.Kind})
	}
	if filter.Project != "" {
		q = q.Where(sq.Eq{"project": filter.Project})
	}

	if len(filter.Labels) > 0 {
		tags := labelMapToTags(filter.Labels)
		q = q.Join("resource_tags USING (resource_id)").
			Where(sq.Eq{"tag": tags}).
			GroupBy("resource_id")
	}

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(st.db).QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var res []resource.Resource
	for rows.Next() {
		var urn string
		if err := rows.Scan(&urn); err != nil {
			return nil, err
		}

		r, err := st.GetByURN(ctx, urn)
		if err != nil {
			return nil, err
		}
		res = append(res, *r)
	}

	return res, nil
}

func (st *Store) Create(ctx context.Context, r resource.Resource, hooks ...resource.MutationHook) error {
	insertResource := func(ctx context.Context, tx *sqlx.Tx) error {
		id, err := insertResourceRecord(ctx, tx, r)
		if err != nil {
			return err
		}

		if err := setResourceTags(ctx, tx, id, r.Labels); err != nil {
			return err
		}

		if err := insertResourceDependencies(ctx, tx, id, r.Spec.Dependencies); err != nil {
			return err
		}

		return runAllHooks(ctx, hooks)
	}

	txErr := withinTx(ctx, st.db, false, insertResource)
	if txErr != nil {
		return txErr
	}
	return nil
}

func (st *Store) Update(ctx context.Context, r resource.Resource, hooks ...resource.MutationHook) error {
	updateResource := func(ctx context.Context, tx *sqlx.Tx) error {
		panic("not implemented")
	}

	txErr := withinTx(ctx, st.db, false, updateResource)
	if txErr != nil {
		return txErr
	}
	return nil
}

func (st *Store) Delete(ctx context.Context, urn string, hooks ...resource.MutationHook) error {
	deleteFn := func(ctx context.Context, tx *sqlx.Tx) error {
		id, err := translateURNToID(ctx, tx, urn)
		if err != nil {
			return err
		}

		_, err = sq.Delete(tableResourceDependencies).
			Where(sq.Eq{"resource_id": id}).
			PlaceholderFormat(sq.Dollar).
			RunWith(tx).
			ExecContext(ctx)
		if err != nil {
			return err
		}

		_, err = sq.Delete(tableResourceTags).
			Where(sq.Eq{"resource_id": id}).
			PlaceholderFormat(sq.Dollar).
			RunWith(tx).
			ExecContext(ctx)
		if err != nil {
			return err
		}

		return runAllHooks(ctx, hooks)
	}

	return withinTx(ctx, st.db, false, deleteFn)
}

func (st *Store) DoPending(ctx context.Context, fn resource.PendingHandler) error {
	return nil
}

func insertResourceRecord(ctx context.Context, runner sq.BaseRunner, r resource.Resource) (int64, error) {
	q := sq.Insert(tableResources).
		Columns("urn", "kind", "project", "name", "created_at", "updated_at",
			"spec_configs", "state_status", "state_output", "state_module_data").
		Values(r.URN, r.Kind, r.Project, r.Name, r.CreatedAt, r.UpdatedAt,
			r.Spec.Configs, r.State.Status, r.State.Output, r.State.ModuleData).
		Suffix(`RETURNING "id"`).
		PlaceholderFormat(sq.Dollar)

	var id int64
	if err := q.RunWith(runner).QueryRowContext(ctx).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func setResourceTags(ctx context.Context, runner sq.BaseRunner, id int64, labels map[string]string) error {
	deleteOld := sq.Delete(tableResourceTags).Where(sq.Eq{"resource_id": id}).PlaceholderFormat(sq.Dollar)
	insertTags := sq.Insert(tableResourceTags).Columns("resource_id", "tag").PlaceholderFormat(sq.Dollar)
	for _, tag := range labelMapToTags(labels) {
		insertTags = insertTags.Values(id, tag)
	}

	if _, err := deleteOld.RunWith(runner).ExecContext(ctx); err != nil {
		return err
	}

	if _, err := insertTags.RunWith(runner).ExecContext(ctx); err != nil {
		return err
	}

	return nil
}

func insertResourceDependencies(ctx context.Context, runner sq.BaseRunner, id int64, deps map[string]string) error {
	insertDeps := sq.Insert(tableResourceDependencies).
		Columns("resource_id", "dependency_key", "depends_on").
		PlaceholderFormat(sq.Dollar)
	for key, val := range deps {
		insertDeps = insertDeps.Values(id, key, val)
	}

	_, err := insertDeps.RunWith(runner).ExecContext(ctx)
	return err
}
