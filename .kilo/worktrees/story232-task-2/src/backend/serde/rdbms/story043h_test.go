package rdbms

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	c "chronocone/backend/constants"
	"chronocone/backend/serde/rdbms/gen/postgres/public/enum"
	"chronocone/backend/serde/rdbms/gen/postgres/public/model"
	"chronocone/backend/serde/rdbms/gen/postgres/public/table"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- MetricDim LockableEntity wrapper ---

type metricDimWrapper struct {
	*model.MetricDim
}

func (w metricDimWrapper) EntityID() int64 {
	return w.MetricDim.MetricID
}

func (w metricDimWrapper) NaturalKey() string {
	return fmt.Sprintf("%d:%d", w.MetricDim.MetricID, w.MetricDim.DimensionID)
}

func init() {
	BitemporalTableRegistry[reflect.TypeOf(metricDimWrapper{})] = registryEntry(
		table.MetricDim, TableTypeMetricDim,
		table.MetricDim.MetricID, table.MetricDim.DimensionID,
	)
}

// seedMetricDim inserts a metric_dim row using explicit timestamps.
func seedMetricDim(t *testing.T, ctx context.Context, db Executor, metricID, dimID int64, dimOrder int16, vs, ve, txStart time.Time) {
	t.Helper()
	stmt := table.MetricDim.INSERT(
		table.MetricDim.MetricID,
		table.MetricDim.DimensionID,
		table.MetricDim.DimensionOrder,
		table.MetricDim.ValidityStart,
		table.MetricDim.ValidityEnd,
		table.MetricDim.TransactionStart,
		table.MetricDim.InsertionMetadata,
	).VALUES(metricID, dimID, dimOrder, vs, ve, txStart, "{}")
	_, err := overridingInsert{stmt}.ExecContext(ctx, db)
	require.NoError(t, err)
}

