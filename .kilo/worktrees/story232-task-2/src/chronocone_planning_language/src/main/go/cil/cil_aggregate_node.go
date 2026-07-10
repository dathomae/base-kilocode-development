// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// CILAggregateNode represents an aggregation operation in the CIL expression tree.
// It captures the source dimensions, target dimension, aggregation matrix,
// cell operation, and dimensional unit effect.
type CILAggregateNode struct {
	// SourceCIL is the CIL subtree representing the metric being aggregated.
	// Populated during Increment 4 (aggregator integration into the metric DAG).
	SourceCIL CILNode

	// FromDimensions is the ordered list of source concept names for the aggregation.
	FromDimensions []string

	// ToDimension is the target concept name (empty when aggregating to scalar).
	ToDimension string

	// AggregationMatrix is the 2D weight matrix for tensor dimension reduction.
	// Rows correspond to values of the concept being collapsed.
	// Columns correspond to values of the concept being preserved.
	AggregationMatrix *AggregationMatrix

	// CellOperation is the name of the aggregation function (e.g., "sum").
	CellOperation string

	// DimensionalUnitEffect is the string representation of the dimensional unit
	// effect of this aggregation (e.g., "Country", "1/Country").
	DimensionalUnitEffect string

	// SourceLocation records where this aggregator was defined in the source CPL.
	SourceLocation symboltable.SourceLocation
}

// cilNode is a marker method that satisfies the CILNode interface.
func (n *CILAggregateNode) cilNode() {}
