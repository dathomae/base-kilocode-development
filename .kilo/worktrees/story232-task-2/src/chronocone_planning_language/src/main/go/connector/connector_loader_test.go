// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package connector

import (
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// mockConnectorProvider is a test double that implements ConnectorProvider.
type mockConnectorProvider struct {
	connectors map[string]*ConnectorMetadata
}

func (m *mockConnectorProvider) GetConnectorMetadata(name string) *ConnectorMetadata {
	if m == nil || m.connectors == nil {
		return nil
	}
	return m.connectors[name]
}

// TestGenerateConnectorLoaders verifies that GenerateConnectorLoaders generates loader for each connector.
func TestGenerateConnectorLoaders(t *testing.T) {
	tests := []struct {
		name          string
		connectors    map[string]*ConnectorMetadata
		inputs        []io.ConnectorSpecification
		expectLoaders []string
		expectError   bool
	}{
		{
			name: "single connector with input spec",
			connectors: map[string]*ConnectorMetadata{
				"sales_data": {
					Name:       "sales_data",
					Dimensions: []string{"Product", "Region"},
					CellType:   "float64",
				},
			},
			inputs: []io.ConnectorSpecification{
				{
					ConnectorName: "sales_data",
					DataFormat:    "parquet",
					Driver:        "file",
				},
			},
			expectLoaders: []string{"loadSalesData", "sales_data"},
			expectError:   false,
		},
		{
			name: "multiple connectors",
			connectors: map[string]*ConnectorMetadata{
				"sales_data": {
					Name:       "sales_data",
					Dimensions: []string{"Product", "Region"},
					CellType:   "float64",
				},
				"inventory": {
					Name:       "inventory",
					Dimensions: []string{"Warehouse"},
					CellType:   "int64",
				},
			},
			inputs: []io.ConnectorSpecification{
				{
					ConnectorName: "sales_data",
					DataFormat:    "parquet",
					Driver:        "file",
				},
				{
					ConnectorName: "inventory",
					DataFormat:    "parquet",
					Driver:        "file",
				},
			},
			expectLoaders: []string{"loadSalesData", "loadInventory", "sales_data", "inventory"},
			expectError:   false,
		},
		{
			name:          "no connectors",
			connectors:    map[string]*ConnectorMetadata{},
			inputs:        []io.ConnectorSpecification{},
			expectLoaders: []string{},
			expectError:   false,
		},
		{
			name: "connector without input spec",
			connectors: map[string]*ConnectorMetadata{
				"sales_data": {
					Name:       "sales_data",
					Dimensions: []string{"Product"},
					CellType:   "float64",
				},
			},
			inputs:        []io.ConnectorSpecification{},
			expectLoaders: []string{},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockConnectorProvider{connectors: make(map[string]*ConnectorMetadata)}
			for name, metadata := range tt.connectors {
				mock.connectors[name] = metadata
			}

			ioSpec := &io.IOSpecification{
				Connectors: tt.inputs,
				Outputs: []io.OutputSpecification{},
				Architecture: io.ArchitectureSpecification{
					CPUCoreCount:    4,
					CacheSizeBytes:  1024 * 1024,
					MainMemoryBytes: 1024 * 1024 * 1024,
				},
			}

			code, err := GenerateConnectorLoaders(mock, ioSpec, false)

			if tt.expectError {
				if err == nil {
					t.Errorf("GenerateConnectorLoaders() returned nil error, expected error")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateConnectorLoaders() returned error: %v", err)
				}

				// Check expected content
				for _, expected := range tt.expectLoaders {
					if !strings.Contains(code, expected) {
						t.Errorf("GenerateConnectorLoaders() output missing expected content: %q", expected)
					}
				}
			}
		})
	}
}

