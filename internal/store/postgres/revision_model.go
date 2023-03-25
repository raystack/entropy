package postgres

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type revisionModel struct {
	ID          int64     `db:"id"`
	Reason      string    `db:"reason"`
	CreatedAt   time.Time `db:"created_at"`
	ResourceID  int64     `db:"resource_id"`
	SpecConfigs []byte    `db:"spec_configs"`
}

func readRevisionTags(ctx context.Context, r sq.BaseRunner, revisionID int64, into *[]string) error {
	return readTags(ctx, r, tableRevisionTags, columnRevisionID, revisionID, into)
}
