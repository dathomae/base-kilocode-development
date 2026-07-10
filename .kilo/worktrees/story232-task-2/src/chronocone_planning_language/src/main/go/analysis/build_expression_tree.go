package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// BuildExpressionTree evaluates a postfix sequence of symbol definitions
// into a resolved Expression tree.
//
// The parser pushes expression components to the symbol table's Definitions
// slice in postfix order. This function uses a stack-based algorithm to
// construct the expression tree:
//
//   - UnitReferenceDefinition: looked up in units map, pushed as UnitReference
//   - NumberDefinition: pushed as NumberNode
//   - Binary operators (Multiply, Divide, Plus, Minus): pop right and left
//     operands, push the corresponding binary node
//   - PowerOperatorDefinition: pop exponent (NumberNode) and base, push PowerNode
//
// After processing all definitions, exactly one item must remain on the stack.
// Any other count indicates a malformed expression.
//
// Parameters:
//   - defs: postfix-ordered symbol definitions from the parser
//   - units: lookup map from unit name to its analyzed definition
//   - vars: optional map of pattern-bound variable names; when non-nil, names
//     found in vars are pushed as VariableRef instead of UnitReference
//
// Returns:
//   - The root Expression node of the constructed tree
//   - An error if the expression is malformed or references an undefined unit
func BuildExpressionTree(defs []symboltable.SymbolDefinition, units map[string]AnalysisUnitDefinition, vars map[string]bool) (Expression, error) {
	var stack []Expression

	for _, def := range defs {
		switch def.DefinitionType {
		case symboltable.UnitReferenceDefinition:
			ref, ok := def.Data.(parser.ParserUnitReference)
			if !ok {
				return nil, fmt.Errorf("invalid unit reference data")
			}
			if vars != nil && vars[ref.Name] {
				stack = append(stack, &VariableRef{Name: ref.Name})
			} else if unit, exists := units[ref.Name]; exists {
				stack = append(stack, &UnitReference{Name: ref.Name, UnitGroup: &unit})
			} else if vars != nil {
				return nil, fmt.Errorf("undefined identifier %q in expression (must be a pattern-bound variable or a declared unit)", ref.Name)
			} else {
				// Unit not yet defined; create reference without group for now.
				// Validation happens later after all units are collected.
				stack = append(stack, &UnitReference{Name: ref.Name, UnitGroup: nil})
			}

		case symboltable.NumberDefinition:
			fd, ok := def.Data.(symboltable.FloatData)
			if !ok {
				return nil, fmt.Errorf("invalid number data")
			}
			stack = append(stack, &NumberNode{Value: fd.Value})

		case symboltable.MultiplyOperatorDefinition:
			right, left, err := popTwo(&stack)
			if err != nil {
				return nil, err
			}
			stack = append(stack, &MultiplicationNode{Left: left, Right: right})

		case symboltable.DivideOperatorDefinition:
			right, left, err := popTwo(&stack)
			if err != nil {
				return nil, err
			}
			stack = append(stack, &DivisionNode{Left: left, Right: right})

		case symboltable.PowerOperatorDefinition:
			right, left, err := popTwo(&stack)
			if err != nil {
				return nil, err
			}
			expNode, ok := right.(*NumberNode)
			if !ok {
				return nil, fmt.Errorf("power exponent must be a number")
			}
			stack = append(stack, &PowerNode{Base: left, Exponent: int(expNode.Value)})

		case symboltable.PlusOperatorDefinition:
			right, left, err := popTwo(&stack)
			if err != nil {
				return nil, err
			}
			stack = append(stack, &AdditionNode{Left: left, Right: right})

		case symboltable.MinusOperatorDefinition:
			right, left, err := popTwo(&stack)
			if err != nil {
				return nil, err
			}
			stack = append(stack, &SubtractionNode{Left: left, Right: right})
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("malformed expression: %d items remaining on stack", len(stack))
	}
	return stack[0], nil
}

// popTwo pops the top two expressions from the stack.
// For binary operators, the right operand is popped first, then the left.
// Returns an error if fewer than two items are on the stack.
func popTwo(stack *[]Expression) (right Expression, left Expression, err error) {
	if len(*stack) < 2 {
		return nil, nil, fmt.Errorf("insufficient operands for operator")
	}
	right = (*stack)[len(*stack)-1]
	left = (*stack)[len(*stack)-2]
	*stack = (*stack)[:len(*stack)-2]
	return right, left, nil
}
