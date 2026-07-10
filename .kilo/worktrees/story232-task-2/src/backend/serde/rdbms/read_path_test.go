package rdbms

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	c "chronocone/backend/constants"
	"chronocone/backend/serde/rdbms/gen/postgres/public/enum"
	"chronocone/backend/serde/rdbms/gen/postgres/public/model"
	"chronocone/backend/serde/rdbms/gen/postgres/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/stretchr/testify/require"
)

// TestgetAuditTrailFor_LocksValidityAxis verifies that getAuditTrailFor
// constrains results to the validity window containing effectiveT. Rows
// belonging to other validity windows must not appear regardless of how many
// mutations the entity has undergone.
func TestGetAuditTrailFor_LocksValidityAxis(t *testing.T) {
	ctx := context.Background()
	nfTbl := table.NamespaceFamily

	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	testKey := randomTestKey(t)
	id := randomTestID(t)

	// Three non-overlapping validity windows expressed with static dates so
	// that the test is independent of wall-clock time.
	T0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) // window 1 start
	T1 := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC) // window 2 start
	T2 := time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC) // window 3 start

	// Transaction-time epochs at which beliefs about these windows were formed.
	txA := time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)
	txB := time.Date(2020, 2, 2, 0, 0, 0, 0, time.UTC)
	txC := time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC)
	txD := time.Date(2020, 2, 4, 0, 0, 0, 0, time.UTC)
	txE := time.Date(2020, 2, 5, 0, 0, 0, 0, time.UTC)
	txF := time.Date(2020, 2, 6, 0, 0, 0, 0, time.UTC)

	// effectiveT falls inside window 1 [T0, T1).
	effectiveT := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)

	insertHistorical := func(vStart, vEnd, txStart, txEnd time.Time) {
		stmt := nfTbl.INSERT(
			nfTbl.NsFamID, nfTbl.CustomerSubjectID, nfTbl.OwnerSubjectID,
			nfTbl.NsFamTypeID, nfTbl.Key,
			nfTbl.ValidityStart, nfTbl.ValidityEnd,
			nfTbl.TransactionStart, nfTbl.TransactionEnd,
			nfTbl.Mode, nfTbl.InsertionMetadata,
		).VALUES(id, -1, -1, 1, testKey,
			vStart, vEnd, txStart, txEnd, enum.BitemporalMode.Historical, "{}")
		_, ie := overridingInsert{stmt}.ExecContext(ctx, tx)
		require.NoError(t, ie)
	}

	// Active row for window 3 [T2, Infinity) — current belief.
	activeStmt := nfTbl.INSERT(
		nfTbl.NsFamID, nfTbl.CustomerSubjectID, nfTbl.OwnerSubjectID,
		nfTbl.NsFamTypeID, nfTbl.Key,
		nfTbl.ValidityStart, nfTbl.ValidityEnd,
		nfTbl.TransactionStart, nfTbl.InsertionMetadata,
	).VALUES(id, -1, -1, 1, testKey, T2, c.InfinityTime, txF, "{}")
	_, err = overridingInsert{activeStmt}.ExecContext(ctx, tx)
	require.NoError(t, err)

	// Window 1 [T0, T1): two successive beliefs.
	insertHistorical(T0, T1, txA, txB)
	insertHistorical(T0, T1, txB, txC)

	// Window 2 [T1, T2): two successive beliefs.
	insertHistorical(T1, T2, txC, txD)
	insertHistorical(T1, T2, txD, txE)

	var trail []model.NamespaceFamily
	err = getAuditTrailFor(ctx, tx, nfTbl, id, effectiveT, &trail)
	require.NoError(t, err)

	// Only the two window-1 rows must be returned.
	require.Len(t, trail, 2, "getAuditTrailFor must return only rows where validity_start <= effectiveT AND validity_end > effectiveT")

	for _, r := range trail {
		require.False(t, r.ValidityStart.After(effectiveT),
			"validity_start %v must be <= effectiveT %v", r.ValidityStart, effectiveT)
		require.True(t, r.ValidityEnd.After(effectiveT),
			"validity_end %v must be > effectiveT %v", r.ValidityEnd, effectiveT)
	}

	// Rows must be ordered by transaction_start ascending.
	for i := 1; i < len(trail); i++ {
		require.False(t, trail[i].TransactionStart.Before(trail[i-1].TransactionStart),
			"transaction_start must be non-decreasing; got %v before %v",
			trail[i].TransactionStart, trail[i-1].TransactionStart)
	}

	// The distinct transaction_start values must be strictly increasing
	// (each assertion in the audit trail was formed at a unique moment).
	require.True(t, trail[1].TransactionStart.After(trail[0].TransactionStart),
		"transaction_start values must be strictly increasing across the audit trail")
}

