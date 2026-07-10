package rdbms

import (
	"context"
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

// seedNsFam inserts a namespace_family row using OVERRIDING SYSTEM VALUE so
// callers may pin the bitemporal triad explicitly. Used by the bitemporal
// honesty tests below to set up scenarios that the live INSERT path cannot.
func seedNsFam(t *testing.T, ctx context.Context, db Executor, id int64, key string, vs, ve, txStart time.Time) {
	t.Helper()
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
	).VALUES(id, -1, -1, 1, key, vs, ve, txStart, model.BitemporalMode_Active, "{}")
	_, err := overridingInsert{stmt}.ExecContext(ctx, db)
	require.NoError(t, err)
}

// nsFamInsertion produces the new-assertion INSERT a builder hands back to EBM.
// It always asserts [eff, ∞) at CURRENT_TIMESTAMP, mirroring how a real caller
// would emit a fresh belief.
func nsFamInsertion(id int64, key string, eff postgres.Expression) postgres.InsertStatement {
	return table.NamespaceFamily.INSERT(
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
	).VALUES(id, -1, -1, 1, key, eff, c.InfinityTime, postgres.CURRENT_TIMESTAMP(), model.BitemporalMode_Active, "{}")
}

// TestEBM_FutureRecordIsFlushedBeforePurge pins the data-loss bug in the
// FLUSH/PURGE pair: a future-dated active row [t2, ∞) is currently dropped
// outright by PURGE because the FLUSH filter (validity_start < eff AND
// validity_end > eff) excludes any row whose validity_start is at or after
// the mutation's effective time. The correct pipeline must preserve the
// original belief in the historical partition before purging it.
func TestEBM_FutureRecordIsFlushedBeforePurge(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	testID := randomTestID(t)
	testKey := randomTestKey(t)
	now := time.Now().UTC().Truncate(time.Second)
	t0 := now.Add(-time.Hour)    // belief formation time (in the past)
	t1 := now.Add(time.Hour)     // mutation effective time (future)
	t2 := now.Add(2 * time.Hour) // pre-existing future row's validity_start

	entityID := testID
	key := testKey

	// Seed only a future-dated active row [t2, ∞), tx_start=t0.
	seedNsFam(t, ctx, tx, entityID, key, t2, c.InfinityTime, t0)

	known := nfWrapper{&model.NamespaceFamily{
		NsFamID:           int32(entityID),
		CustomerSubjectID: -1,
		OwnerSubjectID:    -1,
		NsFamTypeID:       1,
		Key:               key,
		TransactionStart:  t0,
	}}

	_, err = ExecuteBitemporalMutation(ctx, tx, t0, &t1, []nfWrapper{known},
		func(_ []nfWrapper, eff postgres.Expression) []postgres.InsertStatement {
			return []postgres.InsertStatement{nsFamInsertion(entityID, key, eff)}
		})
	require.NoError(t, err)

	var rows []model.NamespaceFamily
	err = postgres.SELECT(table.NamespaceFamily.AllColumns).
		FROM(table.NamespaceFamily).
		WHERE(postgres.AND(
			table.NamespaceFamily.NsFamID.EQ(postgres.Int(entityID)),
			table.NamespaceFamily.Mode.EQ(enum.BitemporalMode.Historical),
		)).
		QueryContext(ctx, tx, &rows)
	require.NoError(t, err)
	require.Equal(t, 1, len(rows),
		"expected exactly one historical row preserving the pre-existing future belief")
	assert.True(t, rows[0].ValidityStart.Equal(t2),
		"historical row's validity_start should equal the original t2 (got %v)", rows[0].ValidityStart)
	assert.True(t, rows[0].ValidityEnd.Equal(c.InfinityTime),
		"historical row's validity_end should equal the original ∞ (got %v)", rows[0].ValidityEnd)
	assert.WithinDuration(t, time.Now().UTC(), rows[0].TransactionEnd, 30*time.Second,
		"historical row's transaction_end should be approximately now")
}

