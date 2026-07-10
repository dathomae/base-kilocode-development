// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

// AggregationMatrix is a 2D weight matrix used for tensor dimension
// reduction and expansion operations in the CIL tree.
type AggregationMatrix struct {
	Rows      int
	Columns   int
	Data      [][]float64
	RowLabels []string
	ColLabels []string
}
