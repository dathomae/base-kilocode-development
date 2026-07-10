// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file defines the aggregator postfix expression evaluator and dimensional
// unit validation for aggregator definitions.
package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EvaluateAggregatorExpression evaluates a postfix sequence of symbol definitions
// from an AggregatorScope into a resolved DimensionalUnitExpression and cell operation.
//
// The parser pushes expression components to the symbol table's Definitions slice
// in postfix order. This function uses a stack-based algorithm:
//
//   - Identifier (not in registry, in fromConcepts): concept reference → push {Numerator: [name]}
//   - Identifier (in registry): function call → pop arity params, evaluate, push result
//   - Identifier (neither): returns "unknown function or concept" error
//   - UnitReferenceDefinition: push {Numerator: [name]}
//   - MultiplyOperatorDefinition: pop two, push left.Multiply(right)
//   - DivideOperatorDefinition: pop two, invert right, push left.Multiply(inverted)
//   - PlusOperatorDefinition / MinusOperatorDefinition: pop two, verify equal dims, push same
//   - AggregatorScopeDefinition: skipped (not part of expression)
//
// After processing all definitions, exactly one item must remain on the stack.
func EvaluateAggregatorExpression(
	defs []symboltable.SymbolDefinition,
	fromConcepts []*AnalysisConceptDefinition,
	registry *FunctionRegistry,
) (*DimensionalUnitExpression, string, error) {
	var stack []*DimensionalUnitExpression
	var cellOp string

	for _, def := range defs {
		switch def.DefinitionType {
		case symboltable.Identifier:
			strData, ok := def.Data.(symboltable.StringData)
			if !ok {
				return nil, "", fmt.Errorf("invalid identifier data in aggregator expression")
			}
			name := strData.Value

			fnInfo, exists := registry.Functions[name]
			if exists {
				// Function call
				if len(stack) < fnInfo.Arity {
					return nil, "", fmt.Errorf("insufficient operands for function %s", name)
				}
				params := make([]string, fnInfo.Arity)
				for i := 0; i < fnInfo.Arity; i++ {
					idx := len(stack) - fnInfo.Arity + i
					expr := stack[idx]
					if len(expr.Numerator) != 1 || len(expr.Denominator) != 0 {
						return nil, "", fmt.Errorf("invalid params to function %s", name)
					}
					params[i] = expr.Numerator[0]
				}
				stack = stack[:len(stack)-fnInfo.Arity]

				result := fnInfo.DimUnits(params)
				if result == nil {
					return nil, "", fmt.Errorf("invalid params to function %s", name)
				}
				stack = append(stack, result)
				cellOp = fnInfo.Name
			} else if isValidConceptReference(name, fromConcepts) {
				// Concept reference
				stack = append(stack, &DimensionalUnitExpression{
					Numerator: []string{name},
				})
			} else {
				return nil, "", fmt.Errorf("unknown function or concept: %s", name)
			}

		case symboltable.UnitReferenceDefinition:
			ref, ok := def.Data.(parser.ParserUnitReference)
			if !ok {
				return nil, "", fmt.Errorf("invalid unit reference data in aggregator expression")
			}
			stack = append(stack, &DimensionalUnitExpression{
				Numerator: []string{ref.Name},
			})

		case symboltable.MultiplyOperatorDefinition:
			right, left, err := popTwoDimUnits(&stack)
			if err != nil {
				return nil, "", err
			}
			stack = append(stack, left.Multiply(right))

		case symboltable.DivideOperatorDefinition:
			right, left, err := popTwoDimUnits(&stack)
			if err != nil {
				return nil, "", err
			}
			inverted := &DimensionalUnitExpression{
				Numerator:   right.Denominator,
				Denominator: right.Numerator,
			}
			stack = append(stack, left.Multiply(inverted))

		case symboltable.PlusOperatorDefinition, symboltable.MinusOperatorDefinition:
			right, left, err := popTwoDimUnits(&stack)
			if err != nil {
				return nil, "", err
			}
			if !left.Equals(right) {
				return nil, "", fmt.Errorf("dimensional units mismatch for %s operator: %s vs %s", def.DefinitionType, left, right)
			}
			stack = append(stack, left)

		case symboltable.AggregatorScopeDefinition:
			// Skip the synthetic scope definition; it's not part of the expression.
			continue

		default:
			return nil, "", fmt.Errorf("unexpected definition type in aggregator expression: %s", def.DefinitionType)
		}
	}

	if len(stack) != 1 {
		return nil, "", fmt.Errorf("malformed expression: %d items remaining on stack", len(stack))
	}

	return stack[0], cellOp, nil
}

// isValidConceptReference checks whether the given name matches a concept
// in the provided fromConcepts slice.
func isValidConceptReference(name string, fromConcepts []*AnalysisConceptDefinition) bool {
	for _, c := range fromConcepts {
		if c.ConceptName == name {
			return true
		}
	}
	return false
}

// popTwoDimUnits pops the top two DimensionalUnitExpression values from the stack.
// For binary operators, the right operand is popped first, then the left.
func popTwoDimUnits(stack *[]*DimensionalUnitExpression) (right *DimensionalUnitExpression, left *DimensionalUnitExpression, err error) {
	if len(*stack) < 2 {
		return nil, nil, fmt.Errorf("insufficient operands for operator")
	}
	right = (*stack)[len(*stack)-1]
	left = (*stack)[len(*stack)-2]
	*stack = (*stack)[:len(*stack)-2]
	return right, left, nil
}

// validateAggregatorDimensionalUnits validates that the dimensional units of an
// aggregator definition are consistent.
//
// It computes the expected output dimensional units from the from/to concepts
// and verifies that:
//   fromDimensionalUnits * expressionDimensionalUnits == toDimensionalUnits
//
// Returns a descriptive error if the units do not match.
func ValidateAggregatorDimensionalUnits(info *AggregatorInfo) error {
	info.FromDimensionalUnits = dimensionalUnitsFromConceptSet(info.FromConcepts)

	if info.ToEmpty {
		info.ToDimensionalUnits = &DimensionalUnitExpression{}
	} else {
		info.ToDimensionalUnits = dimensionalUnitsFromConceptSet([]*AnalysisConceptDefinition{info.ToConcept})
	}

	expected := info.FromDimensionalUnits.Multiply(info.ExpressionDimUnits)
	if !expected.Equals(info.ToDimensionalUnits) {
		return fmt.Errorf(
			"dimensional unit mismatch: from %s * expression %s = %s, expected to %s",
			info.FromDimensionalUnits.String(),
			info.ExpressionDimUnits.String(),
			expected.String(),
			info.ToDimensionalUnits.String(),
		)
	}

	return nil
}
