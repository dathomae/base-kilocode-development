// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "fmt"

// CheckMetricOutputUnit verifies that the unit computed from a metric's CIL
// expression tree matches the unit required by the IO spec output.
//
// It walks the CIL tree using ComputeUnitExpression, simplifies the resulting
// UnitExpression, and compares it against ioSpecOutputUnit. When the units do
// not match, it checks the provided conversions for a bridging pair. Detailed,
// actionable error messages are returned on mismatch.
func CheckMetricOutputUnit(
	metricCILRoot CILNode,
	ioSpecOutputUnit string,
	conversions []CILConversionPair,
) error {
	if metricCILRoot == nil {
		return fmt.Errorf("metric CIL root is nil")
	}

	computedUE, err := ComputeUnitExpression(metricCILRoot)
	if err != nil {
		return err
	}

	simplifiedUE := computedUE.Simplify()

	if unitExpressionsMatch(simplifiedUE, ioSpecOutputUnit) {
		return nil
	}

	computedName := unitExpressionToName(simplifiedUE)
	expectedName := ioSpecOutputUnit

	// If IO spec does not declare a unit, accept any computed unit.
	// Unit checking is only enforced when the IO spec explicitly requires one.
	if expectedName == "" {
		return nil
	}

	// If computed unit is dimensionless but expected is not, check if the
	// metric root is a connector reference with an attached unit name.
	// This handles identity passthrough metrics where the connector's celltype
	// measure declares a unit that may differ from the IO spec output unit.
	if simplifiedUE.IsDimensionless() {
		if ref, ok := metricCILRoot.(*CILInputRef); ok && ref.UnitName != "" {
			computedUE = unitRefExpression(ref.UnitName)
			simplifiedUE = computedUE.Simplify()
			computedName = unitExpressionToName(simplifiedUE)
			if unitExpressionsMatch(simplifiedUE, ioSpecOutputUnit) {
				return nil
			}
		} else {
			return fmt.Errorf(
				"computed unit dimensionless but IO spec requires %s; add a conversion or change the formula",
				expectedName,
			)
		}
	}

	// Attempt to find a conversion bridging the gap.
	unitGroups := make(map[string]string)
	for _, pair := range conversions {
		if pair.SourceGroup != "" {
			unitGroups[pair.SourceName] = pair.SourceGroup
		}
		if pair.TargetGroup != "" {
			unitGroups[pair.TargetName] = pair.TargetGroup
		}
	}
	if err := ValidateConversion(computedName, ioSpecOutputUnit, conversions, unitGroups); err != nil {
		return fmt.Errorf(
			"computed unit %s but IO spec requires %s; add a conversion or change the formula: %w",
			computedName,
			expectedName,
			err,
		)
	}

	return nil
}

// unitExpressionsMatch returns true when the UnitExpression matches the
// expected unit string. An empty expected string means dimensionless.
func unitExpressionsMatch(ue UnitExpression, expected string) bool {
	if expected == "" {
		return ue.IsDimensionless()
	}
	expectedUE := unitRefExpression(expected)
	return ue.Equals(expectedUE)
}

// unitExpressionToName extracts a human-readable unit name from a
// UnitExpression for use in error messages and conversion lookup.
//
// If the simplified expression is dimensionless, it returns "".
// If the simplified expression has exactly one unit in the numerator and
// none in the denominator, it returns that unit name.
// Otherwise it returns the String() representation of the expression.
func unitExpressionToName(ue UnitExpression) string {
	if ue.IsDimensionless() {
		return ""
	}
	simplified := ue.Simplify()
	if len(simplified.Numerator) == 1 && len(simplified.Denominator) == 0 {
		for unit := range simplified.Numerator {
			return unit
		}
	}
	return simplified.String()
}
