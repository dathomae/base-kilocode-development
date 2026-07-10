// Package analysis provides the analysis phase data structures for the Chronocone Planning Language.
//
// This file contains the measure conversion analysis phase implementation.
// It processes measure conversion definitions from the symbol table, builds
// conversion pair structures, and derives inverse conversions algebraically.
package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// AnalyzeMeasureConversions walks MeasureScope children for MeasureConversionScope
// entries and delegates to the shared analyzeConversions core.
//
// In addition to the core validation, it ensures that a measure does not
// declare a self-conversion.
func (a *CPLAnalyzer) AnalyzeMeasureConversions(measureScope *symboltable.SymbolTable, measureDef *AnalysisMeasureDefinition) error {
	lookup := make(map[string]*AnalysisMeasureDefinition, len(a.measures))
	for name, m := range a.measures {
		mCopy := m
		lookup[name] = &mCopy
	}

	validatePair := func(targetDef, sourceDef *AnalysisMeasureDefinition) error {
		if targetDef.Name == sourceDef.Name {
			return fmt.Errorf("conversion creates self-conversion %q = %q", targetDef.Name, sourceDef.Name)
		}
		return nil
	}

	return analyzeConversions(measureScope, symboltable.MeasureConversionScope, lookup, measureDef.Name, measureDef, a.units, "measure", validatePair)
}

// invertExpression derives the inverse of a conversion calculation expression.
// It replaces occurrences of oldVar with newVar and inverts the arithmetic
// operations to solve for the new variable.
//
// Supported inversion patterns (where source is the variable and N is a constant):
//
//	source * N → target / N
//	N * source → target / N
//	source / N → target * N
//	source + N → target - N
//	N + source → target - N
//	source - N → target + N
//	N - source → N - target
//
// Unsupported patterns:
//
//	multiplication with non-constant operands
//	addition/subtraction with non-constant operands
//	division with non-constant divisor
//	power expressions (^)
func invertExpression(expr Expression, oldVar string, newVar string) (Expression, error) {
	switch e := expr.(type) {
	case *VariableRef:
		if e.Name == oldVar {
			return &VariableRef{Name: newVar}, nil
		}
		return &VariableRef{Name: e.Name}, nil
	case *UnitReference:
		if e.Name == oldVar {
			return &UnitReference{Name: newVar, UnitGroup: e.UnitGroup}, nil
		}
		return &UnitReference{Name: e.Name, UnitGroup: e.UnitGroup}, nil
	case *NumberNode:
		return &NumberNode{Value: e.Value}, nil
	case *MultiplicationNode:
		// source * N  =>  target / N
		// N * source  =>  target / N
		if isNumberNode(e.Right) {
			leftInv, err := invertExpression(e.Left, oldVar, newVar)
			if err != nil {
				return nil, err
			}
			return &DivisionNode{Left: leftInv, Right: e.Right}, nil
		}
		if isNumberNode(e.Left) {
			rightInv, err := invertExpression(e.Right, oldVar, newVar)
			if err != nil {
				return nil, err
			}
			return &DivisionNode{Left: rightInv, Right: e.Left}, nil
		}
		return nil, fmt.Errorf("cannot invert multiplication with non-constant operand")
	case *DivisionNode:
		// source / N  =>  target * N
		if isNumberNode(e.Right) {
			leftInv, err := invertExpression(e.Left, oldVar, newVar)
			if err != nil {
				return nil, err
			}
			return &MultiplicationNode{Left: leftInv, Right: e.Right}, nil
		}
		return nil, fmt.Errorf("cannot invert division with non-constant divisor")
	case *AdditionNode:
		// source + N  =>  target - N
		// N + source  =>  target - N
		if isNumberNode(e.Right) {
			leftInv, err := invertExpression(e.Left, oldVar, newVar)
			if err != nil {
				return nil, err
			}
			return &SubtractionNode{Left: leftInv, Right: e.Right}, nil
		}
		if isNumberNode(e.Left) {
			rightInv, err := invertExpression(e.Right, oldVar, newVar)
			if err != nil {
				return nil, err
			}
			return &SubtractionNode{Left: rightInv, Right: e.Left}, nil
		}
		return nil, fmt.Errorf("cannot invert addition with non-constant operand")
	case *SubtractionNode:
		// source - N  =>  target + N
		if isNumberNode(e.Right) {
			leftInv, err := invertExpression(e.Left, oldVar, newVar)
			if err != nil {
				return nil, err
			}
			return &AdditionNode{Left: leftInv, Right: e.Right}, nil
		}
		// N - source  =>  N - target  (left is constant)
		if isNumberNode(e.Left) {
			// target = N - source  =>  source = N - target
			return &SubtractionNode{Left: e.Left, Right: &VariableRef{Name: newVar}}, nil
		}
		return nil, fmt.Errorf("cannot invert subtraction with non-constant operand")
	case *PowerNode:
		return nil, fmt.Errorf("cannot invert power expressions")
	default:
		return nil, fmt.Errorf("unknown expression type for inversion")
	}
}

// isNumberNode returns true if the expression is a *NumberNode.
func isNumberNode(expr Expression) bool {
	_, ok := expr.(*NumberNode)
	return ok
}
