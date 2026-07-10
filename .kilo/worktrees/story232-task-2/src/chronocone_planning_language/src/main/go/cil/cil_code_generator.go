// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
	"github.com/chronocone/chronocone_planning_language/src/main/go/connector"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
	"github.com/chronocone/chronocone_planning_language/src/main/go/tensorops"
)

// ConnectorMetadataProvider defines read-only access to connector metadata.
// CILRegistry implements this interface, decoupling CILCodeGenerator from
// the registry's internal map structure so the generator does not hold a
// reference to shared mutable state.
type ConnectorMetadataProvider interface {
	// GetConnectorMetadata retrieves metadata for a single connector by name.
	GetConnectorMetadata(name string) *connector.ConnectorMetadata

	// AllConnectorMetadata returns a map of all registered connectors.
	AllConnectorMetadata() map[string]*connector.ConnectorMetadata
}

// CILCodeGenerator generates Go code from CIL abstract representations.
// It delegates connector metadata access to a ConnectorMetadataProvider
// rather than holding a direct reference to the registry's internal map.
type CILCodeGenerator struct {
	// connectorMetadata provides read-only access to connector metadata,
	// owned by an external CILRegistry.
	connectorMetadata ConnectorMetadataProvider
}

// NewCILCodeGenerator creates a new code generator with the given connector metadata provider.
//
// Parameters:
//   - provider: A ConnectorMetadataProvider for accessing connector metadata.
//     Typically a *CILRegistry.
//
// Returns:
//   - *CILCodeGenerator: A new code generator instance.
func NewCILCodeGenerator(provider ConnectorMetadataProvider) *CILCodeGenerator {
	return &CILCodeGenerator{
		connectorMetadata: provider,
	}
}

// GenerateConnectorMetadataMap generates the Go map declaration for all registered connectors.
// This produces code using package-qualified type names, and emits Grain and GrainToConcept
// fields when present:
//
//	var connectorMetadata = map[string]connector.ConnectorMetadata{
//	    "sales_data": {Name: "sales_data", Dimensions: []string{"Product", "Region"}, CellType: "float64"},
//	}
//
// Returns:
//   - string: The generated Go code for the metadata map.
func (g *CILCodeGenerator) GenerateConnectorMetadataMap() string {
	if g == nil || g.connectorMetadata == nil {
		return "var connectorMetadata = map[string]connector.ConnectorMetadata{}"
	}

	connectors := g.connectorMetadata.AllConnectorMetadata()
	if len(connectors) == 0 {
		return "var connectorMetadata = map[string]connector.ConnectorMetadata{}"
	}

	var sb strings.Builder
	sb.WriteString("var connectorMetadata = map[string]connector.ConnectorMetadata{")

	// Collect keys for deterministic output
	names := make([]string, 0, len(connectors))
	for name := range connectors {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		metadata := connectors[name]
		sb.WriteString(fmt.Sprintf("\n\t%q: connector.ConnectorMetadata{", name))
		sb.WriteString(fmt.Sprintf("Name: %q, ", metadata.Name))
		sb.WriteString("Dimensions: []string{")
		for j, dim := range metadata.Dimensions {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q", dim))
		}
		sb.WriteString("}, ")
		// Emit Grain when present
		if len(metadata.Grain) > 0 {
			sb.WriteString("Grain: []string{")
			for j, gr := range metadata.Grain {
				if j > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%q", gr))
			}
			sb.WriteString("}, ")
		}
		sb.WriteString(fmt.Sprintf("CellType: %q", metadata.CellType))
		if len(metadata.GrainToConcept) > 0 {
			sb.WriteString(", GrainToConcept: map[string]string{")
			firstG2C := true
			// Sort keys for deterministic output
			grainKeys := make([]string, 0, len(metadata.GrainToConcept))
			for gmc := range metadata.GrainToConcept {
				grainKeys = append(grainKeys, gmc)
			}
			sort.Strings(grainKeys)
			for _, gmc := range grainKeys {
				if !firstG2C {
					sb.WriteString(", ")
				}
				firstG2C = false
				sb.WriteString(fmt.Sprintf("%q: %q", gmc, metadata.GrainToConcept[gmc]))
			}
			sb.WriteString("}")
		}
		sb.WriteString("},")
	}

	sb.WriteString("\n}")
	return sb.String()
}

