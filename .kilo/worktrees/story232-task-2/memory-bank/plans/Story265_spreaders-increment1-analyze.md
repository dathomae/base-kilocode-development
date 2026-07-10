# Plan: Story265 — Spreaders Increment 1 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story265: **spreaders-increment1-analyze** — implement the analysis phase for spreader declarations (Increment 1 of Epic008). Covers `AnalyzeSpreaderScope`, `SpreaderInfo`, `SpreaderFunctionRegistry` (all functions registered from start), postfix expression evaluation with identifier resolution to concepts only, dimensional unit validation (`input_units * expression_units = output_units`), `CILSpreadNode` generation, and pre-built `SpreadingMatrix` construction from concept value sets.

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `spreaders-increment1-analyze`                                                               |
| description  | Implement AnalyzeSpreaderScope, SpreaderInfo, SpreaderFunctionRegistry (all functions), expression evaluator with identifier resolution, dimensional unit validation, CILSpreadNode, SpreadingMatrix building |

## Source EROs

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Increment 1 sketch
- [ERO008: Spreaders Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md) — Full analysis data structures, FunctionRegistry, expression evaluation, CIL node types

## Dependencies

- Story264 (spreaders-increment1-parse) — parser must produce `ParserSpreaderDefinition` under synthetic keys
- Story256 (aggregators-increment2-analyze) — spreader analysis mirrors the aggregator analysis pattern (postfix evaluation, dimensional unit algebra, FunctionRegistry pattern); dimensional unit types (`DimensionalUnitExpression`, `Multiply`, `cancel`) already exist
- Story254 (aggregators-increment1-matrix-mult) — `AggregationMatrix` struct used for `SpreadingMatrix` construction
- Epic001 (concepts) — concept definitions must be available for `from`/`to` resolution and matrix construction
- Epic005 (metrics) — metrics map must exist (spreader analysis may reference it for future resolution, but Increment 1 uses concept-only resolution)

## Files to Create or Modify

| File                                                    | Action | Purpose                                                                          |
| ------------------------------------------------------- | ------ | -------------------------------------------------------------------------------- |
| `src/main/go/analysis/spreader_analyzer.go`             | Create | `AnalyzeSpreaderScope`, `SpreaderInfo`, `SpreaderFunctionRegistry`, `SpreaderFunctionInfo`, expression evaluator, dimensional unit validator, `BuildSpreadingMatrix` |
| `src/main/go/analysis/cil_nodes.go`                     | Modify | Add `CILSpreadNode` struct to the CIL node hierarchy                              |
| `src/main/go/analysis/analysis_definitions.go`          | Modify | Add `Spreaders map[string]*SpreaderInfo` to `AnalysisDefinitions`                  |
| `src/main/go/analysis/spreader_analyzer_test.go`        | Create | Unit tests: FunctionRegistry, expression evaluation, dimensional unit validation, matrix construction, error cases |

### SpreaderInfo Struct (Increment 1)

```go
type SpreaderInfo struct {
    FromConcepts         []*AnalysisConceptDefinition
    ToConcepts           []*AnalysisConceptDefinition
    FromEmpty            bool
    NewDimension         *AnalysisConceptDefinition     // The concept being introduced
    FromDimensionalUnits *DimensionalUnitExpression
    ToDimensionalUnits   *DimensionalUnitExpression
    ExpressionDimUnits   *DimensionalUnitExpression
    SourceLocation       SourceLocation
}
```

### Spreaders Lookup Map

```go
type AnalysisDefinitions struct {
    // ... existing maps ...
    Aggregators map[string]*AggregatorInfo
    Spreaders   map[string]*SpreaderInfo    // NEW
    CILTrees    map[string]CILNode
}
```

The Spreaders map uses a compound string key derived from sorted concept sets:

```go
func SpreaderMapKey(fromConcepts, toConcepts []string) string {
    fromSorted := sorted(unique(fromConcepts))
    toSorted := sorted(unique(toConcepts))
    return fmt.Sprintf("[%s]->[%s]", strings.Join(fromSorted, ","), strings.Join(toSorted, ","))
}
```

### SpreaderFunctionRegistry (Increment 1 — all functions)

```go
type SpreaderFunctionInfo struct {
    Name        string
    CILNodeType string
    DimUnits    func(params []string) *DimensionalUnitExpression
    Description string
}

type SpreaderFunctionRegistry struct {
    Functions map[string]*SpreaderFunctionInfo
}
```

Registered functions (all from start, unlike aggregators which started with only `sum`):

