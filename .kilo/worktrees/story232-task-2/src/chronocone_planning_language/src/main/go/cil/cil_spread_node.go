// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// CILSpreadNode represents a spreader operation in the CIL tree.
// It defines a dimension-expanding transformation from a set of source dimensions
// to a larger set of target dimensions by introducing exactly one new dimension.
type CILSpreadNode struct {
	// SourceCIL is the CIL subtree representing the metric being spread.
	// Populated during Increment 4 (spreader integration into the metric DAG).
	SourceCIL CILNode

	// FromDimensions is the ordered list of source concept names for the spread.
	FromDimensions []string

	// ToDimensions is the ordered list of target concept names for the spread.
	ToDimensions []string

	// NewDimension is the name of the concept being introduced by this spreader.
	NewDimension string

	// SpreadingMatrix is the 2D weight matrix for tensor dimension expansion.
	// For Increment 1, this is a single row with equal-proportion weights.
	SpreadingMatrix *AggregationMatrix

	// DistributionExpression is the CIL subtree for the distribution expression.
	// nil in Increment 1 (filled in Increment 3).
	DistributionExpression CILNode

	// ReferencedMetrics is the list of metrics referenced by this spreader.
	// nil in Increment 1 (filled in Increment 3).
	ReferencedMetrics []CILNode

	// DimensionalUnitEffect is the string representation of the dimensional unit
	// effect of this spreader (e.g., "Country", "1/Country").
	DimensionalUnitEffect string

	// SourceLocation records where this spreader was defined in the source CPL.
	SourceLocation symboltable.SourceLocation
}

// cilNode is a marker method that satisfies the CILNode interface.
func (n *CILSpreadNode) cilNode() {}