// GenerateSetupFunction generates the Setup() function for initialization.
// Generated code uses package-qualified type names (e.g., cplio.ArchitectureSpecification,
// calculation.NewChronoconeComputationController).
//
// Returns:
//   - string: The generated Setup function code.
func (g *CILCodeGenerator) GenerateSetupFunction() string {
	return `// Setup initializes the computation environment.
func Setup(architecture *cplio.ArchitectureSpecification) error {
	// Initialize computation controller
	controller = calculation.NewChronoconeComputationController()
	if architecture != nil {
		controller.SetArchitectureConfig(architecture)
	}
	return nil
}`
}

// GenerateComputeFunction generates the Compute() function that orchestrates execution.
// The generated function registers all input tensors as metric computations, then launches
// a goroutine for each metric in the provided topological order, waits for all goroutines
// to complete using a sync.WaitGroup, then returns. Each goroutine executes
// computeX(controller, inputs) where X is the metric name.
//
// The dagOrder parameter must be a topologically sorted slice of metric names
// (dependencies before dependents) so that goroutines are launched in the correct
// order for the MetricComputation synchronization model.
//
// Parameters:
//   - dagOrder: The names of the metrics in topological evaluation order.
//
// Returns:
//   - string: The generated Compute function code.
func (g *CILCodeGenerator) GenerateComputeFunction(dagOrder []string) string {
	var sb strings.Builder
	sb.WriteString("// Compute executes all metric computations.\n")
	sb.WriteString("// Registers all input tensors as metric computations, then launches a goroutine\n")
	sb.WriteString("// for each metric in topological order and waits for all to complete.\n")
	sb.WriteString("func Compute() error {\n")
	if len(dagOrder) > 0 {
		sb.WriteString("\t// Register all input tensors as metric computations for postfix evaluation\n")
		sb.WriteString("\tfor name, tensor := range inputs {\n")
		sb.WriteString("\t\tmc := calculation.NewMetricComputation(0)\n")
		sb.WriteString("\t\tif err := controller.RegisterMetricComputation(name, mc); err != nil {\n")
		sb.WriteString("\t\t\treturn err\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t\tif err := mc.SetTensor(tensor); err != nil {\n")
		sb.WriteString("\t\t\treturn err\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n\n")
		sb.WriteString("\tvar wg sync.WaitGroup\n")
		for _, name := range dagOrder {
			sb.WriteString(fmt.Sprintf("\twg.Add(1)\n"))
			sb.WriteString(fmt.Sprintf("\tgo func() {\n"))
			sb.WriteString(fmt.Sprintf("\t\tdefer wg.Done()\n"))
			sb.WriteString(fmt.Sprintf("\t\tcompute%s(controller, inputs)\n", name))
			sb.WriteString(fmt.Sprintf("\t}()\n"))
		}
		sb.WriteString("\twg.Wait()\n")
	}
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n")
	return sb.String()
}

// GenerateCleanupFunction generates the Cleanup() function for resource release.
// Generated code references the controller variable (not a type), so no package
// qualification is needed for type references in this function.
//
// Returns:
//   - string: The generated Cleanup function code.
func (g *CILCodeGenerator) GenerateCleanupFunction() string {
	return `// Cleanup releases all resources.
func Cleanup() error {
	// Stop workers and release resources
	if controller != nil {
		controller.StopWorkers()
	}
	return nil
}`
}

