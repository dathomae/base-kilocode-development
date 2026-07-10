// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package analysis

import "github.com/chronocone/chronocone_planning_language/src/main/go/cil"

// AnalysisMetricDefinition is the analysis-phase representation of a metric definition.
type AnalysisMetricDefinition struct {
	Name    string                    // Metric identifier
	Inputs  []AnalysisInputDefinition // Input declarations with semantic info
	CILRoot *cil.CILMetricNode        // Root of the CIL tree for this metric
}
