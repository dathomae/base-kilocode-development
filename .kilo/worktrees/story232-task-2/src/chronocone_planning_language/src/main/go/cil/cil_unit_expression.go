// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

const unitEpsilon = 1e-9

// UnitExpression represents a composite unit as the product of unit references
// with exponents. It is used for tracking dimensional analysis through CIL
// expression trees.
//
// A dimensionless UnitExpression has empty Numerator and Denominator maps.
//
// For example, "m/s^2" is represented as:
//
//	Numerator:   {"m": 1}
//	Denominator: {"s": 2}
//
// Units with fractional exponents are supported (e.g., Meter^1.5).
type UnitExpression struct {
	Numerator   map[string]float64 // unit name -> exponent in numerator
	Denominator map[string]float64 // unit name -> exponent in denominator
}

// NewUnitExpression returns a dimensionless UnitExpression.
func NewUnitExpression() UnitExpression {
	return UnitExpression{
		Numerator:   make(map[string]float64),
		Denominator: make(map[string]float64),
	}
}

// IsDimensionless returns true if the UnitExpression has no units.
func (u UnitExpression) IsDimensionless() bool {
	net := make(map[string]float64)
	for k, v := range u.Numerator {
		net[k] += v
	}
	for k, v := range u.Denominator {
		net[k] -= v
	}
	for _, v := range net {
		if math.Abs(v) > unitEpsilon {
			return false
		}
	}
	return true
}

// Equals returns true if two UnitExpressions represent the same composite unit.
func (u UnitExpression) Equals(other UnitExpression) bool {
	// Compare net exponents so un-normalized forms still match correctly.
	netU := make(map[string]float64)
	for k, v := range u.Numerator {
		netU[k] += v
	}
	for k, v := range u.Denominator {
		netU[k] -= v
	}

	netOther := make(map[string]float64)
	for k, v := range other.Numerator {
		netOther[k] += v
	}
	for k, v := range other.Denominator {
		netOther[k] -= v
	}

	if len(netU) != len(netOther) {
		return false
	}
	for k, v := range netU {
		if math.Abs(netOther[k]-v) > unitEpsilon {
			return false
		}
	}
	return true
}

// normalize returns a new UnitExpression where:
//   - negative exponents in the numerator are moved to the denominator with a positive sign
//   - negative exponents in the denominator are moved to the numerator with a positive sign
//   - units whose net exponent is zero are removed entirely.
func (u UnitExpression) normalize() UnitExpression {
	result := NewUnitExpression()
	net := make(map[string]float64)
	for k, v := range u.Numerator {
		net[k] += v
	}
	for k, v := range u.Denominator {
		net[k] -= v
	}
	for k, v := range net {
		if v > 0 {
			result.Numerator[k] = v
		} else if v < 0 {
			result.Denominator[k] = -v
		}
	}
	return result
}

// Simplify returns a normalized copy of the UnitExpression.
// Matching numerator/denominator pairs are cancelled and zero-exponent
// entries are removed. The receiver is not mutated.
func (u UnitExpression) Simplify() UnitExpression {
	return u.normalize()
}

// String returns a human-readable representation of the UnitExpression.
// The format is "unit1^e1*unit2^e2/unit3^e3*unit4^e4".
// Dimensionless expressions return "1".
func (u UnitExpression) String() string {
	if u.IsDimensionless() {
		return "1"
	}

	var parts []string

	// Format numerator units, sorted for determinism.
	if len(u.Numerator) > 0 {
		numKeys := make([]string, 0, len(u.Numerator))
		for k := range u.Numerator {
			numKeys = append(numKeys, k)
		}
		sort.Strings(numKeys)
		for _, k := range numKeys {
			exp := u.Numerator[k]
			if exp == 1 {
				parts = append(parts, k)
			} else {
				parts = append(parts, fmt.Sprintf("%s^%g", k, exp))
			}
		}
	}

	// Format denominator units, sorted for determinism.
	if len(u.Denominator) > 0 {
		denKeys := make([]string, 0, len(u.Denominator))
		for k := range u.Denominator {
			denKeys = append(denKeys, k)
		}
		sort.Strings(denKeys)
		var denParts []string
		for _, k := range denKeys {
			exp := u.Denominator[k]
			if exp == 1 {
				denParts = append(denParts, k)
			} else {
				denParts = append(denParts, fmt.Sprintf("%s^%g", k, exp))
			}
		}
		if len(parts) == 0 {
			parts = append(parts, "1")
		}
		parts = append(parts, strings.Join(denParts, "*"))
		return strings.Join(parts, "/")
	}

	return strings.Join(parts, "*")
}

