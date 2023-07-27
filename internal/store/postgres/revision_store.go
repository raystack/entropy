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
	var revs []resource.Revision
	txFn := func(ctx context.Context, tx *sqlx.Tx) error {
		resourceID, err := translateURNToID(ctx, tx, selector.URN)
		if err != nil {
			return err
		}

		deps := map[string]string{}
		if err := readResourceDeps(ctx, tx, resourceID, deps); err != nil {
			return err
		}

		builder := sq.Select("*").
			From(tableRevisions).
			Where(sq.Eq{"resource_id": resourceID}).
			OrderBy("created_at DESC")

		q, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return err
		}

		rows, err := tx.QueryxContext(ctx, q, args...)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var rm revisionModel
			if err := rows.StructScan(&rm); err != nil {
				return err
			}

			revs = append(revs, resource.Revision{
				ID:        rm.ID,
				URN:       selector.URN,
				Reason:    rm.Reason,
				CreatedAt: rm.CreatedAt,
				CreatedBy: rm.CreatedBy,
				Spec: resource.Spec{
					Configs:      rm.SpecConfigs,
					Dependencies: deps,
				},
			})
		}
		_ = rows.Close()

		for i, rev := range revs {
			var tags []string
			if err := readRevisionTags(ctx, tx, rev.ID, &tags); err != nil {
				return err
			}
			revs[i].Labels = tagsToLabelMap(tags)
		}

		return nil
	}

	if err := withinTx(ctx, st.db, true, txFn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return revs, nil
}

func insertRevision(ctx context.Context, tx *sqlx.Tx, resID int64, rev resource.Revision) error {
	q := sq.Insert(tableRevisions).
		Columns("resource_id", "reason", "spec_configs", "created_by").
		Values(resID, rev.Reason, rev.Spec.Configs, rev.CreatedBy).
		Suffix(`RETURNING "id"`).
		PlaceholderFormat(sq.Dollar)

	var revisionID int64
	if err := q.RunWith(tx).QueryRowContext(ctx).Scan(&revisionID); err != nil {
		return err
	}

	return setRevisionTags(ctx, tx, revisionID, rev.Labels)
}

func setRevisionTags(ctx context.Context, runner sq.BaseRunner, id int64, labels map[string]string) error {
	return setTags(ctx, runner, tableRevisionTags, "revision_id", id, labels)
}
