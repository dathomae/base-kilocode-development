// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file defines SpreaderFunctionInfo and SpreaderFunctionRegistry, which map
// spreader function names to their CIL node types and dimensional unit effects.
// It also defines SpreaderInfo and the core spreader analysis methods.
package analysis

import (
	"fmt"
	"sort"
	"strings"

	cil "github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// SpreaderFunctionInfo describes a function available in spreader expressions.
type SpreaderFunctionInfo struct {
	Name        string
	CILNodeType string
	Arity       int
	DimUnits    func(params []string) *DimensionalUnitExpression
	Description string
}

// SpreaderFunctionRegistry provides lookups for spreader expression evaluation.
type SpreaderFunctionRegistry struct {
	Functions map[string]*SpreaderFunctionInfo
}

// NewSpreaderFunctionRegistry creates a SpreaderFunctionRegistry with the predefined
// spreader functions. All six spreader functions are registered from the start.
func NewSpreaderFunctionRegistry() *SpreaderFunctionRegistry {
	reg := &SpreaderFunctionRegistry{
		Functions: make(map[string]*SpreaderFunctionInfo),
	}

	reg.Functions["size"] = &SpreaderFunctionInfo{
		Name:        "size",
		CILNodeType: "CILSpreadSizeNode",
		Arity:       1,
		DimUnits: func(params []string) *DimensionalUnitExpression {
			if len(params) != 1 {
				return nil
			}
			return &DimensionalUnitExpression{}
		},
		Description: "Number of distinct values in concept",
	}

	reg.Functions["1/size"] = &SpreaderFunctionInfo{
		Name:        "1/size",
		CILNodeType: "CILSpreadInverseSizeNode",
		Arity:       1,
		DimUnits: func(params []string) *DimensionalUnitExpression {
			if len(params) != 1 {
				return nil
			}
			return &DimensionalUnitExpression{}
		},
		Description: "1/N where N is number of values in concept",
	}

	reg.Functions["sum"] = &SpreaderFunctionInfo{
		Name:        "sum",
		CILNodeType: "CILSpreadSumNode",
		Arity:       1,
		DimUnits: func(params []string) *DimensionalUnitExpression {
			if len(params) != 1 {
				return nil
			}
			return &DimensionalUnitExpression{
				Numerator:   []string{params[0]},
				Denominator: []string{},
			}
		},
		Description: "Sums values across a dimension (implicit aggs)",
	}

	reg.Functions["count"] = &SpreaderFunctionInfo{
		Name:        "count",
		CILNodeType: "CILSpreadCountNode",
		Arity:       1,
		DimUnits: func(params []string) *DimensionalUnitExpression {
			if len(params) != 1 {
				return nil
			}
			return &DimensionalUnitExpression{
				Numerator:   []string{params[0]},
				Denominator: []string{},
			}
		},
		Description: "Counts values in a dimension",
	}

	reg.Functions["1/sum"] = &SpreaderFunctionInfo{
		Name:        "1/sum",
		CILNodeType: "CILSpreadInverseSumNode",
		Arity:       1,
		DimUnits: func(params []string) *DimensionalUnitExpression {
			if len(params) != 1 {
				return nil
			}
			return &DimensionalUnitExpression{
				Numerator:   []string{},
				Denominator: []string{params[0]},
			}
		},
		Description: "Inverse of sum across a dimension",
	}

	reg.Functions["1/count"] = &SpreaderFunctionInfo{
		Name:        "1/count",
		CILNodeType: "CILSpreadInverseCountNode",
		Arity:       1,
		DimUnits: func(params []string) *DimensionalUnitExpression {
			if len(params) != 1 {
				return nil
			}
			return &DimensionalUnitExpression{
				Numerator:   []string{},
				Denominator: []string{params[0]},
			}
		},
		Description: "Inverse of count across a dimension",
	}

	return reg
}

// SpreaderInfo holds the analyzed definition of a single spreader declaration.
type SpreaderInfo struct {
	FromConcepts         []*AnalysisConceptDefinition
	ToConcepts           []*AnalysisConceptDefinition
	FromEmpty            bool
	NewDimension         *AnalysisConceptDefinition
	FromDimensionalUnits *DimensionalUnitExpression
	ToDimensionalUnits   *DimensionalUnitExpression
	ExpressionDimUnits   *DimensionalUnitExpression
	SourceLocation       symboltable.SourceLocation
}

// SpreaderMapKey constructs the compound key for the Spreaders map.
// Sorted concept names ensure deterministic keys.
func SpreaderMapKey(fromConcepts, toConcepts []string) string {
	fromSorted := make([]string, len(fromConcepts))
	copy(fromSorted, fromConcepts)
	sort.Strings(fromSorted)
	toSorted := make([]string, len(toConcepts))
	copy(toSorted, toConcepts)
	sort.Strings(toSorted)
	return fmt.Sprintf("[%s]->[%s]", strings.Join(fromSorted, ","), strings.Join(toSorted, ","))
}

// AnalyzeSpreaderScope analyzes a single spreader scope from the symbol table.
//
// Parameters:
//   - scope: The SpreaderScope symbol table whose Definitions contain the
//     spreader declaration and postfix expression.
//   - context: The analysis context (unused for spreaders but required for handler signature).
//   - analyzer: The CPLAnalyzer containing concepts, spreaders map, and function registry.
//
// Returns:
//   - nil if the spreader is successfully processed
//   - error if any validation fails
func AnalyzeSpreaderScope(scope *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
	var pDef *parser.ParserSpreaderDefinition
	var exprDefs []symboltable.SymbolDefinition

	for _, def := range scope.Definitions {
		if def.DefinitionType == symboltable.SpreaderScopeDefinition {
			pDef = def.Data.(*parser.ParserSpreaderDefinition)
		} else {
			exprDefs = append(exprDefs, def)
		}
	}

	if pDef == nil {
		return fmt.Errorf("spreader scope has no SpreaderScopeDefinition")
	}

	// 1. Resolve concept references in from and to
	fromConcepts, err := analyzer.resolveConceptList(pDef.FromConcepts)
	if err != nil {
		return fmt.Errorf("at line %d, column %d: %w", pDef.SourceLocation.Line, pDef.SourceLocation.Column, err)
	}
	toConcepts, err := analyzer.resolveConceptList(pDef.ToConcepts)
	if err != nil {
		return fmt.Errorf("at line %d, column %d: %w", pDef.SourceLocation.Line, pDef.SourceLocation.Column, err)
	}

	// 2. Validate from-to containment and exactly-one-dimension-added
	if err := analyzer.validateSpreaderDimensionalRules(fromConcepts, toConcepts); err != nil {
		return fmt.Errorf("at line %d, column %d: %w", pDef.SourceLocation.Line, pDef.SourceLocation.Column, err)
	}

	// 3. Identify new dimension
	newDim, err := analyzer.findNewDimension(fromConcepts, toConcepts)
	if err != nil {
		return fmt.Errorf("at line %d, column %d: %w", pDef.SourceLocation.Line, pDef.SourceLocation.Column, err)
	}

	// 4. Evaluate postfix expression to compute dimensional unit effect
	exprDimUnits, err := analyzer.evaluateSpreaderExpression(exprDefs, fromConcepts, toConcepts)
	if err != nil {
		return fmt.Errorf("at line %d, column %d: %w", pDef.SourceLocation.Line, pDef.SourceLocation.Column, err)
	}

	// 5. Compute dimensional units and validate compatibility
	fromDimUnits := analyzer.dimUnitsFromConcepts(fromConcepts)
	toDimUnits := analyzer.dimUnitsFromConcepts(toConcepts)

	if err := analyzer.validateSpreaderDimUnits(fromDimUnits, exprDimUnits, toDimUnits); err != nil {
		return fmt.Errorf("at line %d, column %d: %w", pDef.SourceLocation.Line, pDef.SourceLocation.Column, err)
	}

	// 6. Build SpreaderInfo
	info := &SpreaderInfo{
		FromConcepts:         fromConcepts,
		ToConcepts:           toConcepts,
		FromEmpty:            pDef.FromEmpty,
		NewDimension:         newDim,
		FromDimensionalUnits: fromDimUnits,
		ToDimensionalUnits:   toDimUnits,
		ExpressionDimUnits:   exprDimUnits,
		SourceLocation:       pDef.SourceLocation,
	}

	// 7. Store in spreaders map (detect duplicates)
	key := SpreaderMapKey(pDef.FromConcepts, pDef.ToConcepts)
	if existing, exists := analyzer.spreaders[key]; exists {
		return fmt.Errorf("duplicate spreader definition: %s at line %d, column %d (first declared at line %d, column %d)",
			key, pDef.SourceLocation.Line, pDef.SourceLocation.Column, existing.SourceLocation.Line, existing.SourceLocation.Column)
	}
	analyzer.spreaders[key] = info

	// 8. Build spreading matrix and CIL node (Increment 1: equal-proportion weights)
	newDimValues := make([]string, len(info.NewDimension.ResolvedValues))
	for i, v := range info.NewDimension.ResolvedValues {
		newDimValues[i] = v.Name
	}
	// Equal-proportion weights: 1/N for each value
	weights := make(map[string]float64)
	for _, v := range newDimValues {
		weights[v] = 1.0 / float64(len(newDimValues))
	}
	spreadingMatrix := BuildSpreadingMatrix(newDimValues, weights)

	cilNode := &cil.CILSpreadNode{
		SourceCIL:              nil, // no source CIL in Increment 1
		FromDimensions:         sortedNames(info.FromConcepts),
		ToDimensions:           sortedNames(info.ToConcepts),
		NewDimension:           info.NewDimension.ConceptName,
		SpreadingMatrix:        AnalysisMatrixToCIL(spreadingMatrix),
		DistributionExpression: nil, // no distribution expression CIL in Increment 1
		ReferencedMetrics:      nil,
		DimensionalUnitEffect:  info.ExpressionDimUnits.String(),
		SourceLocation:         info.SourceLocation,
	}

	analyzer.metricAnalyzer.AddCILTree("spreader:"+key, cilNode)

	return nil
}

// validateSpreaderDimUnits enforces input_units * expression_units = output_units.
func (a *CPLAnalyzer) validateSpreaderDimUnits(from, expr, to *DimensionalUnitExpression) error {
	predicted := from.Multiply(expr)
	if !predicted.Equals(to) {
		return fmt.Errorf(
			"dimensional unit mismatch: input %s * expression %s = %s, expected %s",
			from.String(), expr.String(), predicted.String(), to.String(),
		)
	}
	return nil
}

// dimUnitsFromConcepts constructs a DimensionalUnitExpression with each concept name
// placed in the Denominator.
func (a *CPLAnalyzer) dimUnitsFromConcepts(concepts []*AnalysisConceptDefinition) *DimensionalUnitExpression {
	names := make([]string, len(concepts))
	for i, c := range concepts {
		names[i] = c.ConceptName
	}
	return &DimensionalUnitExpression{Denominator: names}
}

// resolveConceptList resolves raw concept name strings to AnalysisConceptDefinition pointers.
func (a *CPLAnalyzer) resolveConceptList(names []string) ([]*AnalysisConceptDefinition, error) {
	concepts := make([]*AnalysisConceptDefinition, len(names))
	for i, name := range names {
		concept, exists := a.conceptAnalyzer.GetConcepts()[name]
		if !exists {
			return nil, fmt.Errorf("unknown concept %q in spreader declaration", name)
		}
		concepts[i] = &concept
	}
	return concepts, nil
}

// validateSpreaderDimensionalRules verifies that all from concepts appear in to
// and that exactly one dimension is added.
func (a *CPLAnalyzer) validateSpreaderDimensionalRules(
	fromConcepts []*AnalysisConceptDefinition,
	toConcepts []*AnalysisConceptDefinition,
) error {
	if len(toConcepts) != len(fromConcepts)+1 {
		return fmt.Errorf("spreader must add exactly one dimension: from has %d, to has %d", len(fromConcepts), len(toConcepts))
	}
	for _, from := range fromConcepts {
		found := false
		for _, to := range toConcepts {
			if from.ConceptName == to.ConceptName {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("spreader from concept %q not found in to concepts", from.ConceptName)
		}
	}
	return nil
}

// findNewDimension identifies which concept in to is not in from.
func (a *CPLAnalyzer) findNewDimension(
	fromConcepts []*AnalysisConceptDefinition,
	toConcepts []*AnalysisConceptDefinition,
) (*AnalysisConceptDefinition, error) {
	for _, to := range toConcepts {
		found := false
		for _, from := range fromConcepts {
			if to.ConceptName == from.ConceptName {
				found = true
				break
			}
		}
		if !found {
			return to, nil
		}
	}
	return nil, fmt.Errorf("could not find new dimension in to concepts")
}

// evaluateSpreaderExpression evaluates a postfix sequence of symbol definitions
// from a SpreaderScope into a resolved DimensionalUnitExpression.
func (a *CPLAnalyzer) evaluateSpreaderExpression(
	definitions []symboltable.SymbolDefinition,
	fromConcepts []*AnalysisConceptDefinition,
	toConcepts []*AnalysisConceptDefinition,
) (*DimensionalUnitExpression, error) {
	var stack []*DimensionalUnitExpression

	for _, def := range definitions {
		switch def.DefinitionType {
		case symboltable.Identifier:
			strData, ok := def.Data.(symboltable.StringData)
			if !ok {
				return nil, fmt.Errorf("invalid identifier data in spreader expression")
			}
			name := strData.Value

			fnInfo, exists := a.spreaderFunctionRegistry.Functions[name]
			if exists {
				// Function call
				if len(stack) < fnInfo.Arity {
					return nil, fmt.Errorf("insufficient operands for function %s", name)
				}
				params := make([]string, fnInfo.Arity)
				for i := 0; i < fnInfo.Arity; i++ {
					idx := len(stack) - fnInfo.Arity + i
					expr := stack[idx]
					if len(expr.Numerator) != 1 || len(expr.Denominator) != 0 {
						return nil, fmt.Errorf("invalid params to function %s", name)
					}
					params[i] = expr.Numerator[0]
				}
				stack = stack[:len(stack)-fnInfo.Arity]

				result := fnInfo.DimUnits(params)
				if result == nil {
					return nil, fmt.Errorf("invalid params to function %s", name)
				}
				stack = append(stack, result)
			} else if isValidConceptReference(name, fromConcepts) || isValidConceptReference(name, toConcepts) {
				// Concept reference
				stack = append(stack, &DimensionalUnitExpression{
					Numerator: []string{name},
				})
			} else {
				return nil, fmt.Errorf("unknown identifier %q in spreader expression (Increment 1 supports concepts only)", name)
			}

		case symboltable.UnitReferenceDefinition:
			ref, ok := def.Data.(parser.ParserUnitReference)
			if !ok {
				return nil, fmt.Errorf("invalid unit reference data in spreader expression")
			}
			return nil, fmt.Errorf("unit references not supported in spreader expressions (Increment 1): %q", ref.Name)

		case symboltable.MultiplyOperatorDefinition:
			right, left, err := popTwoDimUnits(&stack)
			if err != nil {
				return nil, err
			}
			stack = append(stack, left.Multiply(right))

		case symboltable.DivideOperatorDefinition:
			right, left, err := popTwoDimUnits(&stack)
			if err != nil {
				return nil, err
			}
			inverted := &DimensionalUnitExpression{
				Numerator:   right.Denominator,
				Denominator: right.Numerator,
			}
			stack = append(stack, left.Multiply(inverted))

		case symboltable.PlusOperatorDefinition, symboltable.MinusOperatorDefinition:
			_, left, err := popTwoDimUnits(&stack)
			if err != nil {
				return nil, err
			}
			// Addition/subtraction doesn't change dimensional units.
			stack = append(stack, left)

		case symboltable.SpreaderScopeDefinition:
			// Skip the synthetic scope definition; it's not part of the expression.
			continue

		default:
			return nil, fmt.Errorf("unexpected definition type in spreader expression: %s", def.DefinitionType)
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("malformed expression: %d items remaining on stack", len(stack))
	}

	return stack[0], nil
}

// BuildSpreadingMatrix constructs an AggregationMatrix for spreading.
// It creates a single-row matrix with one column per new dimension value,
// where each entry is the weight for that value from the weights map.
// Callers must ensure every value in newDimValues has a corresponding entry
// in weights — missing entries default to 0.0.
func BuildSpreadingMatrix(newDimValues []string, weights map[string]float64) *AggregationMatrix {
	cols := len(newDimValues)
	data := make([][]float64, 1)
	data[0] = make([]float64, cols)
	for j, val := range newDimValues {
		data[0][j] = weights[val]
	}
	return &AggregationMatrix{
		Rows:      1,
		Columns:   cols,
		Data:      data,
		RowLabels: []string{"source"},
		ColLabels: newDimValues,
	}
}

// sortedNames extracts and sorts concept names from a slice of definitions.
func sortedNames(concepts []*AnalysisConceptDefinition) []string {
	names := make([]string, len(concepts))
	for i, c := range concepts {
		names[i] = c.ConceptName
	}
	sort.Strings(names)
	return names
}
