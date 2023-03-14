package postgres

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
)

func (st *Store) Revisions(ctx context.Context, selector resource.RevisionsSelector) ([]resource.Revision, error) {
	q := sq.Select("id").
		From(tableRevisions).
		Where(sq.Eq{"urn": selector.URN})

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(st.db).QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var revs []resource.Revision
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		r, err := st.getRevisionByID(ctx, id)
		if err != nil {
			return nil, err
		}
		revs = append(revs, *r)
	}

	return revs, rows.Err()
}

func (st *Store) getRevisionByID(ctx context.Context, id int64) (*resource.Revision, error) {
	var rec revisionModel
	var tags []string
	deps := map[string]string{}

	readRevisionParts := func(ctx context.Context, tx *sqlx.Tx) error {
		if err := readRevisionRecord(ctx, tx, id, &rec); err != nil {
			return err
		}

		if err := readRevisionTags(ctx, tx, rec.ID, &tags); err != nil {
			return err
		}

		resourceID, err := translateURNToID(ctx, tx, rec.URN)
		if err != nil {
			return err
		}

		if err := readResourceDeps(ctx, tx, resourceID, deps); err != nil {
			return err
		}

		return nil
	}

	if txErr := withinTx(ctx, st.db, true, readRevisionParts); txErr != nil {
		return nil, txErr
	}

	return &resource.Revision{
		ID:        rec.ID,
		URN:       rec.URN,
		Reason:    rec.Reason,
		Labels:    tagsToLabelMap(tags),
		CreatedAt: rec.CreatedAt,
		Spec: resource.Spec{
			Configs:      rec.SpecConfigs,
			Dependencies: deps,
		},
	}, nil
}

func insertRevision(ctx context.Context, tx *sqlx.Tx, rev resource.Revision) error {
	revisionID, err := insertRevisionRecord(ctx, tx, rev)
	if err != nil {
		return err
	}

	if err := setRevisionTags(ctx, tx, revisionID, rev.Labels); err != nil {
		return err
	}
	return nil
}

func insertRevisionRecord(ctx context.Context, runner sq.BaseRunner, r resource.Revision) (int64, error) {
	q := sq.Insert(tableRevisions).
		Columns("urn", "reason", "spec_configs").
		Values(r.URN, r.Reason, r.Spec.Configs).
		Suffix(`RETURNING "id"`).
		PlaceholderFormat(sq.Dollar)

	var id int64
	if err := q.RunWith(runner).QueryRowContext(ctx).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func setRevisionTags(ctx context.Context, runner sq.BaseRunner, id int64, labels map[string]string) error {
	return setTags(ctx, runner, tableRevisionTags, "revision_id", id, labels)
}
