// Package analysis_test contains tests for the MetricAnalyzer.
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	cpio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Metric Registration and Retrieval Tests
// ============================================================================

// TestNewMetricAnalyzer_InitializesMaps verifies that NewMetricAnalyzer creates
// a MetricAnalyzer with initialized internal maps.
func TestNewMetricAnalyzer_InitializesMaps(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()

	require.NotNil(t, ma, "NewMetricAnalyzer should not return nil")
	assert.Empty(t, ma.GetMetrics(), "metrics map should be empty initially")
	assert.Empty(t, ma.GetCILTrees(), "cilTrees map should be empty initially")
	assert.Nil(t, ma.GetIOSpec(), "ioSpec should be nil initially")
}

// TestMetricAnalyzer_MetricRegistration verifies that metrics can be registered
// and retrieved directly on the MetricAnalyzer.
func TestMetricAnalyzer_MetricRegistration(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()

	ma.AddMetric("test_metric", analysis.AnalysisMetricDefinition{
		Name:   "test_metric",
		Inputs: []analysis.AnalysisInputDefinition{},
	})

	require.True(t, ma.HasMetric("test_metric"))
	assert.Equal(t, "test_metric", ma.GetMetric("test_metric").Name)
	assert.Empty(t, ma.GetMetric("test_metric").Inputs)
}

// TestMetricAnalyzer_MultipleMetrics verifies that multiple metrics can be
// registered and retrieved.
func TestMetricAnalyzer_MultipleMetrics(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()

	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{Name: "revenue"})
	ma.AddMetric("profit", analysis.AnalysisMetricDefinition{Name: "profit"})
	ma.AddMetric("margin", analysis.AnalysisMetricDefinition{Name: "margin"})

	assert.Len(t, ma.GetMetrics(), 3)
	assert.True(t, ma.HasMetric("revenue"))
	assert.True(t, ma.HasMetric("profit"))
	assert.True(t, ma.HasMetric("margin"))
}

// ============================================================================
// CIL Tree Building and Storage Tests
// ============================================================================

// TestMetricAnalyzer_CILTreeStorage verifies that CIL trees can be stored and
// retrieved on the MetricAnalyzer.
func TestMetricAnalyzer_CILTreeStorage(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()

	cilNode := &cil.CILConnectorNode{ConnectorName: "SalesData"}
	ma.AddCILTree("SalesData", cilNode)

	require.True(t, ma.HasCILTree("SalesData"))
	assert.Equal(t, "SalesData", ma.GetCILTrees()["SalesData"].(*cil.CILConnectorNode).ConnectorName)
}

// TestMetricAnalyzer_CILTreeMultipleEntries verifies that multiple CIL tree
// entries can be stored.
func TestMetricAnalyzer_CILTreeMultipleEntries(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()

	ma.AddCILTree("SalesData", &cil.CILConnectorNode{ConnectorName: "SalesData"})
	ma.AddCILTree("Inventory", &cil.CILConnectorNode{ConnectorName: "Inventory"})

	assert.Len(t, ma.GetCILTrees(), 2)
	assert.True(t, ma.HasCILTree("SalesData"))
	assert.True(t, ma.HasCILTree("Inventory"))
}

// ============================================================================
// IO Spec Validation and Storage Tests
// ============================================================================

// TestMetricAnalyzer_ValidateAndStoreIOSpec_ValidOutputs verifies that an IO
// spec whose output metric names all exist in declared metrics is accepted and stored.
func TestMetricAnalyzer_ValidateAndStoreIOSpec_ValidOutputs(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("revenue_forecast", analysis.AnalysisMetricDefinition{Name: "revenue_forecast"})
	ma.AddMetric("profit_margin", analysis.AnalysisMetricDefinition{Name: "profit_margin"})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{MetricName: "revenue_forecast"},
			{MetricName: "profit_margin"},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.NoError(t, err, "Validation should succeed when all output metrics exist")
	stored := ma.GetIOSpec()
	require.NotNil(t, stored, "IO spec should be stored after successful validation")
	assert.Len(t, stored.Outputs, 2, "Stored IO spec should have 2 outputs")
	assert.Equal(t, "revenue_forecast", stored.Outputs[0].MetricName)
	assert.Equal(t, "profit_margin", stored.Outputs[1].MetricName)
}