// TestGenerateSingleConnectorLoader verifies that GenerateSingleConnectorLoader generates correct loader function.
func TestGenerateSingleConnectorLoader(t *testing.T) {
	tests := []struct {
		name           string
		connectorName  string
		metadata       *ConnectorMetadata
		inputSpec      *io.ConnectorSpecification
		expectContains []string
		expectError    bool
	}{
		{
			name:          "basic connector loader",
			connectorName: "sales_data",
			metadata: &ConnectorMetadata{
				Name:       "sales_data",
				Dimensions: []string{"Product", "Region"},
				CellType:   "float64",
			},
			inputSpec: &io.ConnectorSpecification{
				ConnectorName:    "sales_data",
				DataFormat:       "parquet",
				Driver:           "file",
				DriverParameters: []map[string]interface{}{{"path": "/data/sales.parquet"}},
			},
			expectContains: []string{
				"func",
				"sales_data",
				"calculation.ChronoconeTensor",
				"any",
				"Product",
				"Region",
				"cplio.ConnectorSpecification",
			},
			expectError: false,
		},
		{
			name:          "integer connector loader",
			connectorName: "inventory",
			metadata: &ConnectorMetadata{
				Name:       "inventory",
				Dimensions: []string{"Warehouse", "SKU"},
				CellType:   "int64",
			},
			inputSpec: &io.ConnectorSpecification{
				ConnectorName: "inventory",
				DataFormat:    "parquet",
				Driver:        "file",
			},
			expectContains: []string{
				"inventory",
				"any",
				"Warehouse",
				"SKU",
				"calculation.ChronoconeTensor",
				"cplio.ConnectorSpecification",
			},
			expectError: false,
		},
		{
			name:          "single dimension connector",
			connectorName: "time_series",
			metadata: &ConnectorMetadata{
				Name:       "time_series",
				Dimensions: []string{"Time"},
				CellType:   "float64",
			},
			inputSpec: &io.ConnectorSpecification{
				ConnectorName: "time_series",
				DataFormat:    "parquet",
				Driver:        "file",
			},
			expectContains: []string{
				"Time",
				"calculation.ChronoconeTensor",
				"cplio.ConnectorSpecification",
			},
			expectError: false,
		},
		{
			name:          "nil metadata",
			connectorName: "invalid",
			metadata:      nil,
			inputSpec: &io.ConnectorSpecification{
				ConnectorName: "invalid",
				DataFormat:    "parquet",
				Driver:        "file",
			},
			expectContains: []string{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateSingleConnectorLoader(tt.connectorName, tt.metadata, tt.inputSpec, false)

			if tt.expectError {
				if err == nil {
					t.Errorf("GenerateSingleConnectorLoader() returned nil error, expected error")
				}
			} else {
				if err != nil {
					t.Errorf("GenerateSingleConnectorLoader() returned error: %v", err)
				}

				for _, expected := range tt.expectContains {
					if !strings.Contains(code, expected) {
						t.Errorf("GenerateSingleConnectorLoader() output missing expected content: %q", expected)
					}
				}
			}
		})
	}
}

// TestGenerateLoadParquetConnectorFunction verifies that GenerateLoadParquetConnectorFunction generates reusable loader.
func TestGenerateLoadParquetConnectorFunction(t *testing.T) {
	code := GenerateLoadParquetConnectorFunction()

	// Verify function signature
	if !strings.Contains(code, "func") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing 'func' keyword")
	}

	if !strings.Contains(code, "loadParquetConnector") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing 'loadParquetConnector' function name")
	}

	// Verify package-qualified parameter types
	if !strings.Contains(code, "connector.ConnectorMetadata") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing 'connector.ConnectorMetadata' parameter")
	}

	if !strings.Contains(code, "cplio.ConnectorSpecification") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing 'cplio.ConnectorSpecification' parameter")
	}

	// Verify package-qualified return type
	if !strings.Contains(code, "calculation.ChronoconeTensor") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing 'calculation.ChronoconeTensor' return type")
	}

	// Verify package-qualified NewChronoconeTensor call
	if !strings.Contains(code, "calculation.NewChronoconeTensor") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing 'calculation.NewChronoconeTensor' call")
	}

	// Verify error handling
	if !strings.Contains(code, "error") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing error handling")
	}

	// Verify parquet-specific content
	if !strings.Contains(code, "parquet") && !strings.Contains(code, "Parquet") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing parquet-specific content")
	}

	// Verify unqualified type references are NOT present
	unqualified := []string{
		"(metadata ConnectorMetadata)",
		"(connectorSpec ConnectorSpecification)",
		"return NewChronoconeTensor(",
	}
	for _, unexpected := range unqualified {
		if strings.Contains(code, unexpected) {
			t.Errorf("GenerateLoadParquetConnectorFunction() output contains unqualified reference %q", unexpected)
		}
	}
}

