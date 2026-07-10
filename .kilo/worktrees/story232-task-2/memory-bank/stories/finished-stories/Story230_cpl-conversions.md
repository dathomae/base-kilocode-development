# Story230: CPL Conversions (Simple)

## Goal

Implement unit and measure conversions using pattern/calculation pairs as part of [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md). Covers conversion clause grammar for both `CPLUnits.g4` and `CPLMeasures.g4`, `ConversionScope`/`PatternScope`/`CalculationScope` symbol table scopes, conversion pair assembly via symbol table walking, CIL conversion node types, and conversion code synthesis. Simple (non-conditional) conversions only — single calculation expression, no `when` clauses. Combines what were Increments 4 and 5 of Epic003.

## References

- [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md)
- [ERO003: Units and Measures Parsing Phase — Increments 4-5](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md)
- [ERO003: Units and Measures Analysis Phase (Revised) — Conversion Scope Walking](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md)
- [ERO003: Units and Measures Synthesis Phase (Revised)](../specifications/design/cpl/eros/ERO003_units-measures-synthesis-phase.md)
- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — provides expression grammar, BuildExpressionTree, CIL expression types and synthesis
- [Story229_cpl-compound-units.md](Story229_cpl-compound-units.md) — provides expression-based measure definitions
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md)
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — _Done_ — provides expression grammar, `BuildExpressionTree`, CIL expression types and synthesis
- [Story229_cpl-compound-units.md](Story229_cpl-compound-units.md) — _Not Started_ — provides expression-based measure definitions that conversion calculations reference

## Dependent Stories

- [Story231_cpl-unit-type-checking.md](Story231_cpl-unit-type-checking.md) — _Not Started_ — depends on conversion pairs in CIL for validation
- [Story232_cpl-conversion-injection.md](Story232_cpl-conversion-injection.md) — _Not Started_ — depends on conversion pairs in CIL for injection

## Tasks

### Task Sizing

- **Task 1** (Add conversion scope types to symbol table): `[Small]` — single file (`symbol_table.go`) adding enum values and scope types. Straightforward.
- **Task 2** (Update CPLUnits.g4 with conversion grammar and regenerate): Medium — adds multiple new rules and tokens, requires regeneration.
- **Task 3** (Update CPLMeasures.g4 with conversion grammar and regenerate): Medium — same pattern as Task 2 for measures.
- **Task 4** (Implement conversion listener methods): Medium — introduces new listener methods for scope push/pop and pattern-bound variable capture.
- **Task 5** (Implement AnalyzeUnitConversions): Medium — walks symbol table, pairs PatternScope with CalculationScope, validates units and variables.
- **Task 6** (Implement algebraic reversal and AnalyzeMeasureConversions): Medium — derives inverse conversions, applies same pattern to measures.
- **Task 7** (Define CIL conversion types and implement synthesis): Medium — new CIL structs, synthesis code generation for conversion tables and evaluator functions.
- **Task 8** (Verification and regression): Medium — end-to-end tests with target CPL examples, regression suite.

### Parallel Execution

- **Group A** (Tasks 2, 3): Both tasks modify grammar files (`CPLUnits.g4` and `CPLMeasures.g4`). Disjoint files, no shared compilation. Neither depends on the other's output.
- **Group B** (Tasks 5, 6): Task 5 (unit conversions) and Task 6 (measure conversions) touch different analyzer functions but share the same analysis package. Same-package compilation risk means worktree isolation is required, but they can execute in parallel.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4 → [Task 5 ‖ Task 6] → Task 7 → Task 8
```

- Task 1 must complete first (scope types must exist before grammar or listeners can reference them).
- Tasks 2 and 3 are parallel grammar updates.
- Task 4 depends on Tasks 2 and 3 (listener methods reference new grammar rules and scope types).
- Tasks 5 and 6 are parallel analyzer work (unit vs measure conversions share the same pattern but touch different data structures).
- Task 7 depends on Tasks 5 and 6 (needs analysis conversion types to build CIL).
- Task 8 depends on all preceding tasks.

### Reintegration

**Group A (Tasks 2, 3):**
- Merge order: any order (disjoint grammar files)
- Integration tests: `go build ./src/main/go/parsing/...` — verify both generated parsers compile
- Watch for conflicts in: none expected

**Group B (Tasks 5, 6):**
- Merge order: any order (disjoint functions in same package — `AnalyzeUnitConversions` vs `AnalyzeMeasureConversions`)
- Integration tests: `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — verify combined analyzer behavior
- Watch for conflicts in: `analysis/analysis.go` or shared registration code if both tasks touch `WalkSymbolTable` callbacks

