package domain

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"chronocone/backend/serde/rdbms"
	"chronocone/backend/serde/rdbms/gen/postgres/public/table"
	"github.com/go-jet/jet/v2/postgres"
)

func init() {
	rdbms.RegisterEntity(
		reflect.TypeOf(NamespaceFamily{}),
		table.NamespaceFamily,
		rdbms.TableTypeNamespaceFamily,
		table.NamespaceFamily.NsFamID,
	)
}

// CreateOrUpdateFamily creates a new namespace family or updates an existing
// one. If a family with the given key exists, it is updated; otherwise a new
// one is created. effectiveTime and systemTime allow querying at specific
// points in bitemporal time.
func CreateOrUpdateFamily(
	ctx context.Context,
	db rdbms.Executor,
	key string,
	validFrom, validTo time.Time,
	effectiveTime, systemTime *time.Time,
) (*NamespaceFamily, error) {
	// Check if a family with this key already exists in the active partition.
	var existingList []struct {
		NsFamID int32
	}
	stmt := postgres.SELECT(table.NamespaceFamily.NsFamID).
		FROM(table.NamespaceFamily).
		WHERE(postgres.AND(
			table.NamespaceFamily.Key.EQ(postgres.String(key)),
			rdbms.IsActive(),
		))

	if err := stmt.QueryContext(ctx, db, &existingList); err != nil {
		return nil, fmt.Errorf("query by key failed: %w", err)
	}

	var id int32
	if len(existingList) > 0 {
		id = existingList[0].NsFamID
	}

	fam := NamespaceFamily{
		NsFamID:           id,
		NsFamTypeID:       1,  // TODO: determine type
		CustomerSubjectID: -1, // TODO: from auth context
		OwnerSubjectID:    -1, // TODO: from auth context
		InsertionMetadata: "{}",
	}
	fam.Key = key
	fam.ValidityStart = validFrom
	fam.ValidityEnd = validTo

	// Pass zero maxTransactionStart to skip staleness detection; EBM uses
	// PostgreSQL CURRENT_TIMESTAMP for all transaction-time boundaries.
	result, err := rdbms.ExecuteBitemporalMutation(ctx, db, time.Time{}, nil, []NamespaceFamily{fam}, nil)
	if err != nil {
		return nil, fmt.Errorf("mutation failed: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("mutation returned no result")
	}

	return &result[0], nil
}

func ListFamilies(ctx context.Context, db rdbms.Executor, effectiveTime, systemTime *time.Time) ([]NamespaceFamily, error) {
	var families []NamespaceFamily
	stmt := postgres.SELECT(table.NamespaceFamily.AllColumns).
		FROM(table.NamespaceFamily).
		WHERE(rdbms.ActiveAt(effectiveTime, systemTime))

	if err := stmt.QueryContext(ctx, db, &families); err != nil {
		return nil, fmt.Errorf("ListFamilies: %w", err)
	}
	return families, nil
}

// GetFamilyByKey retrieves an active namespace family by its natural key at
// the given effective and system time. Returns nil if no active family
// matches.
func GetFamilyByKey(
	ctx context.Context,
	db rdbms.Executor,
	key string,
	effectiveTime, systemTime *time.Time,
) (*NamespaceFamily, error) {
	var families []NamespaceFamily
	stmt := postgres.SELECT(table.NamespaceFamily.AllColumns).
		FROM(table.NamespaceFamily).
		WHERE(postgres.AND(
			table.NamespaceFamily.Key.EQ(postgres.String(key)),
			rdbms.ActiveAt(effectiveTime, systemTime),
		))

	if err := stmt.QueryContext(ctx, db, &families); err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	if len(families) == 0 {
		return nil, nil
	}

	return &families[0], nil
}