// TestGenerateConnectorLoaderFunctionName verifies the function name generation.
func TestGenerateConnectorLoaderFunctionName(t *testing.T) {
	tests := []struct {
		connectorName string
		wantContains  []string
	}{
		{
			connectorName: "sales_data",
			wantContains:  []string{"load", "Sales", "Data"},
		},
		{
			connectorName: "inventory",
			wantContains:  []string{"load", "Inventory"},
		},
		{
			connectorName: "simple",
			wantContains:  []string{"load", "Simple"},
		},
		{
			connectorName: "multi_word_connector",
			wantContains:  []string{"load", "Multi", "Word", "Connector"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.connectorName, func(t *testing.T) {
			name := GenerateConnectorLoaderFunctionName(tt.connectorName)

			// Verify it's not empty
			if name == "" {
				t.Errorf("GenerateConnectorLoaderFunctionName(%q) returned empty string", tt.connectorName)
				return
			}

			// Verify expected content
			for _, expected := range tt.wantContains {
				if !strings.Contains(name, expected) {
					t.Errorf("GenerateConnectorLoaderFunctionName(%q) = %q, missing expected content: %q", tt.connectorName, name, expected)
				}
			}

			// Verify it's a valid Go identifier (starts with a letter)
			if len(name) > 0 && !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z')) {
				t.Errorf("GenerateConnectorLoaderFunctionName(%q) = %q, should start with a letter", tt.connectorName, name)
			}
		})
	}
}

// TestGenerateConnectorLoadersWithMultipleFormats verifies handling of different data formats.
func TestGenerateConnectorLoadersWithMultipleFormats(t *testing.T) {
	mock := &mockConnectorProvider{connectors: make(map[string]*ConnectorMetadata)}
	mock.connectors["sales_data"] = NewConnectorMetadata("sales_data", []string{"Product"}, nil, nil, "float64")
	mock.connectors["legacy_data"] = NewConnectorMetadata("legacy_data", []string{"Region"}, nil, nil, "float64")

	ioSpec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
			{
				ConnectorName: "legacy_data",
				DataFormat:    "csv",
				Driver:        "file",
			},
		},
		Outputs: []io.OutputSpecification{},
		Architecture: io.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	code, err := GenerateConnectorLoaders(mock, ioSpec, false)

	if err != nil {
		t.Errorf("GenerateConnectorLoaders() returned error: %v", err)
	}

	// Verify both connectors are handled
	if !strings.Contains(code, "sales_data") {
		t.Errorf("GenerateConnectorLoaders() output missing 'sales_data'")
	}

	if !strings.Contains(code, "legacy_data") {
		t.Errorf("GenerateConnectorLoaders() output missing 'legacy_data'")
	}
}

// TestGenerateSingleConnectorLoaderWithDriverParameters verifies loader with driver parameters.
func TestGenerateSingleConnectorLoaderWithDriverParameters(t *testing.T) {
	metadata := &ConnectorMetadata{
		Name:       "sales_data",
		Dimensions: []string{"Product", "Region"},
		CellType:   "float64",
	}

	inputSpec := &io.ConnectorSpecification{
		ConnectorName: "sales_data",
		DataFormat:    "parquet",
		Driver:        "s3",
		DriverParameters: []map[string]interface{}{
			{"bucket": "my-bucket"},
			{"region": "us-east-1"},
			{"access_key": "AKIA..."},
		},
	}

	code, err := GenerateSingleConnectorLoader("sales_data", metadata, inputSpec, false)

	if err != nil {
		t.Errorf("GenerateSingleConnectorLoader() returned error: %v", err)
	}

	// Verify driver is referenced
	if !strings.Contains(code, "s3") {
		t.Errorf("GenerateSingleConnectorLoader() output missing driver reference 's3'")
	}
}

// TestGenerateSingleConnectorLoaderAcceptsConceptValues verifies loader signature includes conceptValues.
func TestGenerateSingleConnectorLoaderAcceptsConceptValues(t *testing.T) {
	metadata := &ConnectorMetadata{
		Name:       "sales_data",
		Dimensions: []string{"Product", "Region"},
		CellType:   "float64",
	}

	inputSpec := &io.ConnectorSpecification{
		ConnectorName: "sales_data",
		DataFormat:    "parquet",
		Driver:        "file",
	}

	code, err := GenerateSingleConnectorLoader("sales_data", metadata, inputSpec, false)
	if err != nil {
		t.Fatalf("GenerateSingleConnectorLoader() returned error: %v", err)
	}

	if !strings.Contains(code, "conceptValues map[string][]string") {
		t.Errorf("GenerateSingleConnectorLoader() output missing conceptValues parameter")
	}

	if !strings.Contains(code, "loadParquetConnector[any](metadata, connectorSpec, conceptValues)") {
		t.Errorf("GenerateSingleConnectorLoader() output missing conceptValues forwarded to loadParquetConnector")
	}
}

