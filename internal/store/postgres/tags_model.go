package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

func readTags(ctx context.Context, r sq.BaseRunner, table string, idColumn string, id int64, into *[]string) error {
	builder := sq.Select("tag").From(table).Where(sq.Eq{idColumn: id})

	rows, err := builder.PlaceholderFormat(sq.Dollar).RunWith(r).QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return err
		}
		*into = append(*into, tag)
	}
	return rows.Err()
}

func setTags(ctx context.Context, runner sq.BaseRunner, table string, idColumn string, id int64, labels map[string]string) error {
	deleteOld := sq.Delete(table).Where(sq.Eq{idColumn: id}).PlaceholderFormat(sq.Dollar)

	if _, err := deleteOld.RunWith(runner).ExecContext(ctx); err != nil {
		return err
	}

	if len(labels) > 0 {
		insertTags := sq.Insert(table).Columns(idColumn, "tag").PlaceholderFormat(sq.Dollar)
		for _, tag := range labelMapToTags(labels) {
			insertTags = insertTags.Values(id, tag)
		}
		if _, err := insertTags.RunWith(runner).ExecContext(ctx); err != nil {
			return err
		}
	}

	return nil
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
