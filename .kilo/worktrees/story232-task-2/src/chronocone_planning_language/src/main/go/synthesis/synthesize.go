// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package synthesis

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/connector"
	cio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/chronocone/chronocone_planning_language/src/main/go/tensorops"
)

// cplToGoTypeMap maps CPL built-in type names to their Go equivalents.
var cplToGoTypeMap = map[string]string{
	"Number":   "float64",
	"String":   "string",
	"Boolean":  "bool",
	"Instant":  "time.Time",
	"Interval": "time.Duration",
}

// Synthesize generates the Go CLI code from the analyzed CPL program
// and an IO specification. This produces a full computation lifecycle program
// with connector loading, computation execution, and output writing.
//
// Parameters:
//   - analysisDefs: The analysis definitions containing connectors, concepts,
//     and other analyzed CPL constructs.
//   - ioSpec: The IO specification defining inputs, outputs, and architecture.
//   - outputFile: The path to the output file where the generated Go code
//     will be written.
//
// Returns:
//   - error: An error if the output file cannot be created or if code
//     generation fails, nil otherwise.
func Synthesize(analysisDefs *analysis.AnalysisDefinitions, ioSpec *cio.IOSpecification, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	input, err := BuildTransformerFromAnalysis(analysisDefs)
	if err != nil {
		return fmt.Errorf("failed to build transformer from analysis: %w", err)
	}

	if err := validateOutputUnits(input, ioSpec); err != nil {
		return err
	}

	input.IOSpec = ioSpec
	code, err := GenerateMainFunction(input)
	if err != nil {
		return fmt.Errorf("failed to generate main function: %w", err)
	}

	_, err = file.WriteString(code)
	if err != nil {
		return fmt.Errorf("failed to write generated code: %w", err)
	}

	return nil
}

// validateOutputUnits verifies that each output metric's computed unit matches
// the unit declared in the IO spec. It is called during Synthesize after
// BuildTransformerFromAnalysis and before code generation.
//
// Parameters:
//   - input: The synthesis input containing metric definitions and conversion pairs.
//   - ioSpec: The IO specification with output metric unit declarations.
//
// Returns:
//   - error: An error if any metric's computed unit does not match its declared unit,
//     nil otherwise.
func validateOutputUnits(input *SynthesisInput, ioSpec *cio.IOSpecification) error {
	if ioSpec == nil {
		return nil
	}

	combinedPairs := append([]cil.CILConversionPair{}, input.UnitConversionPairs...)
	combinedPairs = append(combinedPairs, input.MeasureConversionPairs...)

	for _, outputSpec := range ioSpec.Outputs {
		metricDef, exists := input.Metrics[outputSpec.MetricName]
		if !exists || metricDef.CILRoot == nil || metricDef.CILRoot.CILNode == nil {
			// If no CIL tree, skip unit checking for this metric
			continue
		}

		if err := cil.CheckMetricOutputUnit(metricDef.CILRoot.CILNode, outputSpec.Unit, combinedPairs); err != nil {
			return fmt.Errorf("metric %q unit check failed: %w", outputSpec.MetricName, err)
		}
	}

	return nil
}

// synthesizeConversionsFromDefs extracts conversions from a map of definitions,
// converts each conversion's Calculation expression to CIL via ExpressionToCILNode,
// synthesizes Go code via SynthesizeExpression, and builds a slice of
// CILConversionPair values.
//
// The helper is generic over the definition type Def. It requires two extractor
// functions: getConversions returns the conversions for a given definition, and
// getName returns the name of a definition for populating SourceName and TargetName.
//
// Parameters:
//   - defs: The map of definition pointers keyed by entity name.
//   - entityLabel: A label used in error messages (e.g. "unit" or "measure").
//   - getConversions: A function that extracts the conversions slice from a definition.
//   - getName: A function that extracts the name from a definition pointer.
//
// Returns:
//   - []cil.CILConversionPair: The built conversion pairs, or nil if there are none.
//   - error: An error if CIL conversion or expression synthesis fails.
func synthesizeConversionsFromDefs[Def any](
	defs map[string]Def,
	entityLabel string,
	getConversions func(Def) []analysis.ConversionPair[Def],
	getName func(*Def) string,
	getGroup func(*Def) string,
) ([]cil.CILConversionPair, error) {
	var pairs []cil.CILConversionPair
	for entityName, def := range defs {
		for _, conv := range getConversions(def) {
			cilNode, err := analysis.ExpressionToCILNode(conv.Calculation)
			if err != nil {
				return nil, fmt.Errorf("failed to convert %s %q conversion to CIL: %w", entityLabel, entityName, err)
			}
			goCode, err := SynthesizeExpression(cilNode)
			if err != nil {
				return nil, fmt.Errorf("failed to synthesize %s %q conversion expression: %w", entityLabel, entityName, err)
			}
			pairs = append(pairs, cil.CILConversionPair{
				SourceName:  getName(conv.SourceDef),
				TargetName:  getName(conv.TargetDef),
				SourceVar:   conv.SourceVar,
				TargetVar:   conv.TargetVar,
				Calculation: goCode,
				SourceGroup: getGroup(conv.SourceDef),
				TargetGroup: getGroup(conv.TargetDef),
			})
		}
	}
	return pairs, nil
}

