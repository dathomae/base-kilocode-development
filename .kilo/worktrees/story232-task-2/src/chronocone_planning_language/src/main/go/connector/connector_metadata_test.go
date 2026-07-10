// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package connector

import (
	"testing"
)

// TestNewConnectorMetadata verifies that NewConnectorMetadata creates a struct with correct fields.
func TestNewConnectorMetadata(t *testing.T) {
	tests := []struct {
		name       string
		connName   string
		dimensions []string
		cellType   string
	}{
		{
			name:       "basic connector metadata",
			connName:   "sales_data",
			dimensions: []string{"Product", "Region", "Time"},
			cellType:   "float64",
		},
		{
			name:       "single dimension connector",
			connName:   "time_series",
			dimensions: []string{"Date"},
			cellType:   "int64",
		},
		{
			name:       "empty dimensions",
			connName:   "scalar_value",
			dimensions: []string{},
			cellType:   "string",
		},
		{
			name:       "many dimensions",
			connName:   "complex_data",
			dimensions: []string{"A", "B", "C", "D", "E"},
			cellType:   "float32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := NewConnectorMetadata(tt.connName, tt.dimensions, nil, nil, tt.cellType)

			if metadata == nil {
				t.Errorf("NewConnectorMetadata() returned nil, expected non-nil")
				return
			}

			if metadata.Name != tt.connName {
				t.Errorf("NewConnectorMetadata().Name = %q, expected %q", metadata.Name, tt.connName)
			}

			if len(metadata.Dimensions) != len(tt.dimensions) {
				t.Errorf("NewConnectorMetadata().Dimensions length = %d, expected %d", len(metadata.Dimensions), len(tt.dimensions))
			} else {
				for i, dim := range tt.dimensions {
					if metadata.Dimensions[i] != dim {
						t.Errorf("NewConnectorMetadata().Dimensions[%d] = %q, expected %q", i, metadata.Dimensions[i], dim)
					}
				}
			}

			if metadata.CellType != tt.cellType {
				t.Errorf("NewConnectorMetadata().CellType = %q, expected %q", metadata.CellType, tt.cellType)
			}
		})
	}
}

// TestNewConnectorSynthesisInfo verifies that NewConnectorSynthesisInfo creates struct correctly.
func TestNewConnectorSynthesisInfo(t *testing.T) {
	tests := []struct {
		name           string
		metadata       *ConnectorMetadata
		wantLoaderCode string
	}{
		{
			name: "basic synthesis info",
			metadata: &ConnectorMetadata{
				Name:       "sales_data",
				Dimensions: []string{"Product", "Region"},
				CellType:   "float64",
			},
			wantLoaderCode: "",
		},
		{
			name: "synthesis info with metadata",
			metadata: &ConnectorMetadata{
				Name:       "inventory",
				Dimensions: []string{"Warehouse", "SKU"},
				CellType:   "int64",
			},
			wantLoaderCode: "",
		},
		{
			name:           "nil metadata",
			metadata:       nil,
			wantLoaderCode: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := NewConnectorSynthesisInfo(tt.metadata)

			if info == nil {
				t.Errorf("NewConnectorSynthesisInfo() returned nil, expected non-nil")
				return
			}

			if info.Metadata != tt.metadata {
				t.Errorf("NewConnectorSynthesisInfo().Metadata = %v, expected %v", info.Metadata, tt.metadata)
			}

			if info.GeneratedLoaderCode != tt.wantLoaderCode {
				t.Errorf("NewConnectorSynthesisInfo().GeneratedLoaderCode = %q, expected %q", info.GeneratedLoaderCode, tt.wantLoaderCode)
			}
		})
	}
}

// TestConnectorMetadataImmutability verifies that modifying input slices doesn't affect the struct.
func TestConnectorMetadataImmutability(t *testing.T) {
	dimensions := []string{"Product", "Region"}
	metadata := NewConnectorMetadata("sales_data", dimensions, nil, nil, "float64")

	if metadata == nil {
		t.Fatalf("NewConnectorMetadata() returned nil")
	}

	// Modify original slice
	dimensions[0] = "Modified"

	// Verify metadata wasn't affected
	if metadata.Dimensions[0] == "Modified" {
		t.Errorf("ConnectorMetadata.Dimensions was affected by external slice modification")
	}
}

// TestNewConnectorMetadataGrainToConcept verifies that NewConnectorMetadata correctly
// stores the GrainToConcept map and that it is a copy of the input map.
func TestNewConnectorMetadataGrainToConcept(t *testing.T) {
	grainToConcept := map[string]string{
		"YearGrain": "Year",
		"WeekGrain": "WeekNumber",
	}
	metadata := NewConnectorMetadata("sales", []string{"Region"}, []string{"YearGrain", "WeekGrain"}, grainToConcept, "float64")

	if metadata == nil {
		t.Fatalf("NewConnectorMetadata() returned nil")
	}

	if len(metadata.GrainToConcept) != 2 {
		t.Errorf("ConnectorMetadata.GrainToConcept length = %d, expected 2", len(metadata.GrainToConcept))
	}

	if metadata.GrainToConcept["YearGrain"] != "Year" {
		t.Errorf("ConnectorMetadata.GrainToConcept[\"YearGrain\"] = %q, expected %q", metadata.GrainToConcept["YearGrain"], "Year")
	}

	if metadata.GrainToConcept["WeekGrain"] != "WeekNumber" {
		t.Errorf("ConnectorMetadata.GrainToConcept[\"WeekGrain\"] = %q, expected %q", metadata.GrainToConcept["WeekGrain"], "WeekNumber")
	}

	// Verify the stored map is a different instance from the input map
	if &metadata.GrainToConcept == &grainToConcept {
		t.Errorf("ConnectorMetadata.GrainToConcept is the same map instance as the input")
	}
}

// TestNewConnectorMetadataGrainToConceptImmutability verifies that mutating the
// input grainToConcept map after construction does not affect the stored map.
func TestNewConnectorMetadataGrainToConceptImmutability(t *testing.T) {
	grainToConcept := map[string]string{
		"YearGrain": "Year",
	}
	metadata := NewConnectorMetadata("sales", []string{"Region"}, []string{"YearGrain"}, grainToConcept, "float64")

	if metadata == nil {
		t.Fatalf("NewConnectorMetadata() returned nil")
	}

	// Modify the original map after construction
	grainToConcept["YearGrain"] = "ModifiedYear"
	grainToConcept["WeekGrain"] = "WeekNumber"

	// Verify metadata wasn't affected
	if metadata.GrainToConcept["YearGrain"] != "Year" {
		t.Errorf("ConnectorMetadata.GrainToConcept[\"YearGrain\"] = %q, expected %q after external map modification", metadata.GrainToConcept["YearGrain"], "Year")
	}

	if _, exists := metadata.GrainToConcept["WeekGrain"]; exists {
		t.Errorf("ConnectorMetadata.GrainToConcept should not contain key %q added after construction", "WeekGrain")
	}
}