// TestEBM_CompositeKeyJunction verifies that FLUSH and PURGE correctly address
// metric_dim rows by (metric_id, dimension_id), not by metric_id alone. If only
// metric_id were used, mutating dimension A would also flush/purge dimension B.
func TestEBM_CompositeKeyJunction(t *testing.T) {
	ctx := context.Background()

	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	metricID := int64(90001)
	dimA := int64(1) // will be mutated
	dimB := int64(2) // must remain untouched

	now := time.Now().UTC().Truncate(time.Second)
	future := now.Add(24 * time.Hour)

	// Seed two rows for the same metric but different dimensions.
	seedMetricDim(t, ctx, tx, metricID, dimA, 1, now, parseTime("3000-01-01T00:00:00Z"), now)
	seedMetricDim(t, ctx, tx, metricID, dimB, 2, now, parseTime("3000-01-01T00:00:00Z"), now)

	// Mutate only dimA via EBM, effective tomorrow.
	knownA := metricDimWrapper{&model.MetricDim{
		MetricID:         metricID,
		DimensionID:      dimA,
		DimensionOrder:   1,
		TransactionStart: now,
	}}

	_, err = ExecuteBitemporalMutation(ctx, tx, now, &future, []metricDimWrapper{knownA},
		func(fresh []metricDimWrapper, eff postgres.Expression) []postgres.InsertStatement {
			return []postgres.InsertStatement{
				table.MetricDim.INSERT(
					table.MetricDim.MetricID,
					table.MetricDim.DimensionID,
					table.MetricDim.DimensionOrder,
					table.MetricDim.ValidityStart,
					table.MetricDim.ValidityEnd,
					table.MetricDim.TransactionStart,
					table.MetricDim.InsertionMetadata,
				).VALUES(metricID, dimA, int16(1), eff, c.InfinityTime, postgres.CURRENT_TIMESTAMP(), "{}"),
			}
		})
	require.NoError(t, err)

	// --- Assertions for dimA (mutated) ---

	var dimAActive []model.MetricDim
	err = table.MetricDim.SELECT(table.MetricDim.AllColumns).
		WHERE(postgres.AND(
			table.MetricDim.MetricID.EQ(postgres.Int(metricID)),
			table.MetricDim.DimensionID.EQ(postgres.Int(dimA)),
			table.MetricDim.Mode.EQ(enum.BitemporalMode.Active),
		)).
		ORDER_BY(table.MetricDim.ValidityStart.ASC()).
		QueryContext(ctx, tx, &dimAActive)
	require.NoError(t, err)

	// Expect exactly 2 active rows: [now, future) and [future, ∞).
	require.Equal(t, 2, len(dimAActive),
		"dimA should have 2 active rows after mutation (truncated original + new assertion)")
	assert.True(t, dimAActive[0].ValidityEnd.Equal(future),
		"dimA original row should be truncated at future")
	assert.True(t, dimAActive[1].ValidityStart.Equal(future),
		"dimA new assertion should start at future")

	// dimA must have a historical row (flushed original).
	var dimAHist []model.MetricDim
	err = table.MetricDim.SELECT(table.MetricDim.AllColumns).
		WHERE(postgres.AND(
			table.MetricDim.MetricID.EQ(postgres.Int(metricID)),
			table.MetricDim.DimensionID.EQ(postgres.Int(dimA)),
			table.MetricDim.Mode.EQ(enum.BitemporalMode.Historical),
		)).
		QueryContext(ctx, tx, &dimAHist)
	require.NoError(t, err)
	require.Equal(t, 1, len(dimAHist),
		"dimA must have exactly one historical row (the flushed original)")

	// --- Assertions for dimB (untouched) ---

	var dimBActive []model.MetricDim
	err = table.MetricDim.SELECT(table.MetricDim.AllColumns).
		WHERE(postgres.AND(
			table.MetricDim.MetricID.EQ(postgres.Int(metricID)),
			table.MetricDim.DimensionID.EQ(postgres.Int(dimB)),
			table.MetricDim.Mode.EQ(enum.BitemporalMode.Active),
		)).
		QueryContext(ctx, tx, &dimBActive)
	require.NoError(t, err)

	// dimB must be completely untouched: 1 active row, no historical rows.
	require.Equal(t, 1, len(dimBActive),
		"dimB must have exactly 1 active row — it was NOT the target of the mutation")
	assert.True(t, dimBActive[0].ValidityEnd.Equal(parseTime("3000-01-01T00:00:00Z")),
		"dimB original validity_end must be unchanged (∞)")

	var dimBHist []model.MetricDim
	err = table.MetricDim.SELECT(table.MetricDim.AllColumns).
		WHERE(postgres.AND(
			table.MetricDim.MetricID.EQ(postgres.Int(metricID)),
			table.MetricDim.DimensionID.EQ(postgres.Int(dimB)),
			table.MetricDim.Mode.EQ(enum.BitemporalMode.Historical),
		)).
		QueryContext(ctx, tx, &dimBHist)
	require.NoError(t, err)
	assert.Equal(t, 0, len(dimBHist),
		"dimB must have zero historical rows — FLUSH must not have touched it")
}

// --- resolveEntityID tests ---

