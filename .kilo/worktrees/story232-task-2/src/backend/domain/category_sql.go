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
		reflect.TypeOf(Category{}),
		table.Category,
		rdbms.TableTypeCategory,
		table.Category.CategoryID,
	)
}

func ListCategories(ctx context.Context, db rdbms.Executor, nsID int64, effectiveTime, systemTime *time.Time) ([]Category, error) {
	var categories []Category
	stmt := postgres.SELECT(table.Category.AllColumns).
		FROM(table.Category).
		WHERE(postgres.AND(
			table.Category.NsID.EQ(postgres.Int64(nsID)),
			rdbms.ActiveAt(effectiveTime, systemTime),
		))

	if err := stmt.QueryContext(ctx, db, &categories); err != nil {
		return nil, fmt.Errorf("ListCategories: %w", err)
	}
	return categories, nil
}

func GetCategoryByKey(ctx context.Context, db rdbms.Executor, nsID int64, key string, effectiveTime, systemTime *time.Time) (*Category, error) {
	var categories []Category
	stmt := postgres.SELECT(table.Category.AllColumns).
		FROM(table.Category).
		WHERE(postgres.AND(
			table.Category.NsID.EQ(postgres.Int64(nsID)),
			table.Category.Key.EQ(postgres.String(key)),
			rdbms.ActiveAt(effectiveTime, systemTime),
		))

	if err := stmt.QueryContext(ctx, db, &categories); err != nil {
		return nil, fmt.Errorf("GetCategoryByKey: %w", err)
	}
	if len(categories) == 0 {
		return nil, nil
	}
	return &categories[0], nil
}

func CreateOrUpdateCategory(ctx context.Context, db rdbms.Executor, nsID int64, key string, validFrom, validTo time.Time, effectiveTime, systemTime *time.Time) (*Category, error) {
	var existingList []struct {
		CategoryID int64
	}
	stmt := postgres.SELECT(table.Category.CategoryID).
		FROM(table.Category).
		WHERE(postgres.AND(
			table.Category.NsID.EQ(postgres.Int64(nsID)),
			table.Category.Key.EQ(postgres.String(key)),
			rdbms.IsActive(),
		))

	if err := stmt.QueryContext(ctx, db, &existingList); err != nil {
		return nil, fmt.Errorf("CreateOrUpdateCategory: query failed: %w", err)
	}

	var id int64
	if len(existingList) > 0 {
		id = existingList[0].CategoryID
	}

	cat := Category{
		CategoryID:        id,
		NsID:              nsID,
		InsertionMetadata: "{}",
	}
	cat.Key = key
	cat.ValidityStart = validFrom
	cat.ValidityEnd = validTo

	result, err := rdbms.ExecuteBitemporalMutation(ctx, db, time.Time{}, nil, []Category{cat}, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateOrUpdateCategory: mutation failed: %w", err)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("CreateOrUpdateCategory: mutation returned no result")
	}
	return &result[0], nil
}