// BuildTransformerFromAnalysis creates a SynthesisInput from analyzed
// definitions. It walks the connector and measure definitions, registers
// them with a new CILAbstractTransformer, and returns a SynthesisInput
// that serves as the stable contract between analysis and synthesis.
//
// Parameters:
//   - analysisDefs: The analysis definitions produced by the analysis phase.
//
// Returns:
//   - *SynthesisInput: A synthesis input populated with connector
//     and measure type registrations and all analyzed entities.
//   - error: An error if a type cannot be resolved.
func BuildTransformerFromAnalysis(analysisDefs *analysis.AnalysisDefinitions) (*SynthesisInput, error) {
	input := &SynthesisInput{
		Transformer: cil.NewCILAbstractTransformer(),
	}
	if analysisDefs == nil {
		return input, nil
	}

	input.Connectors = analysisDefs.Connectors
	input.Measures = analysisDefs.Measures
	input.Metrics = analysisDefs.Metrics
	input.Concepts = analysisDefs.Concepts
	input.Types = analysisDefs.Types
	input.CILTrees = analysisDefs.CILTrees
	input.DAG = analysisDefs.DAG
	input.Lookups = analysisDefs.Lookups

	// Walk CIL trees and register connectors, lookups, and measure types
	// in the transformer's registry. The tree walker handles type-aware
	// dispatch for each CIL node type and delegates to the CILRegistry.
	walker := NewCILTreeWalker(input.Transformer)
	if err := walker.WalkCILTrees(analysisDefs); err != nil {
		return nil, err
	}

	// Synthesize measure unit expressions
	// TODO(Story230): MeasureUnitExpressions is populated but not yet emitted into
	// generated code. The dependent story (conversions) will consume this map.
	input.MeasureUnitExpressions = make(map[string]string)
	for measureName, measureDef := range analysisDefs.Measures {
		if measureDef.UnitExpression == nil {
			continue
		}
		cilNode, err := analysis.ExpressionToCILNode(measureDef.UnitExpression)
		if err != nil {
			return nil, fmt.Errorf("failed to convert measure %q unit expression to CIL: %w", measureName, err)
		}
		goCode, err := SynthesizeExpression(cilNode)
		if err != nil {
			return nil, fmt.Errorf("failed to synthesize measure %q unit expression: %w", measureName, err)
		}
		input.MeasureUnitExpressions[measureName] = goCode
	}

	// Synthesize unit conversions
	unitPairs, err := synthesizeConversionsFromDefs(analysisDefs.Units, "unit",
		func(d analysis.AnalysisUnitDefinition) []analysis.ConversionPair[analysis.AnalysisUnitDefinition] {
			return d.Conversions
		},
		func(d *analysis.AnalysisUnitDefinition) string { return d.Name },
		func(d *analysis.AnalysisUnitDefinition) string { return d.ConceptName },
	)
	if err != nil {
		return nil, err
	}
	input.UnitConversionPairs = unitPairs
	if len(unitPairs) > 0 {
		input.UnitConversionCode = cil.SynthesizeConversions(unitPairs, "unitConversions", "convertUnit", "unit")
	}

	// Synthesize measure conversions
	measurePairs, err := synthesizeConversionsFromDefs(analysisDefs.Measures, "measure",
		func(d analysis.AnalysisMeasureDefinition) []analysis.ConversionPair[analysis.AnalysisMeasureDefinition] {
			return d.Conversions
		},
		func(d *analysis.AnalysisMeasureDefinition) string { return d.Name },
		func(d *analysis.AnalysisMeasureDefinition) string { return "" },
	)
	if err != nil {
		return nil, err
	}
	input.MeasureConversionPairs = measurePairs
	if len(measurePairs) > 0 {
		input.MeasureConversionCode = cil.SynthesizeConversions(measurePairs, "measureConversions", "convertMeasure", "measure")
	}

	return input, nil
}

// resolveGoType maps a CPL type name to its Go equivalent by walking the
// type derivation chain to find the built-in root type.
//
// Parameters:
//   - typeName: The CPL type name to resolve.
//   - types: The map of analyzed type definitions.
//
// Returns:
//   - string: The Go type string (e.g., "float64" for Number-derived types).
//   - error: An error if the type is unknown or unresolvable.
func resolveGoType(typeName string, types map[string]analysis.AnalysisTypeDefinition) (string, error) {
	// If the type name is a built-in type, return its Go mapping directly.
	if goType, ok := cplToGoTypeMap[typeName]; ok {
		return goType, nil
	}

	// Look up the type in the types map.
	typeDef, exists := types[typeName]
	if !exists {
		return "", fmt.Errorf("unknown type: %s", typeName)
	}

	// Validate derivation path.
	if len(typeDef.DerivationPath) == 0 {
		return "", fmt.Errorf("type %s has empty derivation path", typeName)
	}

	// Check for cycles within this type's own DerivationPath slice.
	// Note: this does NOT detect cross-type cycles (e.g. type A derives
	// from B, type B derives from A) because the types map stores a flat
	// path per type rather than a linked graph. Cross-type cycles would
	// need to be caught during analysis when DerivationPath is built.
	seen := make(map[string]bool)
	for _, pathEntry := range typeDef.DerivationPath {
		if seen[pathEntry] {
			return "", fmt.Errorf("circular derivation path detected for type %s", typeName)
		}
		seen[pathEntry] = true
	}

	// The last element should be a built-in type.
	rootName := typeDef.DerivationPath[len(typeDef.DerivationPath)-1]
	goType, ok := cplToGoTypeMap[rootName]
	if !ok {
		return "", fmt.Errorf("type %s derivation path does not resolve to a built-in type, terminal: %s", typeName, rootName)
	}

	return goType, nil
}

// ConditionalConceptEntry holds the parent concept name and the conditional
// parent→child value mappings for a single conditional concept.
type ConditionalConceptEntry struct {
	ParentConcept string
	Values        map[string][]string
}

// Logger is a minimal logging interface used by synthesis functions.
type Logger interface {
	Printf(format string, v ...interface{})
}

// BuildConditionalConceptValuesMap extracts the conditional parent→child value
// mappings from each concept in the synthesis input and returns them as a
// map from concept name to ConditionalConceptEntry. This map is used during
// synthesis so that mask computation knows the correct parent dimension.
//
// Only concepts with non-empty ConditionalValues and a non-nil ParentConceptName
// are included in the result. Concepts without conditionals or without a parent
// concept name are omitted. When a concept has conditional values but no parent
// concept name, a warning is logged via the provided logger (if non-nil).
//
// Parameters:
//   - logger: An optional logger for emitting warnings. If nil, warnings are
//     silently dropped.
//
// Returns:
//   - map[string]ConditionalConceptEntry: A map where each key is a concept name and
//     each value is a ConditionalConceptEntry containing the parent concept and
//     the parent→child value mappings.
func (input *SynthesisInput) BuildConditionalConceptValuesMap(logger Logger) map[string]ConditionalConceptEntry {
	result := make(map[string]ConditionalConceptEntry)
	if input == nil {
		return result
	}
	for name, concept := range input.Concepts {
		if !concept.HasConditionalValues() {
			continue
		}
		if concept.ParentConceptName == nil {
			if logger != nil {
				logger.Printf("warning: concept %q has conditional values but no parent concept name; skipping", name)
			}
			continue
		}
		// Deep copy the conditional values map so the caller cannot mutate
		// the analysis definitions' internal state.
		copied := make(map[string][]string, len(concept.ConditionalValues))
		for parentValue, childValues := range concept.ConditionalValues {
			childCopy := make([]string, len(childValues))
			copy(childCopy, childValues)
			copied[parentValue] = childCopy
		}
		result[name] = ConditionalConceptEntry{
			ParentConcept: *concept.ParentConceptName,
			Values:        copied,
		}
	}
	return result
}

