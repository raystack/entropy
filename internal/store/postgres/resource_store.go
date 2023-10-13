package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/goto/entropy/core/resource"
	"github.com/goto/entropy/pkg/errors"
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

	var syncResult resource.SyncResult
	if len(rec.StateSyncResult) > 0 {
		if err := json.Unmarshal(rec.StateSyncResult, &syncResult); err != nil {
			return nil, errors.ErrInternal.
				WithMsgf("failed to json unmarshal state_sync_result").
				WithCausef(err.Error())
		}
	}

	return &resource.Resource{
		URN:       rec.URN,
		Kind:      rec.Kind,
		Name:      rec.Name,
		Project:   rec.Project,
		Labels:    tagsToLabelMap(tags),
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
		CreatedBy: rec.CreatedBy,
		UpdatedBy: rec.UpdatedBy,
		Spec: resource.Spec{
			Configs:      rec.SpecConfigs,
			Dependencies: deps,
		},
		State: resource.State{
			Status:     rec.StateStatus,
			Output:     rec.StateOutput,
			ModuleData: rec.StateModuleData,
			NextSyncAt: rec.StateNextSync,
			SyncResult: syncResult,
		},
	}, nil
}

func (st *Store) List(ctx context.Context, filter resource.Filter, withSpecConfigs bool) ([]resource.Resource, error) {
	var resourceList []ListResourceByFilterRow
	var err error
	if withSpecConfigs {
		resourceList, err = listResourceWithSpecConfigsByFilter(ctx, st.db, filter.Project, filter.Kind)
	} else {
		resourceList, err = listResourceByFilter(ctx, st.db, filter.Project, filter.Kind)
	}
	if err != nil {
		return nil, err
	}

	var result []resource.Resource
	for _, res := range resourceList {
		var nextSyncAt *time.Time
		if res.StateNextSync != nil && !res.StateNextSync.IsZero() {
			nextSyncAt = res.StateNextSync
		}

		var syncResult resource.SyncResult
		if len(res.StateSyncResult) > 0 {
			if err := json.Unmarshal(res.StateSyncResult, &syncResult); err != nil {
				return nil, err
			}
		}

		deps, err := depsBytesToMap(res.Dependencies)
		if err != nil {
			return nil, err
		}

		result = append(result, resource.Resource{
			URN:       res.Urn,
			Kind:      res.Kind,
			Name:      res.Name,
			Project:   res.Project,
			Labels:    tagsToLabelMap(res.Tags),
			CreatedAt: *res.CreatedAt,
			UpdatedAt: *res.UpdatedAt,
			UpdatedBy: res.UpdatedBy,
			CreatedBy: res.CreatedBy,
			Spec: resource.Spec{
				Configs:      res.SpecConfigs,
				Dependencies: deps,
			},
			State: resource.State{
				Status:     res.StateStatus,
				Output:     res.StateOutput,
				ModuleData: res.StateModuleData,
				NextSyncAt: nextSyncAt,
				SyncResult: syncResult,
			},
		})
	}

	return result, nil
}

func (st *Store) Create(ctx context.Context, r resource.Resource, hooks ...resource.MutationHook) error {
	insertResource := func(ctx context.Context, tx *sqlx.Tx) error {
		id, err := insertResourceRecord(ctx, tx, r)
		if err != nil {
			return translateErr(err)
		}

		if err := setResourceTags(ctx, tx, id, r.Labels); err != nil {
			return translateErr(err)
		}

		if err := setDependencies(ctx, tx, id, r.Spec.Dependencies); err != nil {
			return translateErr(err)
		}

		rev := resource.Revision{
			URN:       r.URN,
			Spec:      r.Spec,
			Labels:    r.Labels,
			Reason:    "action:create",
			CreatedBy: r.UpdatedBy,
		}

		if err := insertRevision(ctx, tx, id, rev); err != nil {
			return translateErr(err)
		}

		return runAllHooks(ctx, hooks)
	}

	txErr := withinTx(ctx, st.db, false, insertResource)
	if txErr != nil {
		return txErr
	}
	return nil
}

