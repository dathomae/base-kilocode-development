// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

// Package operator provides the typed Operator enumeration used across the
// CIL transformer, postfix walker, and calculation evaluator. It lives in its
// own package to avoid an import cycle between cil and calculation.
package operator

import "fmt"

// Operator represents a typed enumeration of all supported CIL operators.
// It replaces string-based operator dispatch with a compile-time-safe
// identifier that can be used across the CIL transformer, postfix walker,
// and calculation evaluator.
type Operator int

const (
	Unknown Operator = iota
	Add
	Subtract
	Multiply
	Divide
	Negate
	GreaterThan
	GreaterThanOrEqual
	LessThan
	LessThanOrEqual
	Equal
	NotEqual
	And
	Or
	Not
	Bool // Literal marker for boolean values in postfix tables.
	Nil  // Literal marker for nil values in postfix tables.
	AbsoluteValue // Reserved for future use.
)

// String returns a human-readable name for the operator.
func (o Operator) String() string {
	switch o {
	case Unknown:
		return "Unknown"
	case Add:
		return "Add"
	case Subtract:
		return "Subtract"
	case Multiply:
		return "Multiply"
	case Divide:
		return "Divide"
	case Negate:
		return "Negate"
	case GreaterThan:
		return "GreaterThan"
	case GreaterThanOrEqual:
		return "GreaterThanOrEqual"
	case LessThan:
		return "LessThan"
	case LessThanOrEqual:
		return "LessThanOrEqual"
	case Equal:
		return "Equal"
	case NotEqual:
		return "NotEqual"
	case And:
		return "And"
	case Or:
		return "Or"
	case Not:
		return "Not"
	case Bool:
		return "Bool"
	case Nil:
		return "Nil"
	case AbsoluteValue:
		return "AbsoluteValue"
	default:
		return fmt.Sprintf("Operator(%d)", o)
	}
}

// Symbol returns the canonical symbol for the operator.
// For arithmetic and comparison operators this is the single- or double-character
// token (e.g. "+", "-", "=="). For logical operators it returns the keyword
// (e.g. "and", "or", "not"). For the negate operator it returns "negate".
// For Bool and Nil markers it returns an empty string.
func (o Operator) Symbol() string {
	switch o {
	case Add:
		return "+"
	case Subtract:
		return "-"
	case Multiply:
		return "*"
	case Divide:
		return "/"
	case Negate:
		return "negate"
	case GreaterThan:
		return ">"
	case GreaterThanOrEqual:
		return ">="
	case LessThan:
		return "<"
	case LessThanOrEqual:
		return "<="
	case Equal:
		return "=="
	case NotEqual:
		return "!="
	case And:
		return "and"
	case Or:
		return "or"
	case Not:
		return "not"
	default:
		return ""
	}
}
