// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package expected_xls_generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// ExpressionToExcel translates a CIL expression tree to an Excel formula string
// with cross-sheet references. connectorSheet is the name of the sheet containing
// connector data (e.g., "data_a"). rowIdx is the 0-indexed row index in the
// input sheet where the data row is located. keyColCount is the number of
// key columns (dimensions + grain flattened). colName is a function that
// converts 0-based column index to Excel column letter.
func ExpressionToExcel(node cil.CILNode, connectorSheet string, rowIdx int, keyColCount int, colName func(int) string, dimensionColumnIndexes map[string]int) string {
	return translateExprXLSX(node, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
}

// translateExprXLSX returns the Excel formula expression string with cross-sheet references.
func translateExprXLSX(node cil.CILNode, connectorSheet string, rowIdx int, keyColCount int, colName func(int) string, dimensionColumnIndexes map[string]int) string {
	switch n := node.(type) {
	case *cil.CILLiteral:
		return strconv.FormatFloat(n.Value, 'f', -1, 64)
	case *cil.CILInputRef:
		// Cross-sheet reference: data_a!C2
		return connectorSheet + "!" + colName(keyColCount) + strconv.Itoa(rowIdx+2)
	case *cil.CILBinaryOp:
		return "(" + translateExprXLSX(n.Left, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes) + n.Op.Symbol() + translateExprXLSX(n.Right, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes) + ")"
	case *cil.CILUnaryOp:
		return "-" + translateExprXLSX(n.Operand, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
	case *cil.CILComparisonOp:
		op := n.Op.Symbol()
		if n.Op == operator.Equal {
			op = "="
		} else if n.Op == operator.NotEqual {
			op = "<>"
		}
		return translateExprXLSX(n.Left, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes) + op + translateExprXLSX(n.Right, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
	case *cil.CILLogicalOp:
		funcName := strings.ToUpper(n.Op.Symbol())
		return funcName + "(" + translateExprXLSX(n.Left, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes) + "," + translateExprXLSX(n.Right, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes) + ")"
	case *cil.CILNotOp:
		return "NOT(" + translateExprXLSX(n.Operand, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes) + ")"
	case *cil.CILBoolLiteral:
		if n.Value {
			return "TRUE()"
		}
		return "FALSE()"
	case *cil.CILNilLiteral:
		return "0"
	case *cil.CILLookupNode:
		if len(n.Dimensions) == 0 {
			return ""
		}
		lookupSheet := n.LookupName
		expectedOutputRow := rowIdx + 2
		// Build the lookup key by concatenating dimension cells from expected_output.
		// The lookup sheet's column A stores the same concatenation.
		lookupKey := buildLookupKeyXLSX(n.Dimensions, dimensionColumnIndexes, colName, expectedOutputRow, "expected_output!")
		if lookupKey == "" {
			return ""
		}
		// Range: A through Value column. numDimensions determines the Value column index.
		rangeEnd := colName(len(n.Dimensions) + 1) // key(A=0) + dims(1..N) + Value(N+1)
		valColIdx := len(n.Dimensions) + 2          // 1-based: key=1, dim1=2, ..., Value=N+2
		return fmt.Sprintf("VLOOKUP(%s,%s!A:%s,%d,FALSE)",
			lookupKey, lookupSheet, rangeEnd, valColIdx)
	case *cil.CILConditional:
		return translateConditionalXLSX(n.Branches, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
	case *cil.CILWhenClause:
		return translateWhenClauseXLSX(n.Clauses, n.Default, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
	default:
		return ""
	}
}

// translateConditionalXLSX translates a CILConditional into a chain of nested IFs.
func translateConditionalXLSX(branches []cil.CILConditionalBranch, connectorSheet string, rowIdx int, keyColCount int, colName func(int) string, dimensionColumnIndexes map[string]int) string {
	if len(branches) == 0 {
		return ""
	}
	return buildIfChainXLSX(branches, 0, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
}

func buildIfChainXLSX(branches []cil.CILConditionalBranch, idx int, connectorSheet string, rowIdx int, keyColCount int, colName func(int) string, dimensionColumnIndexes map[string]int) string {
	if idx >= len(branches) {
		return ""
	}
	branch := branches[idx]
	body := translateExprXLSX(branch.Body, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)

	var cond string
	if branch.Condition == nil {
		if idx == len(branches)-1 {
			return body
		}
		cond = ""
	} else {
		cond = translateExprXLSX(branch.Condition, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
	}

	if idx == len(branches)-1 {
		return body
	}

	next := buildIfChainXLSX(branches, idx+1, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
	if branch.Condition == nil {
		return next
	}
	return "IF(" + cond + "," + body + "," + next + ")"
}

func translateWhenClauseXLSX(clauses []cil.CILWhenCase, defaultBody cil.CILNode, connectorSheet string, rowIdx int, keyColCount int, colName func(int) string, dimensionColumnIndexes map[string]int) string {
	if len(clauses) == 0 {
		if defaultBody != nil {
			return translateExprXLSX(defaultBody, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
		}
		return ""
	}

	clause := clauses[0]
	cond := translateExprXLSX(clause.Condition, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
	body := translateExprXLSX(clause.Body, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)

	var next string
	if len(clauses) == 1 {
		next = translateExprXLSX(defaultBody, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
	} else {
		next = translateWhenClauseXLSX(clauses[1:], defaultBody, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
	}

	return "IF(" + cond + "," + body + "," + next + ")"
}

// ExpressionToCellExcel translates a CIL expression tree to an Excel formula string using
// same-sheet cell references. inputColumnIndexes maps input variable names to their
// 0-based index within the input columns block. colName converts 0-based column index to
// Excel letter. row is the 1-based row number for cell references. keyColCount is
// the number of key columns (before input columns).
func ExpressionToCellExcel(node cil.CILNode, inputColumnIndexes map[string]int, colName func(int) string, row int, keyColCount int, dimensionColumnIndexes map[string]int) string {
	return translateExprCellXLSX(node, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
}

func translateExprCellXLSX(node cil.CILNode, inputColumnIndexes map[string]int, colName func(int) string, row int, keyColCount int, dimensionColumnIndexes map[string]int) string {
	switch n := node.(type) {
	case *cil.CILLiteral:
		return strconv.FormatFloat(n.Value, 'f', -1, 64)
	case *cil.CILInputRef:
		return colName(keyColCount+inputColumnIndexes[n.Name]) + strconv.Itoa(row)
	case *cil.CILBinaryOp:
		return "(" + translateExprCellXLSX(n.Left, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes) + n.Op.Symbol() + translateExprCellXLSX(n.Right, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes) + ")"
	case *cil.CILUnaryOp:
		return "-" + translateExprCellXLSX(n.Operand, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
	case *cil.CILComparisonOp:
		op := n.Op.Symbol()
		if n.Op == operator.Equal {
			op = "="
		} else if n.Op == operator.NotEqual {
			op = "<>"
		}
		return translateExprCellXLSX(n.Left, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes) + op + translateExprCellXLSX(n.Right, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
	case *cil.CILLogicalOp:
		funcName := strings.ToUpper(n.Op.Symbol())
		return funcName + "(" + translateExprCellXLSX(n.Left, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes) + "," + translateExprCellXLSX(n.Right, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes) + ")"
	case *cil.CILNotOp:
		return "NOT(" + translateExprCellXLSX(n.Operand, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes) + ")"
	case *cil.CILBoolLiteral:
		if n.Value {
			return "TRUE()"
		}
		return "FALSE()"
	case *cil.CILNilLiteral:
		return "0"
	case *cil.CILLookupNode:
		if len(n.Dimensions) == 0 {
			return ""
		}
		lookupSheet := n.LookupName
		// Build the lookup key by concatenating dimension cells from the current row.
		// The lookup sheet's column A stores the same concatenation.
		lookupKey := buildLookupKeyXLSX(n.Dimensions, dimensionColumnIndexes, colName, row, "")
		if lookupKey == "" {
			return ""
		}
		// Range: A through Value column. numDimensions determines the Value column index.
		rangeEnd := colName(len(n.Dimensions) + 1) // key(A=0) + dims(1..N) + Value(N+1)
		valColIdx := len(n.Dimensions) + 2          // 1-based: key=1, dim1=2, ..., Value=N+2
		return fmt.Sprintf("VLOOKUP(%s,%s!A:%s,%d,FALSE)",
			lookupKey, lookupSheet, rangeEnd, valColIdx)
	case *cil.CILConditional:
		return translateConditionalCellXLSX(n.Branches, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
	case *cil.CILWhenClause:
		return translateWhenClauseCellXLSX(n.Clauses, n.Default, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
	default:
		return ""
	}
}

func translateConditionalCellXLSX(branches []cil.CILConditionalBranch, inputColumnIndexes map[string]int, colName func(int) string, row int, keyColCount int, dimensionColumnIndexes map[string]int) string {
	if len(branches) == 0 {
		return ""
	}
	return buildIfChainCellXLSX(branches, 0, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
}

func buildIfChainCellXLSX(branches []cil.CILConditionalBranch, idx int, inputColumnIndexes map[string]int, colName func(int) string, row int, keyColCount int, dimensionColumnIndexes map[string]int) string {
	if idx >= len(branches) {
		return ""
	}
	branch := branches[idx]
	body := translateExprCellXLSX(branch.Body, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)

	var cond string
	if branch.Condition == nil {
		if idx == len(branches)-1 {
			return body
		}
		cond = ""
	} else {
		cond = translateExprCellXLSX(branch.Condition, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
	}

	if idx == len(branches)-1 {
		return body
	}

	next := buildIfChainCellXLSX(branches, idx+1, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
	if branch.Condition == nil {
		return next
	}
	return "IF(" + cond + "," + body + "," + next + ")"
}

func translateWhenClauseCellXLSX(clauses []cil.CILWhenCase, defaultBody cil.CILNode, inputColumnIndexes map[string]int, colName func(int) string, row int, keyColCount int, dimensionColumnIndexes map[string]int) string {
	if len(clauses) == 0 {
		if defaultBody != nil {
			return translateExprCellXLSX(defaultBody, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
		}
		return ""
	}

	clause := clauses[0]
	cond := translateExprCellXLSX(clause.Condition, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
	body := translateExprCellXLSX(clause.Body, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)

	var next string
	if len(clauses) == 1 {
		next = translateExprCellXLSX(defaultBody, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
	} else {
		next = translateWhenClauseCellXLSX(clauses[1:], defaultBody, inputColumnIndexes, colName, row, keyColCount, dimensionColumnIndexes)
	}

	return "IF(" + cond + "," + body + "," + next + ")"
}

// joinConditions joins condition strings. If only one, returns it directly. Otherwise wraps in AND().
func joinConditions(parts []string) string {
	if len(parts) == 1 {
		return parts[0]
	}
	return "AND(" + strings.Join(parts, ",") + ")"
}

// buildLookupKeyXLSX builds the cell reference(s) for the VLOOKUP lookup value
// by concatenating dimension cell references. For single-dimension, returns
// the single cell reference. For multi-dimension, returns a pipe-concatenated
// expression (e.g., A2&"|"&B2). The sheetPrefix parameter adds a sheet reference
// for cross-sheet formulas (e.g., "expected_output!"); pass empty string for
// same-sheet formulas.
func buildLookupKeyXLSX(dimensions []string, dimensionColumnIndexes map[string]int, colName func(int) string, row int, sheetPrefix string) string {
	if len(dimensions) == 0 {
		return ""
	}
	parts := make([]string, 0, len(dimensions))
	for _, dim := range dimensions {
		dimIndex, ok := dimensionColumnIndexes[dim]
		if !ok {
			return ""
		}
		parts = append(parts, sheetPrefix+colName(dimIndex)+strconv.Itoa(row))
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts, `&"|"&`)
}

// FormulaForBranchXLSX generates the per-branch activation formula for a conditional in Excel syntax.
func FormulaForBranchXLSX(branches []cil.CILConditionalBranch, branchIdx int, connectorSheet string, rowIdx int, keyColCount int, colName func(int) string, dimensionColumnIndexes map[string]int) string {
	if len(branches) == 0 {
		return ""
	}

	branch := branches[branchIdx]
	bodyExpr := translateExprXLSX(branch.Body, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)

	var negations []string
	for i := 0; i < branchIdx; i++ {
		negations = append(negations, "NOT("+translateExprXLSX(branches[i].Condition, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)+")")
	}

	if branch.Condition == nil {
		if len(negations) == 0 {
			return bodyExpr
		}
		activation := joinConditions(negations)
		return "IF(" + activation + "," + bodyExpr + "," + `""` + ")"
	}

	condExpr := translateExprXLSX(branch.Condition, connectorSheet, rowIdx, keyColCount, colName, dimensionColumnIndexes)
	var fullCondition string
	if len(negations) == 0 {
		fullCondition = condExpr
	} else {
		parts := append([]string{}, negations...)
		parts = append(parts, condExpr)
		fullCondition = joinConditions(parts)
	}

	return "IF(" + fullCondition + "," + bodyExpr + "," + `""` + ")"
}
