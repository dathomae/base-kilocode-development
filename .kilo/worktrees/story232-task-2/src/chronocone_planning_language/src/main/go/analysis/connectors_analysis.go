// Package analysis provides the analysis phase data structures for the Chronocone Planning Language.
//
// This file contains the connector analysis phase implementation.
// It processes connector definitions from the symbol table and builds the Connectors lookup map.
// See ERO004: Connectors Analysis Phase for the incremental implementation plan.
package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// AnalysisConnectorDefinition contains the analyzed connector information.
//
// This structure represents a fully-processed connector definition after the analysis phase.
// All references (dimensions, celltype, and grain) have been validated and resolved to their
// corresponding analysis-level definitions. The resolved references enable efficient
// validation and code generation during the synthesis phase.
//
// Connectors define the dimensional structure, data type, and temporal granularity of
// input data sources.
//
// Example connector declaration in CPL:
//
//	connector SalesData
//	dimensions [Country, Product, Date]
//	celltype Amount
//	grain Day
//
// This would result in an AnalysisConnectorDefinition with:
//   - Name: "SalesData"
//   - Dimensions: []*AnalysisConceptDefinition{CountryConcept, ProductConcept, DateConcept}
//   - CellType: *AnalysisMeasureDefinition(AmountMeasure)
//   - Grain: []*AnalysisMeasureDefinition{DayMeasure}
//   - SourceLocation: Location of the connector declaration in source
type AnalysisConnectorDefinition struct {
	// Name is the identifier for this connector (e.g., "SalesData", "Inventory").
	// This name is used to reference the connector in metrics, aggregators, and
	// other CPL constructs. It must be unique within the CPL document.
	Name string

	// Dimensions is a slice of pointers to resolved concept definitions.
	//
	// Each dimension represents a categorical axis of the connector's data.
	// The order of dimensions is significant as it defines the dimensional
	// structure of the data. All dimensions must reference valid concepts
	// that have been previously defined in the CPL document.
	//
	// During analysis, each dimension name is looked up in the Concepts map
	// and replaced with a pointer to the corresponding AnalysisConceptDefinition.
	// This ensures that all dimension references are valid and provides
	// direct access to concept information during code generation.
	Dimensions []*AnalysisConceptDefinition

	// CellType is a pointer to the resolved measure definition for the cell values.
	//
	// The celltype defines the data type and unit of the values stored in the
	// connector. It must reference a valid measure that has been previously
	// defined in the CPL document.
	//
	// During analysis, the celltype measure name is looked up in the Measures map
	// and replaced with a pointer to the corresponding AnalysisMeasureDefinition.
	// This ensures that the celltype reference is valid and provides direct
	// access to measure information during code generation.
	CellType *AnalysisMeasureDefinition

	// Grain is a slice of pointers to the resolved measure definitions for the
	// connector's granularity, ordered as declared.
	//
	// The grain defines the finest level of detail at which data in the connector
	// is recorded. It is typically a temporal measure (e.g., Day, Month, Year) but
	// can be any measure that represents a granularity level. The grain is used
	// during validation and code generation to ensure temporal consistency.
	//
	// During analysis, each grain measure name is looked up in the Measures map
	// and replaced with a pointer to the corresponding AnalysisMeasureDefinition.
	// An empty Grain slice indicates that no explicit grain has been specified.
	Grain []*AnalysisMeasureDefinition

	// SourceLocation records where this connector was defined in the source CPL.
	//
	// This location information is essential for error reporting during analysis
	// and synthesis phases. When validation errors occur (e.g., undefined
	// dimension or celltype), this location allows the compiler to report
	// precise error messages pointing to the connector declaration.
	SourceLocation symboltable.SourceLocation
}