// multiply returns a new UnitExpression representing u * other.
func (u UnitExpression) multiply(other UnitExpression) UnitExpression {
	result := NewUnitExpression()
	for k, v := range u.Numerator {
		result.Numerator[k] += v
	}
	for k, v := range u.Denominator {
		result.Denominator[k] += v
	}
	for k, v := range other.Numerator {
		result.Numerator[k] += v
	}
	for k, v := range other.Denominator {
		result.Denominator[k] += v
	}
	return result.normalize()
}

// divide returns a new UnitExpression representing u / other.
func (u UnitExpression) divide(other UnitExpression) UnitExpression {
	result := NewUnitExpression()
	for k, v := range u.Numerator {
		result.Numerator[k] += v
	}
	for k, v := range u.Denominator {
		result.Denominator[k] += v
	}
	for k, v := range other.Numerator {
		result.Denominator[k] += v
	}
	for k, v := range other.Denominator {
		result.Numerator[k] += v
	}
	return result.normalize()
}

// power returns a new UnitExpression with all net exponents multiplied by n.
// The receiver is normalized first so that negative exponents are moved correctly
// before multiplication.
func (u UnitExpression) power(n float64) UnitExpression {
	// Start from a normalized base so negative exponents move correctly.
	norm := u.normalize()
	result := NewUnitExpression()
	for k, v := range norm.Numerator {
		exp := v * n
		if exp > 0 {
			result.Numerator[k] = exp
		} else if exp < 0 {
			result.Denominator[k] = -exp
		}
	}
	for k, v := range norm.Denominator {
		exp := v * n
		if exp > 0 {
			result.Denominator[k] = exp
		} else if exp < 0 {
			result.Numerator[k] = -exp
		}
	}
	return result
}

// unitRefExpression returns a UnitExpression with a single unit in the numerator.
func unitRefExpression(unit string) UnitExpression {
	ue := NewUnitExpression()
	ue.Numerator[unit] = 1
	return ue
}

// ComputeUnitExpression computes the composite UnitExpression for a CIL node.
//
// Supported node types:
//   - *CILUnitRef: single unit in numerator
//   - *CILNumber: annotated unit if present, otherwise dimensionless
//   - *CILLiteral: dimensionless
//   - *CILBinaryOp: Multiply, Divide, Add, Subtract
//   - *CILUnaryOp: passes through the operand's unit (negation does not change units)
//   - *CILPower: base unit expression raised to exponent
//
// Returns an error for unsupported node types or for Add/Subtract when
// operand units are incompatible.
func ComputeUnitExpression(node CILNode) (UnitExpression, error) {
	if node == nil {
		return NewUnitExpression(), fmt.Errorf("cannot compute unit expression for nil node")
	}

	switch n := node.(type) {
	case *CILUnitRef:
		return unitRefExpression(n.UnitGroup), nil

	case *CILNumber:
		if n.AnnotatedUnit != "" {
			return unitRefExpression(n.AnnotatedUnit), nil
		}
		return NewUnitExpression(), nil

	case *CILLiteral:
		return NewUnitExpression(), nil

	case *CILBinaryOp:
		return computeBinaryOpUnit(n)

	case *CILPower:
		return computePowerUnit(n)

	case *CILUnaryOp:
		return ComputeUnitExpression(n.Operand)

	case *CILInputRef:
		// References to inputs/connectors do not carry unit information in the
		// CIL node itself for arithmetic checking; treat them as dimensionless.
		// Unit checking for connector passthroughs against IO spec units is
		// handled separately in CheckMetricOutputUnit.
		return NewUnitExpression(), nil

	case *CILConnectorNode:
		// Connector nodes do not carry unit information; treat as dimensionless.
		return NewUnitExpression(), nil

	case *CILComparisonOp, *CILLogicalOp, *CILNotOp:
		// Boolean operations are dimensionless.
		return NewUnitExpression(), nil

	case *CILNilLiteral, *CILBoolLiteral:
		// Null and boolean literals are dimensionless.
		return NewUnitExpression(), nil

	case *CILConditional:
		return computeConditionalUnit(n)

	case *CILWhenClause:
		return computeWhenClauseUnit(n)

	case *CILLookupNode:
		// Lookups do not carry unit information in the CIL node.
		return NewUnitExpression(), nil

	default:
		return NewUnitExpression(), fmt.Errorf("unsupported CIL node type %T for unit expression computation", node)
	}
}