// BuildConceptValuesMap extracts the resolved string values from each concept in the
// synthesis input and returns them as a map from concept name to value slice.
// This map is embedded into synthesized code so that runtime connector loaders can
// compute tensor shapes and map parquet row values to concept indices without
// re-parsing CPL source.
//
// Grain-name duplication contract: When a connector declares grain measures,
// the measure's Name typically mirrors the underlying concept's values (via
// UnitName). BuildConceptValuesMap duplicates the concept's resolved values
// under the grain measure's Name so that runtime connector loaders can look
// up grain measure columns by the measure name used in the connector
// declaration. This duplication exists because the loader receives a flat
// conceptValues map and needs to resolve grain column names (e.g., "Daily")
// to their valid value sets without re-parsing the CPL source. The
// UnitName→concept relationship is established during analysis, and
// preserving it here avoids coupling the loader to AnalysisDefinitions.
// This addresses the P3 technical-debt item identified in Issue 17.
//
// Parameters:
//   - input: The synthesis input containing concept definitions with
//     ResolvedValues populated.
//
// Returns:
//   - map[string][]string: A map where each key is a concept name and each value
//     is the ordered list of resolved string values for that concept.
func (input *SynthesisInput) BuildConceptValuesMap() map[string][]string {
	result := make(map[string][]string)
	if input == nil {
		return result
	}
	for name, concept := range input.Concepts {
		values := make([]string, 0, len(concept.ResolvedValues))
		for _, v := range concept.ResolvedValues {
			values = append(values, v.Name)
		}
		if len(values) > 0 {
			result[name] = values
		}
	}

	// Walk connector grain measures and add entries for grain measure names.
	// When a grain measure's UnitName references a concept that has resolved
	// values, duplicate those values under the measure name so that runtime
	// connector loaders can look up grain measure columns by name.
	for _, connectorDef := range input.Connectors {
		for _, g := range connectorDef.Grain {
			if g.GetUnitName() != "" {
				if vals, ok := result[g.GetUnitName()]; ok {
					copied := make([]string, len(vals))
					copy(copied, vals)
					result[g.Name] = copied
				}
			}
		}
	}

	return result
}

// ComputeValidityMask computes a bit-packed validity mask for a tensor.
// Bit 0 = valid, bit 1 = invalid. Nil return = all cells valid.
// dimNames: ordered dimension concept names
// conditionalConceptValues: parent→child value mappings with parent concept metadata
// conceptValues: flat value lists for shape computation
func ComputeValidityMask(dimNames []string, conditionalConceptValues map[string]ConditionalConceptEntry, conceptValues map[string][]string) []uint64 {
	if len(dimNames) == 0 {
		return nil
	}

	// Compute total cells and dimension sizes
	totalCells := 1
	dimSizes := make([]int, len(dimNames))
	for i, dim := range dimNames {
		vals := conceptValues[dim]
		if len(vals) == 0 {
			return nil
		}
		dimSizes[i] = len(vals)
		totalCells *= len(vals)
	}
	if totalCells == 0 {
		return nil
	}

	// Pre-build conditional lookup sets for O(1) contains checks
	condSets := make(map[string]map[string]map[string]struct{})
	parentDims := make(map[string]string)
	for _, dim := range dimNames {
		if entry, ok := conditionalConceptValues[dim]; ok && len(entry.Values) > 0 {
			setMap := make(map[string]map[string]struct{}, len(entry.Values))
			for parentValue, childValues := range entry.Values {
				childSet := make(map[string]struct{}, len(childValues))
				for _, cv := range childValues {
					childSet[cv] = struct{}{}
				}
				setMap[parentValue] = childSet
			}
			condSets[dim] = setMap
			parentDims[dim] = entry.ParentConcept
		}
	}

	// If no dimensions have conditionals, all cells are valid
	hasAnyConditionals := false
	for _, dim := range dimNames {
		if _, ok := condSets[dim]; ok {
			hasAnyConditionals = true
			break
		}
	}
	if !hasAnyConditionals {
		return nil
	}

	mask := make([]uint64, (totalCells+63)/64)
	anyInvalid := false

	// Precompute strides for flat→multi-dimensional index conversion
	strides := make([]int, len(dimNames))
	stride := 1
	for i := len(dimNames) - 1; i >= 0; i-- {
		strides[i] = stride
		stride *= dimSizes[i]
	}

	// Precompute dim name → index for O(1) parent dimension lookups
	dimIndex := make(map[string]int, len(dimNames))
	for i, name := range dimNames {
		dimIndex[name] = i
	}

	indices := make([]int, len(dimNames))

	for cell := 0; cell < totalCells; cell++ {
		remainder := cell
		for i := 0; i < len(dimNames); i++ {
			indices[i] = remainder / strides[i]
			remainder %= strides[i]
		}

		invalid := false
		for j, dim := range dimNames {
			childCondSet, hasCond := condSets[dim]
			if !hasCond {
				continue
			}
			childVal := conceptValues[dim][indices[j]]
			validForThisChild := false

			parentDim := parentDims[dim]
			if parentIdx, ok := dimIndex[parentDim]; ok {
				parentVal := conceptValues[parentDim][indices[parentIdx]]
				if childSet, ok := childCondSet[parentVal]; ok {
					if _, contains := childSet[childVal]; contains {
						validForThisChild = true
					}
				}
			}

			if !validForThisChild {
				invalid = true
				break
			}
		}

		if invalid {
			word := cell >> 6
			bit := uint(cell & 63)
			mask[word] |= 1 << bit
			anyInvalid = true
		}
	}

	if anyInvalid {
		// Set trailing bits beyond totalCells to 1 (invalid) so that any
		// word-level iteration over the mask does not treat these as valid.
		if totalCells%64 != 0 {
			mask[len(mask)-1] |= ^uint64(0) << uint(totalCells%64)
		}
		return mask
	}
	return nil
}

// BuildConnectorValidityMasks computes validity masks for all connectors that
// have grain dimensions and conditional concept hierarchies.
//
// For each connector with grain measures, it collects the grain dimension names
// and calls ComputeValidityMask. Non-nil masks are stored in the result map
// keyed by connector name.
//
// Every grain measure must have a non-empty UnitName; otherwise an error is
// returned. This is required so that ComputeValidityMask can correctly resolve
// grain measure names to the concepts referenced by
// BuildConditionalConceptValuesMap.
//
// Parameters:
//   - conditionalConceptValues: parent→child value mappings from BuildConditionalConceptValuesMap.
//   - conceptValues: flat value lists for each concept.
//
// Returns:
//   - map[string][]uint64: A map from connector name to its validity mask slice.
//   - error: An error if any grain measure has no UnitName.
func (input *SynthesisInput) BuildConnectorValidityMasks(conditionalConceptValues map[string]ConditionalConceptEntry, conceptValues map[string][]string) (map[string][]uint64, error) {
	result := make(map[string][]uint64)
	if input == nil {
		return result, nil
	}
	for _, connectorDef := range input.Connectors {
		if len(connectorDef.Grain) == 0 {
			continue
		}
		grainNames := make([]string, 0, len(connectorDef.Grain))
		for _, g := range connectorDef.Grain {
			if g.GetUnitName() == "" {
				return nil, fmt.Errorf("grain measure %q has no UnitName specified", g.Name)
			}
			grainNames = append(grainNames, g.GetUnitName())
		}
		mask := ComputeValidityMask(grainNames, conditionalConceptValues, conceptValues)
		if mask != nil {
			result[connectorDef.Name] = mask
		}
	}
	return result, nil
}

