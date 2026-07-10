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
		reflect.TypeOf(Namespace{}),
		table.Namespace,
		rdbms.TableTypeNamespace,
		table.Namespace.NsID,
	)
}

func ListNamespaces(ctx context.Context, db rdbms.Executor, famID int32, effectiveTime, systemTime *time.Time) ([]Namespace, error) {
	var namespaces []Namespace
	stmt := postgres.SELECT(table.Namespace.AllColumns).
		FROM(table.Namespace).
		WHERE(postgres.AND(
			table.Namespace.NsFamID.EQ(postgres.Int32(famID)),
			rdbms.ActiveAt(effectiveTime, systemTime),
		))

	if err := stmt.QueryContext(ctx, db, &namespaces); err != nil {
		return nil, fmt.Errorf("ListNamespaces: %w", err)
	}
	return namespaces, nil
}

func GetNamespaceByKey(ctx context.Context, db rdbms.Executor, famID int32, key string, effectiveTime, systemTime *time.Time) (*Namespace, error) {
	var namespaces []Namespace
	stmt := postgres.SELECT(table.Namespace.AllColumns).
		FROM(table.Namespace).
		WHERE(postgres.AND(
			table.Namespace.NsFamID.EQ(postgres.Int32(famID)),
			table.Namespace.Key.EQ(postgres.String(key)),
			rdbms.ActiveAt(effectiveTime, systemTime),
		))

	if err := stmt.QueryContext(ctx, db, &namespaces); err != nil {
		return nil, fmt.Errorf("GetNamespaceByKey: %w", err)
	}
	if len(namespaces) == 0 {
		return nil, nil
	}
	return &namespaces[0], nil
}

func CreateOrUpdateNamespace(ctx context.Context, db rdbms.Executor, famID int32, key string, validFrom, validTo time.Time, effectiveTime, systemTime *time.Time) (*Namespace, error) {
	var existingList []struct {
		NsID int64
	}
	stmt := postgres.SELECT(table.Namespace.NsID).
		FROM(table.Namespace).
		WHERE(postgres.AND(
			table.Namespace.NsFamID.EQ(postgres.Int32(famID)),
			table.Namespace.Key.EQ(postgres.String(key)),
			rdbms.IsActive(),
		))

	if err := stmt.QueryContext(ctx, db, &existingList); err != nil {
		return nil, fmt.Errorf("CreateOrUpdateNamespace: query failed: %w", err)
	}

	var id int64
	if len(existingList) > 0 {
		id = existingList[0].NsID
	}

	ns := Namespace{
		NsID:              id,
		NsFamID:           famID,
		OwnerSubjectID:    -1,
		NsTypeID:          1,
		InsertionMetadata: "{}",
	}
	ns.Key = key
	ns.ValidityStart = validFrom
	ns.ValidityEnd = validTo

	result, err := rdbms.ExecuteBitemporalMutation(ctx, db, time.Time{}, nil, []Namespace{ns}, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateOrUpdateNamespace: mutation failed: %w", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("CreateOrUpdateNamespace: mutation returned no result")
	}
	return &result[0], nil
}
