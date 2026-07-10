package rdbms

import (
	"time"

	"chronocone/backend/serde/rdbms/gen/postgres/public/enum"
	"github.com/go-jet/jet/v2/postgres"
)

// ActiveAt returns a boolean expression filtering for records active at both effective_time and system_time
// on the active partition. If either time is nil, CURRENT_TIMESTAMP is used.
// This combines validity filtering, transaction time filtering, and mode filtering into one expression.
func ActiveAt(effectiveTime *time.Time, systemTime *time.Time) postgres.BoolExpression {
	var effExpr postgres.TimestampzExpression = postgres.CURRENT_TIMESTAMP()
	if effectiveTime != nil {
		effExpr = postgres.TimestampzT(*effectiveTime)
	}

	var sysExpr postgres.TimestampzExpression = postgres.CURRENT_TIMESTAMP()
	if systemTime != nil {
		sysExpr = postgres.TimestampzT(*systemTime)
	}

	return postgres.AND(
		postgres.TimestampzColumn("validity_start").LT_EQ(effExpr),
		postgres.TimestampzColumn("validity_end").GT(effExpr),
		postgres.TimestampzColumn("transaction_start").LT_EQ(sysExpr),
		postgres.TimestampzColumn("transaction_end").GT(sysExpr),
		postgres.StringColumn("mode").EQ(enum.BitemporalMode.Active),
	)
}

// ActiveAtParams returns parameter values in the order they're used by ActiveAt.
// If both times are nil, returns empty slice. Otherwise returns non-nil times as parameters.
func ActiveAtParams(effectiveTime *time.Time, systemTime *time.Time) []interface{} {
	var params []interface{}
	if effectiveTime != nil {
		params = append(params, *effectiveTime)
	}
	if systemTime != nil {
		params = append(params, *systemTime)
	}
	return params
}

// IsValidAt returns a boolean expression that filters for rows that were valid
// at the specified point in bitemporal 'valid_time'. Callers must supply an
// explicit validAt instead of relying on wall-clock time so that logic remains
// deterministic across re-runs.
func IsValidAt(start, end postgres.ColumnTimestampz, validAt time.Time) postgres.BoolExpression {
	return AsOf(start, end, validAt)
}

// AsOf returns a boolean expression that filters for rows that were valid
// at the specified point in bitemporal 'valid_time'.
func AsOf(start, end postgres.ColumnTimestampz, v time.Time) postgres.BoolExpression {
	ts := postgres.TimestampzT(v)
	return postgres.AND(
		start.LT_EQ(ts),
		end.GT(ts),
	)
}

// BelievedAt returns a boolean expression that filters for rows that were
// the latest known state at the specified point in bitemporal 'transaction_time'.
// Locks the transaction axis at s. Applies to either partition because the
// transaction axis spans both.
func BelievedAt(start, end postgres.ColumnTimestampz, s time.Time) postgres.BoolExpression {
	ts := postgres.TimestampzT(s)
	return postgres.AND(
		start.LT_EQ(ts),
		end.GT(ts),
	)
}

// RecordedAfter returns a boolean expression that filters for rows created
// on or after a certain transaction time (useful for audit or sync).
func RecordedAfter(start postgres.ColumnTimestampz, s time.Time) postgres.BoolExpression {
	return start.GT_EQ(postgres.TimestampzT(s))
}

// IsActive returns a boolean expression that filters for the current 'active' partition.
func IsActive() postgres.BoolExpression {
	return postgres.StringColumn("mode").EQ(enum.BitemporalMode.Active)
}
