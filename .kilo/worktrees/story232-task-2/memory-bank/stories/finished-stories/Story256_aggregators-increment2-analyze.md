# Story256: Aggregators Increment 2 — Analyze

## Goal

Implement the analysis phase for basic aggregator declarations (Increment 2 of Epic007). Walks the symbol table to extract parser-level aggregator definitions from synthetic-key entries, resolves concept references, evaluates postfix expressions via a function registry to compute dimensional unit effects, validates `input_units * expression_units = output_units`, builds `AggregatorInfo` entries stored in the Aggregators lookup map, and generates `CILAggregateNode` CIL trees with pre-built aggregation matrices. This story is part of [Epic007 — Aggregators Incremental Implementation](../../epics/Epic007_cpl-aggregators-implementation.md).

## References

- [ERO007 — Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md)
- [ERO007 — Aggregators Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md)
- [ERO007 — Aggregators Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO007_aggregators-parsing-phase.md)
- [Postfix Expression Parsing Pattern](../patterns/postfix-expression-parsing.md)
- [Symbol Table-Driven Parsing Pattern](../patterns/symbol-table-driven-parsing.md)
- [Units and Measures Design](../specifications/design/cpl/units-and-measures.md)

## Dependencies

- [Story255_aggregators-increment2-parse.md](Story255_aggregators-increment2-parse.md) - Not Started — `ParserAggregatorDefinition` struct, `AggregatorScopeDefinition` scope type, synthetic-key symbol table entries, postfix `OperandDefinition` entries
- Epic001 (concepts) — `AnalysisDefinitions.Concepts` map for resolving from/to concept references; `AnalysisConceptDefinition.ValueOrdering` for building aggregation matrices
- Epic003 (units/measures) — dimensional unit algebra reuses `*`, `/`, `^` machinery; `DimensionalUnitExpression` uses the same algebraic pattern

## Dependent Stories

- Story257 (aggregators-increment2-synthesize) - Not Started — needs `AggregatorInfo`, `CILAggregateNode`, and pre-built aggregation matrices generated during this story's analysis phase
- Increment 4 (aggregator selection and integration into metric DAG) — queries the Aggregators map assembled during this story

## Tasks

### Task Sizing

All four tasks introduce new structs, methods spanning multiple files, or complex analysis logic. None qualify as `[Small]` — each touches 3+ files, introduces new infrastructure, and requires reasoning across multiple concerns (dimensional unit algebra, postfix evaluation, concept resolution, CIL node construction). All four tasks route to `code-for-story-implementor`.

Architect-for-story-planning analysis summary: No decomposition needed — all tasks have natural boundaries and are appropriately sized for single-handoff execution. Tasks 1-3 each have 5–6 subtasks spanning 3–4 files, but none exhibit intermediate verification cycles, self-contained preambles, or 8+ subtask steps. Task 4 (6 subtasks, 3 files) has no internal phase transitions — the analyze → store flow is one linear walk. All tasks are at their natural granularity.

### Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 creates `function_registry.go` and `aggregation_matrix.go` (structs + builder functions + tests). Task 3 creates `aggregator_info.go`, `cil_aggregate_node.go`, and implements the postfix evaluator and dimensional unit validator with their tests. All files are disjoint; no file is modified by both tasks. Both tasks are in the same package (`analysis`) and share compilation scope.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4
```

- Task 1 must complete first: `DimensionalUnitExpression` is a foundational type used by every other task. The `Aggregators` field addition to `AnalysisDefinitions` completes the struct shape that Task 4 populates.
- Tasks 2 and 3 run in parallel: disjoint file sets, same package. Task 3's evaluator references `FunctionRegistry` and `AggregationMatrix` types from Task 2 — the evaluator should accept its function registry as a `map[string]*AggregatorFunctionInfo` parameter to compile independently in the Task 3 worktree before Task 2 is merged.
- Task 4 depends on all preceding tasks: `AnalyzeAggregatorScope` uses `DimensionalUnitExpression`, `FunctionRegistry`, `AggregatorInfo`, `CILAggregateNode`, and `AggregationMatrix`.

### Reintegration

**Group A (Tasks 2, 3):**
- Merge order: Task 2 first (defines `FunctionRegistry`, `AggregatorFunctionInfo`, and `AggregationMatrix` types that Task 3's evaluator references), then Task 3
- Integration tests: `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — verify all types compile together, evaluator resolves functions through registry, matrix builders produce correct output
- Watch for conflicts in: none expected (fully disjoint files within the same package)

### Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

### Constraints

- Must work with parser-level aggregator definitions from the symbol table (synthetic keys)
- Must follow the Postfix Expression Parsing Pattern for expression evaluation
- Must follow the Symbol Table-Driven Parsing Pattern for phase separation
- The `DimensionalUnitExpression` algebra (`*`, `/`, `cancel`) must match the same algebraic behavior as the existing unit/measure conversion system
- Aggregation matrices must be pre-built from concept value sets at compile time — no runtime matrix construction
- All concept references in `from`, `to`, and expression must be validated against the `Concepts` map
- `AggregatorMapKey` must be deterministic (sorted concept names) to ensure `[Country, Product]` and `[Product, Country]` produce the same key
- CIL nodes must implement the `cil.CILNode` interface (private `cilNode()` method)
- Only `sum` function is supported in the FunctionRegistry for this increment
- Aggregators are stored but not yet applied to metric calculations (that integration occurs in Increment 4)

### Intent

This story delivers the analysis phase for aggregator declarations. The analyzer walks the symbol table, validates all cross-references, evaluates postfix expressions to compute dimensional unit effects, verifies `input_units * expression_units = output_units`, builds `AggregatorInfo` entries, and generates `CILAggregateNode` CIL trees with pre-built aggregation matrices. The `sum` function only is registered; `count`, `size`, `1/sum`, `1/count` come in Increment 3. The aggregators map uses compound string keys `[sorted-from]->[to]` for O(1) lookup, consumed by Increment 4 during metric DAG analysis.

### Acceptance Criteria

1. `DimensionalUnitExpression` correctly multiplies, cancels common factors, and compares equality
2. `FunctionRegistry` returns correct `DimUnits` for `sum(Concept)` — concept name in numerator
3. `BuildSumAggregationMatrix` produces correct all-ones matrix with given row/column labels
4. Postfix evaluator correctly computes dimensional units for `sum(Country)`, arithmetic expressions (e.g., `(a-b)*sum(Country)`), and mixed operand sequences
5. `AnalyzeAggregatorScope` resolves concepts, evaluates expressions, validates dimensional units, stores `AggregatorInfo` in Aggregators map under correct compound key
6. Missing concept references produce clear compile-time errors (e.g., `aggregator: from concept "X" not found`)
7. `to []` with 2+ from-concepts produces clear error (e.g., `aggregator: to [] requires exactly one concept in from list, got 2`)
8. Dimensional unit mismatches produce clear error messages showing input, expression, and output units
9. `CILAggregateNode` entries are stored in `AnalysisDefinitions.CILTrees` under key `"aggregator:" + mapKey` with pre-built matrices
10. All tests pass with `go test ./src/main/go/analysis/...`

### Task 1: Create DimensionalUnitExpression and analysis definitions — Completed

   a. Create `src/main/go/analysis/dimensional_unit_expression.go` with `DimensionalUnitExpression` struct (fields: `Numerator []string`, `Denominator []string`), `Multiply(other *DimensionalUnitExpression) *DimensionalUnitExpression`, `Equals(other *DimensionalUnitExpression) bool`, `IsUnitless() bool`, `cancel()` (removes elements appearing in both numerator and denominator, then sorts both slices), and `String() string` methods — Not Started
   b. Write unit tests in `src/main/go/analysis/dimensional_unit_expression_test.go` for: multiply (Product × Country produces both in numerator), cancel (common factor Country appears in both and is removed), equals (same numerators/denominators in different order), isUnitless, string format — Not Started
   c. Add `Aggregators map[string]*AggregatorInfo` field to `AnalysisDefinitions` in `src/main/go/analysis/analysis_definitions.go`. Update `GetAnalysisDefinitions()` in `src/main/go/analysis/cpl_analyzer.go` to populate the new field (initialize as empty map — populated later by Task 4). Add `*FunctionRegistry` field to `CPLAnalyzer` struct and initialize in `NewCPLAnalyzer()` — Not Started
   d. Implement `dimensionalUnitsFromConceptSet(concepts []*AnalysisConceptDefinition) *DimensionalUnitExpression` helper in `src/main/go/analysis/dimensional_unit_expression.go`: n=0 returns unitless, n=1 returns `{Numerator: [name]}`, n>1 returns `{Numerator: [last], Denominator: [first...n-1]}`. Implement `AggregatorMapKey(fromConcepts []string, toConcept string) string` helper: sorts from concepts, produces `"[sorted]->[to]"` or `"[sorted]->[]"` — Not Started
   e. Run `go test ./src/main/go/analysis/...` — all tasks pass, package compiles — Not Started