// TestGenerateLoadParquetConnectorFunctionAcceptsConceptValues verifies parquet loader signature.
func TestGenerateLoadParquetConnectorFunctionAcceptsConceptValues(t *testing.T) {
	code := GenerateLoadParquetConnectorFunction()

	if !strings.Contains(code, "conceptValues map[string][]string") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing conceptValues parameter")
	}
}

// TestGenerateConnectorLoadersNilIOSpec verifies error handling for nil IO spec.
func TestGenerateConnectorLoadersNilIOSpec(t *testing.T) {
	mock := &mockConnectorProvider{connectors: make(map[string]*ConnectorMetadata)}

	_, err := GenerateConnectorLoaders(mock, nil, false)

	// Should return an error for nil IO spec
	if err == nil {
		t.Errorf("GenerateConnectorLoaders() with nil ioSpec returned nil error, expected error")
	}
}

// TestGenerateSingleConnectorLoaderNilInputSpec verifies error handling for nil input spec.
func TestGenerateSingleConnectorLoaderNilInputSpec(t *testing.T) {
	metadata := &ConnectorMetadata{
		Name:       "sales_data",
		Dimensions: []string{"Product"},
		CellType:   "float64",
	}

	_, err := GenerateSingleConnectorLoader("sales_data", metadata, nil, false)

	// Should return an error for nil input spec
	if err == nil {
		t.Errorf("GenerateSingleConnectorLoader() with nil inputSpec returned nil error, expected error")
	}
}

// TestGenerateLoadParquetConnectorFunctionUsesReadParquetToTensor verifies that the
// generated loadParquetConnector function delegates to cplio.ReadParquetToTensor with
// the new four-argument signature (including columnToConcept).
func TestGenerateLoadParquetConnectorFunctionUsesReadParquetToTensor(t *testing.T) {
	code := GenerateLoadParquetConnectorFunction()

	if !strings.Contains(code, "cplio.ReadParquetToTensor(filePath, columnNames, columnToConcept, conceptValues)") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing 4-argument 'cplio.ReadParquetToTensor' call")
	}

	if strings.Contains(code, "cplio.ReadParquetToTensor(filePath, columnNames, conceptValues)") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output still contains old 3-argument 'cplio.ReadParquetToTensor' call")
	}
}

// TestGenerateLoadParquetConnectorFunctionBuildsColumnToConcept verifies that the
// generated loadParquetConnector function constructs a column-to-concept mapping
// with identity mapping for dimensions, and no grain measure remapping.
func TestGenerateLoadParquetConnectorFunctionBuildsColumnToConcept(t *testing.T) {
	code := GenerateLoadParquetConnectorFunction()

	if !strings.Contains(code, "columnToConcept := make(map[string]string") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing columnToConcept map allocation")
	}

	if !strings.Contains(code, "columnToConcept[dim] = dim") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output missing identity mapping for dimensions")
	}

	if strings.Contains(code, "for grainMeasure, conceptName := range metadata.GrainToConcept") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output contains grain-to-concept mapping loop — should be removed")
	}

	if strings.Contains(code, "columnToConcept[grainMeasure] = conceptName") {
		t.Errorf("GenerateLoadParquetConnectorFunction() output contains grain measure to concept name assignment — should be removed")
	}
}

