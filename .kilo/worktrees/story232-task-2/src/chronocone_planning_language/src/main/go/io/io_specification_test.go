// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package io

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestLoadIOSpecification_Success tests successful parsing of valid IO specification JSON.
func TestLoadIOSpecification_Success(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Define test cases for successful parsing
	testCases := []struct {
		name     string
		json     string
		expected IOSpecification
	}{
		{
			name: "complete specification with single connector and output",
			json: `{
				"connectors": [
					{
						"connector": "sales_data",
						"data_format": "parquet",
						"driver": "file",
						"driver_parameters": [{"path": "/data/sales.parquet"}]
					}
				],
				"outputs": [
					{
						"metric": "revenue",
						"dimensions": ["Product", "Region"],
						"celltype": "Currency",
						"grain": ["Month"]
					}
				],
				"architecture": {
					"cpu_cores": 8,
					"cache_size": 16777216,
					"main_memory_size": 34359738368
				}
			}`,
			expected: IOSpecification{
				Connectors: []ConnectorSpecification{
					{
						ConnectorName:    "sales_data",
						DataFormat:       "parquet",
						Driver:           "file",
						DriverParameters: []map[string]interface{}{{"path": "/data/sales.parquet"}},
					},
				},
				Outputs: []OutputSpecification{
					{
						MetricName: "revenue",
						Dimensions: []string{"Product", "Region"},
						CellType:   "Currency",
						Grain:      []string{"Month"},
					},
				},
				Architecture: ArchitectureSpecification{
					CPUCoreCount:    8,
					CacheSizeBytes:  16777216,
					MainMemoryBytes: 34359738368,
				},
			},
		},
		{
			name: "specification with multiple connectors and outputs",
			json: `{
				"connectors": [
					{
						"connector": "sales_data",
						"data_format": "parquet",
						"driver": "file",
						"driver_parameters": [{"path": "/data/sales.parquet"}]
					},
					{
						"connector": "inventory_data",
						"data_format": "csv",
						"driver": "s3",
						"driver_parameters": [{"bucket": "mybucket", "key": "inventory.csv"}]
					}
				],
				"outputs": [
					{
						"metric": "revenue",
						"dimensions": ["Product", "Region"],
						"celltype": "Currency",
						"grain": ["Month"]
					},
					{
						"metric": "units_sold",
						"dimensions": ["Product"],
						"celltype": "Integer",
						"grain": ["Day"]
					}
				],
				"architecture": {
					"cpu_cores": 16,
					"cache_size": 33554432,
					"main_memory_size": 68719476736
				}
			}`,
			expected: IOSpecification{
				Connectors: []ConnectorSpecification{
					{
						ConnectorName:    "sales_data",
						DataFormat:       "parquet",
						Driver:           "file",
						DriverParameters: []map[string]interface{}{{"path": "/data/sales.parquet"}},
					},
					{
						ConnectorName:    "inventory_data",
						DataFormat:       "csv",
						Driver:           "s3",
						DriverParameters: []map[string]interface{}{{"bucket": "mybucket", "key": "inventory.csv"}},
					},
				},
				Outputs: []OutputSpecification{
					{
						MetricName: "revenue",
						Dimensions: []string{"Product", "Region"},
						CellType:   "Currency",
						Grain:      []string{"Month"},
					},
					{
						MetricName: "units_sold",
						Dimensions: []string{"Product"},
						CellType:   "Integer",
						Grain:      []string{"Day"},
					},
				},
				Architecture: ArchitectureSpecification{
					CPUCoreCount:    16,
					CacheSizeBytes:  33554432,
					MainMemoryBytes: 68719476736,
				},
			},
		},
		{
			name: "specification with empty connectors and outputs arrays",
			json: `{
				"connectors": [],
				"outputs": [],
				"architecture": {
					"cpu_cores": 4,
					"cache_size": 8388608,
					"main_memory_size": 17179869184
				}
			}`,
			expected: IOSpecification{
				Connectors:  []ConnectorSpecification{},
				Outputs: []OutputSpecification{},
				Architecture: ArchitectureSpecification{
					CPUCoreCount:    4,
					CacheSizeBytes:  8388608,
					MainMemoryBytes: 17179869184,
				},
			},
		},
		{
			name: "specification with empty driver_parameters",
			json: `{
				"connectors": [
					{
						"connector": "sales_data",
						"data_format": "parquet",
						"driver": "file",
						"driver_parameters": []
					}
				],
				"outputs": [
					{
						"metric": "revenue",
						"dimensions": [],
						"celltype": "Currency",
						"grain": ["Month"]
					}
				],
				"architecture": {
					"cpu_cores": 8,
					"cache_size": 16777216,
					"main_memory_size": 34359738368
				}
			}`,
			expected: IOSpecification{
				Connectors: []ConnectorSpecification{
					{
						ConnectorName:    "sales_data",
						DataFormat:       "parquet",
						Driver:           "file",
						DriverParameters: []map[string]interface{}{},
					},
				},
				Outputs: []OutputSpecification{
					{
						MetricName: "revenue",
						Dimensions: []string{},
						CellType:   "Currency",
						Grain:      []string{"Month"},
					},
				},
				Architecture: ArchitectureSpecification{
					CPUCoreCount:    8,
					CacheSizeBytes:  16777216,
					MainMemoryBytes: 34359738368,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary file with the test JSON
			tempFile := filepath.Join(tempDir, tc.name+".json")
			err := os.WriteFile(tempFile, []byte(tc.json), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Call the function under test
			result, err := LoadIOSpecification(tempFile)

			// Verify no error occurred
			if err != nil {
				t.Errorf("LoadIOSpecification() returned unexpected error: %v", err)
				return
			}

			// Verify result is not nil
			if result == nil {
				t.Errorf("LoadIOSpecification() returned nil result")
				return
			}

			// Verify connectors length matches
			if len(result.Connectors) != len(tc.expected.Connectors) {
				t.Errorf("Connectors length mismatch: got %d, want %d", len(result.Connectors), len(tc.expected.Connectors))
			}

			// Verify each connector specification
			for i, connector := range result.Connectors {
				if i >= len(tc.expected.Connectors) {
					break
				}
				expectedConnector := tc.expected.Connectors[i]
				if connector.ConnectorName != expectedConnector.ConnectorName {
					t.Errorf("Connector[%d].ConnectorName mismatch: got %q, want %q", i, connector.ConnectorName, expectedConnector.ConnectorName)
				}
				if connector.DataFormat != expectedConnector.DataFormat {
					t.Errorf("Connector[%d].DataFormat mismatch: got %q, want %q", i, connector.DataFormat, expectedConnector.DataFormat)
				}
				if connector.Driver != expectedConnector.Driver {
					t.Errorf("Connector[%d].Driver mismatch: got %q, want %q", i, connector.Driver, expectedConnector.Driver)
				}
				if len(connector.DriverParameters) != len(expectedConnector.DriverParameters) {
					t.Errorf("Connector[%d].DriverParameters length mismatch: got %d, want %d", i, len(connector.DriverParameters), len(expectedConnector.DriverParameters))
				}
			}

			// Verify outputs length matches
			if len(result.Outputs) != len(tc.expected.Outputs) {
				t.Errorf("Outputs length mismatch: got %d, want %d", len(result.Outputs), len(tc.expected.Outputs))
			}

			// Verify each output specification
			for i, output := range result.Outputs {
				if i >= len(tc.expected.Outputs) {
					break
				}
				expectedOutput := tc.expected.Outputs[i]
				if output.MetricName != expectedOutput.MetricName {
					t.Errorf("Output[%d].MetricName mismatch: got %q, want %q", i, output.MetricName, expectedOutput.MetricName)
				}
				if output.CellType != expectedOutput.CellType {
					t.Errorf("Output[%d].CellType mismatch: got %q, want %q", i, output.CellType, expectedOutput.CellType)
				}
				if !slices.Equal(output.Grain, expectedOutput.Grain) {
					t.Errorf("Output[%d].Grain mismatch: got %v, want %v", i, output.Grain, expectedOutput.Grain)
				}
				if len(output.Dimensions) != len(expectedOutput.Dimensions) {
					t.Errorf("Output[%d].Dimensions length mismatch: got %d, want %d", i, len(output.Dimensions), len(expectedOutput.Dimensions))
				}
			}

			// Verify architecture specification
			if result.Architecture.CPUCoreCount != tc.expected.Architecture.CPUCoreCount {
				t.Errorf("Architecture.CPUCoreCount mismatch: got %d, want %d", result.Architecture.CPUCoreCount, tc.expected.Architecture.CPUCoreCount)
			}
			if result.Architecture.CacheSizeBytes != tc.expected.Architecture.CacheSizeBytes {
				t.Errorf("Architecture.CacheSizeBytes mismatch: got %d, want %d", result.Architecture.CacheSizeBytes, tc.expected.Architecture.CacheSizeBytes)
			}
			if result.Architecture.MainMemoryBytes != tc.expected.Architecture.MainMemoryBytes {
				t.Errorf("Architecture.MainMemoryBytes mismatch: got %d, want %d", result.Architecture.MainMemoryBytes, tc.expected.Architecture.MainMemoryBytes)
			}
		})
	}
}

// TestLoadIOSpecification_FileNotFound tests error handling for missing file.
func TestLoadIOSpecification_FileNotFound(t *testing.T) {
	// Use a non-existent file path
	nonExistentPath := filepath.Join(t.TempDir(), "non_existent_file.json")

	result, err := LoadIOSpecification(nonExistentPath)

	// Verify that an error is returned
	if err == nil {
		t.Errorf("LoadIOSpecification() with non-existent file should return an error, got nil")
	}

	// Verify that result is nil when error occurs
	if result != nil {
		t.Errorf("LoadIOSpecification() with non-existent file should return nil result, got %v", result)
	}
}

// TestLoadIOSpecification_InvalidJSON tests error handling for invalid JSON.
func TestLoadIOSpecification_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name        string
		json        string
		description string
	}{
		{
			name:        "malformed json",
			json:        `{"connectors": [], "outputs": [], "architecture": {`,
			description: "unclosed JSON object",
		},
		{
			name:        "invalid json syntax",
			json:        `this is not json at all`,
			description: "completely invalid JSON",
		},
		{
			name:        "wrong type for cpu_cores",
			json:        `{"connectors": [], "outputs": [], "architecture": {"cpu_cores": "eight", "cache_size": 16777216, "main_memory_size": 34359738368}}`,
			description: "string instead of int for cpu_cores",
		},
		{
			name:        "wrong type for connectors",
			json:        `{"connectors": "not an array", "outputs": [], "architecture": {"cpu_cores": 8, "cache_size": 16777216, "main_memory_size": 34359738368}}`,
			description: "string instead of array for connectors",
		},
		{
			name:        "empty string",
			json:        ``,
			description: "empty file content",
		},
		{
			name:        "just whitespace",
			json:        `   \n\t  `,
			description: "whitespace only content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary file with invalid JSON
			tempFile := filepath.Join(tempDir, tc.name+".json")
			err := os.WriteFile(tempFile, []byte(tc.json), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Call the function under test
			result, err := LoadIOSpecification(tempFile)

			// Verify that an error is returned
			if err == nil {
				t.Errorf("LoadIOSpecification() with %s should return an error, got nil", tc.description)
			}

			// Verify that result is nil when error occurs
			if result != nil {
				t.Errorf("LoadIOSpecification() with %s should return nil result, got %v", tc.description, result)
			}
		})
	}
}