// TestMetricAnalyzer_ValidateAndStoreIOSpec_UnknownMetric verifies that
// referencing an output metric that does not exist in declared metrics returns an error.
func TestMetricAnalyzer_ValidateAndStoreIOSpec_UnknownMetric(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("revenue_forecast", analysis.AnalysisMetricDefinition{Name: "revenue_forecast"})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{MetricName: "revenue_forecast"},
			{MetricName: "unknown_metric"},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.Error(t, err, "Validation should fail for unknown output metric")
	assert.Contains(t, err.Error(), "unknown metric", "Error should mention unknown metric")
	assert.Contains(t, err.Error(), "unknown_metric", "Error should contain the metric name")

	stored := ma.GetIOSpec()
	assert.Nil(t, stored, "IO spec should not be stored after validation failure")
}

// TestMetricAnalyzer_ValidateAndStoreIOSpec_EmptyOutputs verifies that an IO
// spec with an empty outputs list is accepted.
func TestMetricAnalyzer_ValidateAndStoreIOSpec_EmptyOutputs(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.NoError(t, err, "Validation should succeed for empty outputs list")
	stored := ma.GetIOSpec()
	require.NotNil(t, stored, "IO spec should be stored after successful validation")
	assert.Empty(t, stored.Outputs, "Stored IO spec should have empty outputs")
}

// TestMetricAnalyzer_ValidateAndStoreIOSpec_NilSpec verifies that a nil IO spec
// returns an error.
func TestMetricAnalyzer_ValidateAndStoreIOSpec_NilSpec(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()

	err := ma.ValidateAndStoreIOSpec(nil)

	require.Error(t, err, "Validation should fail for nil IO spec")
	assert.Contains(t, err.Error(), "nil", "Error should mention nil")

	stored := ma.GetIOSpec()
	assert.Nil(t, stored, "IO spec should remain nil after validation failure")
}

// TestMetricAnalyzer_GetIOSpec_NilInitially verifies that GetIOSpec returns nil
// when no IO spec has been validated.
func TestMetricAnalyzer_GetIOSpec_NilInitially(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()

	stored := ma.GetIOSpec()
	assert.Nil(t, stored, "GetIOSpec should return nil when no spec has been stored")
}

// ============================================================================
// Full IO Specification Validation Tests
// ============================================================================

// setupValidConnectorAndMeasures creates a ConnectorAnalyzer and measures map
// with a single connector "SalesData" having dimensions [Country, Product],
// grain [Day], and celltype Amount, plus a "Quarter" measure.
func setupValidConnectorAndMeasures(t *testing.T) (*analysis.ConnectorAnalyzer, map[string]analysis.AnalysisMeasureDefinition) {
	t.Helper()

	ca := analysis.NewConnectorAnalyzer()
	ca.AddConnector("SalesData", analysis.AnalysisConnectorDefinition{
		Name: "SalesData",
		Dimensions: []*analysis.AnalysisConceptDefinition{
			{ConceptName: "Country"},
			{ConceptName: "Product"},
		},
		CellType: &analysis.AnalysisMeasureDefinition{Name: "Amount"},
		Grain: []*analysis.AnalysisMeasureDefinition{
			{Name: "Day"},
		},
	})

	measures := map[string]analysis.AnalysisMeasureDefinition{
		"Amount":  {Name: "Amount", TypeName: "Float"},
		"Day":     {Name: "Day", TypeName: "Interval"},
		"Quarter": {Name: "Quarter", TypeName: "Interval"},
	}

	return ca, measures
}

