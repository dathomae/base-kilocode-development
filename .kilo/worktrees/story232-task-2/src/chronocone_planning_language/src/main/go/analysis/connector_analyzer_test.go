// Package analysis_test contains unit tests for the ConnectorAnalyzer.
//
// These tests verify that ConnectorAnalyzer correctly stores, retrieves,
// and validates connector definitions.
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ConnectorAnalyzer Construction Tests
// ============================================================================

// TestNewConnectorAnalyzer_InitializesEmptyMap verifies that NewConnectorAnalyzer
// creates a ConnectorAnalyzer with an empty connectors map.
func TestNewConnectorAnalyzer_InitializesEmptyMap(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()

	require.NotNil(t, ca, "NewConnectorAnalyzer should return non-nil")
	connectors := ca.GetConnectors()
	assert.Empty(t, connectors, "connectors map should be empty")
}

// ============================================================================
// Connector Registration and Retrieval Tests
// ============================================================================

// TestConnectorAnalyzer_AddConnectorAndGetConnector verifies that a connector
// can be added and then retrieved by name.
func TestConnectorAnalyzer_AddConnectorAndGetConnector(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	connector := analysis.AnalysisConnectorDefinition{
		Name:           "SalesData",
		Dimensions:     nil,
		CellType:       nil,
		Grain:          nil,
		SourceLocation: symboltable.SourceLocation{Line: 1, Column: 1},
	}

	ca.AddConnector("SalesData", connector)

	retrieved := ca.GetConnector("SalesData")
	assert.Equal(t, "SalesData", retrieved.Name)
	assert.Equal(t, 1, retrieved.SourceLocation.Line)
}

// TestConnectorAnalyzer_GetConnector_MissingReturnsZeroValue verifies that
// getting a missing connector returns a zero-value AnalysisConnectorDefinition.
func TestConnectorAnalyzer_GetConnector_MissingReturnsZeroValue(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	retrieved := ca.GetConnector("Missing")

	assert.Equal(t, "", retrieved.Name, "missing connector should return zero value")
}

// TestConnectorAnalyzer_GetConnectors_ReturnsMap verifies that
// GetConnectors returns the internal map and modifications
// through it are visible.
func TestConnectorAnalyzer_GetConnectors_ReturnsMap(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	connectors := ca.GetConnectors()

	connectors["DirectAdd"] = analysis.AnalysisConnectorDefinition{
		Name: "DirectAdd",
	}

	retrieved := ca.GetConnector("DirectAdd")
	assert.Equal(t, "DirectAdd", retrieved.Name)
}

// TestConnectorAnalyzer_AddConnector_OverwritesExisting verifies that
// adding a connector with the same name overwrites the previous definition.
func TestConnectorAnalyzer_AddConnector_OverwritesExisting(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	ca.AddConnector("SalesData", analysis.AnalysisConnectorDefinition{
		Name: "SalesData",
	})
	ca.AddConnector("SalesData", analysis.AnalysisConnectorDefinition{
		Name:       "SalesData",
		Dimensions: []*analysis.AnalysisConceptDefinition{{ConceptName: "Country"}},
	})

	retrieved := ca.GetConnector("SalesData")
	require.Len(t, retrieved.Dimensions, 1)
	assert.Equal(t, "Country", retrieved.Dimensions[0].ConceptName)
}

// ============================================================================
// ValidateGrainTypes Tests
// ============================================================================

// TestConnectorAnalyzer_ValidateGrainTypes_NilGrain_Skipped verifies that connectors with
// no explicit grain are ignored by the deferred grain type validation pass.
func TestConnectorAnalyzer_ValidateGrainTypes_NilGrain_Skipped(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	ca.AddConnector("NoGrainConnector", analysis.AnalysisConnectorDefinition{
		Name:  "NoGrainConnector",
		Grain: nil,
	})

	typeAnalyzer := analysis.NewTypeAnalyzer()

	err := ca.ValidateGrainTypes(typeAnalyzer)

	require.NoError(t, err, "connector without explicit grain should be skipped")
}

