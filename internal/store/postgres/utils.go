package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/odpf/entropy/core/resource"
	"github.com/odpf/entropy/pkg/errors"
)

type TxFunc func(ctx context.Context, tx *sqlx.Tx) error

func withinTx(ctx context.Context, db *sqlx.DB, readOnly bool, fns ...TxFunc) error {
	opts := &sql.TxOptions{ReadOnly: readOnly}

	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}

	for _, fn := range fns {
		if err := fn(ctx, tx); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func translateErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errors.ErrNotFound.WithCausef(err.Error())
	}

	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		// Refer http://www.postgresql.org/docs/9.3/static/errcodes-appendix.html
		switch pgErr.Code.Name() {
		case "unique_violation":
			return errors.ErrConflict.WithCausef(err.Error())
		}
	}

	return err
}

func tagsToLabelMap(tags []string) map[string]string {
	const keyValueParts = 2

	labels := map[string]string{}
	for _, tag := range tags {
		parts := strings.SplitN(tag, "=", keyValueParts)
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