func computeBinaryOpUnit(node *CILBinaryOp) (UnitExpression, error) {
	left, err := ComputeUnitExpression(node.Left)
	if err != nil {
		return NewUnitExpression(), fmt.Errorf("left operand: %w", err)
	}
	right, err := ComputeUnitExpression(node.Right)
	if err != nil {
		return NewUnitExpression(), fmt.Errorf("right operand: %w", err)
	}

	switch node.Op {
	case operator.Multiply:
		return left.multiply(right), nil

	case operator.Divide:
		return left.divide(right), nil

	case operator.Add, operator.Subtract:
		if !left.Equals(right) {
			return NewUnitExpression(), fmt.Errorf(
				"incompatible units for %s: %s vs %s",
				node.Op.Symbol(), left.String(), right.String(),
			)
		}
		return left, nil

	default:
		return NewUnitExpression(), fmt.Errorf("unsupported binary operator %v for unit expression computation", node.Op)
	}
}

func computePowerUnit(node *CILPower) (UnitExpression, error) {
	base, err := ComputeUnitExpression(node.Base)
	if err != nil {
		return NewUnitExpression(), fmt.Errorf("power base: %w", err)
	}
	return base.power(float64(node.Exponent)), nil
}

func computeConditionalUnit(node *CILConditional) (UnitExpression, error) {
	if len(node.Branches) == 0 {
		return NewUnitExpression(), fmt.Errorf("conditional has no branches")
	}
	if node.Branches[0].Body == nil {
		return NewUnitExpression(), fmt.Errorf("conditional branch 0 body is nil")
	}
	firstUnit, err := ComputeUnitExpression(node.Branches[0].Body)
	if err != nil {
		return NewUnitExpression(), fmt.Errorf("conditional branch: %w", err)
	}
	for i := 1; i < len(node.Branches); i++ {
		if node.Branches[i].Body == nil {
			return NewUnitExpression(), fmt.Errorf("conditional branch %d body is nil", i)
		}
		branchUnit, err := ComputeUnitExpression(node.Branches[i].Body)
		if err != nil {
			return NewUnitExpression(), fmt.Errorf("conditional branch %d: %w", i, err)
		}
		if !firstUnit.Equals(branchUnit) {
			return NewUnitExpression(), fmt.Errorf(
				"conditional branch %d has unit %s which differs from first branch %s",
				i, branchUnit.String(), firstUnit.String(),
			)
		}
	}
	return firstUnit, nil
}

func computeWhenClauseUnit(node *CILWhenClause) (UnitExpression, error) {
	if len(node.Clauses) == 0 {
		if node.Default == nil {
			return NewUnitExpression(), fmt.Errorf("when clause has no clauses and no default")
		}
		return ComputeUnitExpression(node.Default)
	}
	if node.Clauses[0].Body == nil {
		return NewUnitExpression(), fmt.Errorf("when clause 0 body is nil")
	}
	firstUnit, err := ComputeUnitExpression(node.Clauses[0].Body)
	if err != nil {
		return NewUnitExpression(), fmt.Errorf("when clause: %w", err)
	}
	for i := 1; i < len(node.Clauses); i++ {
		if node.Clauses[i].Body == nil {
			return NewUnitExpression(), fmt.Errorf("when clause %d body is nil", i)
		}
		clauseUnit, err := ComputeUnitExpression(node.Clauses[i].Body)
		if err != nil {
			return NewUnitExpression(), fmt.Errorf("when clause %d: %w", i, err)
		}
		if !firstUnit.Equals(clauseUnit) {
			return NewUnitExpression(), fmt.Errorf(
				"when clause %d has unit %s which differs from first clause %s",
				i, clauseUnit.String(), firstUnit.String(),
			)
		}
	}
	if node.Default != nil {
		defaultUnit, err := ComputeUnitExpression(node.Default)
		if err != nil {
			return NewUnitExpression(), fmt.Errorf("when default: %w", err)
		}
		if !firstUnit.Equals(defaultUnit) {
			return NewUnitExpression(), fmt.Errorf(
				"when default has unit %s which differs from first clause %s",
				defaultUnit.String(), firstUnit.String(),
			)
		}
	}
	return firstUnit, nil
}
