// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"github.com/chronocone/chronocone_planning_language/src/main/go/connector"
)

// LookupMetadata holds synthesis-level metadata for a CPL lookup.
// It is derived from CILLookupNode and stored in the CILRegistry for
// use during preamble generation and code generation.
type LookupMetadata struct {
	// Name is the identifier for this lookup (e.g., "CountryProductIndex").
	Name string

	// Dimensions is the ordered list of resolved concept names for the lookup's dimensions.
	Dimensions []string

	// CellType is the resolved measure name for the lookup's cell values.
	CellType string
}

// CILRegistry holds type registrations for measures, metrics, and connectors.
// It encapsulates the registration concern of the CIL transformer, keeping
// mutation of type mappings separate from code generation and tree walking.
type CILRegistry struct {
	// measureTypes is a map from measure name to its Go type (e.g., "float64").
	measureTypes map[string]string

	// metricTypes is a map from metric name to its element type.
	metricTypes map[string]string

	// connectors is a map from connector name to its runtime metadata.
	// This is populated when visiting CILConnectorNode instances.
	connectors map[string]*connector.ConnectorMetadata

	// lookups is a map from lookup name to its synthesis metadata.
	// This is populated when walking CILLookupNode instances.
	lookups map[string]*LookupMetadata
}

// NewCILRegistry creates a new registry with empty type mappings.
//
// Returns:
//   - *CILRegistry: A new registry instance.
func NewCILRegistry() *CILRegistry {
	return &CILRegistry{
		measureTypes: make(map[string]string),
		metricTypes:  make(map[string]string),
		connectors:   make(map[string]*connector.ConnectorMetadata),
		lookups:      make(map[string]*LookupMetadata),
	}
}

// RegisterMeasureType registers the Go type for a measure name.
//
// Parameters:
//   - measureName: The name of the measure.
//   - goType: The Go type string (e.g., "float64", "int64").
func (r *CILRegistry) RegisterMeasureType(measureName string, goType string) {
	if r != nil {
		r.measureTypes[measureName] = goType
	}
}

// RegisterMetricType registers the element type for a metric name.
//
// Parameters:
//   - metricName: The name of the metric.
//   - elementType: The element type string.
func (r *CILRegistry) RegisterMetricType(metricName string, elementType string) {
	if r != nil {
		r.metricTypes[metricName] = elementType
	}
}

// RegisterConnector registers a connector's metadata.
// Called when visiting a CILConnectorNode.
//
// Parameters:
//   - name: The connector name.
//   - dimensions: The ordered list of dimension names.
//   - grain: The ordered list of grain measure names.
//   - grainToConcept: A map from grain measure names to their underlying concept names.
//   - cellType: The Go cell type.
func (r *CILRegistry) RegisterConnector(name string, dimensions []string, grain []string, grainToConcept map[string]string, cellType string) {
	if r == nil {
		return
	}
	metadata := connector.NewConnectorMetadata(name, dimensions, grain, grainToConcept, cellType)
	r.connectors[name] = metadata
}

// GetConnectorMetadata retrieves the metadata for a registered connector.
//
// Parameters:
//   - name: The connector name.
//
// Returns:
//   - *connector.ConnectorMetadata: The connector metadata, or nil if not found.
func (r *CILRegistry) GetConnectorMetadata(name string) *connector.ConnectorMetadata {
	if r == nil {
		return nil
	}
	return r.connectors[name]
}

// AllConnectorMetadata returns a map of all registered connector metadata.
// The returned map is the registry's internal map; callers must not mutate it.
func (r *CILRegistry) AllConnectorMetadata() map[string]*connector.ConnectorMetadata {
	if r == nil {
		return nil
	}
	return r.connectors
}

// RegisterLookup registers a lookup's metadata.
// Called when visiting a CILLookupNode during synthesis.
//
// Parameters:
//   - name: The lookup name.
//   - dimensions: The ordered list of dimension concept names.
//   - cellType: The measure name for cell values.
func (r *CILRegistry) RegisterLookup(name string, dimensions []string, cellType string) {
	if r == nil {
		return
	}
	// Defensive copy of dimensions slice so caller mutations do not affect registry.
	dimsCopy := make([]string, len(dimensions))
	copy(dimsCopy, dimensions)
	r.lookups[name] = &LookupMetadata{
		Name:       name,
		Dimensions: dimsCopy,
		CellType:   cellType,
	}
}

// GetLookupMetadata retrieves the metadata for a registered lookup.
//
// Parameters:
//   - name: The lookup name.
//
// Returns:
//   - *LookupMetadata: The lookup metadata, or nil if not found.
func (r *CILRegistry) GetLookupMetadata(name string) *LookupMetadata {
	if r == nil {
		return nil
	}
	return r.lookups[name]
}

// AllLookupMetadata returns a map of all registered lookup metadata.
// The returned map is the registry's internal map; callers must not mutate it.
func (r *CILRegistry) AllLookupMetadata() map[string]*LookupMetadata {
	if r == nil {
		return nil
	}
	return r.lookups
}