// TestgetActiveFactAt_NotFoundReturnsFalseNoError verifies that querying for a
// non-existent entity returns (zero, false, nil) — not an error.
func TestGetActiveFactAt_NotFoundReturnsFalseNoError(t *testing.T) {
	ctx := context.Background()
	nfTbl := table.NamespaceFamily

	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	const nonExistentID = int64(999999999)
	result, found, err := getActiveFactAt[model.NamespaceFamily](ctx, tx, nfTbl, nonExistentID, time.Now().UTC())
	require.NoError(t, err)
	require.False(t, found, "found must be false for a non-existent entity")
	require.Equal(t, model.NamespaceFamily{}, result, "result must be the zero value when not found")
}

// TestSelectAll_ErrorsIfTableHasNoAllColumns verifies that selectAll returns a
// non-nil error instead of silently falling back to SELECT * when the table
// value has no AllColumns projection.
func TestSelectAll_ErrorsIfTableHasNoAllColumns(t *testing.T) {
	// postgres.NewTable produces a tableImpl that has no AllColumns field —
	// identical to what the code generator would produce for an empty table.
	tbl := postgres.NewTable("public", "dummy_table", "dummy")
	_, err := selectAll(tbl)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no AllColumns projection")
}

// TestIsValidAt_Determinism verifies that two calls to IsValidAt with the same
// validAt argument produce identical SQL, proving that no ambient clock is used.
func TestIsValidAt_Determinism(t *testing.T) {
	start := postgres.TimestampzColumn("validity_start")
	end := postgres.TimestampzColumn("validity_end")
	validAt := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	nfTbl := table.NamespaceFamily
	sql1, args1 := nfTbl.SELECT(nfTbl.NsFamID).WHERE(IsValidAt(start, end, validAt)).Sql()
	sql2, args2 := nfTbl.SELECT(nfTbl.NsFamID).WHERE(IsValidAt(start, end, validAt)).Sql()

	require.Equal(t, sql1, sql2, "SQL template must be identical across calls")
	require.Equal(t, args1, args2, "SQL arguments must be identical across calls")
}