### Task 2: Create FunctionRegistry and AggregationMatrix builders — Completed

   a. Create `src/main/go/analysis/function_registry.go` with `AggregatorFunctionInfo` struct (Name string, CILNodeType string, DimUnits func(params []string) *DimensionalUnitExpression, Description string) and `FunctionRegistry` struct (Functions map[string]*AggregatorFunctionInfo). Implement `NewAggregatorFunctionRegistry() *FunctionRegistry`: register only `"sum"` with CILNodeType `"CILAggregateNode"`, DimUnits returning `{Numerator: [params[0]]}`, description "Sums values across a concept dimension" — Not Started
   b. Create `src/main/go/analysis/aggregation_matrix.go` with `AggregationMatrix` struct (Rows int, Columns int, Data [][]float64, RowLabels []string, ColLabels []string). Implement `BuildSumAggregationMatrix(collapsedValues []string, targetValues []string) *AggregationMatrix` — all entries 1.0. Implement `BuildSizeAggregationMatrix(collapsedValues []string) *AggregationMatrix` — one column of 1.0 per row. Implement `BuildInverseSumAggregationMatrix(collapsedValues []string, targetValues []string) *AggregationMatrix` — each entry 1.0/len(collapsedValues) — Not Started
   c. Write unit tests in `src/main/go/analysis/function_registry_test.go`: verify `NewAggregatorFunctionRegistry()` returns registry with `"sum"` entry, sum(C) dim units are `{Numerator: [C]}`, unknown function lookup returns nil, invalid params (wrong count) returns nil — Not Started
   d. Write unit tests in `src/main/go/analysis/aggregation_matrix_test.go`: verify `BuildSumAggregationMatrix` produces row×col matrix of all 1.0, row labels match collapsedValues, column labels match targetValues, single-column case (targetValues=nil or empty) produces right shape — Not Started
   e. Run `go test ./src/main/go/analysis/...` — package compiles, all new tests pass — Not Started

### Task 3: Create AggregatorInfo, CILAggregateNode, and postfix evaluator — Completed

   a. Create `src/main/go/analysis/aggregator_info.go` with `AggregatorInfo` struct: FromConcepts []*AnalysisConceptDefinition, ToConcept *AnalysisConceptDefinition, ToEmpty bool, FromDimensionalUnits *DimensionalUnitExpression, ToDimensionalUnits *DimensionalUnitExpression, ExpressionDimUnits *DimensionalUnitExpression, CellOperation string, SourceLocation SourceLocation — Not Started
   b. Create `src/main/go/analysis/cil_aggregate_node.go` with `CILAggregateNode` struct: SourceCIL cil.CILNode, FromDimensions []string, ToDimension string, AggregationMatrix *AggregationMatrix, CellOperation string, DimensionalUnitEffect string, SourceLocation SourceLocation. Implement `func (n *CILAggregateNode) cilNode() {}` — Not Started
   c. Implement `evaluateAggregatorExpression(definitions []*SymbolDefinition, fromConcepts []*AnalysisConceptDefinition, registry *FunctionRegistry) (*DimensionalUnitExpression, string, error)` as a package-level function in `src/main/go/analysis/aggregator_evaluator.go`. For each definition: extract `OperandData` from `def.Data`. Switch on `operand.Kind` — `"concept"` pushes `{Numerator: [name]}` onto a `[]*DimensionalUnitExpression` stack; `"fncall"` looks up `operand.Name` in registry, calls `fnInfo.DimUnits(operand.Params)`, pushes result, sets cellOp; `"operator"` — `"*"` pops two, pushes `left.Multiply(right)`; `"/"` pops two (right, left), inverts right, pushes `left.Multiply(inverted)`. Return resulting dimension expression, cellOp, and any error — Not Started
   d. Implement `validateAggregatorDimensionalUnits(info *AggregatorInfo) error` in `src/main/go/analysis/aggregator_evaluator.go`. Compute `info.FromDimensionalUnits` via `dimensionalUnitsFromConceptSet(info.FromConcepts)`. Compute `info.ToDimensionalUnits`: if ToEmpty, unitless; otherwise via `dimensionalUnitsFromConceptSet([]*AnalysisConceptDefinition{info.ToConcept})`. Validate `info.FromDimensionalUnits.Multiply(info.ExpressionDimUnits).Equals(info.ToDimensionalUnits)`. Return descriptive error on mismatch — Not Started
   e. Write unit tests in `src/main/go/analysis/aggregator_evaluator_test.go`: postfix evaluator for `sum(Country)` produces `{Numerator: [Country]}`; `a-b` evaluates to unitless (same concept dimensions cancel); `(a-b)*sum(Country)` produces `{Numerator: [Country]}`; unknown function name returns error; insufficient operands for operator returns error; invalid params to function returns error — Not Started
   f. Run `go test ./src/main/go/analysis/...` — package compiles, all new tests pass — Not Started

