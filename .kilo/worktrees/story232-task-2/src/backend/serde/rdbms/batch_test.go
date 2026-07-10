package rdbms

import (
	"context"
	"fmt"
	"testing"
	"time"

	c "chronocone/backend/constants"
	"chronocone/backend/serde/rdbms/gen/postgres/public/model"
	"chronocone/backend/serde/rdbms/gen/postgres/public/table"
	"github.com/stretchr/testify/require"
)

func TestBatchBitemporalImport(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	// 1. Prepare 10,000 entities using 1k batches for setup
	count := 10_000
	batchSize := 1000
	now := time.Now().UTC().Truncate(time.Second)

	// Generate random base ID for batch to avoid collisions across test runs
	// We use a base within the reserved range and add offsets for uniqueness
	baseID := randomTestID(t)
	baseKeyPrefix := randomTestKey(t)

	knowns := make([]nfWrapper, count)

	for i := 0; i < count; i += batchSize {
		end := i + batchSize
		if end > count { end = count }

		stmt := table.NamespaceFamily.INSERT(
			table.NamespaceFamily.NsFamID,
			table.NamespaceFamily.CustomerSubjectID,
			table.NamespaceFamily.OwnerSubjectID,
			table.NamespaceFamily.NsFamTypeID,
			table.NamespaceFamily.Key,
			table.NamespaceFamily.ValidityStart,
			table.NamespaceFamily.ValidityEnd,
			table.NamespaceFamily.TransactionStart,
			table.NamespaceFamily.InsertionMetadata,
		)

		for j := i; j < end; j++ {
			// Use random base ID + offset to ensure uniqueness while randomizing
			id := baseID + int64(j)
			key := fmt.Sprintf("%s-%d", baseKeyPrefix, j)
			
			// We set validity_start to 'now'
			stmt = stmt.VALUES(id, -1, -1, 1, key, now, c.InfinityTime, now, "{}")
			
			knowns[j] = nfWrapper{
				&model.NamespaceFamily{
					NsFamID:           int32(id),
					CustomerSubjectID: -1,
					OwnerSubjectID:    -1,
					NsFamTypeID:       1,
					Key:               key,
					ValidityStart:     now,
					ValidityEnd:       c.InfinityTime,
					TransactionStart:  now,
					InsertionMetadata: "{}",
				},
			}
		}
		
		_, err = overridingInsert{stmt}.ExecContext(ctx, tx)
		require.NoError(t, err)
	}

	// 2. Perform a massive batch mutation using the DEFAULT ASSERTOR (nil builder)
	// We update the local models with a new validity date to simulate business logic changes.
	newValidity := now.Add(time.Hour)
	for i := range knowns {
		knowns[i].NamespaceFamily.ValidityStart = newValidity
	}

	// Passing nil builder triggers the DefaultBitemporalAssertor which batches by 1k.
	_, err = ExecuteBitemporalMutation(ctx, tx, now, &newValidity, knowns, nil)
	require.NoError(t, err)

	// 3. Verify distribution
	var h0Count, h1Count int64
	err = tx.QueryRow("SELECT count(*) FROM internal_partitions.namespace_family_historical_0").Scan(&h0Count)
	require.NoError(t, err)
	err = tx.QueryRow("SELECT count(*) FROM internal_partitions.namespace_family_historical_1").Scan(&h1Count)
	require.NoError(t, err)

	t.Logf("Rows in historical_0: %d, historical_1: %d", h0Count, h1Count)
	require.True(t, h0Count > 0)
	require.True(t, h1Count > 0)
	require.Equal(t, int64(count), h0Count+h1Count)
}