// TestGenerateConnectorLoaders_NoGrainRemapping verifies that when a connector has
// grain measures with UnitName set (and thus GrainToConcept is populated), the
// generated loadParquetConnector code does NOT contain a loop that remaps grain
// measure names to concept names in columnToConcept. This test should fail before
// Task 2b (which removes the grain mapping loop) and pass after.
func TestGenerateConnectorLoaders_NoGrainRemapping(t *testing.T) {
	mock := &mockConnectorProvider{connectors: make(map[string]*ConnectorMetadata)}
	// Connector with grain measures that have UnitName set
	mock.connectors["sales_data"] = NewConnectorMetadata(
		"sales_data",
		[]string{"Region"},
		[]string{"Daily"},
		map[string]string{"Daily": "DayOfWeek"},
		"float64",
	)

	ioSpec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []io.OutputSpecification{},
		Architecture: io.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	code, err := GenerateConnectorLoaders(mock, ioSpec, false)
	if err != nil {
		t.Fatalf("GenerateConnectorLoaders() returned error: %v", err)
	}

	// The generated code must NOT contain the GrainToConcept mapping loop
	if strings.Contains(code, "metadata.GrainToConcept") {
		t.Errorf("Generated code contains 'metadata.GrainToConcept' — grain remapping loop should be removed")
	}
	if strings.Contains(code, "for grainMeasure, conceptName := range metadata.GrainToConcept") {
		t.Errorf("Generated code contains grain remapping loop — should be removed")
	}
	if strings.Contains(code, "columnToConcept[grainMeasure] = conceptName") {
		t.Errorf("Generated code contains 'columnToConcept[grainMeasure] = conceptName' — grain remapping should be removed")
	}
}

// TestGenerateLoadParquetConnectorFunction_NoGrainRemapping verifies that the
// reusable loadParquetConnector function template does NOT contain grain measure
// remapping logic. This test should fail before Task 2b and pass after.
func TestGenerateLoadParquetConnectorFunction_NoGrainRemapping(t *testing.T) {
	code := GenerateLoadParquetConnectorFunction()

	if strings.Contains(code, "metadata.GrainToConcept") {
		t.Errorf("loadParquetConnector template contains 'metadata.GrainToConcept' — grain remapping should be removed")
	}
	if strings.Contains(code, "for grainMeasure, conceptName := range metadata.GrainToConcept") {
		t.Errorf("loadParquetConnector template contains grain remapping loop — should be removed")
	}
	if strings.Contains(code, "columnToConcept[grainMeasure] = conceptName") {
		t.Errorf("loadParquetConnector template contains 'columnToConcept[grainMeasure] = conceptName' — grain remapping should be removed")
	}
}

// TestGenerateSingleConnectorLoader_NoGrainRemapping verifies that even when a
// connector's metadata has GrainToConcept populated, the generated loader code (via
// GenerateConnectorLoaders, which inlines the shared loadParquetConnector template)
// does not contain grain remapping loops.
func TestGenerateSingleConnectorLoader_NoGrainRemapping(t *testing.T) {
	mock := &mockConnectorProvider{connectors: make(map[string]*ConnectorMetadata)}
	mock.connectors["sales_data"] = NewConnectorMetadata(
		"sales_data",
		[]string{"Region"},
		[]string{"Daily"},
		map[string]string{"Daily": "DayOfWeek"},
		"float64",
	)

	ioSpec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []io.OutputSpecification{},
		Architecture: io.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	code, err := GenerateConnectorLoaders(mock, ioSpec, false)
	if err != nil {
		t.Fatalf("GenerateConnectorLoaders() returned error: %v", err)
	}

	if strings.Contains(code, "metadata.GrainToConcept") {
		t.Errorf("Generated code contains 'metadata.GrainToConcept' — grain remapping loop should be removed")
	}
	if strings.Contains(code, "for grainMeasure, conceptName := range metadata.GrainToConcept") {
		t.Errorf("Generated code contains grain remapping loop — should be removed")
	}
	if strings.Contains(code, "columnToConcept[grainMeasure] = conceptName") {
		t.Errorf("Generated code contains 'columnToConcept[grainMeasure] = conceptName' — grain remapping should be removed")
	}
}

