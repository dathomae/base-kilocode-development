package rdbms

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"chronocone/backend/serde/rdbms/gen/postgres/public/enum"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

// allCols returns the AllColumns projection list for a generated table struct
// via reflection. Go-Jet's default aliasing already produces "table.column"
// labels that qrm uses to match against naming-aligned destination structs;
// no manual AS() override is required.
func allCols(tbl postgres.Table) postgres.ColumnList {
	v := reflect.ValueOf(tbl)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if f := v.FieldByName("AllColumns"); f.IsValid() {
		if cl, ok := f.Interface().(postgres.ColumnList); ok {
			return cl
		}
	}
	return nil
}

// selectAll returns a SELECT statement over all columns of tbl.
// Returns an error when tbl has no AllColumns projection, preventing a
// silent SELECT * fallback.
func selectAll(tbl postgres.Table) (postgres.SelectStatement, error) {
	cols := allCols(tbl)
	if len(cols) == 0 {
		return nil, fmt.Errorf("selectAll: table %q has no AllColumns projection; refusing SELECT *", tbl.TableName())
	}
	return postgres.SELECT(cols).FROM(tbl), nil
}

// IsNoRows reports whether err signals an empty result set from go-jet's qrm layer.
// Two forms are checked because qrm wraps sql.ErrNoRows only in some code paths.
func IsNoRows(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, qrm.ErrNoRows) ||
		strings.Contains(err.Error(), "qrm: no rows in result set")
}

// registryEntryForTable finds the BitemporalRegistryEntry whose Table matches tbl by
// table name and alias. This is an O(n) scan over the registry; acceptable because the
// registry is small (~35 entries) and read-path functions are not in hot loops.
// Panics if the table is not registered, since calling read-path helpers on an
// unregistered table is a programming error.
func registryEntryForTable(tbl postgres.Table) BitemporalRegistryEntry {
	for _, entry := range BitemporalTableRegistry {
		if entry.Table.TableName() == tbl.TableName() && entry.Table.Alias() == tbl.Alias() {
			return entry
		}
	}
	panic(fmt.Sprintf("registryEntryForTable: table %q (alias %q) not found in BitemporalTableRegistry", tbl.TableName(), tbl.Alias()))
}

// primaryIntColumn returns the first (and for single-key tables, only) IDColumn
// as a ColumnInteger. Panics for composite-key tables, since the read-path helpers
// below accept a single int64 entityID and cannot represent composite keys.
func primaryIntColumn(tbl postgres.Table) postgres.ColumnInteger {
	entry := registryEntryForTable(tbl)
	if len(entry.IDColumns) != 1 {
		panic(fmt.Sprintf("primaryIntColumn: table %q has %d IDColumns; read-path helpers require single-column-key tables", tbl.TableName(), len(entry.IDColumns)))
	}
	return entry.IDColumns[0].(postgres.ColumnInteger)
}

// getActiveFactAt returns the record for an entity that is currently considered
// 'active' at a specific validity time. The active partition is the sole scan
// target; the bool return is true when a row was found and false (with nil error)
// when qrm reports no rows.
func getActiveFactAt[T any](
	ctx context.Context,
	db Executor,
	tbl postgres.Table,
	entityID int64,
	validityTime time.Time,
) (T, bool, error) {
	var dest T

	stmt, err := selectAll(tbl)
	if err != nil {
		var zero T
		return zero, false, err
	}

	idCol := primaryIntColumn(tbl)
	startCol := postgres.TimestampzColumn("validity_start")
	endCol := postgres.TimestampzColumn("validity_end")
	modeCol := postgres.StringColumn("mode")

	err = stmt.
		WHERE(postgres.AND(
			idCol.EQ(postgres.Int(entityID)),
			modeCol.EQ(enum.BitemporalMode.Active),
			AsOf(startCol, endCol, validityTime),
		)).
		LIMIT(1).
		QueryContext(ctx, db, &dest)

	if err != nil {
		if IsNoRows(err) {
			var zero T
			return zero, false, nil
		}
		var zero T
		return zero, false, err
	}
	return dest, true, nil
}

// getHistoricalFactAt returns the record for an entity as it was believed to be
// at a specific point in transaction time, for a specific validity time.
// This query targets the historical hash partitions. The bool return is true when
// a row was found and false (with nil error) when qrm reports no rows.
func getHistoricalFactAt[T any](
	ctx context.Context,
	db Executor,
	tbl postgres.Table,
	entityID int64,
	validityTime time.Time,
	transactionTime time.Time,
) (T, bool, error) {
	var dest T

	stmt, err := selectAll(tbl)
	if err != nil {
		var zero T
		return zero, false, err
	}

	idCol := primaryIntColumn(tbl)
	vStartCol := postgres.TimestampzColumn("validity_start")
	vEndCol := postgres.TimestampzColumn("validity_end")
	txStartCol := postgres.TimestampzColumn("transaction_start")
	txEndCol := postgres.TimestampzColumn("transaction_end")
	modeCol := postgres.StringColumn("mode")

	err = stmt.
		WHERE(postgres.AND(
			idCol.EQ(postgres.Int(entityID)),
			modeCol.EQ(enum.BitemporalMode.Historical),
			AsOf(vStartCol, vEndCol, validityTime),
			BelievedAt(txStartCol, txEndCol, transactionTime),
		)).
		LIMIT(1).
		QueryContext(ctx, db, &dest)

	if err != nil {
		if IsNoRows(err) {
			var zero T
			return zero, false, nil
		}
		var zero T
		return zero, false, err
	}
	return dest, true, nil
}

// getTimelineAsOf returns the full validity timeline (1D slice) for an entity
// as it was believed to be at a specific point in transaction time.
func getTimelineAsOf[T any](
	ctx context.Context,
	db Executor,
	tbl postgres.Table,
	entityID int64,
	transactionTime time.Time,
	dest *[]T,
) error {
	stmt, err := selectAll(tbl)
	if err != nil {
		return err
	}

	idCol := primaryIntColumn(tbl)
	vStartCol := postgres.TimestampzColumn("validity_start")
	txStartCol := postgres.TimestampzColumn("transaction_start")
	txEndCol := postgres.TimestampzColumn("transaction_end")

	return stmt.
		WHERE(postgres.AND(
			idCol.EQ(postgres.Int(entityID)),
			BelievedAt(txStartCol, txEndCol, transactionTime),
		)).
		ORDER_BY(vStartCol.ASC()).
		QueryContext(ctx, db, dest)
}

// getAuditTrailFor returns the full history of assertions (belief history)
// for an entity within the validity window that contains effectiveT, ordered
// by transaction time. Locking the validity axis prevents the caller from
// accidentally reading cross-window audit rows in a single query.
func getAuditTrailFor[T any](
	ctx context.Context,
	db Executor,
	tbl postgres.Table,
	entityID int64,
	effectiveT time.Time,
	dest *[]T,
) error {
	stmt, err := selectAll(tbl)
	if err != nil {
		return err
	}

	idCol := primaryIntColumn(tbl)
	txStartCol := postgres.TimestampzColumn("transaction_start")
	vStartCol := postgres.TimestampzColumn("validity_start")
	vEndCol := postgres.TimestampzColumn("validity_end")

	return stmt.
		WHERE(postgres.AND(
			idCol.EQ(postgres.Int(entityID)),
			AsOf(vStartCol, vEndCol, effectiveT),
		)).
		ORDER_BY(txStartCol.ASC(), vStartCol.ASC()).
		QueryContext(ctx, db, dest)
}