### Task 1: Add conversion scope types to symbol table — Completed [Small]
   a. Add `ConversionScope`, `PatternScope`, `CalculationScope`, `MeasureConversionScope` to `SymbolTableScopeType` enum in `symbol_table.go` — Not Started
   b. Add `DefinitionType` values for conversion scopes: `ConversionDefinition`, `PatternDefinition`, `CalculationDefinition`, `MeasureConversionDefinition` — Not Started
   c. Add lexer tokens: `CONVERSION_KEYWORD`, `PATTERN_KEYWORD`, `CALCULATION_KEYWORD` — Not Started
   d. Verify symbol table package compiles — Not Started

### Task 2: Update CPLUnits.g4 with conversion clause grammar and regenerate ANTLR files — Completed
   a. Extend `unitDeclaration` to include optional `conversionClause*` — Not Started
   b. Add `conversionClause: CONVERSION_KEYWORD conversionPair+` — Not Started
   c. Add `conversionPair: patternClause calculationClause` — Not Started
   d. Add `patternClause: PATTERN_KEYWORD pattern` — Not Started
   e. Add `pattern` rule — equation form only: `expression patternBoundVariable EQ expression patternBoundVariable` — Not Started
   f. Add `patternBoundVariable: LPAREN identifier RPAREN` — Not Started
   g. Add `calculationClause: CALCULATION_KEYWORD expression` — Not Started
   h. Regenerate ANTLR Go files — Not Started
   i. Verify generated files compile — Not Started

### Task 3: Update CPLMeasures.g4 with measure conversion grammar and regenerate ANTLR files — Completed
   a. Extend `measureDeclaration` to include optional `measureConversionClause*` — Not Started
   b. Add `measureConversionClause: CONVERSION_KEYWORD measureConversionPair+` — Not Started
   c. Add `measureConversionPair: patternClause measureCalculationClause` — Not Started
   d. Add `measureCalculationClause: CALCULATION_KEYWORD expression` — Not Started
   e. Regenerate ANTLR Go files — Not Started
   f. Verify generated files compile — Not Started

### Task 4: Implement conversion listener methods — Completed
   a. Implement `EnterConversionClause` / `ExitConversionClause` — push/pop `ConversionScope` — Not Started
   b. Implement `EnterPatternClause` / `ExitPatternClause` — push/pop `PatternScope` — Not Started
   c. Implement `EnterCalculationClause` / `ExitCalculationClause` — push/pop `CalculationScope` — Not Started
   d. Implement measure variants: `EnterMeasureConversionClause` / `ExitMeasureConversionClause` — push/pop `MeasureConversionScope` — Not Started
   e. Implement `ExitPatternBoundVariable` — capture variable name bound to the preceding unit expression — Not Started
   f. Ensure pattern and calculation expressions use existing Story228 listener methods (identifiers, operators pushed to current scope definitions in postfix order) — Not Started
   g. Write listener unit tests verifying correct symbol table tree structure: `ConversionScope` → `PatternScope` + `CalculationScope` — Not Started
   h. Verify listener tests pass — Not Started