// TestConnectorAnalyzer_ValidateGrainTypes_BuiltInInterval_Accepted verifies that a
// connector grain measure whose TypeName is the built-in Interval type is valid.
func TestConnectorAnalyzer_ValidateGrainTypes_BuiltInInterval_Accepted(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "DailyGrain",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	ca.AddConnector("IntervalConnector", analysis.AnalysisConnectorDefinition{
		Name:  "IntervalConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	})

	typeAnalyzer := analysis.NewTypeAnalyzer()

	err := ca.ValidateGrainTypes(typeAnalyzer)

	require.NoError(t, err, "built-in Interval grain type should be accepted")
}

// TestConnectorAnalyzer_ValidateGrainTypes_DerivedInterval_Accepted verifies that a
// connector grain measure whose declared type derives from Interval is valid.
func TestConnectorAnalyzer_ValidateGrainTypes_DerivedInterval_Accepted(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	typeAnalyzer := analysis.NewTypeAnalyzer()
	typeAnalyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "DailyInterval",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.IntervalType),
		DerivationPath: []string{"DailyInterval", "Interval"},
	})
	typeAnalyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "BusinessDayInterval",
		BaseType:       analysis.NewDeclaredTypeRef("DailyInterval"),
		DerivationPath: []string{"BusinessDayInterval", "DailyInterval", "Interval"},
	})

	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "BusinessDayGrain",
		TypeName: "BusinessDayInterval",
		UnitExpression: &analysis.UnitReference{Name: "business_day"},
	}
	ca.AddConnector("DerivedIntervalConnector", analysis.AnalysisConnectorDefinition{
		Name:  "DerivedIntervalConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	})

	err := ca.ValidateGrainTypes(typeAnalyzer)

	require.NoError(t, err, "Interval-derived grain type should be accepted")
}

// TestConnectorAnalyzer_ValidateGrainTypes_BuiltInNumber_Rejected verifies that a
// connector grain measure whose TypeName is the built-in Number type is invalid.
func TestConnectorAnalyzer_ValidateGrainTypes_BuiltInNumber_Rejected(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "CountGrain",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "count"},
	}
	ca.AddConnector("NumberConnector", analysis.AnalysisConnectorDefinition{
		Name:  "NumberConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	})

	typeAnalyzer := analysis.NewTypeAnalyzer()

	err := ca.ValidateGrainTypes(typeAnalyzer)

	require.Error(t, err, "built-in Number grain type should be rejected")
	assert.Contains(t, err.Error(), "NumberConnector", "error should identify the connector")
	assert.Contains(t, err.Error(), "CountGrain", "error should identify the grain measure")
	assert.Contains(t, err.Error(), "Number", "error should identify the invalid grain type")
}

// TestConnectorAnalyzer_ValidateGrainTypes_DerivedNumber_Rejected verifies that a
// connector grain measure whose declared type derives from Number is invalid.
func TestConnectorAnalyzer_ValidateGrainTypes_DerivedNumber_Rejected(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	typeAnalyzer := analysis.NewTypeAnalyzer()
	typeAnalyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		DerivationPath: []string{"Currency", "Number"},
	})
	typeAnalyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "RetailCurrency",
		BaseType:       analysis.NewDeclaredTypeRef("Currency"),
		DerivationPath: []string{"RetailCurrency", "Currency", "Number"},
	})

	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "CurrencyGrain",
		TypeName: "RetailCurrency",
		UnitExpression: &analysis.UnitReference{Name: "dollar"},
	}
	ca.AddConnector("DerivedNumberConnector", analysis.AnalysisConnectorDefinition{
		Name:  "DerivedNumberConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	})

	err := ca.ValidateGrainTypes(typeAnalyzer)

	require.Error(t, err, "Number-derived grain type should be rejected")
	assert.Contains(t, err.Error(), "DerivedNumberConnector", "error should identify the connector")
	assert.Contains(t, err.Error(), "CurrencyGrain", "error should identify the grain measure")
	assert.Contains(t, err.Error(), "RetailCurrency", "error should identify the invalid declared grain type")
}

// TestConnectorAnalyzer_ValidateGrainTypes_UndeclaredType_Rejected verifies that a
// connector grain measure whose TypeName is not a built-in or declared type is invalid.
func TestConnectorAnalyzer_ValidateGrainTypes_UndeclaredType_Rejected(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	grainMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "MysteryGrain",
		TypeName: "UndeclaredType",
		UnitExpression: &analysis.UnitReference{Name: "mystery_unit"},
	}
	ca.AddConnector("UndeclaredTypeConnector", analysis.AnalysisConnectorDefinition{
		Name:  "UndeclaredTypeConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&grainMeasure},
	})

	typeAnalyzer := analysis.NewTypeAnalyzer()

	err := ca.ValidateGrainTypes(typeAnalyzer)

	require.Error(t, err, "undeclared grain type should be rejected")
	assert.Contains(t, err.Error(), "UndeclaredTypeConnector", "error should identify the connector")
	assert.Contains(t, err.Error(), "MysteryGrain", "error should identify the grain measure")
	assert.Contains(t, err.Error(), "UndeclaredType", "error should identify the undeclared grain type")
}

