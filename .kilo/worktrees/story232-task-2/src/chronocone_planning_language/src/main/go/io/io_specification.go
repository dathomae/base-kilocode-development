// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package io

import (
	"encoding/json"
	"fmt"
	"os"
)

// IOSpecification represents the complete IO configuration for a CPL computation.
// It defines all connectors (data sources), outputs (metrics to produce),
// lookups (reference data sources), and the target compute architecture requirements.
type IOSpecification struct {
	Connectors   []ConnectorSpecification  `json:"connectors"`
	Outputs      []OutputSpecification     `json:"outputs"`
	Lookups      []LookupSpecification     `json:"lookups"`
	Architecture ArchitectureSpecification `json:"architecture"`
}

// ConnectorSpecification defines a single input data source for a computation.
// It specifies which connector to use, the data format, driver configuration,
// and any driver-specific parameters required to establish the connection.
type ConnectorSpecification struct {
	ConnectorName    string                   `json:"connector"`
	DataFormat       string                   `json:"data_format"`
	Driver           string                   `json:"driver"`
	DriverParameters []map[string]interface{} `json:"driver_parameters"`
}

// LookupSpecification defines a single lookup reference data source for a computation.
// It specifies which lookup to use, the data format, driver configuration,
// and any driver-specific parameters required to establish the connection.
type LookupSpecification struct {
	LookupName       string                   `json:"lookup"`
	DataFormat       string                   `json:"data_format"`
	Driver           string                   `json:"driver"`
	DriverParameters []map[string]interface{} `json:"driver_parameters"`
}

// OutputSpecification defines a single metric output for a computation.
// It specifies the metric name, dimensions, cell type, and grain for the output.
// Grain is an ordered list (slice) of grain dimensions.
type OutputSpecification struct {
	MetricName string   `json:"metric"`
	Dimensions []string `json:"dimensions"`
	CellType   string   `json:"celltype"`
	Grain      []string `json:"grain"`
	Unit       string   `json:"unit,omitempty"` // expected output unit, e.g. "USD"
}

// ArchitectureSpecification defines the target compute resource requirements
// for executing a CPL computation. It includes CPU cores, cache size, and
// main memory requirements.
type ArchitectureSpecification struct {
	CPUCoreCount    int `json:"cpu_cores"`
	CacheSizeBytes  int `json:"cache_size"`
	MainMemoryBytes int `json:"main_memory_size"`
}

// LoadIOSpecification parses an IO specification from a JSON file.
//
// Parameters:
//   - filePath: The path to the JSON file containing the IO specification.
//
// Returns:
//   - *IOSpecification: The parsed IO specification, or nil if parsing fails.
//   - error: An error if the file cannot be read or parsed.
func LoadIOSpecification(filePath string) (*IOSpecification, error) {
	// Read the file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read IO specification file: %w", err)
	}

	// Parse the JSON into IOSpecification struct
	var spec IOSpecification
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse IO specification JSON: %w", err)
	}

	return &spec, nil
}