// AnalyzeConnectorsScope processes one structured connector declaration from a
// ConnectorScope symbol table and builds the analysis-level connectors lookup map.
//
// This function is called during the symbol table walk for each ConnectorScope
// encountered. It reads connector information from table.Data as a
// *symboltable.ConnectorDeclaration, validates name-based references for the
// first-pass connector shape, and populates the analyzer-wide connectors map
// with fully-resolved definitions.
//
// The parser may still leave generic Identifier or GrainDefinition entries in
// the ConnectorScope's Definitions slice as parse-symbol artifacts, but this
// function does not reconstruct connector roles from that mixed identifier
// sequence. Connector name, dimensions, grain, and celltype are taken only from
// the structured declaration record so a dimension concept and grain measure can
// legally share the same textual name without causing the dimension to be
// skipped.
//
// Symbol Table Structure:
//
//	The parser attaches a *symboltable.ConnectorDeclaration to table.Data before
//	popping the ConnectorScope. That structured record is scoped to one connector
//	declaration and contains:
//	- Name and NameLocation for the connector declaration
//	- DimensionNames and DimensionLocations for dimension concept references
//	- GrainNames and GrainLocations for the mandatory grain measure references
//	- CellTypeName and CellTypeLocation for the required celltype measure
//
//	DimensionNames are resolved against the analyzer's concepts map. GrainNames,
//	when present, are resolved against the analyzer's measures map. CellTypeName is
//	resolved against the analyzer's measures map. The mixed Definitions identifier
//	sequence is not used to infer roles by position or by excluding identifiers
//	whose names match the grain name.
//
// Validation Performed:
//   - Structured data validation: table.Data must contain a *symboltable.ConnectorDeclaration
//   - Duplicate connector detection: Fails if the connector name already exists in the analyzer-wide connectors map
//   - Dimension resolution: Each DimensionNames entry must exist in the concepts map
//   - Celltype resolution: CellTypeName must exist in the measures map
//   - Grain resolution: GrainNames must contain at least one entry and each must exist in the measures map
//
// Grain interval-family type validation is intentionally not performed in this
// first-pass connector analysis. That validation is deferred to a
// post-derivation pass after type derivation paths have been computed.
//
// CIL Tree Generation:
//
// For each valid connector, a one-node CIL tree is created consisting of a
// CILConnectorNode. This tree serves as the root for connector
// references in metrics and aggregators during synthesis.
//
// Parameters:
//   - table: The ConnectorScope symbol table whose Data field contains a structured connector declaration
//   - context: The analysis context (unused for connectors but required for handler signature)
//   - analyzer: The CPLAnalyzer containing the concepts and measures maps for resolution
//
// Returns:
//   - nil if the connector declaration is successfully processed
//   - error if structured data validation or reference validation fails
//
// Example:
//
//	// Connector definition: SalesData = dimensions [Country, Product] celltype Amount grain Day
//	// table.Data: &symboltable.ConnectorDeclaration{
//	//     Name: "SalesData",
//	//     DimensionNames: []string{"Country", "Product"},
//	//     GrainNames: []string{"Day"},
//	//     CellTypeName: "Amount",
//	// }
//	// Result: AnalysisConnectorDefinition{
//	//     Name: "SalesData",
//	//     Dimensions: []*AnalysisConceptDefinition{CountryConcept, ProductConcept},
//	//     CellType: &AmountMeasure,
//	//     Grain: []*AnalysisMeasureDefinition{DayMeasure},
//	// }
//	// CILTrees["SalesData"] = &cil.CILConnectorNode{ConnectorName: "SalesData"}
//
// Error Cases:
//   - "connector scope missing structured connector declaration" - table.Data is nil or not a connector declaration
//   - "duplicate connector definition: <name>" - Connector already defined in the analyzer-wide connectors map
//   - "connector <name> has no grain measure" - GrainNames is empty
//   - "connector <name> references undefined concept dimension: <dimension>"
//   - "connector <name> references undefined celltype measure: <measure>"
//   - "connector <name> references undefined grain measure: <grain>"
//   - Non-interval-family grain measure types are reported by the deferred post-derivation validation pass, not by this function
func AnalyzeConnectorsScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
	// Retrieve structured connector declaration from table.Data
	if table.Data == nil {
		return fmt.Errorf("connector scope has no structured declaration data")
	}

	declaration, ok := table.Data.(*symboltable.ConnectorDeclaration)
	if !ok {
		return fmt.Errorf("connector scope data has unexpected type %T", table.Data)
	}

	// Validate connector name
	if declaration.Name == "" {
		return fmt.Errorf("connector declaration has no name")
	}

	connectorName := declaration.Name
	sourceLocation := declaration.NameLocation

	// Check for duplicate connector definitions
	if analyzer.connectorAnalyzer.HasConnector(connectorName) {
		return fmt.Errorf("duplicate connector definition: %s", connectorName)
	}

	// Resolve celltype measure
	if declaration.CellTypeName == "" {
		return fmt.Errorf("connector %s has no celltype measure", connectorName)
	}

	cellTypeMeasure, exists := analyzer.measures[declaration.CellTypeName]
	if !exists {
		return fmt.Errorf("connector %s references undefined celltype measure: %s",
			connectorName, declaration.CellTypeName)
	}

	// Resolve dimension concepts
	var dimensions []*AnalysisConceptDefinition
	for _, dimName := range declaration.DimensionNames {
		if !analyzer.conceptAnalyzer.HasConcept(dimName) {
			return fmt.Errorf("connector %s references undefined concept dimension: %s",
				connectorName, dimName)
		}
		// Store the concept and take its address
		// We need to store it first because Go doesn't allow taking address of map values directly
		conceptCopy := analyzer.conceptAnalyzer.GetConcept(dimName)
		dimensions = append(dimensions, &conceptCopy)
	}

	// Resolve grain measures (mandatory, at least one)
	if len(declaration.GrainNames) == 0 {
		return fmt.Errorf("connector %s has no grain measure", connectorName)
	}

	var grainMeasures []*AnalysisMeasureDefinition
	for _, grainName := range declaration.GrainNames {
		grain, exists := analyzer.measures[grainName]
		if !exists {
			return fmt.Errorf("connector %s references undefined grain measure: %s",
				connectorName, grainName)
		}
		grainCopy := grain
		grainMeasures = append(grainMeasures, &grainCopy)
	}

	// Grain interval-family type validation is intentionally deferred to a
	// post-derivation pass after ComputeDerivationPaths has been run.
	// Do not validate here.

	// Create analysis-level connector definition
	analysisDef := AnalysisConnectorDefinition{
		Name:           connectorName,
		Dimensions:     dimensions,
		CellType:       &cellTypeMeasure,
		Grain:          grainMeasures,
		SourceLocation: sourceLocation,
	}

	// Store the connector definition
	analyzer.connectorAnalyzer.AddConnector(connectorName, analysisDef)

	// Create a one-node CIL tree for the connector
	analyzer.metricAnalyzer.AddCILTree(connectorName, &cil.CILConnectorNode{ConnectorName: connectorName})

	return nil
}
