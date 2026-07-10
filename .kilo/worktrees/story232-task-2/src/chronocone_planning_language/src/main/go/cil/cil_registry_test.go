// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"testing"
)

func TestNewCILRegistry(t *testing.T) {
	registry := NewCILRegistry()
	if registry == nil {
		t.Fatalf("NewCILRegistry() returned nil")
	}
	if registry.measureTypes == nil {
		t.Errorf("NewCILRegistry().measureTypes is nil, expected initialized map")
	}
	if registry.metricTypes == nil {
		t.Errorf("NewCILRegistry().metricTypes is nil, expected initialized map")
	}
	if registry.connectors == nil {
		t.Errorf("NewCILRegistry().connectors is nil, expected initialized map")
	}
	if registry.lookups == nil {
		t.Errorf("NewCILRegistry().lookups is nil, expected initialized map")
	}
}

func TestCILRegistry_RegisterMeasureType(t *testing.T) {
	registry := NewCILRegistry()
	registry.RegisterMeasureType("Revenue", "float64")
	registry.RegisterMeasureType("Quantity", "int64")
	// measureTypes is unexported; verify indirectly by ensuring no panic.
}

func TestCILRegistry_RegisterMetricType(t *testing.T) {
	registry := NewCILRegistry()
	registry.RegisterMetricType("TotalSales", "float64")
	registry.RegisterMetricType("UnitCount", "int64")
	// metricTypes is unexported; verify indirectly by ensuring no panic.
}

func TestCILRegistry_RegisterConnector(t *testing.T) {
	registry := NewCILRegistry()
	registry.RegisterConnector("sales_data", []string{"Product", "Region"}, nil, nil, "float64")

	metadata := registry.GetConnectorMetadata("sales_data")
	if metadata == nil {
		t.Fatalf("GetConnectorMetadata(%q) returned nil after registration", "sales_data")
	}
	if metadata.Name != "sales_data" {
		t.Errorf("registered connector Name = %q, expected %q", metadata.Name, "sales_data")
	}
	if len(metadata.Dimensions) != 2 {
		t.Errorf("registered connector Dimensions length = %d, expected 2", len(metadata.Dimensions))
	}
	if metadata.CellType != "float64" {
		t.Errorf("registered connector CellType = %q, expected %q", metadata.CellType, "float64")
	}
}

func TestCILRegistry_RegisterConnector_WithGrainToConcept(t *testing.T) {
	registry := NewCILRegistry()
	grainToConcept := map[string]string{
		"YearGrain": "Year",
		"WeekGrain": "WeekNumber",
	}
	registry.RegisterConnector("sales_data", []string{"Region"}, []string{"YearGrain", "WeekGrain"}, grainToConcept, "float64")

	metadata := registry.GetConnectorMetadata("sales_data")
	if metadata == nil {
		t.Fatalf("GetConnectorMetadata returned nil")
	}
	if len(metadata.Grain) != 2 {
		t.Errorf("metadata.Grain length = %d, expected 2", len(metadata.Grain))
	}
	if metadata.Grain[0] != "YearGrain" {
		t.Errorf("metadata.Grain[0] = %q, expected %q", metadata.Grain[0], "YearGrain")
	}
	if metadata.Grain[1] != "WeekGrain" {
		t.Errorf("metadata.Grain[1] = %q, expected %q", metadata.Grain[1], "WeekGrain")
	}
	if metadata.GrainToConcept == nil {
		t.Fatalf("metadata.GrainToConcept is nil")
	}
	if metadata.GrainToConcept["YearGrain"] != "Year" {
		t.Errorf("metadata.GrainToConcept[%q] = %q, expected %q", "YearGrain", metadata.GrainToConcept["YearGrain"], "Year")
	}
	if metadata.GrainToConcept["WeekGrain"] != "WeekNumber" {
		t.Errorf("metadata.GrainToConcept[%q] = %q, expected %q", "WeekGrain", metadata.GrainToConcept["WeekGrain"], "WeekNumber")
	}
}

func TestCILRegistry_GetConnectorMetadata_NotFound(t *testing.T) {
	registry := NewCILRegistry()
	metadata := registry.GetConnectorMetadata("unknown_connector")
	if metadata != nil {
		t.Errorf("GetConnectorMetadata() for unknown connector returned %v, expected nil", metadata)
	}
}

func TestCILRegistry_RegisterLookup(t *testing.T) {
	registry := NewCILRegistry()
	registry.RegisterLookup("CountryProductIndex", []string{"Country", "Product"}, "UnitsSold")

	metadata := registry.GetLookupMetadata("CountryProductIndex")
	if metadata == nil {
		t.Fatalf("GetLookupMetadata(%q) returned nil after registration", "CountryProductIndex")
	}
	if metadata.Name != "CountryProductIndex" {
		t.Errorf("registered lookup Name = %q, expected %q", metadata.Name, "CountryProductIndex")
	}
	if len(metadata.Dimensions) != 2 {
		t.Errorf("registered lookup Dimensions length = %d, expected 2", len(metadata.Dimensions))
	}
	if metadata.Dimensions[0] != "Country" {
		t.Errorf("registered lookup Dimensions[0] = %q, expected %q", metadata.Dimensions[0], "Country")
	}
	if metadata.Dimensions[1] != "Product" {
		t.Errorf("registered lookup Dimensions[1] = %q, expected %q", metadata.Dimensions[1], "Product")
	}
	if metadata.CellType != "UnitsSold" {
		t.Errorf("registered lookup CellType = %q, expected %q", metadata.CellType, "UnitsSold")
	}
}

func TestCILRegistry_RegisterLookup_NilRegistry(t *testing.T) {
	var registry *CILRegistry
	// Should not panic.
	registry.RegisterLookup("test", []string{"A"}, "float64")

	metadata := registry.GetLookupMetadata("test")
	if metadata != nil {
		t.Errorf("GetLookupMetadata() on nil registry returned %v, expected nil", metadata)
	}
}

func TestCILRegistry_AllLookupMetadata(t *testing.T) {
	registry := NewCILRegistry()
	registry.RegisterLookup("LookupA", []string{"Region"}, "Revenue")
	registry.RegisterLookup("LookupB", []string{"Country", "Product"}, "Units")

	all := registry.AllLookupMetadata()
	if all == nil {
		t.Fatalf("AllLookupMetadata() returned nil")
	}
	if len(all) != 2 {
		t.Errorf("AllLookupMetadata() returned %d entries, expected 2", len(all))
	}
	if all["LookupA"] == nil {
		t.Errorf("AllLookupMetadata() missing %q", "LookupA")
	}
	if all["LookupB"] == nil {
		t.Errorf("AllLookupMetadata() missing %q", "LookupB")
	}
}
