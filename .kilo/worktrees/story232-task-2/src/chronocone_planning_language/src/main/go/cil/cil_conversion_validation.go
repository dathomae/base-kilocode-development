// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "fmt"

// ValidateConversion checks whether a conversion path exists between
// sourceUnit and targetUnit using the provided CILConversionPair slice.
//
// Conversion pairs are treated as bidirectional: a pair with SourceName "A"
// and TargetName "B" satisfies a request to convert from "B" to "A" as well.
//
// If sourceUnit equals targetUnit, validation succeeds immediately with no
// error. When no matching conversion pair is found, a descriptive error is
// returned.
func ValidateConversion(
	sourceUnit string,
	targetUnit string,
	conversions []CILConversionPair,
	unitGroups map[string]string,
) error {
	if sourceUnit == targetUnit {
		return nil
	}

	for _, pair := range conversions {
		if (pair.SourceName == sourceUnit && pair.TargetName == targetUnit) ||
			(pair.SourceName == targetUnit && pair.TargetName == sourceUnit) {
			return nil
		}
	}

	// Fallback: build unit groups from conversion pairs when caller didn't supply a map.
	if unitGroups == nil {
		unitGroups = make(map[string]string)
		for _, pair := range conversions {
			if pair.SourceGroup != "" {
				unitGroups[pair.SourceName] = pair.SourceGroup
			}
			if pair.TargetGroup != "" {
				unitGroups[pair.TargetName] = pair.TargetGroup
			}
		}
	}

	srcGroup := unitGroups[sourceUnit]
	tgtGroup := unitGroups[targetUnit]
	if srcGroup != "" && tgtGroup != "" && srcGroup != tgtGroup {
		return fmt.Errorf("no conversion available from %s to %s (different unit groups)", sourceUnit, targetUnit)
	}
	return fmt.Errorf("no conversion available from %s to %s", sourceUnit, targetUnit)
}

// ValidateConversionClosure checks that every pair of units sharing the same
// unit group has at least one conversion pair (forward or reverse) connecting
// them. It groups conversion pairs by their unit group, then for each group,
// verifies that all pairs of units within that group have a connection.
//
// Returns nil if closure is satisfied for all groups.
// Returns an error listing the first missing connection found, with unit group
// context.
func ValidateConversionClosure(
	conversionPairs []CILConversionPair,
	unitGroups map[string]string, // unit name -> unit group name (e.g., "Meter" -> "DistanceUnits")
) error {
	if len(unitGroups) == 0 {
		return nil
	}

	// Build adjacency set: mark both directions as connected.
	connected := make(map[string]bool)
	for _, pair := range conversionPairs {
		connected[pair.SourceName+"->"+pair.TargetName] = true
		connected[pair.TargetName+"->"+pair.SourceName] = true
	}

	// Group units by their unit group name.
	groups := make(map[string][]string)
	for unit, groupName := range unitGroups {
		groups[groupName] = append(groups[groupName], unit)
	}

	// For each unit group with 2+ units, check all pairs.
	for groupName, units := range groups {
		if len(units) < 2 {
			continue
		}
		for i := 0; i < len(units); i++ {
			for j := i + 1; j < len(units); j++ {
				u1, u2 := units[i], units[j]
				if !connected[u1+"->"+u2] && !connected[u2+"->"+u1] {
					return fmt.Errorf("no conversion path between %s and %s in unit group %s", u1, u2, groupName)
				}
			}
		}
	}

	return nil
}

// ValidateConversionWithClosure performs both direct pair validation and
// unit-group closure validation. It first runs ValidateConversionClosure to
// verify group completeness, then runs ValidateConversion for the specific pair.
//
// This is the recommended entry point for comprehensive conversion validation.
func ValidateConversionWithClosure(
	sourceUnit string,
	targetUnit string,
	conversionPairs []CILConversionPair,
	unitGroups map[string]string,
) error {
	if err := ValidateConversionClosure(conversionPairs, unitGroups); err != nil {
		return err
	}
	return ValidateConversion(sourceUnit, targetUnit, conversionPairs, unitGroups)
}
