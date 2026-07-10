package rdbms

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvisoryLockKey(t *testing.T) {
	t.Run("ValidPacking", func(t *testing.T) {
		key := advisoryLockKey(TableTypeNamespace, 123)
		expected := (int64(TableTypeNamespace) << 54) | 123
		assert.Equal(t, expected, key)
	})

	t.Run("PanicOnNegativeTableType", func(t *testing.T) {
		assert.Panics(t, func() { advisoryLockKey(-1, 123) })
	})

	t.Run("PanicOnOverflowTableType", func(t *testing.T) {
		assert.Panics(t, func() { advisoryLockKey(1024, 123) })
	})

	t.Run("PanicOnNegativeEntityID", func(t *testing.T) {
		assert.Panics(t, func() { advisoryLockKey(TableTypeNamespace, -1) })
	})

	t.Run("PanicOnOverflowEntityID", func(t *testing.T) {
		assert.Panics(t, func() { advisoryLockKey(TableTypeNamespace, 1<<54) })
	})
}

func TestAcquireLocks(t *testing.T) {
	ctx := context.Background()

	t.Run("SuccessSingle", func(t *testing.T) {
		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		key := advisoryLockKey(TableTypeNamespace, 1)
		err = acquireLocks(ctx, tx, []int64{key})
		assert.NoError(t, err)
	})

	t.Run("SuccessMultipleSorted", func(t *testing.T) {
		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		keys := []int64{
			advisoryLockKey(TableTypeNamespace, 10),
			advisoryLockKey(TableTypeNamespace, 5),
			advisoryLockKey(TableTypeCategory, 1),
		}
		err = acquireLocks(ctx, tx, keys)
		assert.NoError(t, err)
	})

	t.Run("PanicOnSecondCall", func(t *testing.T) {
		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		key1 := advisoryLockKey(TableTypeNamespace, 100)
		key2 := advisoryLockKey(TableTypeNamespace, 101)

		err = acquireLocks(ctx, tx, []int64{key1})
		require.NoError(t, err)

		assert.Panics(t, func() {
			err = acquireLocks(ctx, tx, []int64{key2})
		}, "Should panic because transaction already holds locks")
	})

	t.Run("EmptyKeys", func(t *testing.T) {
		tx, err := testDB.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		err = acquireLocks(ctx, tx, []int64{})
		assert.NoError(t, err)
	})
}