func TestReadPathHelpers(t *testing.T) {
	ctx := context.Background()
	nfTbl := table.NamespaceFamily

	// 1. Setup: Create an entity with a history
	// We use real-time now to ensure we don't trip the past-dated guard,
	// but we need to ensure some time passes for the transaction window to be valid.
	tStart := time.Now().UTC()
	t1 := tStart.Add(1 * time.Hour)

	key := randomTestKey(t)
	id := randomTestID(t)

	// A. Initial state: Use direct INSERT
	tx1, err := testDB.Begin()
	require.NoError(t, err)
	insertStmt := nfTbl.INSERT(
		nfTbl.NsFamID,
		nfTbl.CustomerSubjectID,
		nfTbl.OwnerSubjectID,
		nfTbl.NsFamTypeID,
		nfTbl.Key,
		nfTbl.ValidityStart,
		nfTbl.ValidityEnd,
		nfTbl.TransactionStart,
		nfTbl.InsertionMetadata,
	).VALUES(id, -1, -1, 1, key, tStart, t1, postgres.CURRENT_TIMESTAMP(), "{}")

	_, err = overridingInsert{insertStmt}.ExecContext(ctx, tx1)
	require.NoError(t, err)

	// Fetch back the transaction_start
	var seed struct {
		TransactionStart time.Time `sql:"transaction_start"`
	}
	err = nfTbl.SELECT(nfTbl.TransactionStart.AS("transaction_start")).
		WHERE(nfTbl.NsFamID.EQ(postgres.Int(id))).
		QueryContext(ctx, tx1, &seed)
	require.NoError(t, err)
	tStart = seed.TransactionStart
	fmt.Printf("SEEDED tStart: %v\n", tStart)
	require.NoError(t, tx1.Commit())

	// Sleep to ensure transaction clock moves
	time.Sleep(200 * time.Millisecond)

	// B. Mutation: Use EBM to supersede the existing record.
	// We MUST use a transaction time that is clearly AFTER the initial tStart.
	// B. Mutation: Use EBM in a NEW transaction
	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	known := nfWrapper{&model.NamespaceFamily{
		NsFamID: int32(id), CustomerSubjectID: -1, OwnerSubjectID: -1, NsFamTypeID: 1,
		Key: key, ValidityStart: tStart, ValidityEnd: t1,
		TransactionStart:  tStart,
		InsertionMetadata: "{}",
	}}

	// Mutate at now + 30m (which is clearly after now and before t1)
	tMid := time.Now().UTC().Add(30 * time.Minute)
	known.NamespaceFamily.ValidityStart = tMid
	known.NamespaceFamily.ValidityEnd = c.InfinityTime
	_, err = ExecuteBitemporalMutation(ctx, tx, tStart, &tMid, []nfWrapper{known}, nil)
	require.NoError(t, err)

	t.Run("getActiveFactAt_Correctness", func(t *testing.T) {
		// Query at tMid + 1 minute (should find the new active record)
		queryTime := tMid.Add(time.Minute)
		dest, found, err := getActiveFactAt[model.NamespaceFamily](ctx, tx, nfTbl, id, queryTime)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, tMid.Format(time.RFC3339), dest.ValidityStart.Format(time.RFC3339))
		require.Equal(t, c.InfinityTime.Format(time.RFC3339), dest.ValidityEnd.Format(time.RFC3339))
	})

	t.Run("getActiveFactAt_PruningProof", func(t *testing.T) {
		// We use EXPLAIN to prove that only the 'active' partition is scanned.
		idColName := registryEntryForTable(nfTbl).IDColumns[0].Name()
		query := fmt.Sprintf(`EXPLAIN SELECT * FROM %s WHERE %s = %d AND mode = 'active' AND validity_start <= '%s' AND validity_end > '%s'`,
			nfTbl.TableName(), idColName, id, tStart.Format(time.RFC3339), tStart.Format(time.RFC3339))

		rows, err := tx.QueryContext(ctx, query)
		require.NoError(t, err)
		defer rows.Close()

		var plan string
		var historicalScanned bool
		for rows.Next() {
			var line string
			rows.Scan(&line)
			plan += line + "\n"
			if strings.Contains(line, "historical") {
				historicalScanned = true
			}
		}

		t.Logf("Execution Plan:\n%s", plan)
		require.False(t, historicalScanned, "The historical partitions should have been pruned from the plan")
	})

	t.Run("getAuditTrailFor", func(t *testing.T) {
		var trail []model.NamespaceFamily
		// effectiveT is inside the first validity window [tStart, tMid) so we
		// expect only rows whose validity window covers that point.
		effectiveT := tStart.Add(15 * time.Minute)
		err := getAuditTrailFor(ctx, tx, nfTbl, id, effectiveT, &trail)
		require.NoError(t, err)

		// Should have 2 entries covering [tStart, tMid):
		// 1. The original active row, now truncated to [tStart, tMid)
		// 2. The historical copy of the original [tStart, t1) row
		require.Equal(t, 2, len(trail))

		for _, r := range trail {
			require.False(t, r.ValidityStart.After(effectiveT),
				"validity_start must be <= effectiveT")
			require.True(t, r.ValidityEnd.After(effectiveT),
				"validity_end must be > effectiveT")
		}

		// Check that we have both modes
		var hasActive, hasHistorical bool
		for _, r := range trail {
			if r.Mode == "active" {
				hasActive = true
			}
			if r.Mode == "historical" {
				hasHistorical = true
			}
		}
		require.True(t, hasActive)
		require.True(t, hasHistorical)
	})

	t.Run("getHistoricalFactAt", func(t *testing.T) {
		// Fetch the historical record to get its transaction window
		var hist []model.NamespaceFamily
		err := selectAllMust(nfTbl).
			WHERE(postgres.AND(
				nfTbl.NsFamID.EQ(postgres.Int(id)),
				nfTbl.Mode.EQ(enum.BitemporalMode.Historical),
			)).QueryContext(ctx, tx, &hist)
		require.NoError(t, err)
		require.NotEmpty(t, hist)
		h := hist[0]

		dest, found, err := getHistoricalFactAt[model.NamespaceFamily](ctx, tx, nfTbl, id, tStart.Add(time.Minute), h.TransactionEnd.Add(-time.Millisecond))
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, tStart.Format(time.RFC3339), dest.ValidityStart.Format(time.RFC3339))
		require.Equal(t, t1.Format(time.RFC3339), dest.ValidityEnd.Format(time.RFC3339))
	})

	t.Run("getTimelineAsOf", func(t *testing.T) {
		var timeline []model.NamespaceFamily
		// Query at a transaction time that is clearly AFTER the mutation
		err := getTimelineAsOf(ctx, tx, nfTbl, id, time.Now().Add(time.Hour), &timeline)
		require.NoError(t, err)

		// Should have 2 active records representing the timeline:
		// 1. [tStart, tMid)
		// 2. [tMid, Infinity)
		require.Equal(t, 2, len(timeline))
		require.Equal(t, tStart.Format(time.RFC3339), timeline[0].ValidityStart.Format(time.RFC3339))
		require.Equal(t, tMid.Format(time.RFC3339), timeline[1].ValidityStart.Format(time.RFC3339))
	})
}