// TestMetricAnalyzer_ValidateIOSpec_WithConnectorsDimensionsCellTypeGrain
// verifies that a valid IO spec with matching dimensions, celltype, and grain
// passes all validations.
func TestMetricAnalyzer_ValidateIOSpec_WithConnectorsDimensionsCellTypeGrain(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// Register a metric with a CIL tree that references the SalesData connector.
	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{
		Name: "revenue",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILBinaryOp{
				Op:    operator.Add,
				Left:  &cil.CILConnectorNode{ConnectorName: "SalesData"},
				Right: &cil.CILLiteral{Value: 1.0},
			},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Country", "Product"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
		Connectors: []cpio.ConnectorSpecification{
			{ConnectorName: "SalesData"},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.NoError(t, err, "Validation should succeed when all constraints match")
	stored := ma.GetIOSpec()
	require.NotNil(t, stored, "IO spec should be stored after successful validation")
	assert.Len(t, stored.Outputs, 1)
}

// TestMetricAnalyzer_ValidateIOSpec_DimensionMismatch verifies that an output
// dimension not declared in the connector's dimensions returns an error.
func TestMetricAnalyzer_ValidateIOSpec_DimensionMismatch(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{
		Name: "revenue",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILConnectorNode{ConnectorName: "SalesData"},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Country", "UnknownDimension"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.Error(t, err, "Validation should fail for unknown dimension")
	assert.Contains(t, err.Error(), "dimension")
	assert.Contains(t, err.Error(), "UnknownDimension")
	assert.Contains(t, err.Error(), "SalesData")
}

// TestMetricAnalyzer_ValidateIOSpec_CellTypeMismatch verifies that an output
// celltype not in the measures map returns an error.
func TestMetricAnalyzer_ValidateIOSpec_CellTypeMismatch(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{
		Name: "revenue",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILConnectorNode{ConnectorName: "SalesData"},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Country"},
				CellType:   "UnknownMeasure",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.Error(t, err, "Validation should fail for unknown celltype")
	assert.Contains(t, err.Error(), "celltype")
	assert.Contains(t, err.Error(), "UnknownMeasure")
}

// TestMetricAnalyzer_ValidateIOSpec_GrainMismatch verifies that an output
// grain not in the connector's grain returns an error.
func TestMetricAnalyzer_ValidateIOSpec_GrainMismatch(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{
		Name: "revenue",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILConnectorNode{ConnectorName: "SalesData"},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Country"},
				CellType:   "Amount",
				Grain:      []string{"Month"}, // Connector has grain Day, not Month
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.Error(t, err, "Validation should fail for unknown grain")
	assert.Contains(t, err.Error(), "grain")
	assert.Contains(t, err.Error(), "Month")
	assert.Contains(t, err.Error(), "SalesData")
}

// TestMetricAnalyzer_ValidateIOSpec_InputConnectorNotFound verifies that
// an input referencing an unknown connector returns an error.
func TestMetricAnalyzer_ValidateIOSpec_InputConnectorNotFound(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{
		Name: "revenue",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILConnectorNode{ConnectorName: "SalesData"},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Country"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
		Connectors: []cpio.ConnectorSpecification{
			{ConnectorName: "UnknownConnector"},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.Error(t, err, "Validation should fail for unknown input connector")
	assert.Contains(t, err.Error(), "unknown connector")
	assert.Contains(t, err.Error(), "UnknownConnector")
}

// TestMetricAnalyzer_ValidateIOSpec_InputMissingPath verifies that an input
// with driver parameters but no valid path returns an error.
func TestMetricAnalyzer_ValidateIOSpec_InputMissingPath(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{
		Name: "revenue",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILConnectorNode{ConnectorName: "SalesData"},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Country"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
		Connectors: []cpio.ConnectorSpecification{
			{
				ConnectorName:    "SalesData",
				DriverParameters: []map[string]interface{}{{ "host": "localhost" }},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.Error(t, err, "Validation should fail when input driver_parameters lack a path")
	assert.Contains(t, err.Error(), "path")
	assert.Contains(t, err.Error(), "SalesData")
}

// TestMetricAnalyzer_ValidateIOSpec_InputWithValidPath verifies that an input
// with a valid path in driver parameters passes validation.
func TestMetricAnalyzer_ValidateIOSpec_InputWithValidPath(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{
		Name: "revenue",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILConnectorNode{ConnectorName: "SalesData"},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Country"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
		Connectors: []cpio.ConnectorSpecification{
			{
				ConnectorName:    "SalesData",
				DriverParameters: []map[string]interface{}{{ "path": "/data/sales.parquet" }},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.NoError(t, err, "Validation should succeed when input driver_parameters contain a valid path")
}

// TestMetricAnalyzer_ValidateIOSpec_MetricWithNoConnectors verifies that an
// output metric with no connectors in its CIL tree skips connector-based
// validations gracefully and still accepts the IO spec.
func TestMetricAnalyzer_ValidateIOSpec_MetricWithNoConnectors(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// Metric with a CILRoot but NO CILConnectorNode in the tree.
	ma.AddMetric("computed_metric", analysis.AnalysisMetricDefinition{
		Name: "computed_metric",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILBinaryOp{
				Op:    operator.Add,
				Left:  &cil.CILLiteral{Value: 1.0},
				Right: &cil.CILLiteral{Value: 2.0},
			},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "computed_metric",
				Dimensions: []string{"Country"}, // No connector to validate against
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.NoError(t, err, "Validation should succeed when metric has no connectors")
	stored := ma.GetIOSpec()
	require.NotNil(t, stored)
}

// TestMetricAnalyzer_ValidateIOSpec_ConnectorNotFoundInTree verifies that
// when a CIL tree references a connector that is not in the connector analyzer,
// the connector-based validations are skipped gracefully.
func TestMetricAnalyzer_ValidateIOSpec_ConnectorNotFoundInTree(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// Metric references a connector not registered in the connector analyzer.
	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{
		Name: "revenue",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILConnectorNode{ConnectorName: "NonExistentConnector"},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Country"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	// Should not error on the connector lookup — skips gracefully.
	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.NoError(t, err, "Validation should skip when connector not found in analyzer")
	stored := ma.GetIOSpec()
	require.NotNil(t, stored)
}

// TestMetricAnalyzer_ValidateIOSpec_EmptyDimensionsAndGrain verifies that
// outputs with empty dimensions and grain arrays are accepted.
func TestMetricAnalyzer_ValidateIOSpec_EmptyDimensionsAndGrain(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{
		Name: "revenue",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILConnectorNode{ConnectorName: "SalesData"},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: nil,
				CellType:   "Amount",
				Grain:      nil,
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.NoError(t, err, "Validation should succeed with empty dimensions and grain")
}

// TestMetricAnalyzer_ValidateIOSpec_MultipleConnectorsInCILTree verifies that
// connector names from all branches of the CIL tree are collected and validated.
func TestMetricAnalyzer_ValidateIOSpec_MultipleConnectorsInCILTree(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca := analysis.NewConnectorAnalyzer()
	ca.AddConnector("SalesData", analysis.AnalysisConnectorDefinition{
		Name: "SalesData",
		Dimensions: []*analysis.AnalysisConceptDefinition{
			{ConceptName: "Country"},
		},
		Grain: []*analysis.AnalysisMeasureDefinition{
			{Name: "Day"},
		},
	})
	ca.AddConnector("Inventory", analysis.AnalysisConnectorDefinition{
		Name: "Inventory",
		Dimensions: []*analysis.AnalysisConceptDefinition{
			{ConceptName: "Product"},
		},
		Grain: []*analysis.AnalysisMeasureDefinition{
			{Name: "Day"},
		},
	})
	measures := map[string]analysis.AnalysisMeasureDefinition{
		"Amount": {Name: "Amount", TypeName: "Float"},
		"Day":    {Name: "Day", TypeName: "Interval"},
	}
	ma.SetValidationContext(ca, measures)

	// Metric with a CIL tree containing two different connectors.
	ma.AddMetric("combined_metric", analysis.AnalysisMetricDefinition{
		Name: "combined_metric",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILBinaryOp{
				Op:    operator.Add,
				Left:  &cil.CILConnectorNode{ConnectorName: "SalesData"},
				Right: &cil.CILConnectorNode{ConnectorName: "Inventory"},
			},
		},
	})

	// Dimensions valid for SalesData but NOT for Inventory.
	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "combined_metric",
				Dimensions: []string{"Country"}, // Valid for SalesData but not Inventory
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.Error(t, err, "Validation should fail when dimension is not in one of the connectors")
	assert.Contains(t, err.Error(), "Country")
	assert.Contains(t, err.Error(), "Inventory")
}

// TestMetricAnalyzer_ValidateIOSpec_InputWithEmptyConnectorName verifies that
// inputs with empty connector names and no driver parameters are skipped during validation.
func TestMetricAnalyzer_ValidateIOSpec_InputWithEmptyConnectorName(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{Name: "revenue"})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{MetricName: "revenue"},
		},
		Connectors: []cpio.ConnectorSpecification{
			{ConnectorName: ""}, // Empty connector name without driver params should be skipped
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.NoError(t, err, "Validation should succeed when input has empty connector name and no driver parameters")
}

// TestMetricAnalyzer_ValidateIOSpec_EmptyConnectorWithDriverParams verifies that
// an input with an empty connector name and driver parameters returns an error.
func TestMetricAnalyzer_ValidateIOSpec_EmptyConnectorWithDriverParams(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{Name: "revenue"})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{MetricName: "revenue"},
		},
		Connectors: []cpio.ConnectorSpecification{
			{
				ConnectorName:    "",
				DriverParameters: []map[string]interface{}{{"path": "/data/file.parquet"}},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.Error(t, err, "Validation should fail when input has driver parameters but no connector name")
	assert.Contains(t, err.Error(), "no connector name specified")
}

// TestMetricAnalyzer_ValidateIOSpec_DuplicateConnectorNamesInCILTree verifies
// that duplicate connector names in the CIL tree are de-duplicated.
func TestMetricAnalyzer_ValidateIOSpec_DuplicateConnectorNamesInCILTree(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// Metric whose CIL tree references the same connector twice.
	ma.AddMetric("revenue", analysis.AnalysisMetricDefinition{
		Name: "revenue",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILBinaryOp{
				Op:    operator.Add,
				Left:  &cil.CILConnectorNode{ConnectorName: "SalesData"},
				Right: &cil.CILConnectorNode{ConnectorName: "SalesData"},
			},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Country", "Product"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)

	require.NoError(t, err, "Validation should succeed with duplicate connector names (de-duplicated)")
}

// ============================================================================
// extractConnectorNamesFromCIL Tests
// ============================================================================

// TestExtractConnectorNamesFromCIL_NilNode verifies that a nil node returns nil.
func TestExtractConnectorNamesFromCIL_NilNode(t *testing.T) {
	t.Parallel()

	// We access the unexported function via a public API.
	// Build a CIL tree with a connector and verify it works through ValidateAndStoreIOSpec.
	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// Metric with NO CILRoot — no connectors to extract.
	ma.AddMetric("simple_metric", analysis.AnalysisMetricDefinition{
		Name:    "simple_metric",
		CILRoot: nil,
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "simple_metric",
				Dimensions: []string{"Country"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	// No error because no connectors to validate against.
	err := ma.ValidateAndStoreIOSpec(ioSpec)
	require.NoError(t, err, "Nil CILRoot should skip connector validations")
}

// TestExtractConnectorNamesFromCIL_DeepNesting verifies that connectors
// deep in the CIL tree are found.
func TestExtractConnectorNamesFromCIL_DeepNesting(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// Deeply nested CIL tree: Add(Add(Sales, Lit), Mul(Sales, Lit))
	ma.AddMetric("deep_metric", analysis.AnalysisMetricDefinition{
		Name: "deep_metric",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILBinaryOp{
				Op: operator.Add,
				Left: &cil.CILBinaryOp{
					Op:    operator.Add,
					Left:  &cil.CILConnectorNode{ConnectorName: "SalesData"},
					Right: &cil.CILLiteral{Value: 1.0},
				},
				Right: &cil.CILBinaryOp{
					Op:      operator.Multiply,
					Left:    &cil.CILConnectorNode{ConnectorName: "SalesData"},
					Right:   &cil.CILLiteral{Value: 2.0},
				},
			},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "deep_metric",
				Dimensions: []string{"Country", "Product"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)
	require.NoError(t, err, "Deeply nested connectors should be found")
}

// TestExtractConnectorNamesFromCIL_ConditionalTree verifies that connectors
// inside conditional branches are found.
func TestExtractConnectorNamesFromCIL_ConditionalTree(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// CILConditional tree: if SalesData > 0 then SalesData else 0
	ma.AddMetric("conditional_metric", analysis.AnalysisMetricDefinition{
		Name: "conditional_metric",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILConditional{
				Branches: []cil.CILConditionalBranch{
					{
						Condition: &cil.CILComparisonOp{
							Op:    operator.GreaterThan,
							Left:  &cil.CILConnectorNode{ConnectorName: "SalesData"},
							Right: &cil.CILLiteral{Value: 0},
						},
						Body: &cil.CILConnectorNode{ConnectorName: "SalesData"},
					},
				},
			},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "conditional_metric",
				Dimensions: []string{"Country", "Product"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)
	require.NoError(t, err, "Connectors in conditional branches should be found")
}

// TestExtractConnectorNamesFromCIL_UnaryOp verifies that connectors inside
// a unary negation expression are found.
func TestExtractConnectorNamesFromCIL_UnaryOp(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// CILUnaryOp (negation) wrapping a connector.
	ma.AddMetric("negated_metric", analysis.AnalysisMetricDefinition{
		Name: "negated_metric",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILUnaryOp{
				Op:      operator.Negate,
				Operand: &cil.CILConnectorNode{ConnectorName: "SalesData"},
			},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "negated_metric",
				Dimensions: []string{"Country", "Product"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)
	require.NoError(t, err, "Connectors in unary negation should be found")
}

// TestExtractConnectorNamesFromCIL_LogicalOp verifies that connectors inside
// logical AND/OR expressions are found.
func TestExtractConnectorNamesFromCIL_LogicalOp(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// CILLogicalOp (AND) containing connectors in both left and right.
	ma.AddMetric("logical_metric", analysis.AnalysisMetricDefinition{
		Name: "logical_metric",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILLogicalOp{
				Op: operator.And,
				Left: &cil.CILComparisonOp{
					Op:    operator.GreaterThan,
					Left:  &cil.CILConnectorNode{ConnectorName: "SalesData"},
					Right: &cil.CILLiteral{Value: 0},
				},
				Right: &cil.CILComparisonOp{
					Op:    operator.LessThan,
					Left:  &cil.CILConnectorNode{ConnectorName: "SalesData"},
					Right: &cil.CILLiteral{Value: 100},
				},
			},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "logical_metric",
				Dimensions: []string{"Country", "Product"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)
	require.NoError(t, err, "Connectors in logical AND expressions should be found")
}

// TestExtractConnectorNamesFromCIL_NotOp verifies that connectors inside
// NOT expressions are found.
func TestExtractConnectorNamesFromCIL_NotOp(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// CILNotOp wrapping a comparison that references a connector.
	ma.AddMetric("not_metric", analysis.AnalysisMetricDefinition{
		Name: "not_metric",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILNotOp{
				Operand: &cil.CILComparisonOp{
					Op:    operator.Equal,
					Left:  &cil.CILConnectorNode{ConnectorName: "SalesData"},
					Right: &cil.CILLiteral{Value: 0},
				},
			},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "not_metric",
				Dimensions: []string{"Country", "Product"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)
	require.NoError(t, err, "Connectors in NOT expressions should be found")
}

// TestExtractConnectorNamesFromCIL_WhenClause verifies that connectors inside
// when/default clauses are found.
func TestExtractConnectorNamesFromCIL_WhenClause(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// CILWhenClause with a when branch referencing SalesData.
	ma.AddMetric("when_metric", analysis.AnalysisMetricDefinition{
		Name: "when_metric",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILWhenClause{
				Clauses: []cil.CILWhenCase{
					{
						Condition: &cil.CILComparisonOp{
							Op:    operator.GreaterThan,
							Left:  &cil.CILConnectorNode{ConnectorName: "SalesData"},
							Right: &cil.CILLiteral{Value: 0},
						},
						Body: &cil.CILConnectorNode{ConnectorName: "SalesData"},
					},
				},
				Default: &cil.CILLiteral{Value: 0},
			},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "when_metric",
				Dimensions: []string{"Country", "Product"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)
	require.NoError(t, err, "Connectors in when/default clauses should be found")
}

// TestExtractConnectorNamesFromCIL_Power verifies that connectors inside
// a power expression are found.
func TestExtractConnectorNamesFromCIL_Power(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// CILPower wrapping a connector reference in the base.
	ma.AddMetric("power_metric", analysis.AnalysisMetricDefinition{
		Name: "power_metric",
		CILRoot: &cil.CILMetricNode{
			CILNode: &cil.CILPower{
				Base:     &cil.CILConnectorNode{ConnectorName: "SalesData"},
				Exponent: 2,
			},
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "power_metric",
				Dimensions: []string{"Country", "Product"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)
	require.NoError(t, err, "Connectors in power expressions should be found")
}

// TestExtractConnectorNamesFromCIL_NilCILNode verifies that a metric with
// CILRoot present but CILNode set to nil does not panic.
func TestExtractConnectorNamesFromCIL_NilCILNode(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ca, measures := setupValidConnectorAndMeasures(t)
	ma.SetValidationContext(ca, measures)

	// CILRoot is present, but the inner CILNode is nil.
	// extractConnectorNamesFromCIL should handle nil CILNode gracefully.
	ma.AddMetric("nil_node_metric", analysis.AnalysisMetricDefinition{
		Name: "nil_node_metric",
		CILRoot: &cil.CILMetricNode{
			CILNode: nil,
		},
	})

	ioSpec := &cpio.IOSpecification{
		Outputs: []cpio.OutputSpecification{
			{
				MetricName: "nil_node_metric",
				Dimensions: []string{"Country"},
				CellType:   "Amount",
				Grain:      []string{"Day"},
			},
		},
	}

	err := ma.ValidateAndStoreIOSpec(ioSpec)
	require.NoError(t, err, "Nil CILNode inside CILRoot should not cause a panic or error")
}

// TestMetricAnalyzer_MetricReferenceValidation verifies that metric references
// can be validated by looking up names in the metrics map.
func TestMetricAnalyzer_MetricReferenceValidation(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddMetric("existing_metric", analysis.AnalysisMetricDefinition{Name: "existing_metric"})

	// Existing metric should be found
	assert.True(t, ma.HasMetric("existing_metric"), "Existing metric should be found")

	// Non-existent metric should not be found
	assert.False(t, ma.HasMetric("missing_metric"), "Missing metric should not be found")
}

// TestMetricAnalyzer_CILTreeReferenceValidation verifies that CIL tree entries
// can be validated by looking up names in the cilTrees map.
func TestMetricAnalyzer_CILTreeReferenceValidation(t *testing.T) {
	t.Parallel()

	ma := analysis.NewMetricAnalyzer()
	ma.AddCILTree("connector_a", &cil.CILConnectorNode{ConnectorName: "connector_a"})

	// Existing CIL tree should be found
	assert.True(t, ma.HasCILTree("connector_a"), "Existing CIL tree should be found")

	// Non-existent CIL tree should not be found
	assert.False(t, ma.HasCILTree("missing_connector"), "Missing CIL tree should not be found")
}
