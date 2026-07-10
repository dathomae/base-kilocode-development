// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package analysis

// AnalysisWhereClause represents an OR-of-AND tree for where clause conditions.
// Each AndGroup is combined with OR; conditions within a group are combined with AND.
type AnalysisWhereClause struct {
	AndGroups []AnalysisWhereAndGroup
}

// AnalysisWhereAndGroup represents a group of conditions combined with AND.
type AnalysisWhereAndGroup struct {
	Conditions []AnalysisWhereCondition
}

// AnalysisWhereCondition represents a single aspect expression within a where clause.
type AnalysisWhereCondition struct {
	Negated bool   // true if the condition is preceded by "not"
	Aspect  string // "ACTUALS" or "FORECAST"
}