// selectAllMust is a test-only helper that panics rather than returning the
// selectAll error, keeping test setup concise where the table is known-valid.
func selectAllMust(tbl postgres.Table) postgres.SelectStatement {
	stmt, err := selectAll(tbl)
	if err != nil {
		panic(err)
	}
	return stmt
}

func TestHistoricalExclusionConstraint(t *testing.T) {
	ctx := context.Background()
	nfTbl := table.NamespaceFamily
	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	testKey := randomTestKey(t)
	id := randomTestID(t)
	tStart := time.Now().UTC()
	tEnd := tStart.Add(1 * time.Hour)

	// Insert first historical record: [tStart, tEnd)
	insert1 := nfTbl.INSERT(
		nfTbl.NsFamID,
		nfTbl.CustomerSubjectID,
		nfTbl.OwnerSubjectID,
		nfTbl.NsFamTypeID,
		nfTbl.Key,
		nfTbl.ValidityStart,
		nfTbl.ValidityEnd,
		nfTbl.TransactionStart,
		nfTbl.TransactionEnd,
		nfTbl.Mode,
		nfTbl.InsertionMetadata,
	).VALUES(id, -1, -1, 1, testKey, tStart, tEnd, tStart, tEnd, enum.BitemporalMode.Historical, "{}")

	_, err = overridingInsert{insert1}.ExecContext(ctx, tx)
	require.NoError(t, err)

	// Attempt to insert an OVERLAPPING historical record in the SAME transaction window
	// Validity: [tStart + 30m, tEnd + 30m) - Overlaps with [tStart, tEnd)
	// Transaction: [tStart, tEnd) - Exactly matches the first record's transaction window
	tOverlapStart := tStart.Add(30 * time.Minute)
	tOverlapEnd := tEnd.Add(30 * time.Minute)

	insert2 := nfTbl.INSERT(
		nfTbl.NsFamID,
		nfTbl.CustomerSubjectID,
		nfTbl.OwnerSubjectID,
		nfTbl.NsFamTypeID,
		nfTbl.Key,
		nfTbl.ValidityStart,
		nfTbl.ValidityEnd,
		nfTbl.TransactionStart,
		nfTbl.TransactionEnd,
		nfTbl.Mode,
		nfTbl.InsertionMetadata,
	).VALUES(id, -1, -1, 1, testKey, tOverlapStart, tOverlapEnd, tStart, tEnd, enum.BitemporalMode.Historical, "{}")

	_, err = overridingInsert{insert2}.ExecContext(ctx, tx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "conflicting key value violates exclusion constraint")
}
