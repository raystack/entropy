package pgsql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

const (
	tableResources            = "resources"
	tableResourceTags         = "resource_tags"
	tableResourceDependencies = "resource_dependencies"
)

func (st *Store) GetByURN(ctx context.Context, urn string) (*resource.Resource, error) {
	var rec resourceRecord
	var tags []string
	deps := map[string]string{}

	readResourceParts := func(ctx context.Context, tx *sql.Tx) error {
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
	// TODO implement me
	panic("implement me")
}

func (st *Store) Update(ctx context.Context, r resource.Resource, hooks ...resource.MutationHook) error {
	// TODO implement me
	panic("implement me")
}

func (st *Store) Delete(ctx context.Context, urn string, hooks ...resource.MutationHook) error {
	deleteFn := func(ctx context.Context, tx *sql.Tx) error {
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
	// TODO implement me
	panic("implement me")
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

func readResourceRecord(ctx context.Context, r sq.BaseRunner, urn string, into *resourceRecord) error {
	cols := []string{
		"id", "urn", "kind", "project", "name", "created_at", "updated_at",
		"spec_configs", "state_status", "state_output", "state_module_data",
	}
	q := sq.Select(cols...).From(tableResources).Where(sq.Eq{"urn": urn})

	row := q.PlaceholderFormat(sq.Dollar).RunWith(r).QueryRowContext(ctx)

	if err := structScan(row, cols, into); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.ErrNotFound
		}
		return err
	}
	return nil
}

func readResourceTags(ctx context.Context, r sq.BaseRunner, id int64, into *[]string) error {
	q := sq.Select("tag").From(tableResourceTags).Where(sq.Eq{"resource_id": id})

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(r).QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return err
		}
		*into = append(*into, tag)
	}
	return nil
}

func readResourceDeps(ctx context.Context, r sq.BaseRunner, id int64, into map[string]string) error {
	q := sq.Select("dependency_key", "depends_on").From(tableResourceDependencies).Where(sq.Eq{"resource_id": id})

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(r).QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	for rows.Next() {
		var key, val string
		if err := rows.Scan(&key, &val); err != nil {
			return err
		}
		into[key] = val
	}
	return nil
}

func tagsToLabelMap(tags []string) map[string]string {
	labels := map[string]string{}
	for _, tag := range tags {
		parts := strings.SplitN(tag, "=", 2)
		key, val := parts[0], parts[1]
		labels[key] = val
	}
	return labels
}

func labelMapToTags(labels map[string]string) []string {
	var res []string
	for k, v := range labels {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}
	return res
}

func runAllHooks(ctx context.Context, hooks []resource.MutationHook) error {
	for _, hook := range hooks {
		if err := hook(ctx); err != nil {
			return err
		}
	}
	return nil
}