// TestEBM_ActiveRowsDoNotLieAboutTransactionStart pins the bitemporal-honesty
// invariant for the in-place TRUNCATE: when EBM shortens an active row's
// validity interval, that row's transaction_start must be refreshed to the
// moment the new belief (the truncated interval) was formed. The current
// implementation rewrites only validity_end, leaving the original
// transaction_start in place — an active row that lies about when its
// current belief began.
func TestEBM_ActiveRowsDoNotLieAboutTransactionStart(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	testID := randomTestID(t)
	testKey := randomTestKey(t)
	testStart := time.Now().UTC()
	now := testStart.Truncate(time.Second)
	stale := now.Add(-time.Hour)
	t1 := now.Add(time.Hour)

	entityID := testID
	key := testKey

	seedNsFam(t, ctx, tx, entityID, key, stale, c.InfinityTime, stale)

	known := nfWrapper{&model.NamespaceFamily{
		NsFamID:           int32(entityID),
		CustomerSubjectID: -1,
		OwnerSubjectID:    -1,
		NsFamTypeID:       1,
		Key:               key,
		TransactionStart:  stale,
	}}

	_, err = ExecuteBitemporalMutation(ctx, tx, stale, &t1, []nfWrapper{known},
		func(_ []nfWrapper, eff postgres.Expression) []postgres.InsertStatement {
			return []postgres.InsertStatement{nsFamInsertion(entityID, key, eff)}
		})
	require.NoError(t, err)

	const epsilon = 30 * time.Second
	threshold := testStart.Add(-epsilon)
	var rows []model.NamespaceFamily
	err = postgres.SELECT(table.NamespaceFamily.AllColumns).
		FROM(table.NamespaceFamily).
		WHERE(postgres.AND(
			table.NamespaceFamily.NsFamID.EQ(postgres.Int(entityID)),
			table.NamespaceFamily.Mode.EQ(enum.BitemporalMode.Active),
		)).
		QueryContext(ctx, tx, &rows)
	require.NoError(t, err)
	require.NotEmpty(t, rows, "expected at least one active row after mutation")
	for _, r := range rows {
		assert.False(t, r.TransactionStart.Before(threshold),
			"active row [%v, %v) carries stale transaction_start %v (threshold %v); the row's current belief was formed by this mutation and must say so",
			r.ValidityStart, r.ValidityEnd, r.TransactionStart, threshold)
	}
}

// TestEBM_BeliefAtUVPointIsUnique is the §3 unicity check across both
// partitions. After a revision, every (validity_T, transaction_U) point in
// the historical region should resolve to exactly one row. The combination
// of (a) preserving every superseded belief in historical and (b) refreshing
// the active row's transaction_start on TRUNCATE is what guarantees this:
// without (b) the historical and active rows overlap on U for U in
// [tx0, mutation_time).
func TestEBM_BeliefAtUVPointIsUnique(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	testID := randomTestID(t)
	testKey := randomTestKey(t)
	now := time.Now().UTC().Truncate(time.Second)
	tx0 := now.Add(-time.Hour)        // original belief formation
	t0 := now.Add(-30 * time.Minute)  // original validity start
	t1 := now.Add(time.Hour)          // revision effective time

	entityID := testID
	key := testKey

	seedNsFam(t, ctx, tx, entityID, key, t0, c.InfinityTime, tx0)

	known := nfWrapper{&model.NamespaceFamily{
		NsFamID:           int32(entityID),
		CustomerSubjectID: -1,
		OwnerSubjectID:    -1,
		NsFamTypeID:       1,
		Key:               key,
		TransactionStart:  tx0,
	}}

	_, err = ExecuteBitemporalMutation(ctx, tx, tx0, &t1, []nfWrapper{known},
		func(_ []nfWrapper, eff postgres.Expression) []postgres.InsertStatement {
			return []postgres.InsertStatement{nsFamInsertion(entityID, key, eff)}
		})
	require.NoError(t, err)

	// Pick a UV point inside the original belief's validity that was known
	// at U just after tx0 but well before the revision landed.
	queryT := t0.Add(15 * time.Minute)
	queryU := tx0.Add(15 * time.Minute)

	var rows []model.NamespaceFamily
	err = postgres.SELECT(table.NamespaceFamily.AllColumns).
		FROM(table.NamespaceFamily).
		WHERE(postgres.AND(
			table.NamespaceFamily.NsFamID.EQ(postgres.Int(entityID)),
			table.NamespaceFamily.ValidityStart.LT_EQ(postgres.TimestampzT(queryT)),
			table.NamespaceFamily.ValidityEnd.GT(postgres.TimestampzT(queryT)),
			table.NamespaceFamily.TransactionStart.LT_EQ(postgres.TimestampzT(queryU)),
			table.NamespaceFamily.TransactionEnd.GT(postgres.TimestampzT(queryU)),
		)).
		QueryContext(ctx, tx, &rows)
	require.NoError(t, err)
	assert.Equal(t, 1, len(rows),
		"exactly one belief must hold at (T=%v, U=%v); got %d (rows: %+v)",
		queryT, queryU, len(rows), rows)
}