// TestLoadIOSpecification_FieldPopulation tests that all fields are correctly populated from JSON.
func TestLoadIOSpecification_FieldPopulation(t *testing.T) {
	tempDir := t.TempDir()

	// Create a comprehensive JSON with all fields populated
	json := `{
		"connectors": [
			{
				"connector": "test_connector",
				"data_format": "test_format",
				"driver": "test_driver",
				"driver_parameters": [
					{"key1": "value1", "key2": 123},
					{"key3": true, "key4": null}
				]
			}
		],
		"outputs": [
			{
				"metric": "test_metric",
				"dimensions": ["dim1", "dim2", "dim3"],
				"celltype": "test_celltype",
				"grain": ["test_grain"]
			}
		],
		"architecture": {
			"cpu_cores": 32,
			"cache_size": 67108864,
			"main_memory_size": 137438953472
		}
	}`

	// Create temporary file
	tempFile := filepath.Join(tempDir, "field_population.json")
	err := os.WriteFile(tempFile, []byte(json), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the specification
	result, err := LoadIOSpecification(tempFile)
	if err != nil {
		t.Errorf("LoadIOSpecification() returned unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Errorf("LoadIOSpecification() returned nil result")
		return
	}

	// Verify Connector fields
	if len(result.Connectors) != 1 {
		t.Errorf("Expected 1 connector, got %d", len(result.Connectors))
		return
	}

	connector := result.Connectors[0]
	if connector.ConnectorName != "test_connector" {
		t.Errorf("ConnectorName mismatch: got %q, want %q", connector.ConnectorName, "test_connector")
	}
	if connector.DataFormat != "test_format" {
		t.Errorf("DataFormat mismatch: got %q, want %q", connector.DataFormat, "test_format")
	}
	if connector.Driver != "test_driver" {
		t.Errorf("Driver mismatch: got %q, want %q", connector.Driver, "test_driver")
	}
	if len(connector.DriverParameters) != 2 {
		t.Errorf("DriverParameters length mismatch: got %d, want %d", len(connector.DriverParameters), 2)
	}

	// Verify first driver parameter map
	if len(connector.DriverParameters) > 0 {
		param1 := connector.DriverParameters[0]
		if param1["key1"] != "value1" {
			t.Errorf("DriverParameters[0][\"key1\"] mismatch: got %v, want %v", param1["key1"], "value1")
		}
		if param1["key2"] != float64(123) {
			t.Errorf("DriverParameters[0][\"key2\"] mismatch: got %v, want %v", param1["key2"], float64(123))
		}
	}

	// Verify second driver parameter map
	if len(connector.DriverParameters) > 1 {
		param2 := connector.DriverParameters[1]
		if param2["key3"] != true {
			t.Errorf("DriverParameters[1][\"key3\"] mismatch: got %v, want %v", param2["key3"], true)
		}
		if param2["key4"] != nil {
			t.Errorf("DriverParameters[1][\"key4\"] mismatch: got %v, want %v", param2["key4"], nil)
		}
	}

	// Verify Output fields
	if len(result.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(result.Outputs))
		return
	}

	output := result.Outputs[0]
	if output.MetricName != "test_metric" {
		t.Errorf("MetricName mismatch: got %q, want %q", output.MetricName, "test_metric")
	}
	if output.CellType != "test_celltype" {
		t.Errorf("CellType mismatch: got %q, want %q", output.CellType, "test_celltype")
	}
	if !slices.Equal(output.Grain, []string{"test_grain"}) {
		t.Errorf("Grain mismatch: got %v, want %v", output.Grain, []string{"test_grain"})
	}
	if len(output.Dimensions) != 3 {
		t.Errorf("Dimensions length mismatch: got %d, want %d", len(output.Dimensions), 3)
	}
	expectedDims := []string{"dim1", "dim2", "dim3"}
	for i, dim := range output.Dimensions {
		if dim != expectedDims[i] {
			t.Errorf("Dimensions[%d] mismatch: got %q, want %q", i, dim, expectedDims[i])
		}
	}

	// Verify Architecture fields
	if result.Architecture.CPUCoreCount != 32 {
		t.Errorf("CPUCoreCount mismatch: got %d, want %d", result.Architecture.CPUCoreCount, 32)
	}
	if result.Architecture.CacheSizeBytes != 67108864 {
		t.Errorf("CacheSizeBytes mismatch: got %d, want %d", result.Architecture.CacheSizeBytes, 67108864)
	}
	if result.Architecture.MainMemoryBytes != 137438953472 {
		t.Errorf("MainMemoryBytes mismatch: got %d, want %d", result.Architecture.MainMemoryBytes, 137438953472)
	}
}

// TestOutputSpecificationGrain_JSON_SingleGrain verifies JSON marshal/unmarshal for a single grain element.
func TestOutputSpecificationGrain_JSON_SingleGrain(t *testing.T) {
	spec := OutputSpecification{
		MetricName: "revenue",
		Dimensions: []string{"Product"},
		CellType:   "Currency",
		Grain:      []string{"Month"},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"grain":["Month"]`) {
		t.Errorf("Marshaled JSON missing expected grain array: %s", jsonStr)
	}

	var result OutputSpecification
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if !slices.Equal(result.Grain, []string{"Month"}) {
		t.Errorf("Grain mismatch after unmarshal: got %v, want %v", result.Grain, []string{"Month"})
	}
}

// TestOutputSpecificationGrain_JSON_MultipleGrains verifies JSON marshal/unmarshal for multiple grain elements.
func TestOutputSpecificationGrain_JSON_MultipleGrains(t *testing.T) {
	spec := OutputSpecification{
		MetricName: "sales",
		Dimensions: []string{"Product", "Region"},
		CellType:   "Integer",
		Grain:      []string{"YearGrain", "WeekGrain", "DayGrain"},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"grain":["YearGrain","WeekGrain","DayGrain"]`) {
		t.Errorf("Marshaled JSON missing expected grain array: %s", jsonStr)
	}

	var result OutputSpecification
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expected := []string{"YearGrain", "WeekGrain", "DayGrain"}
	if !slices.Equal(result.Grain, expected) {
		t.Errorf("Grain mismatch after unmarshal: got %v, want %v", result.Grain, expected)
	}
}

// TestOutputSpecificationGrain_JSON_EmptyArray verifies that an empty grain slice round-trips as [] not null.
func TestOutputSpecificationGrain_JSON_EmptyArray(t *testing.T) {
	spec := OutputSpecification{
		MetricName: "returns",
		Dimensions: []string{},
		CellType:   "Integer",
		Grain:      []string{},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"grain":[]`) {
		t.Errorf("Marshaled JSON missing expected empty grain array: %s", jsonStr)
	}

	var result OutputSpecification
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if result.Grain == nil {
		t.Errorf("Grain is nil after unmarshal, expected non-nil empty slice")
	}
	if len(result.Grain) != 0 {
		t.Errorf("Grain length mismatch: got %d, want 0", len(result.Grain))
	}
}

// TestIOSpecification_JSON_FullRoundTrip_WithMultipleGrainOutputs verifies a full IO spec with mixed grain arrays.
func TestIOSpecification_JSON_FullRoundTrip_WithMultipleGrainOutputs(t *testing.T) {
	spec := IOSpecification{
		Connectors: []ConnectorSpecification{
			{
				ConnectorName:    "sales_data",
				DataFormat:       "parquet",
				Driver:           "file",
				DriverParameters: []map[string]interface{}{{"path": "/data/sales.parquet"}},
			},
		},
		Outputs: []OutputSpecification{
			{
				MetricName: "revenue",
				Dimensions: []string{"Product", "Region"},
				CellType:   "Currency",
				Grain:      []string{"Month"},
			},
			{
				MetricName: "units_sold",
				Dimensions: []string{"Product"},
				CellType:   "Integer",
				Grain:      []string{"YearGrain", "WeekGrain", "DayGrain"},
			},
			{
				MetricName: "returns",
				Dimensions: []string{},
				CellType:   "Integer",
				Grain:      []string{},
			},
		},
		Architecture: ArchitectureSpecification{
			CPUCoreCount:    8,
			CacheSizeBytes:  16777216,
			MainMemoryBytes: 34359738368,
		},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var result IOSpecification
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(result.Outputs) != 3 {
		t.Fatalf("Expected 3 outputs, got %d", len(result.Outputs))
	}

	if !slices.Equal(result.Outputs[0].Grain, []string{"Month"}) {
		t.Errorf("Outputs[0].Grain mismatch: got %v, want %v", result.Outputs[0].Grain, []string{"Month"})
	}

	if !slices.Equal(result.Outputs[1].Grain, []string{"YearGrain", "WeekGrain", "DayGrain"}) {
		t.Errorf("Outputs[1].Grain mismatch: got %v, want %v", result.Outputs[1].Grain, []string{"YearGrain", "WeekGrain", "DayGrain"})
	}

	if result.Outputs[2].Grain == nil {
		t.Errorf("Outputs[2].Grain is nil, expected non-nil empty slice")
	}
	if len(result.Outputs[2].Grain) != 0 {
		t.Errorf("Outputs[2].Grain length mismatch: got %d, want 0", len(result.Outputs[2].Grain))
	}
}

// TestOutputSpecificationGrain_JSON_OrderPreserved verifies that grain slice order is preserved through marshal/unmarshal.
func TestOutputSpecificationGrain_JSON_OrderPreserved(t *testing.T) {
	original := []string{"A", "B", "C", "D", "E"}
	spec := OutputSpecification{
		MetricName: "order_test",
		Dimensions: []string{},
		CellType:   "Integer",
		Grain:      original,
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var result OutputSpecification
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if !slices.Equal(result.Grain, original) {
		t.Errorf("Grain order not preserved: got %v, want %v", result.Grain, original)
	}
}

// TestLoadIOSpecification_LookupsPopulated tests parsing of IO specification with lookups populated.
func TestLoadIOSpecification_LookupsPopulated(t *testing.T) {
	tempDir := t.TempDir()

	json := `{
		"connectors": [],
		"outputs": [],
		"lookups": [
			{
				"lookup": "country_codes",
				"data_format": "csv",
				"driver": "file",
				"driver_parameters": [{"path": "/data/countries.csv"}]
			},
			{
				"lookup": "product_categories",
				"data_format": "json",
				"driver": "http",
				"driver_parameters": [{"url": "https://api.example.com/categories"}]
			}
		],
		"architecture": {
			"cpu_cores": 4,
			"cache_size": 8388608,
			"main_memory_size": 17179869184
		}
	}`

	tempFile := filepath.Join(tempDir, "lookups_populated.json")
	err := os.WriteFile(tempFile, []byte(json), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := LoadIOSpecification(tempFile)
	if err != nil {
		t.Fatalf("LoadIOSpecification() returned unexpected error: %v", err)
	}

	if len(result.Lookups) != 2 {
		t.Fatalf("Expected 2 lookups, got %d", len(result.Lookups))
	}

	lookup1 := result.Lookups[0]
	if lookup1.LookupName != "country_codes" {
		t.Errorf("Lookups[0].LookupName mismatch: got %q, want %q", lookup1.LookupName, "country_codes")
	}
	if lookup1.DataFormat != "csv" {
		t.Errorf("Lookups[0].DataFormat mismatch: got %q, want %q", lookup1.DataFormat, "csv")
	}
	if lookup1.Driver != "file" {
		t.Errorf("Lookups[0].Driver mismatch: got %q, want %q", lookup1.Driver, "file")
	}
	if len(lookup1.DriverParameters) != 1 {
		t.Errorf("Lookups[0].DriverParameters length mismatch: got %d, want %d", len(lookup1.DriverParameters), 1)
	}

	lookup2 := result.Lookups[1]
	if lookup2.LookupName != "product_categories" {
		t.Errorf("Lookups[1].LookupName mismatch: got %q, want %q", lookup2.LookupName, "product_categories")
	}
	if lookup2.DataFormat != "json" {
		t.Errorf("Lookups[1].DataFormat mismatch: got %q, want %q", lookup2.DataFormat, "json")
	}
	if lookup2.Driver != "http" {
		t.Errorf("Lookups[1].Driver mismatch: got %q, want %q", lookup2.Driver, "http")
	}
}

// TestLoadIOSpecification_LookupsEmpty tests parsing of IO specification with empty lookups array.
func TestLoadIOSpecification_LookupsEmpty(t *testing.T) {
	tempDir := t.TempDir()

	json := `{
		"connectors": [],
		"outputs": [],
		"lookups": [],
		"architecture": {
			"cpu_cores": 4,
			"cache_size": 8388608,
			"main_memory_size": 17179869184
		}
	}`

	tempFile := filepath.Join(tempDir, "lookups_empty.json")
	err := os.WriteFile(tempFile, []byte(json), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := LoadIOSpecification(tempFile)
	if err != nil {
		t.Fatalf("LoadIOSpecification() returned unexpected error: %v", err)
	}

	if len(result.Lookups) != 0 {
		t.Errorf("Expected 0 lookups, got %d", len(result.Lookups))
	}
}