// GenerateMainFunction generates the main() function for the synthesized program.
// This function orchestrates the complete computation lifecycle using package-qualified
// type names (e.g., calculation.ChronoconeComputationController, cplio.ArchitectureSpecification):
// 1. Load IO specification
// 2. Initialize computation controller
// 3. Load input tensors from connectors
// 4. Execute computations
// 5. Write outputs
// 6. Cleanup
//
// The implementation builds a GeneratedProgram struct as the intermediate
// representation, separating program construction from rendering. Sub-component
// code from transformer and connector packages is stored as raw source in
// the program's RawSourceBeforeMain slice.
//
// Parameters:
//   - input: The synthesis input containing all analyzed definitions, the
//     CIL transformer, and the IO specification.
//
// Returns:
//   - string: The generated main function code.
//   - error: An error if generation fails.
func GenerateMainFunction(input *SynthesisInput) (string, error) {
	if input == nil {
		return "", fmt.Errorf("input is nil")
	}
	if input.IOSpec == nil {
		return "", fmt.Errorf("ioSpec is nil")
	}
	if input.Transformer == nil {
		return "", fmt.Errorf("transformer is nil")
	}

	ioSpec := input.IOSpec
	transformer := input.Transformer

	// Build topological evaluation order from DAG.
	// If DAG is available, use its topological order filtered to metrics
	// that exist in input.Metrics and have CIL trees. Otherwise fall back
	// to sorted output metric names for deterministic output.
	metricNames := buildDAGOrder(input, collectMetricNamesFromOutputs(ioSpec))

	// Determine whether additional imports are needed.
	needsSync := len(metricNames) > 0
	needsConnectors := len(ioSpec.Connectors) > 0 || len(ioSpec.Lookups) > 0
	if input.Transformer != nil && len(input.Transformer.AllLookupMetadata()) > 0 {
		needsConnectors = true
	}
	needsOperator := false
	if len(metricNames) > 0 {
		for _, metricName := range metricNames {
			metricDef, exists := input.Metrics[metricName]
			if !exists || metricDef.CILRoot == nil || metricDef.CILRoot.CILNode == nil {
				continue
			}
			if cil.HasNonLiteralCILNode(metricDef.CILRoot.CILNode) {
				needsOperator = true
				break
			}
		}
	}
	needsTensorops := false
	for key := range input.CILTrees {
		if strings.HasPrefix(key, "aggregator:") {
			needsTensorops = true
			break
		}
	}

	// Phase 1: Build the GeneratedProgram struct as the intermediate representation.
	program := &GeneratedProgram{
		PackageName: "main",
		Imports:     buildMainImports(needsSync, needsConnectors, needsOperator, needsTensorops),
	}

	// Emit the check() helper function for consistent error handling in
	// generated programs. This function is always emitted so that every
	// error handling site in the generated code uses check(err, "msg")
	// instead of scattered inline fmt.Fprintf(os.Stderr, ...); os.Exit(1)
	// blocks. The function name "check" is chosen for brevity but should
	// not conflict with any user-defined identifier since CPL programs
	// do not define their own functions.
	program.Functions = append(program.Functions, FunctionDecl{
		Name:    "check",
		Params:  "err error, msg string",
		Returns: "",
		Body:    "\tif err != nil {\n\t\tfmt.Fprintf(os.Stderr, \"%s: %v\\n\", msg, err)\n\t\tos.Exit(1)\n\t}\n",
		Comment: "check exits the program with a non-zero status if err is non-nil, printing the message to stderr.",
	})

	// Build concept values map from synthesis input
	conceptValuesMap := input.BuildConceptValuesMap()
	if len(conceptValuesMap) > 0 {
		var valueSb strings.Builder
		valueSb.WriteString("map[string][]string{\n")
		keys := make([]string, 0, len(conceptValuesMap))
		for k := range conceptValuesMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			values := conceptValuesMap[name]
			valueSb.WriteString(fmt.Sprintf("\t%q: {", name))
			for i, v := range values {
				if i > 0 {
					valueSb.WriteString(", ")
				}
				valueSb.WriteString(fmt.Sprintf("%q", v))
			}
			valueSb.WriteString("},\n")
		}
		valueSb.WriteString("}")
		program.Variables = append(program.Variables, VariableDecl{
			Name:    "conceptValues",
			Value:   valueSb.String(),
			Comment: "conceptValues maps concept names to their resolved values.\n// This is used by connector loaders to compute tensor shapes and map data rows.",
		})
	}

	// Build aggregation matrices from aggregator CIL trees.
	// These must appear before metric compute functions in the preamble.
	if needsTensorops {
		aggregatorKeys := make([]string, 0)
		for key := range input.CILTrees {
			if strings.HasPrefix(key, "aggregator:") {
				aggregatorKeys = append(aggregatorKeys, key)
			}
		}
		sort.Strings(aggregatorKeys)
		var aggregationMatrixSources []string
		for _, key := range aggregatorKeys {
			aggNode, ok := input.CILTrees[key].(*cil.CILAggregateNode)
			if !ok || aggNode == nil {
				continue
			}
			matrix := &tensorops.AggregationMatrix{
				Rows:      aggNode.AggregationMatrix.Rows,
				Columns:   aggNode.AggregationMatrix.Columns,
				Data:      aggNode.AggregationMatrix.Data,
				RowLabels: aggNode.AggregationMatrix.RowLabels,
				ColLabels: aggNode.AggregationMatrix.ColLabels,
			}
			// Derive variable name: aggregationMatrix_<collapsedDimName>
			collapsedDimName := cil.GetCollapsedDimensionName(aggNode.FromDimensions, aggNode.ToDimension)
			if collapsedDimName == "" {
				return "", fmt.Errorf("could not derive aggregation matrix variable name for aggregator %q: empty collapsed dimension", key)
			}
			variableName := fmt.Sprintf("aggregationMatrix_%s", collapsedDimName)
			matrixCode := transformer.GenerateAggregationMatrix(variableName, matrix)
			aggregationMatrixSources = append(aggregationMatrixSources, matrixCode+"\n\n")
		}
		// Prepend so matrices appear before compute/cleanup/connector code.
		program.RawSourceBeforeMain = append(aggregationMatrixSources, program.RawSourceBeforeMain...)
	}

	// Build metric names variable if there are any.
	if len(metricNames) > 0 {
		var nameSb strings.Builder
		nameSb.WriteString("[]string{\n")
		for _, name := range metricNames {
			nameSb.WriteString(fmt.Sprintf("\t%q,\n", name))
		}
		nameSb.WriteString("}")
		program.Variables = append(program.Variables, VariableDecl{
			Name:  "metricNames",
			Value: nameSb.String(),
		})
	}

	// Build consumer counts slice matching metricNames order.
	// consumerCount = DAG dependents + 1 if metric is an output.
	if len(metricNames) > 0 {
		var ccSb strings.Builder
		ccSb.WriteString("[]int32{")
		for _, metricName := range metricNames {
			consumerCount := 0
			if input.DAG != nil && input.DAG.Nodes != nil {
				if node, ok := input.DAG.Nodes[metricName]; ok {
					consumerCount = len(node.Dependents)
				}
			}
			for _, out := range ioSpec.Outputs {
				if out.MetricName == metricName {
					consumerCount++
					break
				}
			}
			if metricName != metricNames[0] {
				ccSb.WriteString(", ")
			}
			ccSb.WriteString(fmt.Sprintf("%d", consumerCount))
		}
		ccSb.WriteString("}")
		program.Variables = append(program.Variables, VariableDecl{
			Name:    "consumerCounts",
			Value:   ccSb.String(),
			Comment: "consumerCounts holds the expected consumer count for each metric in metricNames.",
		})
	}

	// Controller and inputs variables
	program.Variables = append(program.Variables, VariableDecl{
		Name: "controller",
		Type: "*calculation.ChronoconeComputationController",
	})
	program.Variables = append(program.Variables, VariableDecl{
		Name: "inputs",
		Type: "map[string]*calculation.ChronoconeTensor[any]",
	})

	// Build Setup function body
	var setupSb strings.Builder
	setupSb.WriteString("\tcontroller = calculation.NewChronoconeComputationController()\n")
	setupSb.WriteString("\t// Populate concept values on the controller for tensor allocation\n")
	setupSb.WriteString("\tcontroller.SetConceptValues(conceptValues)\n")
	if len(metricNames) > 0 {
		setupSb.WriteString("\t// Register metric computations with correct consumer counts\n")
		setupSb.WriteString("\tfor i, metricName := range metricNames {\n")
		setupSb.WriteString("\t\tmc := calculation.NewMetricComputation(int32(consumerCounts[i]))\n")
		setupSb.WriteString("\t\tif err := controller.RegisterMetricComputation(metricName, mc); err != nil {\n")
		setupSb.WriteString("\t\t\treturn fmt.Errorf(\"failed to register metric %s: %w\", metricName, err)\n")
		setupSb.WriteString("\t\t}\n")
		setupSb.WriteString("\t}\n")
	}
	setupSb.WriteString("\tif architecture != nil {\n")
	setupSb.WriteString("\t\tcontroller.SetArchitectureConfig(architecture)\n")
	setupSb.WriteString("\t}\n")
	setupSb.WriteString("\treturn nil\n")
	program.Functions = append(program.Functions, FunctionDecl{
		Name:    "Setup",
		Params:  "architecture *cplio.ArchitectureSpecification",
		Returns: "error",
		Body:    setupSb.String(),
		Comment: "Setup initializes the computation environment.",
	})

	// Phase 2: Build sub-component code (transformer and connector generators).
	// These produce complete Go source strings that are stored as raw source
	// and emitted after the structured variables and Setup function.

	// Compute function from transformer — launched in DAG topological order.
	program.RawSourceBeforeMain = append(program.RawSourceBeforeMain,
		transformer.GenerateComputeFunction(metricNames)+"\n\n")

	// Cleanup function from transformer
	program.RawSourceBeforeMain = append(program.RawSourceBeforeMain,
		transformer.GenerateCleanupFunction()+"\n\n")

	// Connector metadata map from transformer
	program.RawSourceBeforeMain = append(program.RawSourceBeforeMain,
		transformer.GenerateConnectorMetadataMap()+"\n\n")

	// Build conditional concept values map for mask computation.
	conditionalConceptValuesMap := input.BuildConditionalConceptValuesMap(log.Default())

	// Build connector validity masks for all connectors.
	var connectorValidityMasks map[string][]uint64
	var err error
	connectorValidityMasks, err = input.BuildConnectorValidityMasks(conditionalConceptValuesMap, conceptValuesMap)
	if err != nil {
		return "", fmt.Errorf("failed to build connector validity masks: %w", err)
	}
	if len(connectorValidityMasks) > 0 {
		var maskSb strings.Builder
		maskSb.WriteString("// connectorValidityMasks maps connector names to precomputed bit-packed validity masks.\n")
		maskSb.WriteString("// Bit 0 = valid, bit 1 = invalid. Each mask covers the connector's grain dimensions.\n")
		maskSb.WriteString("var connectorValidityMasks = map[string][]uint64{\n")
		maskConnectorNames := make([]string, 0, len(connectorValidityMasks))
		for k := range connectorValidityMasks {
			maskConnectorNames = append(maskConnectorNames, k)
		}
		sort.Strings(maskConnectorNames)
		for _, name := range maskConnectorNames {
			mask := connectorValidityMasks[name]
			maskSb.WriteString(fmt.Sprintf("\t%q: {", name))
			for i, word := range mask {
				if i > 0 {
					maskSb.WriteString(", ")
				}
				maskSb.WriteString(fmt.Sprintf("0x%X", word))
			}
			maskSb.WriteString("},\n")
		}
		maskSb.WriteString("}\n\n")
		program.RawSourceBeforeMain = append(program.RawSourceBeforeMain, maskSb.String())
	}

	// Emit unit conversion tables and helper function when present.
	if input.UnitConversionCode != "" {
		program.RawSourceBeforeMain = append(program.RawSourceBeforeMain,
			input.UnitConversionCode+"\n")
	}

	// Emit measure conversion tables and helper function when present.
	if input.MeasureConversionCode != "" {
		program.RawSourceBeforeMain = append(program.RawSourceBeforeMain,
			input.MeasureConversionCode+"\n")
	}

	// Connector loader functions.
	loadersCode, err := connector.GenerateConnectorLoaders(transformer, ioSpec, len(connectorValidityMasks) > 0)
	if err != nil {
		return "", err
	}
	program.RawSourceBeforeMain = append(program.RawSourceBeforeMain, loadersCode+"\n")

	// Lookup loading functions
	lookupCode, err := generateLookupIntegration(input)
	if err != nil {
		return "", fmt.Errorf("failed to generate lookup integration: %w", err)
	}
	if lookupCode != "" {
		program.RawSourceBeforeMain = append(program.RawSourceBeforeMain, lookupCode+"\n")
	}

	// Metric compute functions by walking each metric in DAG topological order.
	if len(metricNames) > 0 {
		for _, metricName := range metricNames {
			metricDef, exists := input.Metrics[metricName]
			if !exists {
				continue
			}
			if metricDef.CILRoot == nil {
				continue
			}
			computeFuncCode, err := transformer.TransformMetric(metricName, metricDef.CILRoot)
			if err != nil {
				return "", fmt.Errorf("failed to generate compute function for metric %q: %w", metricName, err)
			}
			program.RawSourceBeforeMain = append(program.RawSourceBeforeMain, computeFuncCode+"\n")
		}
	}

	// Phase 3: Build the main function.
	var mainSb strings.Builder
	mainSb.WriteString("func main() {\n")
	if needsConnectors {
		mainSb.WriteString("\tdataDir := flag.String(\"data-dir\", \"\", \"Directory containing input parquet files\")\n")
		mainSb.WriteString("\tflag.Parse()\n")
	}
	mainSb.WriteString("\tdefer Cleanup()\n\n")
	mainSb.WriteString("\t// Configure architecture\n")
	mainSb.WriteString("\tarchitecture := cplio.ArchitectureSpecification{\n")
	mainSb.WriteString(fmt.Sprintf("\t\tCPUCoreCount: %d,\n", ioSpec.Architecture.CPUCoreCount))
	mainSb.WriteString(fmt.Sprintf("\t\tCacheSizeBytes: %d,\n", ioSpec.Architecture.CacheSizeBytes))
	mainSb.WriteString(fmt.Sprintf("\t\tMainMemoryBytes: %d,\n", ioSpec.Architecture.MainMemoryBytes))
	mainSb.WriteString("\t}\n\n")
	mainSb.WriteString("\t// Initialize the computation environment\n")
	mainSb.WriteString("\tif err := Setup(&architecture); err != nil {\n")
	mainSb.WriteString("\t\tcheck(err, \"Setup failed\")\n")
	mainSb.WriteString("\t}\n\n")

	// Initialize the inputs map before loading connectors
	mainSb.WriteString("\tinputs = make(map[string]*calculation.ChronoconeTensor[any])\n\n")

	// Generate connector loading code.
	inputCode := generateConnectorLoadingCode(ioSpec)
	if inputCode != "" {
		mainSb.WriteString("\t// Load input data from connectors\n")
		mainSb.WriteString(inputCode)
		mainSb.WriteString("\n")
	}

	// Generate lookup loading call if lookup code was generated
	if lookupCode != "" {
		mainSb.WriteString("\t// Load lookup data\n")
		mainSb.WriteString("\tif err := loadLookupData(controller, dataDir); err != nil {\n")
		mainSb.WriteString("\t\tcheck(err, \"Failed to load lookup data\")\n")
		mainSb.WriteString("\t}\n\n")
	}

	// Generate computation call
	mainSb.WriteString("\t// Execute all computations\n")
	mainSb.WriteString("\tif err := Compute(); err != nil {\n")
	mainSb.WriteString("\t\tcheck(err, \"Computation failed\")\n")
	mainSb.WriteString("\t}\n")
	mainSb.WriteString("\n")

	// Generate output writing code
	outputCode := generateOutputWritingCode(ioSpec)
	if outputCode != "" {
		mainSb.WriteString("\t// Write output results\n")
		mainSb.WriteString(outputCode)
		mainSb.WriteString("\n")
	}

	mainSb.WriteString("\tfmt.Println(\"Computation completed successfully\")\n")
	mainSb.WriteString("}\n")

	// Phase 4: Render the GeneratedProgram and append the main function.
	result := program.RenderMainSource()
	result += mainSb.String()

	return result, nil
}

