package pgsql

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/odpf/entropy/pkg/errors"
)

type rowScanner interface {
	Scan(targets ...interface{}) error
}

type TxFunc func(ctx context.Context, tx *sql.Tx) error

func withinTx(ctx context.Context, db *sql.DB, readOnly bool, fns ...TxFunc) error {
	opts := &sql.TxOptions{ReadOnly: readOnly}

	tx, err := db.BeginTx(ctx, opts)
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

func structScan(sc rowScanner, columns []string, structPtr interface{}) error {
	rt := reflect.TypeOf(structPtr)
	if !isPtrToStruct(rt) {
		return errors.ErrInternal.WithCausef("structPtr must be a pointer to a struct value, not '%s'", rt.Kind())
	}
	rt = rt.Elem()
	rv := reflect.ValueOf(structPtr).Elem()

	colSet := map[string]bool{}
	for _, column := range columns {
		colSet[column] = false
	}

	var scanTargets []interface{}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		colTarget := f.Tag.Get("db")
		if colTarget == "" {
			return errors.ErrInternal.WithCausef("field '%s' must be annotated with db tag", f.Name)
		}

		if _, exists := colSet[colTarget]; exists {
			scanTargets = append(scanTargets, rv.Field(i).Addr().Interface())
			colSet[colTarget] = true
		}
	}

	for col, targetFound := range colSet {
		if !targetFound {
			return errors.ErrInternal.WithCausef("target field for column '%s' not found", col)
		}
	}

	return sc.Scan(scanTargets...)
}

func isPtrToStruct(rv reflect.Type) bool {
	return rv.Kind() == reflect.Pointer &&
		rv.Elem().Kind() == reflect.Struct
}
