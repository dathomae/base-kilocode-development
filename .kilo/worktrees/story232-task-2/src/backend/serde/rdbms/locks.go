package rdbms

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/lib/pq"
)

// Executor defines the database interaction subset needed for orchestration.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// advisoryLockKey packs a 10-bit table type and a 54-bit entity ID into a 64-bit advisory lock key.
// It panics if tableType or entityID exceed their allotted bits, following the "fail loudly" principle.
func advisoryLockKey(tableType AdvisoryLockTableType, entityID int64) int64 {
	// Assumption: tableType fits in 10 bits (0-1023)
	if tableType < 0 || tableType > 1023 {
		panic(fmt.Sprintf("advisoryLockKey: tableType %d exceeds 10 bits", tableType))
	}
	// Assumption: entityID is positive and fits in 54 bits
	if entityID < 0 {
		panic(fmt.Sprintf("advisoryLockKey: entityID %d is negative", entityID))
	}
	if entityID >= (1 << 54) {
		panic(fmt.Sprintf("advisoryLockKey: entityID %d exceeds 54 bits", entityID))
	}

	return (int64(tableType) << 54) | entityID
}

// acquireLocks acquires PostgreSQL transaction-level advisory locks for the given keys.
// Keys are sorted ascending to prevent deadlocks.
func acquireLocks(ctx context.Context, db Executor, keys []int64) error {
	if len(keys) == 0 {
		return nil
	}

	// Guard against multiple calls in one transaction.
	// All locks must be acquired in a single deterministic call to prevent deadlocks.
	var count int
	// Check for any advisory locks held by the current PID.
	query := "SELECT count(*) FROM pg_locks WHERE locktype = 'advisory' AND pid = pg_backend_pid()"
	err := db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return fmt.Errorf("acquireLocks safety check failed: %w", err)
	}
	if count > 0 {
		panic("acquireLocks: transaction already holds advisory locks. All locks must be acquired in a single call to prevent deadlocks.")
	}

	// Sort keys to prevent deadlocks
	sortedKeys := make([]int64, len(keys))
	copy(sortedKeys, keys)
	sort.Slice(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })

	// Acquire locks in a single round-trip using unnest.
	// pg_advisory_xact_lock is used to ensure locks are released at transaction end.
	_, err = db.ExecContext(ctx, "SELECT pg_advisory_xact_lock(k) FROM unnest($1::bigint[]) AS k", pq.Array(sortedKeys))
	if err != nil {
		return fmt.Errorf("acquireLocks: failed to acquire locks: %w", err)
	}

	return nil
}