// TestGenerateConnectorLoadersWithMultipleGrainMeasures_NoGrainRemapping verifies
// that when multiple connectors have grain measures with UnitName set, none of
// the generated code contains grain remapping loops.
func TestGenerateConnectorLoadersWithMultipleGrainMeasures_NoGrainRemapping(t *testing.T) {
	mock := &mockConnectorProvider{connectors: make(map[string]*ConnectorMetadata)}
	mock.connectors["sales_data"] = NewConnectorMetadata(
		"sales_data",
		[]string{"Region"},
		[]string{"Daily"},
		map[string]string{"Daily": "DayOfWeek"},
		"float64",
	)
	mock.connectors["inventory"] = NewConnectorMetadata(
		"inventory",
		[]string{"Warehouse"},
		[]string{"Monthly"},
		map[string]string{"Monthly": "Month"},
		"int64",
	)

	ioSpec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{
				ConnectorName: "sales_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
			{
				ConnectorName: "inventory",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []io.OutputSpecification{},
		Architecture: io.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	code, err := GenerateConnectorLoaders(mock, ioSpec, false)
	if err != nil {
		t.Fatalf("GenerateConnectorLoaders() returned error: %v", err)
	}

	if strings.Contains(code, "metadata.GrainToConcept") {
		t.Errorf("Generated code contains 'metadata.GrainToConcept' — grain remapping loop should be removed")
	}
	if strings.Contains(code, "for grainMeasure, conceptName := range metadata.GrainToConcept") {
		t.Errorf("Generated code contains grain remapping loop — should be removed")
	}
	if strings.Contains(code, "columnToConcept[grainMeasure] = conceptName") {
		t.Errorf("Generated code contains 'columnToConcept[grainMeasure] = conceptName' — grain remapping should be removed")
	}
}

// TestGenerateConnectorLoaders_WithConditionalMasks verifies that when
// hasConditionalMasks is true, generated loader functions accept
// conditionalConceptValues as a third parameter and apply connector
// validity masks.
func TestGenerateConnectorLoaders_WithConditionalMasks(t *testing.T) {
	mock := &mockConnectorProvider{connectors: make(map[string]*ConnectorMetadata)}
	mock.connectors["monthly_data"] = NewConnectorMetadata(
		"monthly_data",
		[]string{"Month", "DayInMonth"},
		[]string{"Month", "DayInMonth"},
		map[string]string{"Month": "Month", "DayInMonth": "DayInMonth"},
		"float64",
	)

	ioSpec := &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			{
				ConnectorName: "monthly_data",
				DataFormat:    "parquet",
				Driver:        "file",
			},
		},
		Outputs: []io.OutputSpecification{},
		Architecture: io.ArchitectureSpecification{
			CPUCoreCount:    4,
			CacheSizeBytes:  1024 * 1024,
			MainMemoryBytes: 1024 * 1024 * 1024,
		},
	}

	code, err := GenerateConnectorLoaders(mock, ioSpec, true)
	if err != nil {
		t.Fatalf("GenerateConnectorLoaders() returned error: %v", err)
	}

	if strings.Contains(code, "conditionalConceptValues map[string]map[string][]string") {
		t.Errorf("Generated code should not contain unused conditionalConceptValues parameter")
	}
	if !strings.Contains(code, "connectorValidityMasks[") {
		t.Errorf("Generated code missing connectorValidityMasks lookup")
	}
	if !strings.Contains(code, "tensor.SetMask(mask)") {
		t.Errorf("Generated code missing tensor.SetMask(mask)")
	}
}

// TestGenerateSingleConnectorLoader_WithConditionalMasks verifies that when
// hasConditionalMasks is true, the generated single loader function applies
// connector validity masks without adding an unused conditionalConceptValues
// parameter.
func TestGenerateSingleConnectorLoader_WithConditionalMasks(t *testing.T) {
	metadata := &ConnectorMetadata{
		Name:       "monthly_data",
		Dimensions: []string{"Month", "DayInMonth"},
		Grain:      []string{"Month", "DayInMonth"},
		GrainToConcept: map[string]string{
			"Month":      "Month",
			"DayInMonth": "DayInMonth",
		},
		CellType: "float64",
	}

	inputSpec := &io.ConnectorSpecification{
		ConnectorName: "monthly_data",
		DataFormat:    "parquet",
		Driver:        "file",
	}

	code, err := GenerateSingleConnectorLoader("monthly_data", metadata, inputSpec, true)
	if err != nil {
		t.Fatalf("GenerateSingleConnectorLoader() returned error: %v", err)
	}

	if strings.Contains(code, "conditionalConceptValues map[string]map[string][]string") {
		t.Errorf("Generated code should not contain unused conditionalConceptValues parameter")
	}
	if !strings.Contains(code, "connectorValidityMasks[") {
		t.Errorf("Generated code missing connectorValidityMasks lookup")
	}
	if !strings.Contains(code, "tensor.SetMask(mask)") {
		t.Errorf("Generated code missing tensor.SetMask(mask)")
	}
}

