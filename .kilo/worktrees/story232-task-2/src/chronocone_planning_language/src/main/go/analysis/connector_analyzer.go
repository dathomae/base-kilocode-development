// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file contains the ConnectorAnalyzer which handles connector-related analysis
// operations including connector registration and grain type validation.
package analysis

import "fmt"

// ConnectorAnalyzer handles connector-related analysis operations.
//
// It maintains the connectors map and provides methods for connector registration
// and grain type validation.
type ConnectorAnalyzer struct {
	// connectors is a map from connector name to its analyzed definition.
	// This map is built during analysis by processing connector definitions
	// from the symbol table. It provides O(1) access to connector information
	// for validation and code generation.
	connectors map[string]AnalysisConnectorDefinition
}

// NewConnectorAnalyzer creates a new ConnectorAnalyzer with initialized data structures.
//
// Returns:
//   - A pointer to a new, initialized ConnectorAnalyzer.
func NewConnectorAnalyzer() *ConnectorAnalyzer {
	return &ConnectorAnalyzer{
		connectors: make(map[string]AnalysisConnectorDefinition),
	}
}

// AddConnector adds a connector definition to the analyzer's connectors map.
// This is primarily useful for testing, allowing direct population of the connectors map.
func (ca *ConnectorAnalyzer) AddConnector(name string, def AnalysisConnectorDefinition) {
	ca.connectors[name] = def
}

// GetConnector returns the connector definition with the given name.
// Returns a zero-value AnalysisConnectorDefinition if the connector is not found.
func (ca *ConnectorAnalyzer) GetConnector(name string) AnalysisConnectorDefinition {
	return ca.connectors[name]
}

// GetConnectors returns the connectors map.
func (ca *ConnectorAnalyzer) GetConnectors() map[string]AnalysisConnectorDefinition {
	return ca.connectors
}

// HasConnector returns true if a connector with the given name exists.
func (ca *ConnectorAnalyzer) HasConnector(name string) bool {
	_, exists := ca.connectors[name]
	return exists
}

// ValidateGrainTypes validates the semantic type family of each connector grain
// after type derivation paths have been computed.
//
// This deferred validation pass iterates the connectors map and examines each
// AnalysisConnectorDefinition's Grain slice. Connectors with an empty Grain slice
// are skipped because they have no explicit grain measure to validate.
//
// A connector grain is accepted when the grain measure's TypeName is the built-in
// "Interval" type. A connector grain is also accepted when the grain measure's
// TypeName names a declared type in the typeAnalyzer whose AnalysisTypeDefinition
// DerivationPath terminates in "Interval", which indicates that the declared type
// ultimately derives from the built-in interval type.
//
// All other connector grain measure types are rejected. The validation error must
// name the connector, the grain measure, and the grain measure type so callers can
// identify the exact connector declaration and grain definition that failed
// interval-family validation.
func (ca *ConnectorAnalyzer) ValidateGrainTypes(typeAnalyzer *TypeAnalyzer) error {
	for _, connector := range ca.connectors {
		if len(connector.Grain) == 0 {
			continue
		}

		for grainIdx, grain := range connector.Grain {
			typeName := grain.TypeName
			if typeName == "Interval" {
				continue
			}

			typeDef := typeAnalyzer.GetType(typeName)
			if typeDef.Name != "" && len(typeDef.DerivationPath) > 0 && typeDef.DerivationPath[len(typeDef.DerivationPath)-1] == "Interval" {
				continue
			}

			return fmt.Errorf("connector %s grain[%d] measure %s has type %s which is not Interval or derived from Interval", connector.Name, grainIdx, grain.Name, grain.TypeName)
		}
	}

	return nil
}
