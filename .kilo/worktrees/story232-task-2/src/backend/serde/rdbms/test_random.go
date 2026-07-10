package rdbms

import (
	"testing"

	"github.com/google/uuid"
)

// randomTestKey returns a deterministic test key with UUID suffix.
// Keys are 6-64 characters and suitable for use in database tests.
func randomTestKey(t *testing.T) string {
	u := uuid.New().String()
	// Use "key-" prefix (4 chars) + first 8 chars of UUID = 12 chars total
	key := "key-" + u[:8]
	if len(key) > 64 {
		t.Fatalf("test key exceeds 64 characters: %d", len(key))
	}
	return key
}

// randomTestID returns an ID in the reserved test range [1_000_000, 1_999_999].
// These IDs avoid collision with seed fixtures and other test data.
func randomTestID(t *testing.T) int64 {
	u := uuid.New().ID()
	// Map UUID's 32-bit value into our reserved range
	min := int64(1_000_000)
	max := int64(1_999_999)
	idRange := max - min
	return min + (int64(u) % idRange)
}
