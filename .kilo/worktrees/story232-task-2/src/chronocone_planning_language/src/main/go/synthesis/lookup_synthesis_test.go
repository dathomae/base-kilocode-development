// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package synthesis

import (
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	cio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// TestGenerateLookupLoading_SingleLookup verifies that generateLookupLoading
// emits valid Go code for a single lookup with one dimension.
func TestGenerateLookupLoading_SingleLookup(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterLookup("product_categories", []string{"Product"}, "CategoryLabel")

	lookups := []cio.LookupSpecification{
		{
			LookupName:       "product_categories",
			DataFormat:       "parquet",
			Driver:           "parquet",
			DriverParameters: []map[string]interface{}{{"path": "/data/lookups/product_categories.parquet"}},
		},
	}

	var sb strings.Builder
	err := generateLookupLoading(transformer, lookups, &sb)
	if err != nil {
		t.Fatalf("generateLookupLoading() returned error: %v", err)
	}

	code := sb.String()

	// Verify function signature
	if !strings.Contains(code, "func loadLookupData(controller *calculation.ChronoconeComputationController, dataDir *string) error") {
		t.Error("Generated code missing expected function signature")
	}

	// Verify ReadParquetToTensor call with correct path (prepends dataDir)
	if !strings.Contains(code, `cplio.ReadParquetToTensor(filepath.Join(*dataDir, "/data/lookups/product_categories.parquet"`) {
		t.Errorf("Generated code missing expected ReadParquetToTensor call with correct path:\n%s", code)
	}

	// Verify dimension names
	if !strings.Contains(code, `[]string{"Product"}`) {
		t.Errorf("Generated code missing expected dimension slice:\n%s", code)
	}

	// Verify tensor construction
	if !strings.Contains(code, `calculation.NewChronoconeTensor(typedData, []string{"Product"}, shape)`) {
		t.Errorf("Generated code missing expected NewChronoconeTensor call:\n%s", code)
	}

	// Verify MetricComputation creation and registration
	if !strings.Contains(code, "calculation.NewMetricComputation(0)") {
		t.Errorf("Generated code missing expected NewMetricComputation call:\n%s", code)
	}
	if !strings.Contains(code, `controller.RegisterMetricComputation("product_categories", mc)`) {
		t.Errorf("Generated code missing expected RegisterMetricComputation call:\n%s", code)
	}

	// Verify SetTensor call
	if !strings.Contains(code, "mc.SetTensor(tensor)") {
		t.Errorf("Generated code missing expected mc.SetTensor call:\n%s", code)
	}
}

// TestGenerateLookupLoading_MultipleDimensions verifies that generateLookupLoading
// emits correct Go code for a lookup with multiple dimension concepts.
func TestGenerateLookupLoading_MultipleDimensions(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterLookup("country_product_index", []string{"Country", "Product"}, "IndexValue")

	lookups := []cio.LookupSpecification{
		{
			LookupName:       "country_product_index",
			DataFormat:       "parquet",
			Driver:           "parquet",
			DriverParameters: []map[string]interface{}{{"path": "/data/lookups/country_product.parquet"}},
		},
	}

	var sb strings.Builder
	err := generateLookupLoading(transformer, lookups, &sb)
	if err != nil {
		t.Fatalf("generateLookupLoading() returned error: %v", err)
	}

	code := sb.String()

	// Verify multiple dimensions in ReadParquetToTensor call
	if !strings.Contains(code, `[]string{"Country", "Product"}`) {
		t.Errorf("Generated code missing expected multi-dimension slice:\n%s", code)
	}

	// Verify tensor construction with multiple dimensions
	if !strings.Contains(code, `calculation.NewChronoconeTensor(typedData, []string{"Country", "Product"}, shape)`) {
		t.Errorf("Generated code missing expected NewChronoconeTensor with multiple dims:\n%s", code)
	}

	// Verify registration with correct lookup name
	if !strings.Contains(code, `controller.RegisterMetricComputation("country_product_index", mc)`) {
		t.Errorf("Generated code missing expected RegisterMetricComputation call:\n%s", code)
	}

	// Verify error message includes lookup name
	if !strings.Contains(code, `failed to load lookup 'country_product_index'`) {
		t.Errorf("Generated code missing expected error message for lookup name:\n%s", code)
	}
}

// TestGenerateLookupLoading_EmptyLookups verifies that when no lookups are
// declared in the registry, generateLookupLoading emits a no-op function
// that returns nil.
func TestGenerateLookupLoading_EmptyLookups(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	// No lookups registered

	lookups := []cio.LookupSpecification{}

	var sb strings.Builder
	err := generateLookupLoading(transformer, lookups, &sb)
	if err != nil {
		t.Fatalf("generateLookupLoading() returned error: %v", err)
	}

	code := sb.String()

	// Verify function signature exists
	if !strings.Contains(code, "func loadLookupData(controller *calculation.ChronoconeComputationController, dataDir *string) error") {
		t.Errorf("Generated code missing expected function signature:\n%s", code)
	}

	// Verify it's a no-op with return nil
	if !strings.Contains(code, "\treturn nil\n") {
		t.Errorf("Generated code missing expected 'return nil' for empty lookups:\n%s", code)
	}

	// Verify no lookup-specific code is emitted
	if strings.Contains(code, "ReadParquetToTensor") {
		t.Error("Generated code should not contain ReadParquetToTensor for empty lookups")
	}
	if strings.Contains(code, "NewChronoconeTensor") {
		t.Error("Generated code should not contain NewChronoconeTensor for empty lookups")
	}
	if strings.Contains(code, "RegisterMetricComputation") {
		t.Error("Generated code should not contain RegisterMetricComputation for empty lookups")
	}
}

// TestGenerateLookupLoading_MissingIOSpecEntry verifies that an error is
// returned when a lookup metadata entry has no matching IO spec entry.
func TestGenerateLookupLoading_MissingIOSpecEntry(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterLookup("unmapped_lookup", []string{"Dim1"}, "Value")

	lookups := []cio.LookupSpecification{}

	var sb strings.Builder
	err := generateLookupLoading(transformer, lookups, &sb)
	if err == nil {
		t.Fatal("generateLookupLoading() expected error for missing IO spec entry, got nil")
	}
	if !strings.Contains(err.Error(), "unmapped_lookup") {
		t.Errorf("Error message should mention lookup name, got: %v", err)
	}
}

// TestGenerateLookupLoading_MultipleLookups verifies deterministic ordering
// when multiple lookups are registered.
func TestGenerateLookupLoading_MultipleLookups(t *testing.T) {
	transformer := cil.NewCILAbstractTransformer()
	transformer.RegisterLookup("region_mapping", []string{"Region"}, "RegionCode")
	transformer.RegisterLookup("product_categories", []string{"Product"}, "CategoryLabel")

	lookups := []cio.LookupSpecification{
		{
			LookupName:       "product_categories",
			DataFormat:       "parquet",
			Driver:           "parquet",
			DriverParameters: []map[string]interface{}{{"path": "/data/lookups/product_categories.parquet"}},
		},
		{
			LookupName:       "region_mapping",
			DataFormat:       "parquet",
			Driver:           "parquet",
			DriverParameters: []map[string]interface{}{{"path": "/data/lookups/regions.parquet"}},
		},
	}

	var sb strings.Builder
	err := generateLookupLoading(transformer, lookups, &sb)
	if err != nil {
		t.Fatalf("generateLookupLoading() returned error: %v", err)
	}

	code := sb.String()

	// Both lookups should be present
	if !strings.Contains(code, `"product_categories"`) {
		t.Error("Generated code missing product_categories lookup")
	}
	if !strings.Contains(code, `"region_mapping"`) {
		t.Error("Generated code missing region_mapping lookup")
	}

	// Verify deterministic ordering: product_categories should appear before region_mapping
	if strings.Index(code, `"product_categories"`) > strings.Index(code, `"region_mapping"`) {
		t.Error("Expected product_categories to appear before region_mapping in generated code")
	}

	// Both paths should be present (prepended with dataDir)
	if !strings.Contains(code, `cplio.ReadParquetToTensor(filepath.Join(*dataDir, "/data/lookups/product_categories.parquet"`) {
		t.Error("Generated code missing product_categories path")
	}
	if !strings.Contains(code, `cplio.ReadParquetToTensor(filepath.Join(*dataDir, "/data/lookups/regions.parquet"`) {
		t.Error("Generated code missing region_mapping path")
	}

	// Each should have its own dimension
	if !strings.Contains(code, `[]string{"Product"}`) {
		t.Error("Generated code missing Product dimension")
	}
	if !strings.Contains(code, `[]string{"Region"}`) {
		t.Error("Generated code missing Region dimension")
	}
}