// generateConnectorLoadingCode generates the code block that loads all input tensors.
// This iterates through the IO spec connectors and generates loader dispatch code
// using package-qualified types (e.g., cplio.ConnectorSpecification).
//
// For each connector, the generated code includes:
//   - A cplio.ConnectorSpecification literal with ConnectorName, DataFormat, Driver,
//     and DriverParameters (a []map[string]interface{} literal when non-empty).
//   - A loader function call that passes both the connector spec and the package-level
//     conceptValues map (e.g., loadConnectorName(connectorSpec, conceptValues)).
//
// The loaded tensor is stored in the inputs map using the connector name as the key
// (e.g., inputs["revenue_amt"] = revenue_amtTensor). This is required because
// CIL trees and the postfix evaluator reference connectors by their connector name,
// not the CellType measure name.
//
// DriverParameters values are formatted according to their Go type: strings are
// quoted with %q, numeric values use %d/%g, and booleans use %t. Map iteration
// keys are sorted to produce deterministic output.
//
// Parameters:
//   - ioSpec: The IO specification with input connectors.
//
// Returns:
//   - string: The generated connector loading code.
func generateConnectorLoadingCode(ioSpec *cio.IOSpecification) string {
	if ioSpec == nil || len(ioSpec.Connectors) == 0 {
		return ""
	}

	var sb strings.Builder

	for _, connSpec := range ioSpec.Connectors {
		loaderFunc := connector.GenerateConnectorLoaderFunctionName(connSpec.ConnectorName)
		sb.WriteString(fmt.Sprintf("\t// Load connector %q\n", connSpec.ConnectorName))

		// Pre-resolve any "path" driver parameters so that absolute paths are
		// preserved and relative paths are joined with the -data-dir flag value.
		resolvedPaths := make(map[string]string)
		if len(connSpec.DriverParameters) > 0 {
			for _, params := range connSpec.DriverParameters {
				for key, value := range params {
					if key == "path" {
						if pathStr, ok := value.(string); ok {
							varName := fmt.Sprintf("%sPath", connSpec.ConnectorName)
							sb.WriteString(fmt.Sprintf("\t%s := %q\n", varName, pathStr))
							sb.WriteString(fmt.Sprintf("\tif !filepath.IsAbs(%s) {\n", varName))
							sb.WriteString(fmt.Sprintf("\t\t%s = filepath.Join(*dataDir, %s)\n", varName, varName))
							sb.WriteString("\t}\n")
							resolvedPaths[pathStr] = varName
						}
					}
				}
			}
		}

		sb.WriteString(fmt.Sprintf("\t%sConnectorSpec := cplio.ConnectorSpecification{\n", connSpec.ConnectorName))
		sb.WriteString(fmt.Sprintf("\t\tConnectorName: %q,\n", connSpec.ConnectorName))
		sb.WriteString(fmt.Sprintf("\t\tDataFormat: %q,\n", connSpec.DataFormat))
		sb.WriteString(fmt.Sprintf("\t\tDriver: %q,\n", connSpec.Driver))
		if len(connSpec.DriverParameters) > 0 {
			sb.WriteString("\t\tDriverParameters: []map[string]interface{}{\n")
			for _, params := range connSpec.DriverParameters {
				sb.WriteString("\t\t\t{\n")
				keys := make([]string, 0, len(params))
				for k := range params {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, key := range keys {
					value := params[key]
					switch v := value.(type) {
					case string:
						if key == "path" {
							if varName, ok := resolvedPaths[v]; ok {
								sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %s,\n", key, varName))
							} else {
								sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %q,\n", key, v))
							}
						} else {
							sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %q,\n", key, v))
						}
					case int:
						sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %d,\n", key, v))
					case int8:
						sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %d,\n", key, v))
					case int16:
						sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %d,\n", key, v))
					case int32:
						sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %d,\n", key, v))
					case int64:
						sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %d,\n", key, v))
					case float32:
						sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %g,\n", key, v))
					case float64:
						sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %g,\n", key, v))
					case bool:
						sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %t,\n", key, v))
					default:
						sb.WriteString(fmt.Sprintf("\t\t\t\t%q: %q,\n", key, fmt.Sprintf("%v", v)))
					}
				}
				sb.WriteString("\t\t\t},\n")
			}
			sb.WriteString("\t\t},\n")
		}
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\t%sTensor, err := %s(%sConnectorSpec, conceptValues)\n", connSpec.ConnectorName, loaderFunc, connSpec.ConnectorName))
		sb.WriteString(fmt.Sprintf("\tif err != nil {\n"))
		sb.WriteString(fmt.Sprintf("\t\tcheck(err, \"Failed to load %s\")\n", connSpec.ConnectorName))
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tinputs[%q] = %sTensor\n", connSpec.ConnectorName, connSpec.ConnectorName))
	}

	return sb.String()
}

// generateOutputWritingCode generates the code block that writes each output tensor
// to its own parquet file. For each output metric, it calls cplio.WriteMetricsToParquet
// with a single-element slice so the "Metric" column is preserved.
//
// Parameters:
//   - ioSpec: The IO specification with output metrics.
//
// Returns:
//   - string: The generated output writing code.
func generateOutputWritingCode(ioSpec *cio.IOSpecification) string {
	if ioSpec == nil || len(ioSpec.Outputs) == 0 {
		return ""
	}

	var sb strings.Builder

	for _, output := range ioSpec.Outputs {
		sb.WriteString(fmt.Sprintf("\t// Collect and write output metric %q\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t%sComp := controller.GetMetricComputation(%q)\n", output.MetricName, output.MetricName))
		sb.WriteString(fmt.Sprintf("\t%sTensor := %sComp.GetTensor()\n", output.MetricName, output.MetricName))
		sb.WriteString(fmt.Sprintf("\t%sRaw := %sTensor.Data()\n", output.MetricName, output.MetricName))
		sb.WriteString(fmt.Sprintf("\t%sData := make([]float64, len(%sRaw))\n", output.MetricName, output.MetricName))
		sb.WriteString(fmt.Sprintf("\tfor i, v := range %sRaw {\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\tfv, ok := v.(float64)\n"))
		sb.WriteString(fmt.Sprintf("\t\tif !ok {\n"))
		sb.WriteString(fmt.Sprintf("\t\t\tcheck(fmt.Errorf(\"unexpected type in tensor for %s: %%T\", v), \"Type assertion failed\")\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\t}\n"))
		sb.WriteString(fmt.Sprintf("\t\t%sData[i] = fv\n", output.MetricName))
		sb.WriteString("\t}\n")

		// 3a/3c. Generate validation code for output metric dimensions and grain from IO spec
		sb.WriteString(fmt.Sprintf("\t// Verify dimensions and grain match IO spec for %q\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t%sExpectedDims := []string{", output.MetricName))
		for i, dim := range output.Dimensions {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q", dim))
		}
		sb.WriteString("}\n")
		if len(output.Grain) > 0 {
			sb.WriteString(fmt.Sprintf("\t%sExpectedGrain := []string{", output.MetricName))
			for i, g := range output.Grain {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%q", g))
			}
			sb.WriteString("}\n")
		} else {
			sb.WriteString(fmt.Sprintf("\t%sExpectedGrain := []string(nil)\n", output.MetricName))
		}
		sb.WriteString(fmt.Sprintf("\t%sActualDims := %sTensor.DimensionNames()\n", output.MetricName, output.MetricName))
		sb.WriteString(fmt.Sprintf("\tif len(%sExpectedDims)+len(%sExpectedGrain) != len(%sActualDims) {\n", output.MetricName, output.MetricName, output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\tcheck(fmt.Errorf(\"total dimension count mismatch for %s: expected %%d, got %%d\", len(%sExpectedDims)+len(%sExpectedGrain), len(%sActualDims)), \"dimension mismatch\")\n", output.MetricName, output.MetricName, output.MetricName, output.MetricName))
		sb.WriteString(fmt.Sprintf("\t}\n"))
		sb.WriteString(fmt.Sprintf("\tfor i, expected := range %sExpectedDims {\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\tif expected != %sActualDims[i] {\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\t\tcheck(fmt.Errorf(\"dimension name mismatch for %s at index %%d: expected %%q, got %%q\", i, expected, %sActualDims[i]), \"dimension name mismatch\")\n", output.MetricName, output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\t}\n"))
		sb.WriteString(fmt.Sprintf("\t}\n"))
		sb.WriteString(fmt.Sprintf("\tfor i, expected := range %sExpectedGrain {\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\tif expected != %sActualDims[len(%sExpectedDims)+i] {\n", output.MetricName, output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\t\tcheck(fmt.Errorf(\"grain name mismatch for %s at index %%d: expected %%q, got %%q\", i, expected, %sActualDims[len(%sExpectedDims)+i]), \"grain mismatch\")\n", output.MetricName, output.MetricName, output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\t}\n"))
		sb.WriteString(fmt.Sprintf("\t}\n"))

		// 3b. Generate validation code / comment for output celltype from IO spec.
		//
		// KNOWN LIMITATION: All tensor data is currently normalized to float64 regardless
		// of the IO spec celltype. The parquet Value column is always Float64. Handling
		// heterogeneous cell types (int64, bool, etc.) in both tensor storage and parquet
		// schema generation is deferred to future stories.
		sb.WriteString(fmt.Sprintf("\t// IO spec celltype for %q is %q; tensor data converted to float64 (known limitation)\n", output.MetricName, output.CellType))

		// Grain parameter for MetricTensorData
		var grainParam string
		if len(output.Grain) > 0 {
			grainParam = "[]string{"
			for i, g := range output.Grain {
				if i > 0 {
					grainParam += ", "
				}
				grainParam += fmt.Sprintf("%q", g)
			}
			grainParam += "}"
			sb.WriteString(fmt.Sprintf("\t// IO spec grain for %q is %s\n", output.MetricName, grainParam))
		} else {
			grainParam = "nil"
			sb.WriteString(fmt.Sprintf("\t// IO spec grain for %q is nil\n", output.MetricName))
		}

		// Write this metric to its own parquet file using WriteMetricsToParquet with a single-element slice.
		sb.WriteString(fmt.Sprintf("\tif err := cplio.WriteMetricsToParquet(%q, []cplio.MetricTensorData{\n", output.MetricName+".parquet"))
		sb.WriteString(fmt.Sprintf("\t\t{\n"))
		sb.WriteString(fmt.Sprintf("\t\t\tMetricName: %q,\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\t\tData: %sData,\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\t\tDimensionNames: %sTensor.DimensionNames(),\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\t\tShape: %sTensor.Shape(),\n", output.MetricName))
		sb.WriteString(fmt.Sprintf("\t\t\tGrain: %s,\n", grainParam))
		sb.WriteString(fmt.Sprintf("\t\t},\n"))
		sb.WriteString(fmt.Sprintf("\t}, conceptValues); err != nil {\n"))
		sb.WriteString(fmt.Sprintf("\t\tcheck(err, \"Failed to write output metric %s\")\n", output.MetricName))
		sb.WriteString("\t}\n")
	}

	return sb.String()
}

// collectMetricNamesFromOutputs collects metric names from the IO specification's
// Outputs slice. Each output in the IO spec has a MetricName field that identifies
// the metric to be computed. These names are used during synthesis to register
// MetricComputation objects in the Setup function and to pass to the Compute
// function generation.
//
// Parameters:
//   - ioSpec: The IO specification containing output metric definitions.
//
// Returns:
//   - []string: The collected metric names, or nil if ioSpec is nil or has no outputs.
func collectMetricNamesFromOutputs(ioSpec *cio.IOSpecification) []string {
	if ioSpec == nil || len(ioSpec.Outputs) == 0 {
		return nil
	}
	names := make([]string, 0, len(ioSpec.Outputs))
	for _, output := range ioSpec.Outputs {
		names = append(names, output.MetricName)
	}
	return names
}

// walkCILForConnectors recursively walks a CIL expression tree and collects
// connector names referenced by CILConnectorNode and CILInputRef nodes whose
// Kind is "connector".
func walkCILForConnectors(node cil.CILNode, needed map[string]bool) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *cil.CILConnectorNode:
		needed[n.ConnectorName] = true
	case *cil.CILInputRef:
		if n.Kind == "connector" {
			needed[n.Name] = true
		}
	case *cil.CILBinaryOp:
		walkCILForConnectors(n.Left, needed)
		walkCILForConnectors(n.Right, needed)
	case *cil.CILUnaryOp:
		walkCILForConnectors(n.Operand, needed)
	case *cil.CILComparisonOp:
		walkCILForConnectors(n.Left, needed)
		walkCILForConnectors(n.Right, needed)
	case *cil.CILLogicalOp:
		walkCILForConnectors(n.Left, needed)
		walkCILForConnectors(n.Right, needed)
	case *cil.CILNotOp:
		walkCILForConnectors(n.Operand, needed)
	case *cil.CILConditional:
		for _, branch := range n.Branches {
			walkCILForConnectors(branch.Condition, needed)
			walkCILForConnectors(branch.Body, needed)
		}
	case *cil.CILWhenClause:
		for _, clause := range n.Clauses {
			walkCILForConnectors(clause.Condition, needed)
			walkCILForConnectors(clause.Body, needed)
		}
		walkCILForConnectors(n.Default, needed)
	case *cil.CILPower:
		walkCILForConnectors(n.Base, needed)
	case *cil.CILAggregateNode:
		walkCILForConnectors(n.SourceCIL, needed)
	case *cil.CILLiteral, *cil.CILNumber, *cil.CILNilLiteral, *cil.CILBoolLiteral, *cil.CILUnitRef:
		// Leaf nodes — nothing to do.
	case *cil.CILLookupNode:
		// Lookup nodes do not reference connectors.
	}
}

// collectConnectorsForReachableMetrics walks the CIL trees of all metrics in
// dagOrder and returns the set of connector names they reference. A connector
// is considered needed if any reachable metric's CIL tree contains a
// CILConnectorNode or a CILInputRef whose Kind is "connector".
//
// Parameters:
//   - input: The synthesis input containing metric definitions.
//   - dagOrder: The topological order of reachable metrics (from buildDAGOrder).
//
// Returns:
//   - map[string]bool: A set of connector names needed by reachable metrics.
func collectConnectorsForReachableMetrics(input *SynthesisInput, dagOrder []string) map[string]bool {
	needed := make(map[string]bool)
	if input == nil || len(dagOrder) == 0 {
		return needed
	}

	for _, metricName := range dagOrder {
		metricDef, exists := input.Metrics[metricName]
		if !exists || metricDef.CILRoot == nil {
			continue
		}

		walkCILForConnectors(metricDef.CILRoot.CILNode, needed)

		for _, ref := range metricDef.CILRoot.InputRefs {
			if ref.Kind == "connector" {
				needed[ref.Name] = true
			}
		}
	}

	return needed
}

// buildDAGOrder derives the topological evaluation order from the synthesis input.
// If a DAG is present, it returns DAG.Order filtered to metrics that exist in
// input.Metrics and have CIL trees. If no DAG is present, it falls back to
// sorted metricNames for deterministic output.
//
// Parameters:
//   - input: The synthesis input containing the DAG and metric definitions.
//   - metricNames: Fallback metric names when no DAG is available.
//
// Returns:
//   - []string: Metric names in topological order, or sorted metricNames if no DAG.
func buildDAGOrder(input *SynthesisInput, metricNames []string) []string {
	if input != nil && input.DAG != nil && len(input.DAG.Order) > 0 {
		order := make([]string, 0, len(input.DAG.Order))
		for _, name := range input.DAG.Order {
			metricDef, exists := input.Metrics[name]
			if !exists || metricDef.CILRoot == nil {
				continue
			}
			order = append(order, name)
		}
		return order
	}
	// Fallback: sort metricNames for deterministic output.
	if len(metricNames) > 0 {
		sorted := make([]string, len(metricNames))
		copy(sorted, metricNames)
		sort.Strings(sorted)
		return sorted
	}
	return nil
}

// buildMainImports builds the import declarations for the main program in the
// exact order required for byte-identical output with the pre-refactoring
// import generation. The order is:
//  1. "flag" (conditional on needsInputs)
//  2. "fmt"
//  3. "os"
//  4. "path/filepath" (conditional on needsInputs)
//  5. "sync" (conditional on needsSync)
//  6. calculation package
//  7. connector package
//  8. operator package (conditional on needsOperator)
//  9. cplio (aliased import)
//
// This ordering groups standard library imports before third-party imports,
// which matches the pre-refactoring import generation output and Go
// conventions.
//
// Parameters:
//   - needsSync: Whether sync.WaitGroup is needed (true when there are metric outputs).
//   - needsInputs: Whether flag and path/filepath are needed (true when there are inputs).
//   - needsOperator: Whether the operator package is needed (true when a metric has
//     non-literal CIL expressions).
//   - needsTensorops: Whether the tensorops package is needed (true when aggregators
//     are present in the CIL trees).
//
// Returns:
//   - []ImportDecl: Import declarations in the order described above.
func buildMainImports(needsSync bool, needsInputs bool, needsOperator bool, needsTensorops bool) []ImportDecl {
	calcPath := "github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
	connPath := "github.com/chronocone/chronocone_planning_language/src/main/go/connector"
	opPath := "github.com/chronocone/chronocone_planning_language/src/main/go/operator"
	tensoropsPath := "github.com/chronocone/chronocone_planning_language/src/main/go/tensorops"
	ioPath := "github.com/chronocone/chronocone_planning_language/src/main/go/io"

	imports := make([]ImportDecl, 0, 10)

	if needsInputs {
		imports = append(imports, ImportDecl{Path: "flag"})
	}
	imports = append(imports,
		ImportDecl{Path: "fmt"},
		ImportDecl{Path: "os"},
	)
	if needsInputs {
		imports = append(imports, ImportDecl{Path: "path/filepath"})
	}
	if needsSync {
		imports = append(imports, ImportDecl{Path: "sync"})
	}
	imports = append(imports,
		ImportDecl{Path: calcPath},
		ImportDecl{Path: connPath},
	)
	if needsOperator {
		imports = append(imports, ImportDecl{Path: opPath})
	}
	if needsTensorops {
		imports = append(imports, ImportDecl{Path: tensoropsPath})
	}
	imports = append(imports,
		ImportDecl{Alias: "cplio", Path: ioPath},
	)

	return imports
}
