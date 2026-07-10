// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file contains the MetricAnalyzer which handles metric-related analysis
// operations including metric registration, CIL tree building, IO spec validation,
// and reference resolution.
package analysis

import (
	"fmt"
	"sort"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	cpio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// MetricAnalyzer handles metric-related analysis operations.
//
// It maintains the metrics map, CIL trees map, and IO specification,
// providing methods for metric registration, IO spec validation, and retrieval.
type MetricAnalyzer struct {
	// metrics is a map from metric name to its analyzed definition.
	// This map is built during analysis by processing metric definitions
	// from the symbol table. It provides O(1) access to metric information
	// for validation and code generation.
	metrics map[string]AnalysisMetricDefinition

	// cilTrees is a map from connector/metric name to its CIL tree root node.
	// This map stores the Compiler Intermediate Language representation for
	// each connector and metric. The CIL tree is built during analysis and
	// transformed into executable code during synthesis.
	cilTrees map[string]cil.CILNode

	// ioSpec holds the validated IO specification for the synthesis phase.
	// It is set via ValidateAndStoreIOSpec after metric declarations have
	// been analyzed, ensuring all output metric references are valid.
	// If nil, no IO spec has been validated for this compilation unit.
	ioSpec *cpio.IOSpecification

	// dag is the metric dependency DAG built as a second pass after all metric
	// definitions are assembled. It captures inter-metric dependencies and is
	// used for topological sorting, cycle detection, and dead-code elimination.
	dag *MetricDAG

	// connectorAnalyzer provides access to the connector definitions map for
	// IO spec validation. Set via SetValidationContext before calling
	// ValidateAndStoreIOSpec. If nil, connector-based validations are skipped.
	connectorAnalyzer *ConnectorAnalyzer

	// measures provides access to the measures map for CellType validation.
	// Set via SetValidationContext before calling ValidateAndStoreIOSpec.
	// If nil, measure-based validations are skipped.
	measures map[string]AnalysisMeasureDefinition
}

// NewMetricAnalyzer creates a new MetricAnalyzer with initialized data structures.
//
// Returns:
//   - A pointer to a new, initialized MetricAnalyzer.
func NewMetricAnalyzer() *MetricAnalyzer {
	return &MetricAnalyzer{
		metrics:  make(map[string]AnalysisMetricDefinition),
		cilTrees: make(map[string]cil.CILNode),
	}
}

// SetValidationContext sets the validation dependencies used by
// ValidateAndStoreIOSpec for connector- and measure-based validations.
//
// Parameters:
//   - connectorAnalyzer: The connector analyzer with the connectors map, or nil.
//   - measures: The measures map, or nil.
func (ma *MetricAnalyzer) SetValidationContext(connectorAnalyzer *ConnectorAnalyzer, measures map[string]AnalysisMeasureDefinition) {
	ma.connectorAnalyzer = connectorAnalyzer
	ma.measures = measures
}

// extractConnectorNamesFromCIL recursively walks a CIL expression tree and
// collects the ConnectorName values from every CILConnectorNode in the tree.
//
// The returned names are de-duplicated. Nil nodes are silently skipped.
//
// Parameters:
//   - node: The root of the CIL expression tree to walk.
//
// Returns:
//   - A slice of unique connector names found in the tree.
func extractConnectorNamesFromCIL(node cil.CILNode) []string {
	if node == nil {
		return nil
	}

	seen := make(map[string]bool)
	var names []string

	var walk func(n cil.CILNode)
	walk = func(n cil.CILNode) {
		if n == nil {
			return
		}
		switch node := n.(type) {
		case *cil.CILConnectorNode:
			if !seen[node.ConnectorName] {
				seen[node.ConnectorName] = true
				names = append(names, node.ConnectorName)
			}
		case *cil.CILBinaryOp:
			walk(node.Left)
			walk(node.Right)
		case *cil.CILUnaryOp:
			walk(node.Operand)
		case *cil.CILComparisonOp:
			walk(node.Left)
			walk(node.Right)
		case *cil.CILLogicalOp:
			walk(node.Left)
			walk(node.Right)
		case *cil.CILNotOp:
			walk(node.Operand)
		case *cil.CILConditional:
			for _, branch := range node.Branches {
				walk(branch.Condition)
				walk(branch.Body)
			}
		case *cil.CILWhenClause:
			for _, clause := range node.Clauses {
				walk(clause.Condition)
				walk(clause.Body)
			}
			walk(node.Default)
		case *cil.CILPower:
			walk(node.Base)
		case *cil.CILAggregateNode:
			walk(node.SourceCIL)
		case *cil.CILLiteral, *cil.CILNumber, *cil.CILInputRef,
			*cil.CILNilLiteral, *cil.CILBoolLiteral, *cil.CILUnitRef,
			*cil.CILLookupNode:
			// Leaf nodes — no children to recurse into.
		default:
			panic(fmt.Sprintf("extractConnectorNamesFromCIL: unhandled CIL node type %T", n))
		}
	}

	walk(node)
	return names
}

// ValidateAndStoreIOSpec validates an IO specification's output metric references,
// dimension/celltype/grain constraints, and input connector mappings against the
// analyzer's declared metrics, connectors, and measures.
//
// This method should be called after all metric declarations have been analyzed
// and SetValidationContext has been called to inject connector and measure maps.
//
// Validations performed:
//   - Nil spec check
//   - Output metric name existence in declared metrics
//   - Output celltype validity against measure declarations
//     (skipped when measures map is nil — backwards-compatible)
//   - Output dimension/grain validity against connector declarations
//     (skipped when connector analyzer is nil or no connectors found in CIL tree)
//   - Input connector name existence in declared connectors
//     (skipped when connector analyzer is nil — backwards-compatible)
//
// Parameters:
//   - ioSpec: The IO specification to validate and store.
//
// Returns:
//   - error: An error if any validation fails, or nil on success.
func (ma *MetricAnalyzer) ValidateAndStoreIOSpec(ioSpec *cpio.IOSpecification) error {
	if ioSpec == nil {
		return fmt.Errorf("IO specification is nil")
	}

	for _, output := range ioSpec.Outputs {
		metricDef, exists := ma.metrics[output.MetricName]
		if !exists {
			return fmt.Errorf("IO spec references unknown metric: %s", output.MetricName)
		}

		// Walk the metric's CIL tree to discover connector names.
		// This is needed for both connector-based validations and for
		// determining whether connector checks are relevant.
		var connectorNames []string
		if metricDef.CILRoot != nil {
			connectorNames = extractConnectorNamesFromCIL(metricDef.CILRoot.CILNode)
		}

		// Validate CellType exists in the measures map.
		// Skipped when measures map is nil (backwards-compatible).
		if output.CellType != "" && ma.measures != nil {
			if _, measureExists := ma.measures[output.CellType]; !measureExists {
				return fmt.Errorf("IO spec output %q references unknown celltype measure: %s",
					output.MetricName, output.CellType)
			}
		}

		// If no connectors were found in the CIL tree, skip connector-based
		// dimension and grain validations.
		if len(connectorNames) == 0 {
			continue
		}

		// Connector-based dimension and grain validations are skipped when the
		// connector analyzer has not been set (backwards-compatible).
		if ma.connectorAnalyzer == nil {
			continue
		}

		// For each connector referenced by the metric, validate dimensions and grain.
		for _, connectorName := range connectorNames {
			connectorDef := ma.connectorAnalyzer.GetConnector(connectorName)
			if connectorDef.Name == "" {
				// Connector not found — skip (could be validated elsewhere).
				continue
			}

			// Build a set of concept names from the connector's dimensions.
			connectorDimNames := make(map[string]bool, len(connectorDef.Dimensions))
			for _, dim := range connectorDef.Dimensions {
				connectorDimNames[dim.ConceptName] = true
			}

			// Validate each output dimension exists in the connector's dimensions.
			for _, outputDim := range output.Dimensions {
				if !connectorDimNames[outputDim] {
					return fmt.Errorf("IO spec output %q dimension %q not declared in connector %q dimensions",
						output.MetricName, outputDim, connectorName)
				}
			}

			// Build a set of grain measure names from the connector's grain.
			connectorGrainNames := make(map[string]bool, len(connectorDef.Grain))
			for _, grain := range connectorDef.Grain {
				connectorGrainNames[grain.Name] = true
			}

			// Validate each output grain exists in the connector's grain.
			for _, outputGrain := range output.Grain {
				if !connectorGrainNames[outputGrain] {
					return fmt.Errorf("IO spec output %q grain %q not declared in connector %q grain",
						output.MetricName, outputGrain, connectorName)
				}
			}
		}
	}

	// Validate input connector names and driver parameter paths.
	if ma.connectorAnalyzer != nil {
		for _, input := range ioSpec.Connectors {
			// Empty connector name with driver parameters is an error.
			if input.ConnectorName == "" && len(input.DriverParameters) > 0 {
				return fmt.Errorf("IO spec input has driver_parameters but no connector name specified")
			}
			if input.ConnectorName != "" && !ma.connectorAnalyzer.HasConnector(input.ConnectorName) {
				return fmt.Errorf("IO spec input references unknown connector: %s", input.ConnectorName)
			}
			// Validate that inputs with driver parameters specify a non-empty path.
			if len(input.DriverParameters) > 0 {
				foundPath := false
				for _, params := range input.DriverParameters {
					if path, ok := params["path"].(string); ok && path != "" {
						foundPath = true
						break
					}
				}
				if !foundPath {
					return fmt.Errorf("IO spec input for connector %q has no valid path in driver_parameters", input.ConnectorName)
				}
			}
		}
	}

	ma.ioSpec = ioSpec
	return nil
}

// GetMetrics returns the metrics map.
func (ma *MetricAnalyzer) GetMetrics() map[string]AnalysisMetricDefinition {
	return ma.metrics
}

// GetMetric returns the metric definition with the given name.
// Returns a zero-value AnalysisMetricDefinition if the metric is not found.
func (ma *MetricAnalyzer) GetMetric(name string) AnalysisMetricDefinition {
	return ma.metrics[name]
}

// HasMetric returns true if a metric with the given name exists.
func (ma *MetricAnalyzer) HasMetric(name string) bool {
	_, exists := ma.metrics[name]
	return exists
}

// GetCILTrees returns the CIL trees map.
func (ma *MetricAnalyzer) GetCILTrees() map[string]cil.CILNode {
	return ma.cilTrees
}

// AddMetric adds a metric definition to the metrics map.
func (ma *MetricAnalyzer) AddMetric(name string, metric AnalysisMetricDefinition) {
	ma.metrics[name] = metric
}

// AddCILTree adds a CIL tree to the cilTrees map.
func (ma *MetricAnalyzer) AddCILTree(name string, tree cil.CILNode) {
	ma.cilTrees[name] = tree
}

// HasCILTree returns true if a CIL tree with the given name exists.
func (ma *MetricAnalyzer) HasCILTree(name string) bool {
	_, exists := ma.cilTrees[name]
	return exists
}

// GetIOSpec returns the validated IO specification stored on the analyzer.
//
// This method provides access to the IO spec for the synthesis phase.
// It returns nil if no IO spec has been validated.
//
// Returns:
//   - *cpio.IOSpecification: The validated IO specification, or nil.
func (ma *MetricAnalyzer) GetIOSpec() *cpio.IOSpecification {
	return ma.ioSpec
}

// GetDAG returns the metric dependency DAG built by BuildMetricDependencyDAG.
//
// Returns:
//   - *MetricDAG: The dependency DAG, or nil if it has not been built.
func (ma *MetricAnalyzer) GetDAG() *MetricDAG {
	return ma.dag
}

// CollectMetricDependencies returns the names of all metric-type dependencies
// from the given metric's CIL tree InputRefs.
//
// This method walks the CILMetricNode.InputRefs slice and filters for entries
// with Kind == "metric". The returned names are unique and ordered by first
// appearance in the InputRefs slice.
//
// Parameters:
//   - metric: The metric definition whose CIL tree InputRefs to inspect.
//
// Returns:
//   - A slice of metric names this metric depends on via CIL InputRefs.
func CollectMetricDependencies(metric *AnalysisMetricDefinition) []string {
	if metric == nil || metric.CILRoot == nil {
		return nil
	}

	seen := make(map[string]bool)
	var deps []string

	for _, ref := range metric.CILRoot.InputRefs {
		if ref.Kind == "metric" && !seen[ref.Name] {
			seen[ref.Name] = true
			deps = append(deps, ref.Name)
		}
	}

	return deps
}

// BuildMetricDependencyDAG constructs the metric dependency DAG as a second pass.
//
// This method should be called after all metric definitions have been assembled.
// It creates a node for every metric, then iterates over each metric's Inputs
// and CIL InputRefs to discover metric-type dependencies. Bidirectional edges
// are added: node.Dependencies gets the target metric name, and the target's
// Dependents gets the current metric name. RootNodes are identified as metrics
// with zero dependencies on other metrics. Unknown metric names are silently
// skipped (treated as non-metric references).
//
// Returns:
//   - error: Always nil; unknown metric names are silently skipped.
func (ma *MetricAnalyzer) BuildMetricDependencyDAG() error {
	dag := &MetricDAG{
		Nodes: make(map[string]*DAGNode),
	}

	// Step 1: Create a node for every metric.
	for name := range ma.metrics {
		dag.Nodes[name] = &DAGNode{
			MetricName: name,
		}
	}

	// Step 2: Add dependency edges from both Inputs and CIL InputRefs.
	for name, metric := range ma.metrics {
		node := dag.Nodes[name]
		seen := make(map[string]bool)

		// Collect from explicit Inputs — only add edge if the target is another metric.
		// Skip time-shifted self-references: these are not same-period dependencies.
		for _, input := range metric.Inputs {
			if input.Kind == "metric" && !seen[input.Name] {
				if _, isMetric := ma.metrics[input.Name]; isMetric {
					// Time-shifted self-references are permitted and should not form
					// edges in the same-period dependency DAG.
					if input.Name == name && input.TimeShift != nil {
						continue
					}
					seen[input.Name] = true
					dag.addDependencyEdge(node, input.Name)
				}
			}
		}

		// Collect from CIL InputRefs — only add edge if the target is another metric.
		// NOTE: Self-reference skipping here relies on hasTimeShiftedSelfReference
		// scanning metric.Inputs, which is consistent with CIL InputRefs via normal
		// parsing. If Inputs and CIL become inconsistent (e.g., a CIL self-reference
		// without a corresponding time-shifted Input), the edge will not be skipped
		// and a cycle will be reported.
		cilDeps := CollectMetricDependencies(&metric)
		for _, depName := range cilDeps {
			if !seen[depName] {
				if _, isMetric := ma.metrics[depName]; isMetric {
					// Skip time-shifted self-references in CIL InputRefs as well.
					if depName == name {
						if hasTimeShiftedSelfReference(metric, depName) {
							continue
						}
					}
					seen[depName] = true
					dag.addDependencyEdge(node, depName)
				}
			}
		}
	}

	// Step 3: Identify root nodes (metrics with zero dependencies).
	for name, node := range dag.Nodes {
		if len(node.Dependencies) == 0 {
			dag.RootNodes = append(dag.RootNodes, name)
		}
	}
	sort.Strings(dag.RootNodes)

	ma.dag = dag
	return nil
}

// addDependencyEdge adds a bidirectional dependency edge between two metric nodes.
//
// The caller guarantees that toName exists in dag.Nodes (it is a known metric).
//
// Parameters:
//   - fromNode: The node that depends on the target metric.
//   - toName: The name of the target metric.
func (dag *MetricDAG) addDependencyEdge(fromNode *DAGNode, toName string) {
	targetNode := dag.Nodes[toName]

	fromNode.Dependencies = append(fromNode.Dependencies, toName)
	targetNode.Dependents = append(targetNode.Dependents, fromNode.MetricName)
}

// EliminateDeadCode removes metrics that are not reachable from the given
// output metric names by traversing the DAG in reverse. Only metrics reachable
// from the IO spec output metrics are retained for synthesis.
//
// This method:
//  1. Returns an error if the DAG has not been built (ma.dag is nil).
//  2. Calls ma.dag.MarkReachable(outputMetricNames) to identify reachable metrics.
//  3. Iterates over ma.metrics and deletes any metric NOT in the reachable set.
//  4. Also removes corresponding CIL tree entries from ma.cilTrees.
//
// Unknown output metric names are silently skipped.
//
// Parameters:
//   - outputMetricNames: The names of output metrics from the IO specification.
//
// Returns:
//   - error: An error if the DAG has not been built, or nil on success.
func (ma *MetricAnalyzer) EliminateDeadCode(outputMetricNames []string) error {
	if ma.dag == nil {
		return fmt.Errorf("cannot eliminate dead code: metric dependency DAG has not been built")
	}

	// Identify all metrics reachable from the output metrics.
	reachable := ma.dag.MarkReachable(outputMetricNames)

	// Delete unreachable metrics from ma.metrics and ma.cilTrees.
	for name := range ma.metrics {
		if !reachable[name] {
			delete(ma.metrics, name)
			delete(ma.cilTrees, name)
		}
	}

	// Rebuild the DAG from surviving metrics. This ensures the DAG is
	// fully consistent: Nodes, Dependencies, Dependents, RootNodes, and
	// topological order all reflect only the surviving metrics.
	return ma.BuildMetricDependencyDAG()
}

// hasTimeShiftedSelfReference checks whether the given metric has an input
// that is a time-shifted self-reference to the specified dependency name.
//
// Parameters:
//   - metric: The metric definition to inspect.
//   - depName: The dependency name to check for a time-shifted self-reference.
//
// Returns:
//   - true if the metric has an input with Kind="metric", Name=depName, and TimeShift != nil.
func hasTimeShiftedSelfReference(metric AnalysisMetricDefinition, depName string) bool {
	for _, input := range metric.Inputs {
		if input.Kind == "metric" && input.Name == depName && input.TimeShift != nil {
			return true
		}
	}
	return false
}