| Function     | CILNodeType                  | DimUnits Effect          | Description                                    |
| ------------ | ---------------------------- | ------------------------ | ---------------------------------------------- |
| `size(C)`    | `CILSpreadSizeNode`          | Unitless (empty expr)    | Number of distinct values in concept           |
| `1/size(C)`  | `CILSpreadInverseSizeNode`   | Unitless (empty expr)    | 1/N where N is number of values in concept     |
| `sum(C)`     | `CILSpreadSumNode`           | `C` (numerator)          | Sums values across a dimension (implicit aggs) |
| `count(C)`   | `CILSpreadCountNode`         | `C` (numerator)          | Counts values in a dimension                   |
| `1/sum(C)`   | `CILSpreadInverseSumNode`    | `1/C` (denominator)      | Inverse of sum across a dimension              |
| `1/count(C)` | `CILSpreadInverseCountNode`  | `1/C` (denominator)      | Inverse of count across a dimension            |

Note: The spreader FunctionRegistry is SEPARATE from the aggregator FunctionRegistry. While they share many function names, the CIL node types and dimensional unit semantics differ in the spreading context.

### AnalyzeSpreaderScope Function

```go
func (a *CPLAnalyzer) AnalyzeSpreaderScope(scope *SymbolScope) error {
    for _, def := range scope.Definitions {
        if def.DefinitionType != SpreaderScopeDefinition { continue }

        pDef := def.Data.(*ParserSpreaderDefinition)

        // 1. Resolve concept references in from and to
        fromConcepts := resolveConceptList(pDef.FromConcepts)
        toConcepts := resolveConceptList(pDef.ToConcepts)

        // 2. Validate from-to containment and exactly-one-dimension-added
        //    len(to) == len(from) + 1, all from concepts are in to
        validateSpreaderDimensionalRules(fromConcepts, toConcepts)

        // 3. Identify new dimension
        newDim := findNewDimension(fromConcepts, toConcepts)

        // 4. Evaluate postfix expression to compute dimensional unit effect
        exprDimUnits := evaluateSpreaderExpression(def.Scope.Definitions, fromConcepts, toConcepts)

        // 5. Compute and validate dimensional unit compatibility
        //    input_units * expression_units = output_units
        fromDimUnits := dimUnitsFromConcepts(fromConcepts)
        toDimUnits := dimUnitsFromConcepts(toConcepts)
        validateSpreaderDimUnits(fromDimUnits, exprDimUnits, toDimUnits)

        // 6. Build SpreaderInfo (no metric references in Increment 1)
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

        // 7. Store in Spreaders map
        key := SpreaderMapKey(pDef.FromConcepts, pDef.ToConcepts)
        a.spreaders[key] = info
    }
    return nil
}
```

### Expression Evaluation for Spreaders (Increment 1)

Follows the same postfix evaluation pattern as aggregators but handles identifiers (resolved to concepts in Increment 1, metrics in Increment 3):

```go
func (a *CPLAnalyzer) evaluateSpreaderExpression(
    definitions []*SymbolDefinition,
    fromConcepts, toConcepts []*AnalysisConceptDefinition,
) (*DimensionalUnitExpression, error) {

    var stack []*DimensionalUnitExpression

    for _, def := range definitions {
        operand := def.Data.(*OperandData)

        switch operand.Kind {
        case "concept":
            // Push concept's dimensional unit: {Numerator: [conceptName]}
            stack = append(stack, &DimensionalUnitExpression{
                Numerator: []string{operand.Name},
            })

        case "identifier":
            // Increment 1: resolve to concept only (metric/lookup resolution deferred to Increment 3)
            concept, exists := a.concepts[operand.Name]
            if !exists {
                return nil, fmt.Errorf("unknown identifier %q in spreader expression (Increment 1 supports concepts only)", operand.Name)
            }
            stack = append(stack, &DimensionalUnitExpression{
                Numerator: []string{operand.Name},
            })

        case "fncall":
            fnInfo, exists := a.spreaderFunctionRegistry.Functions[operand.Name]
            if !exists {
                return nil, fmt.Errorf("unknown spreader function %q", operand.Name)
            }
            dimUnits := fnInfo.DimUnits(operand.Params)
            stack = append(stack, dimUnits)

        case "operator":
            // Same arithmetic handling as aggregators: * and / operators
            // Pop two, apply Multiply or invert+Multiply, push result
        }
    }

    return stack[0], nil
}
```

### Dimensional Unit Validation for Spreaders

The validation formula is: `input_units * expression_units = output_units`

