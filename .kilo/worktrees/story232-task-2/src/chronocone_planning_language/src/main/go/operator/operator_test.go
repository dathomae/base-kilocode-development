// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package operator

import "testing"

func TestOperator_String(t *testing.T) {
	tests := []struct {
		op       Operator
		expected string
	}{
		{Unknown, "Unknown"},
		{Add, "Add"},
		{Subtract, "Subtract"},
		{Multiply, "Multiply"},
		{Divide, "Divide"},
		{Negate, "Negate"},
		{GreaterThan, "GreaterThan"},
		{GreaterThanOrEqual, "GreaterThanOrEqual"},
		{LessThan, "LessThan"},
		{LessThanOrEqual, "LessThanOrEqual"},
		{Equal, "Equal"},
		{NotEqual, "NotEqual"},
		{And, "And"},
		{Or, "Or"},
		{Not, "Not"},
		{Bool, "Bool"},
		{Nil, "Nil"},
		{AbsoluteValue, "AbsoluteValue"},
		{Operator(999), "Operator(999)"},
	}

	for _, tt := range tests {
		got := tt.op.String()
		if got != tt.expected {
			t.Errorf("%v.String() = %q, want %q", tt.op, got, tt.expected)
		}
	}
}

func TestOperator_Symbol(t *testing.T) {
	tests := []struct {
		op       Operator
		expected string
	}{
		{Unknown, ""},
		{Add, "+"},
		{Subtract, "-"},
		{Multiply, "*"},
		{Divide, "/"},
		{Negate, "negate"},
		{GreaterThan, ">"},
		{GreaterThanOrEqual, ">="},
		{LessThan, "<"},
		{LessThanOrEqual, "<="},
		{Equal, "=="},
		{NotEqual, "!="},
		{And, "and"},
		{Or, "or"},
		{Not, "not"},
		{Bool, ""},
		{Nil, ""},
		{AbsoluteValue, ""},
		{Operator(999), ""},
	}

	for _, tt := range tests {
		got := tt.op.Symbol()
		if got != tt.expected {
			t.Errorf("%v.Symbol() = %q, want %q", tt.op, got, tt.expected)
		}
	}
}
