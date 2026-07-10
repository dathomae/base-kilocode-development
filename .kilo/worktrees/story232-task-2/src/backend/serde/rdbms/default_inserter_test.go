package rdbms

import (
	"context"
	"reflect"
	"testing"
	"time"

	c "chronocone/backend/constants"
	"chronocone/backend/serde/rdbms/gen/postgres/public/model"
	"chronocone/backend/serde/rdbms/gen/postgres/public/table"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nfWrapperDefaultInserter struct {
	*model.NamespaceFamily
}

func (w nfWrapperDefaultInserter) EntityID() int64    { return int64(w.NsFamID) }
func (w nfWrapperDefaultInserter) NaturalKey() string { return w.Key }

func init() {
	BitemporalTableRegistry[reflect.TypeOf(nfWrapperDefaultInserter{})] = registryEntry(
		table.NamespaceFamily, TableTypeNamespaceFamily,
		table.NamespaceFamily.NsFamID,
	)
}

func TestDefaultBitemporalInserter(t *testing.T) {
	ctx := context.Background()

	t.Run("ExcludesSystemColumnsFromInsertStatement", func(t *testing.T) {
		// PRINCIPLE: Mode and TransactionEnd must be excluded from the INSERT column list.
		// This enforces the API contract at the software layer, preventing callers from
		// inadvertently controlling bitemporal system fields.

		testKey := randomTestKey(t)
		entity := nfWrapperDefaultInserter{
			&model.NamespaceFamily{
				NsFamID:           0, // new entity
				Key:               testKey,
				ValidityStart:     time.Now().UTC(),
				ValidityEnd:       c.InfinityTime,
				TransactionStart:  time.Now().UTC(),
				CustomerSubjectID: -1,
				OwnerSubjectID:    -1,
				NsFamTypeID:       1,
				InsertionMetadata: "{}",
			},
		}

		stmts := DefaultBitemporalInserter([]nfWrapperDefaultInserter{entity}, nil)
		require.Len(t, stmts, 1)

		// Extract the SQL to verify Mode and TransactionEnd are not in the column list
		sql, _ := stmts[0].Sql()
		sqlStr := sql

		// The INSERT statement should NOT explicitly list mode or transaction_end in its column list
		// (they'll use database defaults instead)
		assert.NotContains(t, sqlStr, "mode", "mode column should not be in INSERT column list")
		assert.NotContains(t, sqlStr, "transaction_end", "transaction_end column should not be in INSERT column list")
	})

	t.Run("IgnoresCallerSetSystemFieldValues", func(t *testing.T) {
		// PRINCIPLE: Even if a caller inadvertently sets Mode or TransactionEnd in the model,
		// DefaultBitemporalInserter must prevent those values from reaching the database.

		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		testKey := randomTestKey(t)
		now := time.Now().UTC().Truncate(time.Second)

		entity := nfWrapperDefaultInserter{
			&model.NamespaceFamily{
				NsFamID:           0,
				Key:               testKey,
				ValidityStart:     now,
				ValidityEnd:       c.InfinityTime,
				TransactionStart:  now,
				CustomerSubjectID: -1,
				OwnerSubjectID:    -1,
				NsFamTypeID:       1,
				InsertionMetadata: "{}",
				// Caller illegally sets these (should be ignored)
				Mode:              model.BitemporalMode_Historical, // WRONG - should be Active
				TransactionEnd:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // WRONG - should be c.InfinityTime
			},
		}

		stmts := DefaultBitemporalInserter([]nfWrapperDefaultInserter{entity}, nil)
		require.Len(t, stmts, 1)

		// Execute the statement
		_, err = overridingInsert{stmts[0]}.ExecContext(ctx, tx)
		require.NoError(t, err, "insert should succeed even with illegal system field values")

		// Verify the database has the CORRECT values, not what the caller set
		var result []model.NamespaceFamily
		err = table.NamespaceFamily.SELECT(
			table.NamespaceFamily.Mode,
			table.NamespaceFamily.TransactionEnd,
		).
			WHERE(table.NamespaceFamily.Key.EQ(postgres.String(testKey))).
			QueryContext(ctx, tx, &result)
		require.NoError(t, err)
		require.Len(t, result, 1)

		// Database defaults must override caller's illegal values
		assert.Equal(t, model.BitemporalMode_Active, result[0].Mode,
			"database should have Mode='active' from DEFAULT, not caller's value")
		assert.True(t, result[0].TransactionEnd.Equal(c.InfinityTime),
			"database should have TransactionEnd=c.InfinityTime from DEFAULT, not caller's value")
	})

	t.Run("IncludesAllNonSystemColumns", func(t *testing.T) {
		// PRINCIPLE: All domain columns must be included in the INSERT statement.
		// Only Mode and TransactionEnd are excluded.

		testKey := randomTestKey(t)
		entity := nfWrapperDefaultInserter{
			&model.NamespaceFamily{
				NsFamID:           0,
				Key:               testKey,
				ValidityStart:     time.Now().UTC(),
				ValidityEnd:       c.InfinityTime,
				TransactionStart:  time.Now().UTC(),
				CustomerSubjectID: 42,
				OwnerSubjectID:    99,
				NsFamTypeID:       2,
				InsertionMetadata: `{"custom":"data"}`,
			},
		}

		stmts := DefaultBitemporalInserter([]nfWrapperDefaultInserter{entity}, nil)
		require.Len(t, stmts, 1)

		sql, _ := stmts[0].Sql()
		sqlStr := sql

		// All domain columns should be present
		assert.Contains(t, sqlStr, "key", "key column should be in INSERT")
		assert.Contains(t, sqlStr, "validity_start", "validity_start should be in INSERT")
		assert.Contains(t, sqlStr, "validity_end", "validity_end should be in INSERT")
		assert.Contains(t, sqlStr, "transaction_start", "transaction_start should be in INSERT")
		assert.Contains(t, sqlStr, "customer_subject_id", "customer_subject_id should be in INSERT")
		assert.Contains(t, sqlStr, "owner_subject_id", "owner_subject_id should be in INSERT")
		assert.Contains(t, sqlStr, "ns_fam_type_id", "ns_fam_type_id should be in INSERT")
		assert.Contains(t, sqlStr, "insertion_metadata", "insertion_metadata should be in INSERT")
	})

	t.Run("BatchesLargeInserts", func(t *testing.T) {
		// PRINCIPLE: DefaultBitemporalInserter should batch large result sets into
		// multiple statements (1000 rows per batch) for efficiency.

		entities := make([]nfWrapperDefaultInserter, 2500)
		for i := 0; i < 2500; i++ {
			entities[i] = nfWrapperDefaultInserter{
				&model.NamespaceFamily{
					NsFamID:           0,
					Key:               randomTestKey(nil), // would panic on real test, but OK for statement generation
					ValidityStart:     time.Now().UTC(),
					ValidityEnd:       c.InfinityTime,
					TransactionStart:  time.Now().UTC(),
					CustomerSubjectID: -1,
					OwnerSubjectID:    -1,
					NsFamTypeID:       1,
					InsertionMetadata: "{}",
				},
			}
		}

		stmts := DefaultBitemporalInserter(entities, nil)

		// 2500 entities should produce 3 statements (1000 + 1000 + 500)
		assert.Equal(t, 3, len(stmts), "2500 entities should produce 3 batched INSERT statements")
	})

	t.Run("PreservesAllDomainColumnValues", func(t *testing.T) {
		// PRINCIPLE: All domain column values set by the caller must be preserved
		// in the INSERT statement (only system columns are excluded).

		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		testKey := randomTestKey(t)
		now := time.Now().UTC().Truncate(time.Second)
		validity := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

		entity := nfWrapperDefaultInserter{
			&model.NamespaceFamily{
				NsFamID:           0,
				Key:               testKey,
				ValidityStart:     validity,
				ValidityEnd:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				TransactionStart:  now,
				CustomerSubjectID: -1,
				OwnerSubjectID:    -1,
				NsFamTypeID:       3,
				InsertionMetadata: `{"test":"metadata"}`,
			},
		}

		stmts := DefaultBitemporalInserter([]nfWrapperDefaultInserter{entity}, nil)
		require.Len(t, stmts, 1)

		_, err = overridingInsert{stmts[0]}.ExecContext(ctx, tx)
		require.NoError(t, err)

		var result []model.NamespaceFamily
		err = table.NamespaceFamily.SELECT(table.NamespaceFamily.AllColumns).
			WHERE(table.NamespaceFamily.Key.EQ(postgres.String(testKey))).
			QueryContext(ctx, tx, &result)
		require.NoError(t, err)
		require.Len(t, result, 1)

		row := result[0]
		assert.Equal(t, testKey, row.Key)
		assert.True(t, row.ValidityStart.Equal(validity), "ValidityStart should be preserved")
		assert.True(t, row.ValidityEnd.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)), "ValidityEnd should be preserved")
		assert.True(t, row.TransactionStart.Equal(now), "TransactionStart should be preserved")
		assert.Equal(t, int32(-1), row.CustomerSubjectID, "CustomerSubjectID should be preserved")
		assert.Equal(t, int32(-1), row.OwnerSubjectID, "OwnerSubjectID should be preserved")
		assert.Equal(t, int16(3), row.NsFamTypeID, "NsFamTypeID should be preserved")
		assert.Equal(t, `{"test": "metadata"}`, row.InsertionMetadata, "InsertionMetadata should be preserved")
	})
}
