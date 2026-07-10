// Package parsing provides data structures for the CPL parsing phase.
//
// This file defines typed postfix expression elements that replace the
// previous []interface{} representation. Each concrete element type
// satisfies the PostfixElement interface, enabling type-safe dispatch
// in the postfix evaluator (buildCILTreeFromPostfix) without relying
// on type switches over bare Go types.
package parsing

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// PostfixElement is the interface for typed postfix expression elements.
// Concrete implementations are PostfixLiteral, PostfixIdentifier, and
// PostfixOperator.
type PostfixElement interface {
	isPostfixElement()
}

// PostfixLiteral represents a numeric literal (int, int64, or float64).
// All numeric values are stored as float64 for uniform handling in the
// CIL tree builder.
type PostfixLiteral struct {
	Value float64
}

func (PostfixLiteral) isPostfixElement() {}

// PostfixIdentifier represents a named reference (metric, dimension, connector).
// The Name field holds the identifier text as it appears in the source.
type PostfixIdentifier struct {
	Name string
}

func (PostfixIdentifier) isPostfixElement() {}

// PostfixOperator represents an operator or boolean/nil literal carried as a
// symboltable.DefinitionType constant. The DefType field is dispatched in
// buildCILTreeFromPostfix to produce the appropriate CIL node (binary op,
// unary op, comparison, logical, boolean literal, or nil literal).
type PostfixOperator struct {
	DefType symboltable.DefinitionType
}

func (PostfixOperator) isPostfixElement() {}