### Task 5: Implement AnalyzeUnitConversions — Completed
   a. Define `AnalysisConversionPair` struct: `SourceUnit`, `TargetUnit *AnalysisUnitDefinition`, `TargetVar`, `SourceVar string`, `Calculation Expression` — Not Started
   b. Implement `AnalyzeUnitConversions(unitScope *SymbolTable, unitDef *AnalysisUnitDefinition)`: walk `unitScope.Children` for `ConversionScope` entries — Not Started
   c. For each `ConversionScope`, walk children and pair `PatternScope` with `CalculationScope` by order — Not Started
   d. Extract pattern-bound variables from `PatternScope`: first unit expression + variable = target, second = source — Not Started
   e. Call `BuildExpressionTree` on `CalculationScope`'s `Definitions` slice for the calculation expression — Not Started
   f. Resolve pattern unit references against units lookup map — Not Started
   g. Validate: source and target units belong to the same unit group — Not Started
   h. Validate: all variables in calculation expression are pattern-bound variables or annotated unit constants — Not Started
   i. Attach `AnalysisConversionPair` to `unitDef.Conversions` — Not Started
   j. Write analyzer unit tests for unit conversions — Not Started
   k. Verify analyzer tests pass — Not Started

### Task 6: Implement algebraic reversal and AnalyzeMeasureConversions — Completed
   a. Implement algebraic reversal: for each forward `AnalysisConversionPair`, derive inverse by swapping source/target and inverting calculation. Add inverse as second pair — Not Started
   b. Define `AnalysisMeasureConversionPair` struct: `SourceMeasure`, `TargetMeasure *AnalysisMeasureDefinition`, `TargetVar`, `SourceVar string`, `Calculation Expression` — Not Started
   c. Implement `AnalyzeMeasureConversions(measureScope *SymbolTable, measureDef *AnalysisMeasureDefinition)`: walk `MeasureScope` children for `MeasureConversionScope` — Not Started
   d. Apply same pairing, variable extraction, `BuildExpressionTree`, and validation logic as Task 5 — Not Started
   e. Attach `AnalysisMeasureConversionPair` to `measureDef.Conversions` — Not Started
   f. Write analyzer unit tests for measure conversions and algebraic reversal — Not Started
   g. Verify analyzer tests pass — Not Started

### Task 7: Define CIL conversion types and implement synthesis — Completed
   a. Define `CILConversionPair`: `SourceUnit`, `TargetUnit`, `TargetVar`, `SourceVar string`, `Calculation CILExpression` — Not Started
   b. Define `CILMeasureConversionPair`: `SourceMeasure`, `TargetMeasure`, `TargetVar`, `SourceVar string`, `Calculation CILExpression` — Not Started
   c. Add `Conversions []CILConversionPair` to unit CIL definitions — Not Started
   d. Add `Conversions []CILMeasureConversionPair` to measure CIL definitions — Not Started
   e. Implement `SynthesizeUnitConversions(conversions []CILConversionPair) string`: emits Go conversion table keyed by `(sourceUnit, targetUnit)` pair — Not Started
   f. Implement `SynthesizeMeasureConversions` — same pattern for measure conversions — Not Started
   g. Wire conversion synthesis into main synthesis pipeline — Not Started
   h. Write synthesis unit tests verifying emitted Go conversion functions compile and behave correctly — Not Started
   i. Verify synthesis tests pass — Not Started