### Task 4: Implement AnalyzeAggregatorScope and CIL tree generation — Completed

   a. Create `src/main/go/analysis/analyze_aggregators.go` with `AnalyzeAggregatorScope(table *symboltable.SymbolTable, analyzer *CPLAnalyzer) error` function — Not Started
   b. Implement the full algorithm in `AnalyzeAggregatorScope`: (1) find `AggregatorScopeDefinition` entry in `table.Definitions`, extract `*ParserAggregatorDefinition` from `def.Data`; (2) resolve each from-concept name against `analyzer.conceptAnalyzer.GetConcepts()` — error if not found; (3) resolve to-concept name (unless `ToEmpty`) — error if not found; (4) validate to-concept is in from list (unless `ToEmpty`); validate `ToEmpty` requires exactly 1 from-concept; (5) evaluate aggregator expression via the postfix evaluator created in Task 3; (6) build `AggregatorInfo`, call `validateAggregatorDimensionalUnits` from Task 3; (7) construct compound key via `AggregatorMapKey`, store `AggregatorInfo` in `analyzer.aggregators[key]`; (8) build `AggregationMatrix` using `BuildSumAggregationMatrix` with collapsed concept's `ValueOrdering`; (9) create `CILAggregateNode` and store in `analyzer.cilTrees["aggregator:"+key]` — Not Started
   c. Register `AnalyzeAggregatorScope` in the `WalkSymbolTableWithHandlers` setup in `src/main/go/analysis/cpl_analyzer.go`. Add a new entry to the `entryHandlers` map for `symboltable.AggregatorScope` dispatching to `AnalyzeAggregatorScope`. Ensure the `Aggregators` map and `cilTrees` are initialized on `CPLAnalyzer` — Not Started
   d. Write integration tests in `src/main/go/analysis/analyze_aggregators_test.go`: parse a CPL document with a simple aggregator definition, run full analysis, verify `Aggregators` map populated with correct compound key, verify `CILTrees` contains `"aggregator:"+key` entry with correct `CILAggregateNode`, verify node's `AggregationMatrix` has correct shape and labels. Test both single-concept aggregation and scalar `to []` reduction — Not Started
   e. Write error case tests: non-existent from-concept produces clear error, non-existent to-concept produces clear error, to-concept not in from list produces clear error, `to []` with 2+ from-concepts produces clear error, unknown function name produces clear error, dimensional unit mismatch produces clear error message showing input/expression/output units — Not Started
   f. Run `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — all tests pass, no compilation errors, acceptance criteria 1–10 verified — Not Started

## Notes

- **Story255 dependency**: This story assumes `ParserAggregatorDefinition`, `AggregatorScopeDefinition` scope type, and `OperandDefinition` are already defined by Story255 (aggregators-increment2-parse). The `AnalyzeAggregatorScope` handler is registered for `symboltable.AggregatorScope` — Story255 will add this constant to the `ScopeType` enum. If Story255 has not been completed, adjust the scope type constant accordingly.
- **FunctionRegistry extent**: Only `"sum"` is registered in this increment. The `AggregatorFunctionInfo.CILNodeType` is set to `"CILAggregateNode"` (generic). In Increment 3, this becomes `"CILAggregateSumNode"` and operation-specific node types replace the generic `CILAggregateNode`.
- **Matrix construction**: `BuildSumAggregationMatrix` is the only builder used in this increment. `BuildSizeAggregationMatrix` and `BuildInverseSumAggregationMatrix` are implemented in Task 2 for forward compatibility with Increment 3, but only `BuildSumAggregationMatrix` is called in Task 4.
- **Cell unit validation**: Deferred to Increment 4 when metric cell type information is available in the DAG analysis context.
- **Aggregator disambiguation**: Duplicate aggregators with the same `(from, to)` pattern are rejected with a compile-time error. Full disambiguation via specificity scoring is deferred to Increment 5.