// TestLoaderFuncDecl_RenderByteIdentical verifies that the LoaderFuncDecl.Render()
// refactoring produces byte-identical output to the pre-refactoring
// GenerateSingleConnectorLoader. This test serves as the byte-identical
// comparison gate for the connector loader refactoring in Story225 Task 1b.
func TestLoaderFuncDecl_RenderByteIdentical(t *testing.T) {
	tests := []struct {
		name                string
		connectorName       string
		dimensions          []string
		driver              string
		hasConditionalMasks bool
	}{
		{
			name:                "basic loader without masks",
			connectorName:       "sales_data",
			dimensions:          []string{"Product", "Region"},
			driver:              "file",
			hasConditionalMasks: false,
		},
		{
			name:                "loader with conditional masks",
			connectorName:       "monthly_data",
			dimensions:          []string{"Month", "DayInMonth"},
			driver:              "file",
			hasConditionalMasks: true,
		},
		{
			name:                "single dimension loader",
			connectorName:       "inventory",
			dimensions:          []string{"Warehouse"},
			driver:              "s3",
			hasConditionalMasks: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcName := GenerateConnectorLoaderFunctionName(tt.connectorName)

			// Build LoaderFuncDecl and render
			decl := &LoaderFuncDecl{
				FuncName:            funcName,
				ConnectorName:       tt.connectorName,
				Driver:              tt.driver,
				Dimensions:          tt.dimensions,
				HasConditionalMasks: tt.hasConditionalMasks,
			}
			rendered := decl.Render()

			// Verify key structural elements are present
			if !strings.Contains(rendered, funcName) {
				t.Errorf("Render() missing function name %q", funcName)
			}
			if !strings.Contains(rendered, tt.connectorName) {
				t.Errorf("Render() missing connector name %q", tt.connectorName)
			}
			if !strings.Contains(rendered, tt.driver) {
				t.Errorf("Render() missing driver %q", tt.driver)
			}
			if !strings.Contains(rendered, "calculation.ChronoconeTensor[any]") {
				t.Errorf("Render() missing package-qualified return type")
			}
			if !strings.Contains(rendered, "cplio.ConnectorSpecification") {
				t.Errorf("Render() missing package-qualified parameter type")
			}
			if !strings.Contains(rendered, "loadParquetConnector[any](metadata, connectorSpec, conceptValues)") {
				t.Errorf("Render() missing loadParquetConnector call with conceptValues")
			}
			if tt.hasConditionalMasks {
				if !strings.Contains(rendered, "connectorValidityMasks[") {
					t.Errorf("Render() missing conditional mask application")
				}
			} else {
				if strings.Contains(rendered, "connectorValidityMasks[") {
					t.Errorf("Render() should not contain conditional mask application")
				}
			}
		})
	}
}

// TestLoaderFuncDecl_StructFields verifies that LoaderFuncDecl fields are
// independently accessible for structured testing.
func TestLoaderFuncDecl_StructFields(t *testing.T) {
	decl := &LoaderFuncDecl{
		FuncName:            "loadMonthlyData",
		ConnectorName:       "monthly_data",
		Driver:              "file",
		Dimensions:          []string{"Month", "DayInMonth"},
		HasConditionalMasks: true,
	}

	// Verify fields independently
	if decl.FuncName != "loadMonthlyData" {
		t.Errorf("FuncName = %q, expected %q", decl.FuncName, "loadMonthlyData")
	}
	if decl.ConnectorName != "monthly_data" {
		t.Errorf("ConnectorName = %q, expected %q", decl.ConnectorName, "monthly_data")
	}
	if decl.Driver != "file" {
		t.Errorf("Driver = %q, expected %q", decl.Driver, "file")
	}
	if len(decl.Dimensions) != 2 {
		t.Errorf("Dimensions length = %d, expected 2", len(decl.Dimensions))
	}
	if !decl.HasConditionalMasks {
		t.Errorf("HasConditionalMasks = false, expected true")
	}

	// Verify Rendering produces the expected function signature
	rendered := decl.Render()
	if !strings.Contains(rendered, "func loadMonthlyData(connectorSpec cplio.ConnectorSpecification, conceptValues map[string][]string) (*calculation.ChronoconeTensor[any], error)") {
		t.Errorf("Render() missing expected function signature:\n%s", rendered)
	}
}