// TestConnectorAnalyzer_ValidateGrainTypes_MultipleIntervalGrains_Accepted verifies that
// a connector with multiple grains, all of which have the built-in Interval type,
// passes deferred grain type validation.
func TestConnectorAnalyzer_ValidateGrainTypes_MultipleIntervalGrains_Accepted(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()

	yearMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Year",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "year"},
	}
	monthMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Month",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "month"},
	}
	dayMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	ca.AddConnector("MultiGrainConnector", analysis.AnalysisConnectorDefinition{
		Name:  "MultiGrainConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&yearMeasure, &monthMeasure, &dayMeasure},
	})

	typeAnalyzer := analysis.NewTypeAnalyzer()

	err := ca.ValidateGrainTypes(typeAnalyzer)

	require.NoError(t, err, "Multiple Interval grains should be accepted")
}

// TestConnectorAnalyzer_ValidateGrainTypes_MultipleGrains_FirstInvalid_Rejected verifies
// that when the first grain in a multi-grain connector has an invalid type, the
// deferred validation returns an error identifying the connector, grain index [0],
// grain measure name, and invalid type.
func TestConnectorAnalyzer_ValidateGrainTypes_MultipleGrains_FirstInvalid_Rejected(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()

	numberMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "CountGrain",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "count"},
	}
	intervalMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	ca.AddConnector("MixedGrainConnector", analysis.AnalysisConnectorDefinition{
		Name:  "MixedGrainConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&numberMeasure, &intervalMeasure},
	})

	typeAnalyzer := analysis.NewTypeAnalyzer()

	err := ca.ValidateGrainTypes(typeAnalyzer)

	require.Error(t, err, "Should reject connector when first grain has invalid type")
	assert.Contains(t, err.Error(), "MixedGrainConnector", "Error should identify the connector")
	assert.Contains(t, err.Error(), "[0]", "Error should identify grain index [0]")
	assert.Contains(t, err.Error(), "CountGrain", "Error should identify the grain measure")
	assert.Contains(t, err.Error(), "Number", "Error should identify the invalid type")
}

// TestConnectorAnalyzer_ValidateGrainTypes_MultipleGrains_SecondInvalid_Rejected verifies
// that when the second grain in a multi-grain connector has an invalid type, the
// deferred validation returns an error identifying grain index [1].
func TestConnectorAnalyzer_ValidateGrainTypes_MultipleGrains_SecondInvalid_Rejected(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()

	intervalMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	numberMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "CountGrain",
		TypeName: "Number",
		UnitExpression: &analysis.UnitReference{Name: "count"},
	}

	ca.AddConnector("MixedGrainConnector", analysis.AnalysisConnectorDefinition{
		Name:  "MixedGrainConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&intervalMeasure, &numberMeasure},
	})

	typeAnalyzer := analysis.NewTypeAnalyzer()

	err := ca.ValidateGrainTypes(typeAnalyzer)

	require.Error(t, err, "Should reject connector when second grain has invalid type")
	assert.Contains(t, err.Error(), "[1]", "Error should identify grain index [1]")
}

// TestConnectorAnalyzer_ValidateGrainTypes_MultipleGrains_DerivedInterval_Accepted verifies
// that a connector with multiple grains where one derives from Interval and
// another is the built-in Interval type passes deferred validation.
func TestConnectorAnalyzer_ValidateGrainTypes_MultipleGrains_DerivedInterval_Accepted(t *testing.T) {
	t.Parallel()

	ca := analysis.NewConnectorAnalyzer()
	typeAnalyzer := analysis.NewTypeAnalyzer()
	typeAnalyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Daily",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.IntervalType),
		DerivationPath: []string{"Daily", "Interval"},
	})

	derivedMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "DailyGrain",
		TypeName: "Daily",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}
	intervalMeasure := analysis.AnalysisMeasureDefinition{
		Name:     "Day",
		TypeName: "Interval",
		UnitExpression: &analysis.UnitReference{Name: "day"},
	}

	ca.AddConnector("DerivedMultiGrainConnector", analysis.AnalysisConnectorDefinition{
		Name:  "DerivedMultiGrainConnector",
		Grain: []*analysis.AnalysisMeasureDefinition{&derivedMeasure, &intervalMeasure},
	})

	err := ca.ValidateGrainTypes(typeAnalyzer)

	require.NoError(t, err, "Derived Interval and built-in Interval grains should both be accepted")
}
