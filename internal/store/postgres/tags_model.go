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