// generateConnectorRegistrationCode generates Go code that iterates over the
// inputs map and registers each loaded tensor as a MetricComputation on the
// controller. This ensures connector inputs are available to the postfix
// evaluator via controller.GetMetricComputation.
//
// Returns:
//   - string: The generated Go code for the registration block.
func (g *CILCodeGenerator) generateConnectorRegistrationCode() string {
	return "\tfor name, tensor := range inputs {\n" +
		"\t\tmc := calculation.NewMetricComputation(0)\n" +
		"\t\tif err := controller.RegisterMetricComputation(name, mc); err != nil {\n" +
		"\t\t\tcheck(err, \"Failed to register metric computation for \" + name)\n" +
		"\t\t}\n" +
		"\t\tif err := mc.SetTensor(tensor); err != nil {\n" +
		"\t\t\tcheck(err, \"Failed to set tensor for \" + name)\n" +
		"\t\t}\n" +
		"\t}\n"
}

// GeneratePostfixTableLiteral formats a slice of PostfixOp entries into a
// Go struct literal string suitable for embedding in generated code. The
// output uses the package-qualified type calculation.PostfixOp and includes
// proper field quoting and standard Go formatting with one entry per line.
//
// Example output:
//
//	[]calculation.PostfixOp{
//	    {Kind: "push", Name: "input_a"},
//	    {Kind: "push", Name: "input_b"},
//	    {Kind: "apply", Op: cil.Subtract},
//	}
//
// Parameters:
//   - ops: The slice of PostfixOp entries to format.
//
// Returns:
//   - string: The generated Go struct literal string.
func (g *CILCodeGenerator) GeneratePostfixTableLiteral(ops []calculation.PostfixOp) string {
	return g.generatePostfixTableLiteralInternal(ops, "\t")
}

// generatePostfixTableLiteralInternal recursively formats a slice of PostfixOp
// entries into a Go struct literal string with the given indentation prefix.
func (g *CILCodeGenerator) generatePostfixTableLiteralInternal(ops []calculation.PostfixOp, indent string) string {
	if len(ops) == 0 {
		return "[]calculation.PostfixOp{}"
	}

	var sb strings.Builder
	sb.WriteString("[]calculation.PostfixOp{\n")

	nextIndent := indent + "\t"
	for _, op := range ops {
		sb.WriteString(indent)
		sb.WriteString("{")
		sb.WriteString(fmt.Sprintf("Kind: %q", op.Kind))
		if op.Name != "" {
			sb.WriteString(fmt.Sprintf(", Name: %q", op.Name))
		} else if op.Value != 0 || op.Kind == "push" || op.Op == operator.Bool || op.Op == operator.Nil {
			sb.WriteString(fmt.Sprintf(", Value: %s", strconv.FormatFloat(op.Value, 'f', -1, 64)))
		}
		if op.Op != operator.Unknown {
			sb.WriteString(fmt.Sprintf(", Op: operator.%s", op.Op.String()))
		}
		if op.Kind == "conditional" && len(op.Branches) > 0 {
			sb.WriteString(fmt.Sprintf(", Branches: %s", g.generatePostfixBranchLiteral(op.Branches, nextIndent)))
		}
		sb.WriteString("},\n")
	}

	sb.WriteString(strings.TrimSuffix(indent, "\t"))
	sb.WriteString("}")
	return sb.String()
}

// generatePostfixBranchLiteral formats a slice of PostfixBranch entries into a
// Go struct literal string suitable for embedding in generated code.
func (g *CILCodeGenerator) generatePostfixBranchLiteral(branches []calculation.PostfixBranch, indent string) string {
	if len(branches) == 0 {
		return "[]calculation.PostfixBranch{}"
	}

	var sb strings.Builder
	sb.WriteString("[]calculation.PostfixBranch{\n")

	nextIndent := indent + "\t"
	for _, branch := range branches {
		sb.WriteString(indent)
		sb.WriteString("{")
		if branch.Condition != nil {
			sb.WriteString(fmt.Sprintf("Condition: %s, ", g.generatePostfixTableLiteralInternal(branch.Condition, nextIndent)))
		}
		sb.WriteString(fmt.Sprintf("Body: %s", g.generatePostfixTableLiteralInternal(branch.Body, nextIndent)))
		sb.WriteString("},\n")
	}

	sb.WriteString(strings.TrimSuffix(indent, "\t"))
	sb.WriteString("}")
	return sb.String()
}

