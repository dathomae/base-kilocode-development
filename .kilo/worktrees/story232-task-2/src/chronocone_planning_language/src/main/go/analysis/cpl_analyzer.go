// Package analysis provides CPL analysis functionality.
//
// This file contains the Analyze function that bridges the parser and analysis
// phases. It takes a symbol table produced by parsing and produces analysis
// definitions ready for code generation.
package analysis

import (
	"fmt"

	cpio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// CPLAnalyzer orchestrates the analysis phase of CPL compilation.
//
// The analyzer maintains state during the analysis process, collecting
// concept, type, unit, measure, connector, and metric definitions from
// the symbol table. It uses WalkSymbolTableWithHandlers to traverse the
// symbol table tree and dispatch to scope-specific analysis functions.
//
// Sub-analyzers (ConceptAnalyzer, TypeAnalyzer, MetricAnalyzer,
// ConnectorAnalyzer) own their respective maps and provide scoped
// methods. CPLAnalyzer delegates to these sub-components and aggregates
// results via GetAnalysisDefinitions.
//
// Usage:
//
//	analyzer := NewCPLAnalyzer()
//	handlers := map[symboltable.ScopeType]symboltable.AnalysisFunc{
//	    symboltable.ConceptsScope: analyzer.AnalyzeConceptsScope,
//	}
//	err := symboltable.WalkSymbolTableWithHandlers(rootTable, handlers)
//	if err != nil {
//	    return err
//	}
//	defs := analyzer.GetAnalysisDefinitions()
type CPLAnalyzer struct {
	// conceptAnalyzer handles all concept-related analysis operations.
	// It maintains the concepts map and provides methods for concept registration,
	// extension merging, and concept range generator resolution.
	conceptAnalyzer *ConceptAnalyzer

	// typeAnalyzer handles all type-related analysis operations.
	// It maintains the types map and provides methods for type registration,
	// derivation path computation, and constraint validation.
	typeAnalyzer *TypeAnalyzer

	// units is a map from unit name to its analyzed definition.
	// This map is built during analysis by processing unit definitions
	// from the symbol table. It provides O(1) access to unit information
	// for validation and code generation.
	units map[string]AnalysisUnitDefinition

	// measures is a map from measure name to its analyzed definition.
	// This map is built during analysis by processing measure definitions
	// from the symbol table. It provides O(1) access to measure information
	// for validation and code generation.
	measures map[string]AnalysisMeasureDefinition

	// connectorAnalyzer handles all connector-related analysis operations.
	// It maintains the connectors map and provides methods for connector
	// registration and grain type validation.
	connectorAnalyzer *ConnectorAnalyzer

	// lookups is a map from lookup name to its analyzed definition.
	// This map is built during analysis by processing lookup definitions
	// from the symbol table. It provides O(1) access to lookup information
	// for validation and code generation.
	lookups map[string]*LookupInfo

	// metricAnalyzer handles all metric-related analysis operations.
	// It maintains the metrics map, CIL trees map, and IO specification,
	// providing methods for metric registration, IO spec validation, and retrieval.
	metricAnalyzer *MetricAnalyzer

	// aggregators is a map from compound key to analyzed aggregator definition.
	// This map is built during analysis by processing aggregator definitions
	// from the symbol table. It provides O(1) access to aggregator information
	// for validation and code generation.
	aggregators map[string]*AggregatorInfo

	// functionRegistry maps aggregator function names to their predefined
	// definitions. It is used during postfix expression evaluation to resolve
	// fncall operands and compute dimensional unit effects.
	functionRegistry *FunctionRegistry

	// spreaderFunctionRegistry maps spreader function names to their predefined
	// definitions. It is used during spreader expression evaluation to resolve
	// fncall operands and compute dimensional unit effects.
	spreaderFunctionRegistry *SpreaderFunctionRegistry

	// spreaders is a map from compound key to analyzed spreader definition.
	// This map is built during analysis by processing spreader definitions
	// from the symbol table. It provides O(1) access to spreader information
	// for validation and code generation.
	spreaders map[string]*SpreaderInfo
}

// NewCPLAnalyzer creates a new analyzer with initialized data structures.
//
// This constructor prepares an empty analyzer ready to process symbol tables.
// The analyzer's internal maps are initialized but empty - the actual
// population happens during WalkSymbolTableWithHandlers traversal.
//
// Returns:
//   - A pointer to a new, initialized CPLAnalyzer.
//
// Example:
//
//	analyzer := NewCPLAnalyzer()
//	// analyzer is ready to use with WalkSymbolTableWithHandlers
func NewCPLAnalyzer() *CPLAnalyzer {
	return &CPLAnalyzer{
		conceptAnalyzer:          NewConceptAnalyzer(),
		typeAnalyzer:             NewTypeAnalyzer(),
		units:                    make(map[string]AnalysisUnitDefinition),
		measures:                 make(map[string]AnalysisMeasureDefinition),
		connectorAnalyzer:        NewConnectorAnalyzer(),
		lookups:                  make(map[string]*LookupInfo),
		metricAnalyzer:           NewMetricAnalyzer(),
		aggregators:              make(map[string]*AggregatorInfo),
		functionRegistry:         NewAggregatorFunctionRegistry(),
		spreaderFunctionRegistry: NewSpreaderFunctionRegistry(),
		spreaders:                make(map[string]*SpreaderInfo),
	}
}

// GetAnalysisDefinitions returns the completed analysis definitions.
//
// This method should be called after WalkSymbolTableWithHandlers has completed
// successfully. It packages all analyzed entities into an AnalysisDefinitions
// structure suitable for the synthesis phase.
//
// Returns:
//   - A pointer to AnalysisDefinitions containing all analyzed concepts, types, and units.
//   - Never returns nil (always returns a valid struct, even if empty).
//
// Example:
//
//	analyzer := NewCPLAnalyzer()
//	// ... run WalkSymbolTableWithHandlers ...
//	defs := analyzer.GetAnalysisDefinitions()
//	country, exists := defs.Concepts["Country"]
//	currency, exists := defs.Types["Currency"]
//	currencyUnit, exists := defs.Units["Currency"]
func (a *CPLAnalyzer) GetAnalysisDefinitions() *AnalysisDefinitions {
	return &AnalysisDefinitions{
		Concepts:    a.conceptAnalyzer.GetConcepts(),
		Types:       a.typeAnalyzer.GetTypes(),
		Units:       a.units,
		Measures:    a.measures,
		Connectors:  a.connectorAnalyzer.GetConnectors(),
		Lookups:     a.lookups,
		Metrics:     a.metricAnalyzer.GetMetrics(),
		Aggregators: a.aggregators,
		Spreaders:   a.spreaders,
		CILTrees:    a.metricAnalyzer.GetCILTrees(),
		IOSpec:      a.metricAnalyzer.GetIOSpec(),
		DAG:         a.metricAnalyzer.GetDAG(),
	}
}

// ValidateAndStoreIOSpec validates an IO specification's output metric references
// against the analyzer's declared metrics and stores it for synthesis.
//
// Sets the MetricAnalyzer's validation context before delegating.
func (a *CPLAnalyzer) ValidateAndStoreIOSpec(ioSpec *cpio.IOSpecification) error {
	a.metricAnalyzer.SetValidationContext(a.connectorAnalyzer, a.measures)
	return a.metricAnalyzer.ValidateAndStoreIOSpec(ioSpec)
}

// GetIOSpec returns the validated IO specification stored on the analyzer.
//
// Delegates to the MetricAnalyzer.
func (a *CPLAnalyzer) GetIOSpec() *cpio.IOSpecification {
	return a.metricAnalyzer.GetIOSpec()
}

// AddConcept adds a concept definition to the analyzer's concepts map.
// Delegates to the ConceptAnalyzer.
func (a *CPLAnalyzer) AddConcept(concept *AnalysisConceptDefinition) error {
	return a.conceptAnalyzer.AddConcept(concept)
}

// AddOrMergeConcept adds a new concept or merges a concept extension.
// Delegates to the ConceptAnalyzer.
func (a *CPLAnalyzer) AddOrMergeConcept(concept *AnalysisConceptDefinition) error {
	return a.conceptAnalyzer.AddOrMergeConcept(concept)
}

// ResolveConceptRanges processes ConceptRangeGenerators in dependency order.
// Delegates to the ConceptAnalyzer.
func (a *CPLAnalyzer) ResolveConceptRanges() error {
	return a.conceptAnalyzer.ResolveConceptRanges()
}

// AddNewConcept adds a concept definition directly to the concepts map.
// Delegates to the ConceptAnalyzer.
func (a *CPLAnalyzer) AddNewConcept(concept *AnalysisConceptDefinition) {
	a.conceptAnalyzer.AddNewConcept(concept)
}

// GetConcept returns the concept definition with the given name.
// Delegates to the ConceptAnalyzer.
func (a *CPLAnalyzer) GetConcept(Name string) AnalysisConceptDefinition {
	return a.conceptAnalyzer.GetConcept(Name)
}

// GetConcepts returns the concepts map.
// Delegates to the ConceptAnalyzer.
func (a *CPLAnalyzer) GetConcepts() map[string]AnalysisConceptDefinition {
	return a.conceptAnalyzer.GetConcepts()
}

// HasConcept returns true if a concept with the given name exists.
// Delegates to the ConceptAnalyzer.
func (a *CPLAnalyzer) HasConcept(name string) bool {
	return a.conceptAnalyzer.HasConcept(name)
}

// SetConcept updates a concept definition in the concepts map.
// Delegates to the ConceptAnalyzer.
func (a *CPLAnalyzer) SetConcept(name string, concept AnalysisConceptDefinition) {
	a.conceptAnalyzer.SetConcept(name, concept)
}

// ComputeDerivationPaths computes the full derivation path for each type definition.
// Delegates to the TypeAnalyzer.
func (a *CPLAnalyzer) ComputeDerivationPaths() error {
	return a.typeAnalyzer.ComputeDerivationPaths(a.conceptAnalyzer.GetConcepts())
}

// AddType adds a type definition to the analyzer's types map.
// Delegates to the TypeAnalyzer.
func (a *CPLAnalyzer) AddType(typeDef AnalysisTypeDefinition) {
	a.typeAnalyzer.AddType(typeDef)
}

// ValidateConnectorGrainTypes validates connector grain measure types after type derivation paths are computed.
func (a *CPLAnalyzer) ValidateConnectorGrainTypes() error {
	return a.connectorAnalyzer.ValidateGrainTypes(a.typeAnalyzer)
}

// Analyze performs the analysis phase on a parsed CPL document.
//
// This function takes the root symbol table produced by parsing and performs
// semantic analysis to build analysis-level structures. It orchestrates the
// symbol table walk using scope-specific handlers to process different
// declaration types.
//
// Parameters:
//   - rootTable: The root symbol table produced by parsing. This is typically
//     obtained from the listener's symbol stack after parsing completes.
//
// Returns:
//   - A pointer to AnalysisDefinitions containing all analyzed entities.
//   - An error if analysis fails (e.g., duplicate definitions, invalid references).
//
// Example:
//
//	// After parsing
//	listener := NewCPLListener(logger)
//	// ... parse document ...
//	rootTable, _ := listener.GetSymbolStack().Peek()
//	defs, err := Analyze(rootTable)
//	if err != nil {
//	    log.Fatalf("Analysis failed: %v", err)
//	}
//	// defs.Concepts contains all analyzed concepts
//	// defs.Types contains all analyzed types
func Analyze(rootTable *symboltable.SymbolTable) (*AnalysisDefinitions, error) {
	return AnalyzeWithIO(rootTable, nil)
}

// AnalyzeWithIO performs the analysis phase on a parsed CPL document with an
// optional IO specification for dead-code elimination.
//
// This function is identical to Analyze except that it accepts an IO
// specification up-front. When ioSpec is non-nil and has outputs, dead-code
// elimination is performed as part of the analysis pipeline, pruning metrics
// that are not reachable from the specified output metrics.
//
// Parameters:
//   - rootTable: The root symbol table produced by parsing.
//   - ioSpec: An optional IO specification. If nil, dead-code elimination is skipped.
//
// Returns:
//   - A pointer to AnalysisDefinitions containing all analyzed entities.
//   - An error if analysis fails.
func AnalyzeWithIO(rootTable *symboltable.SymbolTable, ioSpec *cpio.IOSpecification) (*AnalysisDefinitions, error) {
	analyzer := NewCPLAnalyzer()

	// Register handlers for each scope type
	// WalkSymbolTableWithHandlers calls the appropriate handler for each scope
	entryHandlers := map[symboltable.ScopeType]AnalysisFunc{
		// Concepts scope handler
		symboltable.ConceptsScope: EnterAnalyzeConceptsScope,
		// Concept scope handlers
		symboltable.ConceptScope:          EnterAnalyzeConceptScope,
		symboltable.ValueGeneratorScope:   AnalyzeValueGeneratorScope,
		symboltable.IntGeneratorScope:     AnalyzeIntGeneratorScope,
		symboltable.ConceptGeneratorScope: AnalyzeConceptRangeGeneratorScope,
		// Type scope handler
		symboltable.TypeScope: EnterAnalyzeTypeScope,
		// Unit scope handler
		symboltable.UnitScope: AnalyzeUnitScope,
		// Measure scope handler
		symboltable.MeasureScope: AnalyzeMeasureScope,
		// Connector scope handler
		symboltable.ConnectorScope: AnalyzeConnectorsScope,
		// Lookups scope handler
		symboltable.LookupsScope: AnalyzeLookupScope,
		// Aggregator scope handler
		symboltable.AggregatorScope: AnalyzeAggregatorScope,
		// Spreader scope handler
		symboltable.SpreaderScope: AnalyzeSpreaderScope,
		// Metrics scope handlers
		symboltable.MetricsScope: AnalyzeMetricsScope,
		symboltable.MetricScope:  AnalyzeMetricScope,
	}

	exitHandlers := map[symboltable.ScopeType]AnalysisFunc{
		// Concept scope handlers
		symboltable.ConceptScope: ExitAnalyzeConceptScope,
	}

	// Single walk with all handlers
	if err := WalkSymbolTableWithHandlers(rootTable, entryHandlers, exitHandlers, analyzer); err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// Analyze unit conversions after all units have been collected.
	// This post-processing step walks the symbol table again to find
	// ConversionScope entries within each UnitScope.
	if err := analyzeAllUnitConversions(rootTable, analyzer); err != nil {
		return nil, fmt.Errorf("unit conversion analysis failed: %w", err)
	}

	// Analyze measure conversions after all measures have been collected.
	// This post-processing step walks the symbol table again to find
	// MeasureConversionScope entries within each MeasureScope.
	if err := analyzeAllMeasureConversions(rootTable, analyzer); err != nil {
		return nil, fmt.Errorf("measure conversion analysis failed: %w", err)
	}

	// Build metric dependency DAG as a second pass after all metric definitions are assembled.
	// This must happen after WalkSymbolTableWithHandlers so that all metrics are registered.
	if err := analyzer.metricAnalyzer.BuildMetricDependencyDAG(); err != nil {
		return nil, fmt.Errorf("metric dependency DAG construction failed: %w", err)
	}

	// Detect cycles in the metric dependency DAG.
	// Time-shifted self-references (e.g., A referencing A with backward(day)) are permitted.
	if err := analyzer.metricAnalyzer.GetDAG().DetectCycles(analyzer.metricAnalyzer.GetMetrics()); err != nil {
		return nil, fmt.Errorf("metric dependency cycle detected: %w", err)
	}

	// Compute topological ordering of metrics for deterministic evaluation.
	if _, err := analyzer.metricAnalyzer.GetDAG().TopologicalSort(); err != nil {
		return nil, fmt.Errorf("metric dependency topological sort failed: %w", err)
	}

	// Validate and store IO spec if provided.
	if ioSpec != nil {
		if err := analyzer.ValidateAndStoreIOSpec(ioSpec); err != nil {
			return nil, fmt.Errorf("IO spec validation failed: %w", err)
		}
	}

	// Eliminate dead code — prune metrics not reachable from IO spec output metrics.
	// This reduces the set of metrics passed to synthesis, improving performance
	// and reducing generated code size.
	if ioSpec := analyzer.metricAnalyzer.GetIOSpec(); ioSpec != nil {
		outputNames := make([]string, len(ioSpec.Outputs))
		for i, output := range ioSpec.Outputs {
			outputNames[i] = output.MetricName
		}
		if err := analyzer.metricAnalyzer.EliminateDeadCode(outputNames); err != nil {
			return nil, fmt.Errorf("dead-code elimination failed: %w", err)
		}

		// Re-compute topological sort after pruning to populate dag.Order
		// with the correct evaluation order for the surviving metrics.
		if _, err := analyzer.metricAnalyzer.GetDAG().TopologicalSort(); err != nil {
			return nil, fmt.Errorf("metric dependency topological sort after pruning failed: %w", err)
		}
	}

	// Resolve ConceptRangeGenerators after all concepts are initially processed.
	// ConceptRangeGenerators reference other concepts which must be fully resolved
	// before values can be generated. Deferred resolution happens in rounds to
	// handle dependency chains, with circular dependency detection.
	if err := analyzer.ResolveConceptRanges(); err != nil {
		return nil, fmt.Errorf("concept range resolution failed: %w", err)
	}

	if err := analyzer.VerifyMeasureUnitReferences(); err != nil {
		return nil, fmt.Errorf("measure unit reference validation failed: %w", err)
	}

	// Compute derivation paths for all types after all type definitions have been collected
	if err := analyzer.ComputeDerivationPaths(); err != nil {
		return nil, fmt.Errorf("failed to compute type derivation paths: %w", err)
	}

	// Validate Interval type cross-references after derivation paths are computed
	if err := ValidateIntervalTypes(analyzer.typeAnalyzer.GetTypes(), analyzer.units); err != nil {
		return nil, fmt.Errorf("interval type validation failed: %w", err)
	}

	// Verify that all concept references in unit declarations exist
	if err := VerifyUnitConceptReferences(analyzer.units, analyzer.conceptAnalyzer.GetConcepts()); err != nil {
		return nil, fmt.Errorf("unit concept validation failed: %w", err)
	}

	// Validate connector grain measure types after derivation paths are computed
	if err := analyzer.ValidateConnectorGrainTypes(); err != nil {
		return nil, fmt.Errorf("connector grain validation failed: %w", err)
	}

	return analyzer.GetAnalysisDefinitions(), nil
}

// analyzeAllUnitConversions walks the symbol table to find all UnitScope entries
// and calls AnalyzeUnitConversions for each one. This must run after the main
// WalkSymbolTableWithHandlers pass has populated the analyzer's units map.
func analyzeAllUnitConversions(rootTable *symboltable.SymbolTable, analyzer *CPLAnalyzer) error {
	var walkErr error
	symboltable.WalkSymbolTable(rootTable, func(table *symboltable.SymbolTable) {
		if walkErr != nil {
			return
		}
		if table.ScopeType != symboltable.UnitScope {
			return
		}
		if len(table.Definitions) < 2 {
			return
		}
		unitName := *table.Definitions[0].Name
		unitDef, exists := analyzer.units[unitName]
		if !exists {
			return
		}
		if err := AnalyzeUnitConversions(table, analyzer.units, &unitDef); err != nil {
			walkErr = err
			return
		}
		analyzer.units[unitName] = unitDef
	})
	return walkErr
}

// analyzeAllMeasureConversions walks the symbol table to find all MeasureScope entries
// and calls AnalyzeMeasureConversions for each one. This must run after the main
// WalkSymbolTableWithHandlers pass has populated the analyzer's measures map.
// Deferred analysis avoids ordering issues where a conversion references a measure
// that hasn't been processed yet during the initial walk.
func analyzeAllMeasureConversions(rootTable *symboltable.SymbolTable, analyzer *CPLAnalyzer) error {
	var walkErr error
	symboltable.WalkSymbolTable(rootTable, func(table *symboltable.SymbolTable) {
		if walkErr != nil {
			return
		}
		if table.ScopeType != symboltable.MeasureScope {
			return
		}
		if len(table.Definitions) < 2 {
			return
		}
		measureName := *table.Definitions[0].Name
		measureDef, exists := analyzer.measures[measureName]
		if !exists {
			return
		}
		if err := analyzer.AnalyzeMeasureConversions(table, &measureDef); err != nil {
			walkErr = err
			return
		}
		analyzer.measures[measureName] = measureDef
	})
	return walkErr
}