```go
func validateSpreaderDimUnits(from, expr, to *DimensionalUnitExpression) error {
    predicted := from.Multiply(expr)
    if !predicted.Equals(to) {
        return fmt.Errorf(
            "dimensional unit mismatch: input %s * expression %s = %s, expected %s",
            from.String(), expr.String(), predicted.String(), to.String(),
        )
    }
    return nil
}
```

### CILSpreadNode (Increment 1)

```go
type CILSpreadNode struct {
    SourceCIL               CILNode
    FromDimensions          []string
    ToDimensions            []string
    NewDimension            string
    SpreadingMatrix         *AggregationMatrix
    DistributionExpression  CILNode         // CIL tree for the distribution expression
    ReferencedMetrics       []CILNode       // nil in Increment 1 (filled in Increment 3)
    DimensionalUnitEffect   string
    SourceLocation          SourceLocation
}
func (n *CILSpreadNode) cilNode() {}
```

### SpreadingMatrix Construction

```go
func BuildSpreadingMatrix(newDimValues []string, expressionWeights map[string]float64) *AggregationMatrix {
    cols := len(newDimValues)
    data := make([][]float64, 1)  // One source row per source cell
    data[0] = make([]float64, cols)
    for j, val := range newDimValues {
        data[0][j] = expressionWeights[val]
    }
    return &AggregationMatrix{
        Rows:      1,
        Columns:   cols,
        Data:      data,
        ColLabels: newDimValues,
    }
}
```

### Validation During Analysis (Increment 1)

1. **From-To Containment**: All concepts in `from` must appear in `to`
2. **Exactly One Dimension Added**: `len(to) == len(from) + 1`
3. **Concept Existence**: All concept names in `from`, `to`, and expression must exist in concepts map
4. **Dimensional Unit Compatibility**: `input_units * expression_units = output_units`
5. **No Metric References**: In Increment 1, identifiers in expressions must resolve to concepts (not metrics)
6. **No Duplicate Spreader Patterns**: If two spreaders declare the same `(from, to)` pattern, report error

### Limitations (Increment 1)

- Metrics cannot be referenced in spreader expressions (Increment 3)
- Spreaders not integrated into metric DAG (Increment 2)
- Only concept-based distribution expressions (no weighted distributions via metric data)
- No `CILSpreadNode` insertion into metric CIL trees (manual application in test code only)

### Test Cases

- **FunctionRegistry tests**: Each registered function returns correct `DimUnits`; unknown function returns error
- **Expression evaluation**: `1/size(Country)` → unitless, `Product * 1/size(Country)` → `Product/Unitless` = `Product` numerator, multi-concept expression
- **Dimensional unit validation**: Valid: `from [Product] * expression {1/size(Country)*Country} = to [Country,Product]`; Invalid: `from [Product] * expression {Country} = to [Country,Product]` (missing Product)
- **From/to validation**: `from [A,B] to [A,B,C]` (valid), `from [A] to [A,B,C]` (invalid — adds 2 dims), `from [A] to [B,C]` (invalid — from not subset of to)
- **SpreadingMatrix construction**: Equal-proportion weights, manual weight map, empty new dimension
- **CILSpreadNode**: Generated node has correct FromDimensions, ToDimensions, NewDimension, non-nil SpreadingMatrix

## Task Decomposition (for technical writer)

### Task 1: Define SpreaderFunctionRegistry with all functions — Not Started
   a. Create `src/main/go/analysis/spreader_analyzer.go` — define `SpreaderFunctionInfo`, `SpreaderFunctionRegistry` structs — Not Started
   b. Implement `NewSpreaderFunctionRegistry()` with all 6 functions: `size`, `1/size`, `sum`, `count`, `1/sum`, `1/count` — each returning correct `DimUnits` — Not Started
   c. Write unit tests: verify each function name is registered, verify `DimUnits` output for each function with various params, verify unknown function lookup returns error — Not Started
   d. Verify registry tests pass — Not Started

### Task 2: Implement AnalyzeSpreaderScope with expression evaluation — Not Started
   a. Implement `AnalyzeSpreaderScope` — iterate scope definitions for `SpreaderScopeDefinition` entries — Not Started
   b. Implement `resolveConceptList` — resolve raw concept names to `*AnalysisConceptDefinition` (error on missing) — Not Started
   c. Implement `validateSpreaderDimensionalRules` — check containment and exactly-one-dimension-added — Not Started
   d. Implement `findNewDimension` — identify which concept in `to` is not in `from` — Not Started
   e. Implement `evaluateSpreaderExpression` — postfix stack evaluator handling `concept`, `identifier` (concepts only in Increment 1), `fncall`, `operator` operands — Not Started
   f. Implement `validateSpreaderDimUnits` — `input * expression == output` check — Not Started
   g. Build `SpreaderInfo` and store in `AnalysisDefinitions.Spreaders` map — Not Started
   h. Write unit tests: valid spreader produces correct SpreaderInfo, expression evaluation produces correct dim units, dimensional unit validation catches mismatches — Not Started
   i. Write error case tests: missing concept in from/to, from not subset of to, wrong dim count added, unknown identifier, unknown function — Not Started
   j. Verify analysis tests pass — Not Started