// TestOverridingInsert_ValuesSubstringCollision verifies that overridingInsert.Sql()
// places "OVERRIDING SYSTEM VALUE" before the statement's actual VALUES ( clause
// and not inside a string literal that happens to contain the word "values".
func TestOverridingInsert_ValuesSubstringCollision(t *testing.T) {
	// Build an INSERT whose constant value contains " values (" as a substring.
	// Go-Jet will render this as a $N placeholder; the literal text ends up in
	// the args slice, not the query string itself. But we also inject it as an
	// inline Raw expression so the text appears in the SQL body.
	stmt := table.NamespaceFamily.INSERT(
		table.NamespaceFamily.NsFamID,
		table.NamespaceFamily.CustomerSubjectID,
		table.NamespaceFamily.OwnerSubjectID,
		table.NamespaceFamily.NsFamTypeID,
		table.NamespaceFamily.Key,
		table.NamespaceFamily.ValidityStart,
		table.NamespaceFamily.ValidityEnd,
		table.NamespaceFamily.TransactionStart,
		table.NamespaceFamily.Mode,
		table.NamespaceFamily.InsertionMetadata,
	).VALUES(
		int64(99999), int64(-1), int64(-1), int64(1),
		// Key contains " values (" — the substring we want to not pivot on.
		postgres.String("tricky values (injection)"),
		postgres.CURRENT_TIMESTAMP(), postgres.CURRENT_TIMESTAMP(), postgres.CURRENT_TIMESTAMP(),
		postgres.String("active"),
		postgres.String("{}"),
	)

	wrapped := overridingInsert{stmt}
	query, _ := wrapped.Sql()

	// "OVERRIDING SYSTEM VALUE" must appear exactly once.
	count := strings.Count(query, "OVERRIDING SYSTEM VALUE")
	assert.Equal(t, 1, count, "OVERRIDING SYSTEM VALUE must appear exactly once in: %s", query)

	// The injected marker must appear before the actual VALUES keyword.
	overridingPos := strings.Index(query, "OVERRIDING SYSTEM VALUE")
	valuesPos := strings.Index(query, "VALUES")
	assert.Less(t, overridingPos, valuesPos,
		"OVERRIDING SYSTEM VALUE (%d) must precede VALUES (%d) in: %s",
		overridingPos, valuesPos, query)
}

// --- resolveEntityID tests ---

// TestresolveEntityID_ActiveHit verifies that an active row is found and returned
// with ResolveActive status.
func TestResolveEntityID_ActiveHit(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	now := time.Now().UTC().Truncate(time.Second)
	id := int64(80001)
	key := "resolve-active-hit"

	stmt := table.NamespaceFamily.INSERT(
		table.NamespaceFamily.NsFamID,
		table.NamespaceFamily.CustomerSubjectID,
		table.NamespaceFamily.OwnerSubjectID,
		table.NamespaceFamily.NsFamTypeID,
		table.NamespaceFamily.Key,
		table.NamespaceFamily.ValidityStart,
		table.NamespaceFamily.ValidityEnd,
		table.NamespaceFamily.TransactionStart,
		table.NamespaceFamily.Mode,
		table.NamespaceFamily.InsertionMetadata,
	).VALUES(id, -1, -1, 1, key, now, c.InfinityTime, now, model.BitemporalMode_Active, "{}")
	_, err = overridingInsert{stmt}.ExecContext(ctx, tx)
	require.NoError(t, err)

	nfTbl := table.NamespaceFamily
	gotID, status, err := resolveEntityID(ctx, tx, nfTbl, key)
	require.NoError(t, err)
	assert.Equal(t, ResolveActive, status)
	assert.Equal(t, id, gotID)
}

// TestresolveEntityID_TombstonedOnly verifies that a row present only in the
// historical partition is found and returned with ResolveTombstoned status.
func TestResolveEntityID_TombstonedOnly(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	now := time.Now().UTC().Truncate(time.Second)
	id := int64(80002)
	key := "resolve-tombstoned-only"
	txEnd := now.Add(time.Minute)

	// Insert directly into historical partition by specifying mode='historical'
	// and a finite transaction_end.
	stmt := table.NamespaceFamily.INSERT(
		table.NamespaceFamily.NsFamID,
		table.NamespaceFamily.CustomerSubjectID,
		table.NamespaceFamily.OwnerSubjectID,
		table.NamespaceFamily.NsFamTypeID,
		table.NamespaceFamily.Key,
		table.NamespaceFamily.ValidityStart,
		table.NamespaceFamily.ValidityEnd,
		table.NamespaceFamily.TransactionStart,
		table.NamespaceFamily.TransactionEnd,
		table.NamespaceFamily.Mode,
		table.NamespaceFamily.InsertionMetadata,
	).VALUES(id, -1, -1, 1, key, now, c.InfinityTime, now, txEnd, model.BitemporalMode_Historical, "{}")
	_, err = overridingInsert{stmt}.ExecContext(ctx, tx)
	require.NoError(t, err)

	nfTbl := table.NamespaceFamily
	gotID, status, err := resolveEntityID(ctx, tx, nfTbl, key)
	require.NoError(t, err)
	assert.Equal(t, ResolveTombstoned, status)
	assert.Equal(t, id, gotID)
}

