package rdbms

import (
	"context"
	"reflect"
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

// nfWrapper is a test fixture that embeds the generated model so EBM-direct
// tests can construct LockableEntity values without depending on the domain
// entities package (which would create an import cycle). qrm hydrates the
// embedded *model.NamespaceFamily directly — its type name normalizes to
// "namespacefamily" and aligns with the namespace_family table.
type nfWrapper struct {
	*model.NamespaceFamily
}

func (w nfWrapper) EntityID() int64    { return int64(w.NsFamID) }
func (w nfWrapper) NaturalKey() string { return w.Key }

func init() {
	BitemporalTableRegistry[reflect.TypeOf(nfWrapper{})] = registryEntry(
		table.NamespaceFamily, TableTypeNamespaceFamily,
		table.NamespaceFamily.NsFamID,
	)
}

func TestExecuteBitemporalMutation(t *testing.T) {
	ctx := context.Background()

	t.Run("StalenessDetection", func(t *testing.T) {
		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		// 1. Insert initial record with random key and ID
		testID := randomTestID(t)
		testKey := randomTestKey(t)
		now := time.Now().UTC().Truncate(time.Second)
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
		).VALUES(
			testID, -1, -1, 1, testKey, now, c.InfinityTime, now, c.InfinityTime, model.BitemporalMode_Active, "{}",
		)

		_, err = overridingInsert{stmt}.ExecContext(ctx, tx)
		require.NoError(t, err)

		// Verify the row was persisted with the correct key and ID
		var persisted []model.NamespaceFamily
		err = table.NamespaceFamily.SELECT(table.NamespaceFamily.Key, table.NamespaceFamily.NsFamID).
			WHERE(table.NamespaceFamily.NsFamID.EQ(postgres.Int(testID))).
			QueryContext(ctx, tx, &persisted)
		require.NoError(t, err)
		require.Len(t, persisted, 1)
		require.Equal(t, testKey, persisted[0].Key)
		require.Equal(t, int32(testID), persisted[0].NsFamID)

		known := nfWrapper{
			&model.NamespaceFamily{
				NsFamID:           int32(testID),
				CustomerSubjectID: -1,
				OwnerSubjectID:    -1,
				NsFamTypeID:       1,
				Key:               testKey,
				TransactionStart:  now,
			},
		}

		// 2. Simulate someone else updating it (by creating a newer record in the past)
		later := now.Add(time.Second)
		_, err = table.NamespaceFamily.UPDATE(table.NamespaceFamily.TransactionStart).
			SET(postgres.TimestampzT(later)).
			WHERE(table.NamespaceFamily.NsFamID.EQ(postgres.Int(testID))).
			ExecContext(ctx, tx)
		require.NoError(t, err)

		// 3. Attempt mutation with the OLD 'now' timestamp
		_, err = ExecuteBitemporalMutation(ctx, tx, now, nil, []nfWrapper{known},
			func(fresh []nfWrapper, eff postgres.Expression) []postgres.InsertStatement {
				return nil
			})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stale")
	})

	t.Run("KeyImmutabilityViolation", func(t *testing.T) {
		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		testID := randomTestID(t)
		originalKey := randomTestKey(t)
		now := time.Now().UTC().Truncate(time.Second)
		stmt := table.NamespaceFamily.INSERT(
			table.NamespaceFamily.NsFamID,
			table.NamespaceFamily.CustomerSubjectID,
			table.NamespaceFamily.OwnerSubjectID,
			table.NamespaceFamily.Key,
			table.NamespaceFamily.NsFamTypeID,
			table.NamespaceFamily.TransactionStart,
			table.NamespaceFamily.Mode,
			table.NamespaceFamily.InsertionMetadata,
		).VALUES(testID, -1, -1, originalKey, 1, now, model.BitemporalMode_Active, "{}")

		_, err = overridingInsert{stmt}.ExecContext(ctx, tx)
		require.NoError(t, err)

		// Verify the row was persisted
		var persisted []model.NamespaceFamily
		err = table.NamespaceFamily.SELECT(table.NamespaceFamily.Key, table.NamespaceFamily.NsFamID).
			WHERE(table.NamespaceFamily.NsFamID.EQ(postgres.Int(testID))).
			QueryContext(ctx, tx, &persisted)
		require.NoError(t, err)
		require.Len(t, persisted, 1)
		require.Equal(t, originalKey, persisted[0].Key)

		changedKey := randomTestKey(t) // Different key - illegal
		known := nfWrapper{
			&model.NamespaceFamily{
				NsFamID:           int32(testID),
				Key:              changedKey, // ILLEGAL CHANGE
				TransactionStart: now,
			},
		}

		_, err = ExecuteBitemporalMutation(ctx, tx, now.Add(time.Hour), nil, []nfWrapper{known}, func(fresh []nfWrapper, eff postgres.Expression) []postgres.InsertStatement {
			return nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key immutability violation")
	})

	t.Run("FutureDatedMutation", func(t *testing.T) {
		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		testID := randomTestID(t)
		testKey := randomTestKey(t)
		now := time.Now().UTC().Truncate(time.Second)
		stale := now.Add(-time.Hour) // clearly-old seed transaction_start
		tomorrow := now.Add(24 * time.Hour)

		// Insert initial: [now, infinity), transaction_start=stale
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
		).VALUES(testID, -1, -1, 1, testKey, now, parseTime("3000-01-01T00:00:00Z"), stale, model.BitemporalMode_Active, "{}")
		_, err = overridingInsert{stmt}.ExecContext(ctx, tx)
		require.NoError(t, err)

		// Verify persistence
		var persisted []model.NamespaceFamily
		err = table.NamespaceFamily.SELECT(table.NamespaceFamily.Key, table.NamespaceFamily.NsFamID).
			WHERE(table.NamespaceFamily.NsFamID.EQ(postgres.Int(testID))).
			QueryContext(ctx, tx, &persisted)
		require.NoError(t, err)
		require.Len(t, persisted, 1)
		require.Equal(t, testKey, persisted[0].Key)

		known := nfWrapper{
			&model.NamespaceFamily{
				NsFamID:           int32(testID),
				CustomerSubjectID: -1,
				OwnerSubjectID:    -1,
				NsFamTypeID:       1,
				Key:               testKey,
				TransactionStart:  stale,
			},
		}

		// Mutate effective tomorrow
		_, err = ExecuteBitemporalMutation(ctx, tx, stale, &tomorrow, []nfWrapper{known},
			func(fresh []nfWrapper, eff postgres.Expression) []postgres.InsertStatement {
				return []postgres.InsertStatement{
					table.NamespaceFamily.INSERT(
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
					).VALUES(testID, -1, -1, 1, testKey, eff, c.InfinityTime, postgres.CURRENT_TIMESTAMP(), model.BitemporalMode_Active, "{}"),
				}
			})
		require.NoError(t, err)

		// Verify truncation: Old row should now end at tomorrow
		var rows []model.NamespaceFamily
		selectStmt := table.NamespaceFamily.SELECT(
			table.NamespaceFamily.ValidityStart,
			table.NamespaceFamily.ValidityEnd,
			table.NamespaceFamily.TransactionStart,
		).
			WHERE(postgres.AND(
				table.NamespaceFamily.NsFamID.EQ(postgres.Int(testID)),
				table.NamespaceFamily.Mode.EQ(enum.BitemporalMode.Active),
			)).
			ORDER_BY(table.NamespaceFamily.ValidityStart.ASC())
		err = selectStmt.QueryContext(ctx, tx, &rows)
		assert.NoError(t, err)
		require.Equal(t, 2, len(rows))

		assert.True(t, rows[0].ValidityEnd.Equal(tomorrow), "Original row should be truncated at tomorrow")
		assert.True(t, rows[1].ValidityStart.Equal(tomorrow), "New row should start tomorrow")

		// Bitemporal honesty: the truncated original row's transaction_start
		// must be refreshed by EBM, not the seeded `stale` value. Otherwise
		// a UV point at U in [stale, mutation_time) would see both the
		// truncated active row and its historical predecessor.
		assert.True(t, rows[0].TransactionStart.After(stale.Add(30*time.Minute)),
			"truncated active row's transaction_start (%v) should be refreshed to ~now, not the seeded stale value (%v)",
			rows[0].TransactionStart, stale)
	})

	t.Run("SupersedeFutureRecord", func(t *testing.T) {
		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		testID := randomTestID(t)
		testKey := randomTestKey(t)
		t0 := time.Now().UTC().Truncate(time.Second)
		t1 := t0.Add(time.Hour)
		t2 := t0.Add(2 * time.Hour)

		// Setup: [t0, t2) and [t2, infinity)
		stmt1 := table.NamespaceFamily.INSERT(
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
		).VALUES(testID, -1, -1, 1, testKey, t0, t2, t0, model.BitemporalMode_Active, "{}")
		stmt2 := table.NamespaceFamily.INSERT(
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
		).VALUES(testID, -1, -1, 1, testKey, t2, parseTime("3000-01-01T00:00:00Z"), t0, model.BitemporalMode_Active, "{}")

		_, err = overridingInsert{stmt1}.ExecContext(ctx, tx)
		require.NoError(t, err)
		_, err = overridingInsert{stmt2}.ExecContext(ctx, tx)
		require.NoError(t, err)

		// Verify persistence
		var persisted []model.NamespaceFamily
		err = table.NamespaceFamily.SELECT(table.NamespaceFamily.Key, table.NamespaceFamily.NsFamID).
			WHERE(table.NamespaceFamily.NsFamID.EQ(postgres.Int(testID))).
			QueryContext(ctx, tx, &persisted)
		require.NoError(t, err)
		require.Len(t, persisted, 1)
		require.Equal(t, testKey, persisted[0].Key)

		known := nfWrapper{
			&model.NamespaceFamily{
				NsFamID:           int32(testID),
				CustomerSubjectID: -1,
				OwnerSubjectID:    -1,
				NsFamTypeID:       1,
				Key:               testKey,
				TransactionStart:  t0,
			},
		}

		// Update effective T1 (now < T1 < T2)
		// This should truncate [t0, t2) to [t0, t1)
		// And it should effectively "close" [t2, inf) by setting its end to t1
		_, err = ExecuteBitemporalMutation(ctx, tx, t0, &t1, []nfWrapper{known},
			func(fresh []nfWrapper, eff postgres.Expression) []postgres.InsertStatement {
				return []postgres.InsertStatement{
					table.NamespaceFamily.INSERT(
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
					).VALUES(testID, -1, -1, 1, testKey, eff, parseTime("3000-01-01T00:00:00Z"), postgres.CURRENT_TIMESTAMP(), model.BitemporalMode_Active, "{}"),
				}
			})
		require.NoError(t, err)

		var rows []model.NamespaceFamily
		selectStmt := table.NamespaceFamily.SELECT(table.NamespaceFamily.ValidityStart, table.NamespaceFamily.ValidityEnd).
			WHERE(postgres.AND(
				table.NamespaceFamily.NsFamID.EQ(postgres.Int(testID)),
				table.NamespaceFamily.Mode.EQ(enum.BitemporalMode.Active),
			)).
			ORDER_BY(table.NamespaceFamily.ValidityStart.ASC(), table.NamespaceFamily.ValidityEnd.ASC())
		err = selectStmt.QueryContext(ctx, tx, &rows)
		assert.NoError(t, err)

		// Expected active rows:
		// 1. [t0, t1) - truncated original R1
		// 2. [t1, inf) - new assertion R3 (R2 superseded; preserved in historical)
		require.Equal(t, 2, len(rows))

		assert.True(t, rows[0].ValidityEnd.Equal(t1), "Original row R1 should be truncated at t1")
		assert.True(t, rows[1].ValidityStart.Equal(t1), "New row R3 should start at t1")

		// Bitemporal honesty: R2 ([t2, ∞)) was a pre-existing future belief
		// that must have been flushed to historical with a capped
		// transaction_end before being purged from active. Without this,
		// the §3 predicate would return zero matches at any UV point that
		// R2 originally covered.
		var historical []model.NamespaceFamily
		histStmt := table.NamespaceFamily.SELECT(
			table.NamespaceFamily.ValidityStart,
			table.NamespaceFamily.ValidityEnd,
			table.NamespaceFamily.TransactionStart,
			table.NamespaceFamily.TransactionEnd,
		).
			WHERE(postgres.AND(
				table.NamespaceFamily.NsFamID.EQ(postgres.Int(testID)),
				table.NamespaceFamily.Mode.EQ(enum.BitemporalMode.Historical),
				table.NamespaceFamily.ValidityStart.EQ(postgres.TimestampzT(t2)),
			)).
			ORDER_BY(table.NamespaceFamily.TransactionStart.ASC())
		err = histStmt.QueryContext(ctx, tx, &historical)
		require.NoError(t, err)
		require.Equal(t, 1, len(historical),
			"R2 [t2, ∞) must be preserved in historical with capped transaction_end")
		assert.True(t, historical[0].ValidityEnd.Equal(parseTime("3000-01-01T00:00:00Z")),
			"historical R2 should preserve original validity_end (∞)")
		assert.True(t, historical[0].TransactionEnd.Before(parseTime("3000-01-01T00:00:00Z")),
			"historical R2's transaction_end must be capped (got %v)", historical[0].TransactionEnd)
	})

	t.Run("PastDatedGuard", func(t *testing.T) {
		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		testID := randomTestID(t)
		testKey := randomTestKey(t)
		past := time.Now().Add(-time.Hour)
		known := nfWrapper{&model.NamespaceFamily{NsFamID:           int32(testID), Key: testKey}}

		_, err = ExecuteBitemporalMutation(ctx, tx, time.Now(), &past, []nfWrapper{known}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be at or after current time")
	})
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