// generatePostfixComputeBody generates the body of a computeX() function
// using the postfix table evaluation approach. It converts the CIL expression
// tree rooted at the metric's CILNode into a flat postfix operation table,
// then generates the Go source code for that table literal, a connector
// registration block that pre-registers each input tensor as a
// MetricComputation, and the corresponding call to
// calculation.EvaluatePostfixFloat64.
//
// The postfix table approach replaces the older snippet-stack visitor pattern,
// enabling correct evaluation of arbitrarily nested binary and unary
// expressions in a stack-based postfix evaluator.
//
// Parameters:
//   - metricName: The name of the metric for which the compute function body
//     is being generated.
//   - metricNode: The CILMetricNode whose CILNode expression tree is to be
//     converted into a postfix table.
//
// Returns:
//   - string: The generated Go code for the compute function body, containing
//     the postfix table literal, connector registration block, and the
//     EvaluatePostfixFloat64 call.
//   - error: An error if the tree cannot be walked to postfix or if the
//     table literal generation fails, or nil on success.
func (g *CILCodeGenerator) generatePostfixComputeBody(metricName string, metricNode *CILMetricNode) (string, error) {
	if g == nil {
		return "", fmt.Errorf("code generator is nil")
	}
	if metricNode == nil {
		return "", fmt.Errorf("metric node is nil")
	}
	if metricNode.CILNode == nil {
		return "", fmt.Errorf("metric node CILNode is nil")
	}

	ops, err := WalkCILTreeToPostfix(metricNode.CILNode)
	if err != nil {
		return "", fmt.Errorf("failed to walk CIL tree to postfix: %w", err)
	}

	opsLiteral := g.GeneratePostfixTableLiteral(ops)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\tops := %s\n", opsLiteral))
	sb.WriteString("\tresult, err := calculation.EvaluatePostfixFloat64(controller, ops)\n")
	sb.WriteString("\tcheck(err, \"Postfix evaluation failed\")\n")
	sb.WriteString(fmt.Sprintf("\tif err := controller.GetMetricComputation(%q).SetTensor(result); err != nil {\n", metricName))
	sb.WriteString(fmt.Sprintf("\t\tcheck(err, \"Failed to set tensor for %s\")\n", metricName))
	sb.WriteString("\t}\n")

	return sb.String(), nil
}

// GenerateAggregationMatrix formats a pre-built AggregationMatrix as a Go
// package-level variable declaration using &tensorops.AggregationMatrix{...}
// struct literal syntax.
func (g *CILCodeGenerator) GenerateAggregationMatrix(name string, matrix *tensorops.AggregationMatrix) string {
	if matrix == nil || matrix.Rows == 0 || matrix.Columns == 0 || len(matrix.Data) == 0 {
		return fmt.Sprintf("var %s = &tensorops.AggregationMatrix{}", name)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("var %s = &tensorops.AggregationMatrix{\n", name))
	sb.WriteString(fmt.Sprintf("\tRows: %d,\n", matrix.Rows))
	sb.WriteString(fmt.Sprintf("\tColumns: %d,\n", matrix.Columns))

	// RowLabels
	sb.WriteString("\tRowLabels: []string{")
	for i, label := range matrix.RowLabels {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%q", label))
	}
	sb.WriteString("},\n")

	// ColLabels
	sb.WriteString("\tColLabels: []string{")
	for i, label := range matrix.ColLabels {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%q", label))
	}
	sb.WriteString("},\n")

	// Data
	sb.WriteString("\tData: [][]float64{\n")
	for _, row := range matrix.Data {
		sb.WriteString("\t\t{")
		for j, val := range row {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(strconv.FormatFloat(val, 'g', -1, 64))
		}
		sb.WriteString("},\n")
	}
	sb.WriteString("\t},\n")
	sb.WriteString("}")

	return sb.String()
}