func (st *Store) Update(ctx context.Context, r resource.Resource, saveRevision bool, reason string, hooks ...resource.MutationHook) error {
	updateResource := func(ctx context.Context, tx *sqlx.Tx) error {
		id, err := translateURNToID(ctx, tx, r.URN)
		if err != nil {
			return err
		}

		updateSpec := sq.Update(tableResources).
			Where(sq.Eq{"id": id}).
			SetMap(map[string]interface{}{
				"updated_at":        sq.Expr("current_timestamp"),
				"updated_by":        r.UpdatedBy,
				"spec_configs":      r.Spec.Configs,
				"state_status":      r.State.Status,
				"state_output":      r.State.Output,
				"state_module_data": r.State.ModuleData,
				"state_next_sync":   r.State.NextSyncAt,
				"state_sync_result": syncResultAsJSON(r.State.SyncResult),
			}).
			PlaceholderFormat(sq.Dollar)

		if _, err := updateSpec.RunWith(tx).ExecContext(ctx); err != nil {
			return err
		}

		if err := setResourceTags(ctx, tx, id, r.Labels); err != nil {
			return err
		}

		if err := setDependencies(ctx, tx, id, r.Spec.Dependencies); err != nil {
			return err
		}

		if saveRevision {
			rev := resource.Revision{
				URN:       r.URN,
				Spec:      r.Spec,
				Labels:    r.Labels,
				Reason:    reason,
				CreatedBy: r.UpdatedBy,
			}

			if err := insertRevision(ctx, tx, id, rev); err != nil {
				return translateErr(err)
			}
		}

		return runAllHooks(ctx, hooks)
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

func (st *Store) SyncOne(ctx context.Context, syncFn resource.SyncFn) error {
	urn, err := st.fetchResourceForSync(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No resource available for sync.
			return nil
		}
		return err
	}

	cur, err := st.GetByURN(ctx, urn)
	if err != nil {
		return err
	}

	synced, err := st.handleDequeued(ctx, *cur, syncFn)
	if err != nil {
		return err
	}

	return st.Update(ctx, *synced, false, "sync")
}

func (st *Store) handleDequeued(baseCtx context.Context, res resource.Resource, fn resource.SyncFn) (*resource.Resource, error) {
	runCtx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	// Run heartbeat to keep the resource being picked up by some other syncer
	// thread. If heartbeat exits, runCtx will be cancelled and fn should exit.
	go st.runHeartbeat(runCtx, cancel, res.URN)

	return fn(runCtx, res)
}

func (st *Store) fetchResourceForSync(ctx context.Context) (string, error) {
	var urn string

	// find a resource ready for sync, extend it next sync time atomically.
	// this ensures multiple workers do not pick up same resources for sync.
	err := withinTx(ctx, st.db, false, func(ctx context.Context, tx *sqlx.Tx) error {
		builder := sq.
			Select("urn").
			From(tableResources).
			Where(sq.Expr("state_next_sync <= current_timestamp")).
			Suffix("FOR UPDATE SKIP LOCKED")

		query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return err
		}

		if err := st.db.QueryRowxContext(ctx, query, args...).Scan(&urn); err != nil {
			return err
		}

		return st.extendWaitTime(ctx, tx, urn)
	})

	return urn, err
}

func (st *Store) runHeartbeat(ctx context.Context, cancel context.CancelFunc, id string) {
	defer cancel()

	tick := time.NewTicker(st.refreshInterval)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-tick.C:
			if err := st.extendWaitTime(ctx, st.db, id); err != nil {
				return
			}
		}
	}
}

func (st *Store) extendWaitTime(ctx context.Context, r sq.BaseRunner, urn string) error {
	extendTo := sq.Expr("current_timestamp + (? ||' seconds')::interval ", st.extendInterval.Seconds())
	extendQuery := sq.Update(tableResources).
		Set("state_next_sync", extendTo).
		Where(sq.Eq{"urn": urn})

	_, err := extendQuery.PlaceholderFormat(sq.Dollar).RunWith(r).ExecContext(ctx)
	return err
}

func insertResourceRecord(ctx context.Context, runner sqlx.QueryerContext, r resource.Resource) (int64, error) {
	builder := sq.Insert(tableResources).
		Columns("urn", "kind", "project", "name", "created_at", "updated_at", "created_by", "updated_by",
			"spec_configs", "state_status", "state_output", "state_module_data",
			"state_next_sync", "state_sync_result").
		Values(r.URN, r.Kind, r.Project, r.Name, r.CreatedAt, r.UpdatedAt, r.CreatedBy, r.UpdatedBy,
			r.Spec.Configs, r.State.Status, r.State.Output, r.State.ModuleData,
			r.State.NextSyncAt, syncResultAsJSON(r.State.SyncResult)).
		Suffix(`RETURNING "id"`)

	q, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, err
	}

	var id int64
	if err := runner.QueryRowxContext(ctx, q, args...).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func setResourceTags(ctx context.Context, runner sq.BaseRunner, id int64, labels map[string]string) error {
	return setTags(ctx, runner, tableResourceTags, "resource_id", id, labels)
}

func setDependencies(ctx context.Context, runner sq.BaseRunner, id int64, deps map[string]string) error {
	deleteOld := sq.Delete(tableResourceDependencies).Where(sq.Eq{"resource_id": id}).PlaceholderFormat(sq.Dollar)
	if _, err := deleteOld.RunWith(runner).ExecContext(ctx); err != nil {
		return err
	}

	if len(deps) > 0 {
		insertDeps := sq.Insert(tableResourceDependencies).
			Columns("resource_id", "dependency_key", "depends_on").
			PlaceholderFormat(sq.Dollar)

		for depKey, dependsOnURN := range deps {
			dependsOnID, err := translateURNToID(ctx, runner, dependsOnURN)
			if err != nil {
				return err
			}
			insertDeps = insertDeps.Values(id, depKey, dependsOnID)
		}

		if _, err := insertDeps.RunWith(runner).ExecContext(ctx); err != nil {
			return err
		}
	}

	return nil
}

func syncResultAsJSON(syncRes resource.SyncResult) json.RawMessage {
	if syncRes == (resource.SyncResult{}) {
		return nil
	}
	val, err := json.Marshal(syncRes)
	if err != nil {
		panic(err)
	}
	return val
}

func depsBytesToMap(dependencies []byte) (map[string]string, error) {
	deps := map[string]string{}
	if len(dependencies) > 0 {
		if err := json.Unmarshal(dependencies, &deps); err != nil {
			return nil, err
		}

		for k := range deps {
			if k != "" {
				break
			}
			deps = map[string]string{}
		}
	}

	return deps, nil
}
