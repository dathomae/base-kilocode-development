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
		reflect.TypeOf(Element{}),
		table.Element,
		rdbms.TableTypeElement,
		table.Element.ElementID,
	)
}

func ListElements(ctx context.Context, db rdbms.Executor, catID int64, effectiveTime, systemTime *time.Time) ([]Element, error) {
	var elements []Element
	stmt := postgres.SELECT(table.Element.AllColumns).
		FROM(table.Element).
		WHERE(postgres.AND(
			table.Element.CategoryID.EQ(postgres.Int64(catID)),
			rdbms.ActiveAt(effectiveTime, systemTime),
		))

	if err := stmt.QueryContext(ctx, db, &elements); err != nil {
		return nil, fmt.Errorf("ListElements: %w", err)
	}
	return elements, nil
}

func GetElementByKey(ctx context.Context, db rdbms.Executor, catID int64, key string, effectiveTime, systemTime *time.Time) (*Element, error) {
	var elements []Element
	stmt := postgres.SELECT(table.Element.AllColumns).
		FROM(table.Element).
		WHERE(postgres.AND(
			table.Element.CategoryID.EQ(postgres.Int64(catID)),
			table.Element.Key.EQ(postgres.String(key)),
			rdbms.ActiveAt(effectiveTime, systemTime),
		))

	if err := stmt.QueryContext(ctx, db, &elements); err != nil {
		return nil, fmt.Errorf("GetElementByKey: %w", err)
	}
	if len(elements) == 0 {
		return nil, nil
	}
	return &elements[0], nil
}

func CreateOrUpdateElement(ctx context.Context, db rdbms.Executor, catID int64, key string, validFrom, validTo time.Time, effectiveTime, systemTime *time.Time) (*Element, error) {
	var existingList []struct {
		ElementID int64
	}
	stmt := postgres.SELECT(table.Element.ElementID).
		FROM(table.Element).
		WHERE(postgres.AND(
			table.Element.CategoryID.EQ(postgres.Int64(catID)),
			table.Element.Key.EQ(postgres.String(key)),
			rdbms.IsActive(),
		))

	if err := stmt.QueryContext(ctx, db, &existingList); err != nil {
		return nil, fmt.Errorf("CreateOrUpdateElement: query failed: %w", err)
	}

	var id int64
	if len(existingList) > 0 {
		id = existingList[0].ElementID
	}

	elem := Element{
		ElementID:         id,
		CategoryID:        catID,
		InsertionMetadata: "{}",
	}
	elem.Key = key
	elem.ValidityStart = validFrom
	elem.ValidityEnd = validTo

	result, err := rdbms.ExecuteBitemporalMutation(ctx, db, time.Time{}, nil, []Element{elem}, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateOrUpdateElement: mutation failed: %w", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("CreateOrUpdateElement: mutation returned no result")
	}
	return &result[0], nil
}