### Task 3: Add CILSpreadNode and SpreadingMatrix construction — Not Started
   a. Add `CILSpreadNode` struct to `src/main/go/analysis/cil_nodes.go` (with all Increment 1 fields; `ReferencedMetrics` set to nil) — Not Started
   b. Implement `BuildSpreadingMatrix` function in `spreader_analyzer.go` — Not Started
   c. Implement CIL node generation from `SpreaderInfo`: `CILSpreadNode` with pre-built `SpreadingMatrix` using equal-proportion weights from distribution expression — Not Started
   d. Store CILSpreadNode in `CILTrees` map under the spreader's map key — Not Started
   e. Write unit tests: verify CILSpreadNode structure, verify matrix dimensions match concept value counts, verify matrix weights for equal-proportion distribution — Not Started
   f. Verify CIL tests pass — Not Started

### Task 4: End-to-end analysis verification with parse_and_analyze_only examples — Not Started [Small]
   a. Create/update `parse_and_analyze_only/without_errors/` spreader examples — expect successful analysis — Not Started
   b. Create/update `parse_and_analyze_only/with_errors/` spreader examples — expect analysis errors for dimensional mismatches, missing concepts, unknown functions — Not Started
   c. Wire `AnalyzeSpreaderScope` into the main analysis pipeline (`WalkSymbolTableWithHandlers`) — Not Started
   d. Verify full parse→analyze pipeline: valid CPL → no errors, invalid CPL → meaningful error messages with source locations — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

- Task 1 must complete first (FunctionRegistry needed by expression evaluation)
- Task 2 depends on Task 1 (expression evaluation uses FunctionRegistry)
- Task 3 depends on Task 2 (CIL node generation needs SpreaderInfo)
- Task 4 depends on all preceding tasks (end-to-end verification)

No parallel groups — each task builds on the previous.
Each task is standalone in its test scope but sequential in implementation dependency.

## Acceptance Criteria Summary

1. `SpreaderFunctionRegistry` registers all 6 functions with correct `DimUnits`
2. `AnalyzeSpreaderScope` walks symbol table and produces `SpreaderInfo` entries in the Spreaders map
3. From/to containment validated: all `from` concepts must be in `to`; `len(to) == len(from) + 1`
4. Expression postfix evaluation produces correct `DimensionalUnitExpression`
5. Dimensional unit validation enforces `input_units * expression_units = output_units`
6. `CILSpreadNode` generated with pre-built `SpreadingMatrix` from concept value sets
7. Equal-proportion distribution (`1/size(C) * C`) produces uniform weights per concept value
8. Concept references and function calls in expressions are resolved correctly
9. Identifier references resolve to concepts only (metrics deferred to Increment 3)
10. Error cases produce meaningful messages with source locations
11. All tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`

## Intent Summary for Technical Writer

This story delivers the analysis phase for spreader declarations. Spreaders are validated for dimensional correctness: `from` concepts must be a subset of `to` concepts, with exactly one new dimension added. The postfix expression evaluator (same pattern as aggregators) computes the dimensional unit effect of the spreader expression. Dimensional unit validation enforces `input_units * expression_units = output_units` — the inverse of the aggregator formula (`input_units = expression_units * output_units`). The spreader `FunctionRegistry` is separate from the aggregator registry and registers ALL 6 functions from the start (unlike aggregators which started with only `sum`). Expressions in Increment 1 support concept references only — identifier references (for future metric references) are deferred to Increment 3. `CILSpreadNode` is generated with a pre-built `SpreadingMatrix` constructed from concept value sets, using equal-proportion weights. The `SpreadingMatrix` reuses the `AggregationMatrix` struct from Story254 (tensorops). No metric DAG integration happens in this increment — spreaders are validated but not yet applied to metric calculations.

Key design note: The `DimensionalUnitExpression` type and its `Multiply`/`Equals` methods should already exist from the aggregator analysis story (Story256). The spreader analysis reuses these directly.