### Task 8: Verification with target CPL examples and regression — Completed
   a. Create target CPL example with `Length` unit conversions (`Kilometer ↔ Meter ↔ Centimeter`) — Done
   b. Create target CPL example with `TemperatureScale` conversions (`C ↔ F`) — Done
   c. Verify parser produces correct symbol table tree with `ConversionScope` → `PatternScope` + `CalculationScope` — Done
   d. Verify analyzer extracts pattern-bound variables, resolves units, builds calculation expression tree — Done
   e. Verify analyzer derives inverse conversions — Done
   f. Verify analysis rejects: bare numeric constants without unit annotation, unit names reused as pattern variables, conversions between units of different concepts — Done
   g. Verify synthesis emits Go code with conversion tables and evaluator functions — Done
   h. Verify existing unit/measure tests still pass — Done
   i. Run full CPL test suite: `go test ./src/main/go/... -count=1 -timeout 120s` — Done

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Equation-form patterns only — ratio form is explicitly out of scope
3. Pattern-bound variables use `LPAREN identifier RPAREN` syntax
4. All numeric constants in calculation expressions must carry explicit unit annotations
5. Unparenthesized identifiers in calculation expressions are unit references
6. Algebraic reversal performed at analysis time — synthesizer sees both directions as explicit CIL pairs
7. Units and measures share the same conversion machinery, differing only in where conversion pairs are stored and what they reference
8. Go functions that can fail must return `(T, error)`
9. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Add the symbol table infrastructure needed to represent conversion scopes during parsing
- **Task 2 Intent**: Extend the units grammar to accept conversion clauses with pattern/calculation pairs
- **Task 3 Intent**: Extend the measures grammar to accept measure conversion clauses with the same pattern/calculation structure
- **Task 4 Intent**: Implement listener methods that build the conversion scope tree (ConversionScope → PatternScope + CalculationScope)
- **Task 5 Intent**: Build the analyzer for unit conversions — extract pattern variables, build calculation expression trees, validate unit consistency
- **Task 6 Intent**: Derive inverse conversions automatically and apply the same analysis pattern to measure conversions
- **Task 7 Intent**: Generate CIL conversion types and synthesize Go code with conversion tables and evaluator functions
- **Task 8 Intent**: Verify the full pipeline with realistic conversion examples and ensure no regressions

## Acceptance Criteria

1. `CPLUnits.g4` and `CPLMeasures.g4` accept conversion clauses with pattern/calculation pairs
2. Parser produces correct symbol table tree: `ConversionScope` → `PatternScope` + `CalculationScope`
3. `AnalyzeUnitConversions` assembles `AnalysisConversionPair` entries with resolved units and validated calculations
4. `AnalyzeMeasureConversions` assembles `AnalysisMeasureConversionPair` entries with the same validations
5. Algebraic reversal produces inverse conversion pairs for every forward pair
6. Analysis rejects bare numeric constants, variable name collisions, and cross-concept conversions
7. Synthesis emits compilable Go code with conversion tables keyed by `(source, target)`
8. Target CPL examples (Length, TemperatureScale) parse, analyze, and synthesize correctly
9. All existing unit/measure tests pass with zero regressions
10. Full CPL test suite passes

## Footer

### Open Questions

None

### Known Issues

None

### Additional Context

**Design decisions confirmed for this story:**

1. **Pattern: equation form only**. Ratio form dropped as unnecessary syntactic sugar.

2. **Variable binding syntax**: `UnitName(variable_name)` in patterns — the parenthesized name binds a variable to a value of that unit type. This syntax is unambiguous because CPL has no function-call syntax and `LPAREN`/`RPAREN` in other contexts always follows a clear grammatical position.

3. **Unit annotations in calculations**: Unparenthesized identifiers in calculation expressions are unit references. A reference to a name that appears in a unit declaration is always interpreted as the unit. If a unit name and a pattern-bound variable share the same name, it's an error. All numeric constants MUST carry explicit unit annotations (e.g., `32*F`, not bare `32`), verified during analysis.

4. **Algebraic reversibility**: The user writes one pattern per unit pair (e.g., `Kilometer(km) = Meter(m)`). The analyzer infers the inverse conversion at analysis time and emits CIL with both forward and inverse conversion pairs. This keeps synthesis simple (no derivation logic at code-gen time) and avoids requiring users to write both directions.

**Explicitly out of scope:**
- Conditional conversions (`when` clauses, context-dependent calculations)
- Multi-step conversion path resolution (chain: Kilometer→Meter→Centimeter)
- Conversion precedence (specific over general)
- Ratio-form patterns
