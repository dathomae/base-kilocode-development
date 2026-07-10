// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package connector

// ConnectorMetadata holds the runtime information needed to load a connector.
// It is populated during the CIL tree traversal when a cil.CILConnectorNode
// is encountered and synthesized into a map in the generated code.
type ConnectorMetadata struct {
	// Name is the unique identifier for the connector.
	Name string

	// Dimensions is an ordered list of dimension names for the connector.
	Dimensions []string

	// Grain is an ordered list of grain measure names for the connector.
	Grain []string

	// GrainToConcept maps grain measure names to their underlying concept names.
	// This is used by parquet readers to look up concept values when the grain
	// measure name (parquet column name) differs from the concept name.
	GrainToConcept map[string]string

	// CellType is the Go type for the connector's cell values (e.g., "float64", "int64").
	CellType string
}

// ConnectorSynthesisInfo holds the information needed during synthesis for a specific connector.
// This includes both the static metadata and synthesis-time information such as
// the CIL representation and code generation state.
type ConnectorSynthesisInfo struct {
	// Metadata contains the runtime metadata for the connector.
	Metadata *ConnectorMetadata

	// GeneratedLoaderCode contains the Go code generated for loading this connector.
	// This is accumulated during the synthesis phase.
	GeneratedLoaderCode string
}

// ConnectorProvider is an interface for types that can provide connector metadata.
// It is used to break the circular dependency between the cil and connector packages.
type ConnectorProvider interface {
	GetConnectorMetadata(name string) *ConnectorMetadata
}

// NewConnectorMetadata creates a new ConnectorMetadata with the specified properties.
//
// Parameters:
//   - name: The unique identifier for the connector.
//   - dimensions: An ordered list of dimension names.
//   - grain: An ordered list of grain measure names.
//   - grainToConcept: A map from grain measure names to their underlying concept names.
//   - cellType: The Go type for cell values.
//
// Returns:
//   - *ConnectorMetadata: A new metadata instance.
func NewConnectorMetadata(name string, dimensions []string, grain []string, grainToConcept map[string]string, cellType string) *ConnectorMetadata {
	// Copy dimensions slice to ensure immutability
	dimsCopy := make([]string, len(dimensions))
	copy(dimsCopy, dimensions)

	// Copy grain slice to ensure immutability
	grainCopy := make([]string, len(grain))
	copy(grainCopy, grain)

	// Copy grainToConcept map to ensure immutability
	var gtcCopy map[string]string
	if len(grainToConcept) > 0 {
		gtcCopy = make(map[string]string, len(grainToConcept))
		for k, v := range grainToConcept {
			gtcCopy[k] = v
		}
	}

	return &ConnectorMetadata{
		Name:           name,
		Dimensions:     dimsCopy,
		Grain:          grainCopy,
		GrainToConcept: gtcCopy,
		CellType:       cellType,
	}
}

// NewConnectorSynthesisInfo creates a new ConnectorSynthesisInfo for the given metadata.
//
// Parameters:
//   - metadata: The runtime metadata for the connector.
//
// Returns:
//   - *ConnectorSynthesisInfo: A new synthesis info instance.
func NewConnectorSynthesisInfo(metadata *ConnectorMetadata) *ConnectorSynthesisInfo {
	return &ConnectorSynthesisInfo{
		Metadata:            metadata,
		GeneratedLoaderCode: "",
	}
}
