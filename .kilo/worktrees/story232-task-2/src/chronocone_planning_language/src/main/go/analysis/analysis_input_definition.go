// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package analysis

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// AnalysisTimeShift carries validated time shift information for an input declaration.
type AnalysisTimeShift struct {
	Direction string  // "backward" or "forward"
	Field     string  // validated Instant constraint field keyword
	Amount    float64 // shift amount (defaults to 1.0)
}

// AnalysisInputDefinition captures a semantically validated input declaration.
type AnalysisInputDefinition struct {
	Kind           string                // "metric" or "dimension"
	Name           string                // Referenced metric/dimension name
	Alias          string                // Alias from 'as' clause (equals Name if no alias)
	WhereClause    *AnalysisWhereClause  // Where clause (nil if not present)
	TimeShift      *AnalysisTimeShift    // nil if no time shift specified
	SourceLocation symboltable.SourceLocation
}
