// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file defines AnalysisDefinitions, the top-level structure produced by
// the analysis phase that carries all analyzed entities into synthesis.
package analysis

import (
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	cpio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// AnalysisDefinitions holds all analysis-phase definitions for a CPL compilation unit.
//
// This is the top-level structure produced by the analysis phase. It contains
// lookup maps for all defined entities (concepts, units, metrics, etc.), enabling
// efficient validation and code generation.
type AnalysisDefinitions struct {
	// Concepts is a lookup map from concept name to its analyzed definition.
	//
	// This map is built during the analysis phase by processing all concept
	// definitions from the symbol table. It provides O(1) access to concept
	// information for validation of concept references throughout the CPL program.
	Concepts map[string]AnalysisConceptDefinition

	// Types is a lookup map from type name to its analyzed definition.
	//
	// This map is built during the analysis phase by processing all type
	// definitions from the symbol table. It provides O(1) access to type
	// information for validation of type references in metrics and other constructs.
	Types map[string]AnalysisTypeDefinition

	// Units is a lookup map from unit name to its analyzed definition.
	//
	// This map is built during the analysis phase by processing all unit
	// definitions from the symbol table. It provides O(1) access to unit
	// information for validation of unit references in metrics and other constructs.
	Units map[string]AnalysisUnitDefinition

	// Measures is a lookup map from measure name to its analyzed definition.
	//
	// This map is built during the analysis phase by processing all measure
	// definitions from the symbol table. It provides O(1) access to measure
	// information for validation of measure references in metrics and other constructs.
	Measures map[string]AnalysisMeasureDefinition

	// Connectors is a lookup map from connector name to its analyzed definition.
	//
	// This map is built during the analysis phase by processing all connector
	// definitions from the symbol table. It provides O(1) access to connector
	// information for validation of connector references in metrics and aggregators.
	// Each connector definition includes its dimensions (concept references) and
	// cell type (measure reference), all resolved and validated during analysis.
	Connectors map[string]AnalysisConnectorDefinition

	// Lookups is a lookup map from lookup name to its analyzed definition.
	// This map is built during the analysis phase by processing all lookup
	// definitions from the symbol table.
	Lookups map[string]*LookupInfo

	// Metrics is a lookup map from metric name to its analyzed definition.
	//
	// This map is built during the analysis phase by processing all metric
	// definitions from the symbol table. It provides O(1) access to metric
	// information for validation of metric references in other metrics and
	// for code generation during the synthesis phase.
	Metrics map[string]AnalysisMetricDefinition

	// Aggregators is a lookup map from compound key to its analyzed definition.
	//
	// This map is built during the analysis phase by processing all aggregator
	// definitions from the symbol table. It provides O(1) access to aggregator
	// information for validation and code generation. The key is produced by
	// AggregatorMapKey(fromConcepts, toConcept).
	Aggregators map[string]*AggregatorInfo

	// Spreaders is a lookup map from compound key to its analyzed definition.
	//
	// This map is built during the analysis phase by processing all spreader
	// definitions from the symbol table. It provides O(1) access to spreader
	// information for validation and code generation. The key is produced by
	// SpreaderMapKey(fromConcepts, toConcepts).
	Spreaders map[string]*SpreaderInfo

	// CILTrees is a map from connector/metric name to its CIL tree root node.
	//
	// This map stores the Compiler Intermediate Language (CIL) representation
	// for each connector and metric defined in the CPL document. The CIL tree
	// is an abstract representation of calculations that is transformed into
	// executable code during the synthesis phase.
	//
	// For connectors, the CIL tree is typically a single CILConnectorNode
	// that serves as the root. For metrics, the tree represents the calculation
	// expression with operations and operands.
	CILTrees map[string]cil.CILNode

	// IOSpec holds the validated IO specification for the synthesis phase.
	//
	// This field is populated when ValidateAndStoreIOSpec is called after
	// metric declarations have been analyzed. It contains the validated
	// input/output configuration that drives code generation during synthesis.
	// If nil, no IO spec has been validated for this compilation unit.
	IOSpec *cpio.IOSpecification

	// DAG is the metric dependency DAG built during the analysis phase.
	//
	// The DAG captures inter-metric dependencies and provides topological
	// evaluation order via the Order field. It is built as a second pass
	// after all metric definitions are assembled, and is pruned by dead-code
	// elimination to only include metrics reachable from IO spec outputs.
	// If nil, the DAG has not been built (e.g., no metrics analyzed).
	DAG *MetricDAG
}
