// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file defines DimensionalUnitExpression and aggregator-related types
// used for dimensional unit analysis during the aggregator analysis phase.
package analysis

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// DimensionalUnitExpression represents a symbolic dimensional unit.
// It uses the same algebra as regular unit/measure conversions (*, /, ^ only).
// Examples: "Product/Country", "Country", "Unitless", "1/Country".
type DimensionalUnitExpression struct {
	// Numerator holds concept names in the numerator (sorted for canonical form).
	Numerator []string
	// Denominator holds concept names in the denominator (sorted for canonical form).
	Denominator []string
}

// Multiply returns a new DimensionalUnitExpression representing this * other.
// It concatenates numerators and denominators, then calls cancel().
func (d *DimensionalUnitExpression) Multiply(other *DimensionalUnitExpression) *DimensionalUnitExpression {
	result := &DimensionalUnitExpression{
		Numerator:   make([]string, len(d.Numerator)+len(other.Numerator)),
		Denominator: make([]string, len(d.Denominator)+len(other.Denominator)),
	}
	copy(result.Numerator, d.Numerator)
	copy(result.Numerator[len(d.Numerator):], other.Numerator)
	copy(result.Denominator, d.Denominator)
	copy(result.Denominator[len(d.Denominator):], other.Denominator)
	result.cancel()
	return result
}

// Equals compares two DimensionalUnitExpression values after sorting both
// numerator and denominator slices.
func (d *DimensionalUnitExpression) Equals(other *DimensionalUnitExpression) bool {
	dNum := make([]string, len(d.Numerator))
	dDen := make([]string, len(d.Denominator))
	oNum := make([]string, len(other.Numerator))
	oDen := make([]string, len(other.Denominator))
	copy(dNum, d.Numerator)
	copy(dDen, d.Denominator)
	copy(oNum, other.Numerator)
	copy(oDen, other.Denominator)
	sort.Strings(dNum)
	sort.Strings(dDen)
	sort.Strings(oNum)
	sort.Strings(oDen)
	return slices.Equal(dNum, oNum) && slices.Equal(dDen, oDen)
}

// IsUnitless returns true if both numerator and denominator are empty.
func (d *DimensionalUnitExpression) IsUnitless() bool {
	return len(d.Numerator) == 0 && len(d.Denominator) == 0
}

// cancel removes elements that appear in both numerator and denominator
// (multiset cancellation), then sorts both slices for canonical form.
func (d *DimensionalUnitExpression) cancel() {
	// Build a multiset of denominator elements
	denomCounts := make(map[string]int)
	for _, c := range d.Denominator {
		denomCounts[c]++
	}

	// Remove from numerator any elements that appear in denominator
	var newNum []string
	for _, c := range d.Numerator {
		if denomCounts[c] > 0 {
			denomCounts[c]--
		} else {
			newNum = append(newNum, c)
		}
	}
	d.Numerator = newNum

	// Rebuild denominator from remaining counts
	var newDen []string
	for c, count := range denomCounts {
		for i := 0; i < count; i++ {
			newDen = append(newDen, c)
		}
	}
	d.Denominator = newDen

	// Sort for canonical form
	sort.Strings(d.Numerator)
	sort.Strings(d.Denominator)
}

// String returns a human-readable representation of the dimensional unit.
func (d *DimensionalUnitExpression) String() string {
	if d.IsUnitless() {
		return "Unitless"
	}

	numStr := strings.Join(d.Numerator, "*")
	if len(d.Denominator) == 0 {
		return numStr
	}

	denStr := strings.Join(d.Denominator, "*")
	if len(d.Denominator) > 1 {
		denStr = "(" + denStr + ")"
	}
	if len(d.Numerator) == 0 {
		return "1/" + denStr
	}

	return numStr + "/" + denStr
}

// dimensionalUnitsFromConceptSet computes the dimensional units for a metric
// with the given concept dimensions.
//
// For n concepts [D1, D2, ..., Dn]:
//   - n=0 returns unitless
//   - n=1 returns {Numerator: [D1]}
//   - n>1 returns {Numerator: [Dn], Denominator: [D1..Dn-1]}
func dimensionalUnitsFromConceptSet(concepts []*AnalysisConceptDefinition) *DimensionalUnitExpression {
	n := len(concepts)
	if n == 0 {
		return &DimensionalUnitExpression{}
	}

	names := make([]string, n)
	for i, c := range concepts {
		names[i] = c.ConceptName
	}

	if n == 1 {
		return &DimensionalUnitExpression{
			Numerator: []string{names[0]},
		}
	}

	return &DimensionalUnitExpression{
		Numerator:   []string{names[n-1]},
		Denominator: names[:n-1],
	}
}

// AggregatorMapKey constructs the compound key for the Aggregators map.
// The key is deterministic: concept names are sorted to ensure
// [Country, Product]->[Product] and [Product, Country]->[Product] produce the same key.
func AggregatorMapKey(fromConcepts []string, toConcept string) string {
	sorted := make([]string, len(fromConcepts))
	copy(sorted, fromConcepts)
	sort.Strings(sorted)
	if toConcept == "" {
		return fmt.Sprintf("[%s]->[]", strings.Join(sorted, ","))
	}
	return fmt.Sprintf("[%s]->[%s]", strings.Join(sorted, ","), toConcept)
}