// TestresolveEntityID_NotFound verifies that a missing key returns ResolveNotFound
// with a nil error.
func TestResolveEntityID_NotFound(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	nfTbl := table.NamespaceFamily
	gotID, status, err := resolveEntityID(ctx, tx, nfTbl, "key-that-does-not-exist-xyz123")
	require.NoError(t, err, "resolveEntityID must return nil error when nothing is found")
	assert.Equal(t, ResolveNotFound, status)
	assert.Equal(t, int64(0), gotID)
}

// TestresolveEntityID_KeyReuse_PrefersActive verifies that when a tombstoned row and
// an active row exist for the same key (but different entity IDs), the active row
// is returned. This tests the key-reuse / reactivation scenario where the old
// single-pass LIMIT 1 without ORDER BY was non-deterministic.
func TestResolveEntityID_KeyReuse_PrefersActive(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	now := time.Now().UTC().Truncate(time.Second)
	key := "reused-key-xyz789"

	// Entity A (tombstoned): historical row, finite transaction_end.
	idA := int64(80010)
	txEnd := now.Add(time.Minute)
	stmtA := table.NamespaceFamily.INSERT(
		table.NamespaceFamily.NsFamID,
		table.NamespaceFamily.CustomerSubjectID,
		table.NamespaceFamily.OwnerSubjectID,
		table.NamespaceFamily.NsFamTypeID,
		table.NamespaceFamily.Key,
		table.NamespaceFamily.ValidityStart,
		table.NamespaceFamily.ValidityEnd,
		table.NamespaceFamily.TransactionStart,
		table.NamespaceFamily.TransactionEnd,
		table.NamespaceFamily.Mode,
		table.NamespaceFamily.InsertionMetadata,
	).VALUES(idA, -1, -1, 1, key, now, c.InfinityTime, now, txEnd, model.BitemporalMode_Historical, "{}")
	_, err = overridingInsert{stmtA}.ExecContext(ctx, tx)
	require.NoError(t, err)

	// Entity B (active): active row with infinite transaction_end and future validity_end.
	idB := int64(80011)
	stmtB := table.NamespaceFamily.INSERT(
		table.NamespaceFamily.NsFamID,
		table.NamespaceFamily.CustomerSubjectID,
		table.NamespaceFamily.OwnerSubjectID,
		table.NamespaceFamily.NsFamTypeID,
		table.NamespaceFamily.Key,
		table.NamespaceFamily.ValidityStart,
		table.NamespaceFamily.ValidityEnd,
		table.NamespaceFamily.TransactionStart,
		table.NamespaceFamily.Mode,
		table.NamespaceFamily.InsertionMetadata,
	).VALUES(idB, -1, -1, 1, key, now, c.InfinityTime, now, model.BitemporalMode_Active, "{}")
	_, err = overridingInsert{stmtB}.ExecContext(ctx, tx)
	require.NoError(t, err)

	nfTbl := table.NamespaceFamily
	gotID, status, err := resolveEntityID(ctx, tx, nfTbl, key)
	require.NoError(t, err)
	assert.Equal(t, ResolveActive, status, "active entity B must be returned, not tombstoned A")
	assert.Equal(t, idB, gotID, "must return the active entity ID (%d), not the tombstoned one (%d)", idB, idA)
}
