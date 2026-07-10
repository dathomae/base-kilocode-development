# Story229: CPL Compound Units

## Goal

Wire the expression infrastructure from Story228 into measure definitions so that measure unit specifications support compound unit expressions (e.g., `Length/Time^2`). Replace the current `ConstraintScope`-based unit extraction with expression-based parsing. Update the analyzer to build and store `Expression` trees. Update synthesis to emit compound unit Go code. Covers Increment 3 of [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md).

## References

- [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md)
- [ERO003: Units and Measures Parsing Phase — Increment 3](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md)
- [ERO003: Units and Measures Analysis Phase — Increment 3](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md)
- [ERO003: Units and Measures Synthesis Phase (Revised)](../specifications/design/cpl/eros/ERO003_units-measures-synthesis-phase.md)
- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — provides expression grammar rules, BuildExpressionTree, CIL expression types, expression synthesis
- [Story235_unify-cilnode-cilexpression-hierarchies.md](finished-stories/Story235_unify-cilnode-cilexpression-hierarchies.md) — provides unified `CILNode` types that Story229 must target for CIL output
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — _Done_ — provides expression grammar rules, `BuildExpressionTree`, CIL expression types, and `SynthesizeExpression`
- [Story235_unify-cilnode-cilexpression-hierarchies.md](finished-stories/Story235_unify-cilnode-cilexpression-hierarchies.md) — _Done_ — provides unified `CILNode` types that Story229 must target for CIL output

## Dependent Stories

- [Story230_cpl-conversions.md](finished-stories/Story230_cpl-conversions.md) — _Done_ — depends on compound unit expressions being available in measure definitions
- [Story231_cpl-unit-type-checking.md](Story231_cpl-unit-type-checking.md) — _Not Started_ — depends on compound unit CIL trees for unit tracking

## Tasks

### Task Sizing

- **Task 1** (Update CPLMeasures.g4 and regenerate): Medium — grammar file change, regeneration step, verification compile.
- **Task 2** (Update parser listener to push postfix definitions into MeasureScope): Medium — adds `MeasureScope` to expression-defs whitelist, wires `unitParameter` listener methods, removes ConstraintScope creation from measures listener.
- **Task 3** (Update analyzer for expression trees): Medium — updates `AnalysisMeasureDefinition`, `AnalyzeMeasureScope`, and `VerifyMeasureUnitReferences`. Multiple coordinated changes in `analysis/` and `synthesis/` packages.
- **Task 4** (Update measure synthesis): `[Small]` — single synthesis file update to call `SynthesizeExpression` on unit expression. Follows established pattern from Story228.
- **Task 5** (Verification and regression): Medium — creates test CPL files, runs end-to-end pipeline, verifies no regressions.

### Parallel Execution

All tasks are sequential. Each task depends on the output of the previous task.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

- Task 2 depends on Task 1 (grammar must be updated before listener changes compile).
- Task 3 depends on Task 2 (parser must push postfix definitions before the analyzer can build expression trees).
- Task 4 depends on Task 3 (analyzer must have `UnitExpression` field populated before synthesis can emit compound unit code).
- Task 5 depends on all preceding tasks.

### Reintegration

None — no parallel groups. All tasks run sequentially on the feature branch.

### Task 1: Update CPLMeasures.g4 to use expression for unit spec and regenerate ANTLR files — Not Started

#### Intent
Update the measures grammar to accept compound unit expressions using the shared `arithmeticExpr` rules from Story228.

#### Subtasks

a. Read `CPLMeasures.g4` (located at `src/main/antlr/CPLMeasures.g4` relative to source_root). Understand the current `measureDeclaration` rule.

b. Change the grammar to route `unit=` through the shared expression grammar:
   - Replace `constraintList` with a `measureParameterList` that matches `unitParameter` (using `arithmeticExpr`) before generic `constraint` rules.
   - The `unitParameter` rule: `unitParameter: UNIT EQUALS arithmeticExpr;`
   - Keep the existing `constraint` rule as a fallback for non-unit parameters (e.g., `starts_at`).

c. Regenerate ANTLR Go files from updated grammar. The ANTLR grammar files live in `src/main/antlr/` and generate into `src/main/go/parser/antlr/`. Find the regeneration command (check for a Makefile or shell script in the project).

d. Verify generated files compile: `go build ./...` from source_root.

#### Constraints
- Expression grammar rules from Story228 (in CPLExpressions.g4) are NOT modified — only wired into measure definitions.
- Must maintain backward compatibility: simple `unit=Identifier` must still work with the existing constraint-based analysis path until Task 2 and Task 3 wire the expression path.

#### Files to modify
- `src/main/antlr/CPLMeasures.g4`

### Task 2: Update parser listener to push postfix definitions into MeasureScope — Not Started

#### Intent
Wire the parser so that `arithmeticExpr` components inside `unitParameter` (within a `MeasureScope`) accumulate as postfix definitions on the `MeasureScope`'s `Definitions` slice, and remove the obsolete `ConstraintScope`-based unit extraction path.

#### Key Implementation Details

The expression listener code already exists from Story228. The problem is that `shouldPushExpressionDefs()` does not include `MeasureScope` in its whitelist. The following changes are needed:

a. **Add `MeasureScope` to `shouldPushExpressionDefs()`** in `src/main/go/parser/parser_support.go`:
   - Add `symboltable.MeasureScope` to the switch statement in `shouldPushExpressionDefs()`.
   - This causes `ExitIdentifier` (for expression operands), `ExitArithmeticExpr`, `ExitArithmeticTerm`, `ExitArithmeticFactor`, and all operator exits to push postfix definitions to the `MeasureScope`.

b. **Add `EnterUnitParameter` / `ExitUnitParameter` listener methods** if needed by the grammar:
   - The ANTLR-generated base listener will have no-op implementations. Check if additional scaffolding is needed to ensure the `MeasureScope` context is correct when `arithmeticExpr` is entered from `unitParameter`.
   - If the grammar routes `unitParameter` directly to `arithmeticExpr` without scope changes, no special listener methods are needed — the existing expression exits will work because `shouldPushExpressionDefs()` now includes `MeasureScope`.

c. **Remove ConstraintScope push from `EnterConstraint`** when inside a measure context:
   - The `EnterConstraint` listener in `cpl_document_listener.go` pushes a `ConstraintScope` child. For measure unit parameters, the grammar now routes `unit=` through `unitParameter` → `arithmeticExpr`, not through the `constraint` rule. Verify that simple `unit=Identifier` no longer triggers `EnterConstraint`.
   - If the grammar still has `constraint` as an alternative, ensure the parser greedily matches `unitParameter` first (via grammar ordering) so `unit=` never hits the `constraint` path.
   - The `constraint` rule should remain for non-unit measure parameters (e.g., `starts_at`).

d. **Verify the `ExitIdentifier` expression-path logic**: When `shouldPushExpressionDefs()` returns true for `MeasureScope`, `ExitIdentifier` will try to cast its parent to `*antlr.ArithmeticFactorContext` (line 98 of `cpl_document_listener.go`). Verify this works correctly when the identifier is an operand of `arithmeticExpr` inside `unitParameter`.

e. Write listener tests:
   - Parse `Speed = Number(unit=Length)` and verify the `MeasureScope` has 2 identifier definitions (measure name and type name) PLUS postfix definitions: one `UnitReferenceDefinition` for "Length".
   - Parse `Speed = Number(unit=Length/Time)` and verify postfix includes `UnitReferenceDefinition("Length")`, `UnitReferenceDefinition("Time")`, and `DivideOperatorDefinition`.
   - Verify existing measures parser tests still pass.

f. Verify parser tests pass: `go test ./src/main/go/parser/... -count=1 -timeout 60s` from source_root.

#### Constraints
- `ArithmeticExpr` listener code from Story228 is NOT modified — only `shouldPushExpressionDefs()` scope whitelist and new listener wiring.
- Simple unit references must produce the same postfix output as before: a single `UnitReferenceDefinition`.

#### Files to modify
- `src/main/go/parser/parser_support.go` — add `MeasureScope` to `shouldPushExpressionDefs()`
- `src/main/go/parser/cpl_measures_listener.go` — remove/reduce ConstraintScope handling if needed
- `src/main/go/parser/cpl_document_listener.go` — add `unitParameter` listener methods if needed

### Task 3: Update analyzer to build and store expression trees for measures — Completed

#### Intent
Update the analyzer to build expression trees from postfix measure scope definitions and validate unit references. Update `AnalysisMeasureDefinition` to store expression trees instead of (or in addition to) plain unit name strings. Update synthesis code that reads `UnitName` from measure definitions.

#### Key Implementation Details

**Pre-work**: Read these files fully before starting:
- `src/main/go/analysis/measures_analysis.go` — `AnalyzeMeasureScope`, `AnalysisMeasureDefinition`
- `src/main/go/analysis/build_expression_tree.go` — `BuildExpressionTree` signature and mechanics
- `src/main/go/analysis/expression.go` — Expression types (UnitReference, MultiplicationNode, etc.)
- `src/main/go/synthesis/synthesize.go` — all call sites using `g.UnitName` on `*AnalysisMeasureDefinition`
- `src/main/go/parser/parser_support.go` — `shouldPushExpressionDefs()` (should already include MeasureScope from Task 2)

#### Subtasks

a. **Add `UnitExpression Expression` field to `AnalysisMeasureDefinition`** (in `measures_analysis.go`):
   - Keep `UnitName string` field during migration — do NOT remove it yet.
   - Add `UnitExpression Expression` field alongside it.
   - Add a helper function: `func GetUnitNameFromExpression(expr Expression) string` that extracts the unit name from a `*UnitReference`, or returns `""` for compound expressions or nil.

b. **Update `AnalyzeMeasureScope`** to build expression trees from postfix definitions:
   - After Task 2, `MeasureScope.Definitions` will contain: [measureName, typeName, ...postfix components...].
   - When `len(table.Definitions) > 2`, extract `table.Definitions[2:]` and pass to `BuildExpressionTree`.
   - Store the result in `AnalysisMeasureDefinition.UnitExpression`.
   - Also set `UnitName` using `GetUnitNameFromExpression()` for backward compatibility during migration.
   - When `len(table.Definitions) == 2` (no expression), keep the existing ConstraintScope fallback for backward compatibility.
   - Update the validation from `len(table.Definitions) != 2` to `len(table.Definitions) < 2`.

c. **Add `VerifyMeasureUnitReferences`**: A new function that recursively walks expression trees and validates all `UnitReference` nodes resolve to declared units. Called after all definitions are collected during `Analyze`.

d. **Update synthesis code**: There are 4 call sites in `src/main/go/synthesis/synthesize.go` that read `g.UnitName` where `g` is `*AnalysisMeasureDefinition`:
   - Line 117-118 (`grainToConcept[g.Name] = g.UnitName`)
   - Line 295-296 (`result[g.UnitName]`)
   - Line 463 (`g.UnitName == ""`)
   - Line 466 (`g.UnitName`)
   
   Replace each with `GetUnitNameFromExpression(g.UnitExpression)`.

e. **Write analyzer tests**:
   - Test case: Simple unit → single `UnitReference` expression, `GetUnitNameFromExpression` returns correct name.
   - Test case: Compound unit `Length/Time` → `DivisionNode` with two `UnitReference` children.
   - Test case: Power `Length/Time^2` → `DivisionNode` with `PowerNode` child.
   - Test case: Nested `Mass/(Length^2)` → `DivisionNode` with `PowerNode` child.
   - Test case: Invalid unit reference produces error from `BuildExpressionTree`.
   - Test `VerifyMeasureUnitReferences` with valid and invalid trees.

f. **Verify**: `go test ./src/main/go/analysis/... ./src/main/go/synthesis/... -count=1 -timeout 60s` from source_root.

g. **Update test files**: Find and update all test file references to `UnitName` on `AnalysisMeasureDefinition` struct literals. Replace `UnitName: "X"` with `UnitExpression: &analysis.UnitReference{Name: "X"}`. Only fix `AnalysisMeasureDefinition` struct literals — do not change `UnitName` fields on other struct types (e.g., `IntervalConstraints`, `GrainMeasure`, `ConceptScope`).

#### Constraints
- Expression trees preserve units symbolically — no dimensional simplification.
- Simple unit references must continue to work exactly as before.
- `UnitName` field is kept during migration for backward compatibility; removal is deferred.

#### Files to modify
- `src/main/go/analysis/measures_analysis.go` — struct change, `AnalyzeMeasureScope`, `VerifyMeasureUnitReferences`, `GetUnitNameFromExpression`
- `src/main/go/synthesis/synthesize.go` — 4 call sites
- Test files across `analysis/` and `synthesis/` — `UnitName` → `UnitExpression` migration

### Task 4: Update measure synthesis for compound unit expressions — Completed [Small]

#### Intent
Update synthesis to emit Go code for compound unit expressions via the shared `SynthesizeExpression` from Story228.

#### Subtasks

a. Add `UnitExpression cil.CILNode` to the measure's CIL structure (or wherever measure units are represented in the synthesis data model).

b. Update `SynthesizeMeasure` (or equivalent) to call `SynthesizeExpression` on the `UnitExpression` when present. The function from Story228 converts an `Expression` tree into a `CILNode` tree suitable for code generation.

c. Write synthesis unit tests verifying emitted Go code for simple and compound unit expressions.

d. Verify synthesis tests pass: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` from source_root.

#### Constraints
- Follows the existing `SynthesizeExpression` pattern from Story228.

#### Files to modify
- `src/main/go/synthesis/synthesize.go` (or its sub-files)

### Task 5: Verification with compound unit examples and regression — Completed

#### Intent
Create test CPL files exercising a range of compound unit expression forms and verify the full parse → analyze → synthesize pipeline produces correct output with no regressions.

#### Subtasks

a. Create test CPL with simple unit: `Speed = Number(unit=Length)` — verify produces single `UnitReference` expression tree — Not Started
b. Create test CPL with compound unit: `Speed = Number(unit=Length/Time)` — verify produces `DivisionNode` — Not Started
c. Create test CPL with power: `Acceleration = Number(unit=Length/Time^2)` — verify produces `DivisionNode` with `PowerNode` child — Not Started
d. Create test CPL with multiplication: `Momentum = Number(unit=Mass*Length/Time)` — verify correct tree — Not Started
e. Create test CPL with parentheses: `AreaDensity = Number(unit=Mass/(Length^2))` — verify correct tree — Not Started
f. Run parse → analyze → synthesize pipeline for each example and verify output — Not Started
g. Verify all existing measure tests still pass (no regression on Increment 2 behavior) — Not Started
h. Run full CPL test suite: `go test ./... -count=1` from source_root — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Expression grammar rules from Story228 are NOT modified — only wired into measure definitions
3. Expression trees preserve units symbolically — no dimensional simplification (that is Story231's job)
4. Simple unit references (single identifier, no operators) must continue to work exactly as before
5. `ConstraintScope`-based unit extraction is fully removed from measures — no dual paths
6. Go functions that can fail must return `(T, error)`
7. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Update the measures grammar to route `unit=` through the shared expression grammar, keeping `constraint` as fallback for non-unit parameters
- **Task 2 Intent**: Wire the parser so expression postfix definitions accumulate in `MeasureScope` and remove the ConstraintScope-based unit extraction path
- **Task 3 Intent**: Update the analyzer to build expression trees from postfix definitions, add expression-tree-aware unit reference validation, and update synthesis call sites
- **Task 4 Intent**: Update synthesis to emit Go code for compound unit expressions via the shared `SynthesizeExpression`
- **Task 5 Intent**: Verify compound unit parsing, analysis, and synthesis work for a range of expression forms with no regressions

## Acceptance Criteria

1. `CPLMeasures.g4` routes `unit=` through `arithmeticExpr`
2. `MeasureScope` definitions contain postfix expression components for unit specs
3. `AnalysisMeasureDefinition` stores an `Expression` tree in `UnitExpression`
4. `AnalyzeMeasureScope` calls `BuildExpressionTree` and validates all unit references in the tree
5. `SynthesizeMeasure` emits correct Go code for both simple and compound unit expressions
6. Test examples pass for: simple unit, division, power, multiplication, parentheses
7. All existing measure tests pass with zero regressions
8. Full CPL test suite passes

## Footer

### Open Questions

None

### Known Issues

None

### Additional Context

This story is a bridge story: it takes the generic expression infrastructure from Story228 and applies it specifically to measure unit specifications. The key change is replacing `UnitName string` with `UnitExpression Expression` in `AnalysisMeasureDefinition`. This is a breaking change to the analysis data structure, but since Increments 1–2 are already implemented in code, the update is contained within the existing codebase.

The `ConstraintScope` type may still exist in the symbol table for other purposes (e.g., instant constraints in type definitions). Only its use in measures is removed.

### Previous Attempts (Session 2026-06-10)

The original story had `[Task 1 ‖ Task 3] → Task 2` execution order. Three sub-agent attempts on Task 3 failed because:

1. **Task 1** changed the grammar but did NOT add `MeasureScope` to `shouldPushExpressionDefs()` — that's a parser change, not a grammar change. The grammar change alone breaks backward compatibility because it no longer produces `ConstraintScope` children for `unit=`, and the expression postfix defs are silently dropped.
2. **Task 3** was attempted before Task 2, so the `MeasureScope` had no postfix definitions to build expression trees from. The analyzer had no data to work with.
3. **Task 2** (parser wiring — adding `MeasureScope` to `shouldPushExpressionDefs()`, adding listener methods for `unitParameter`, removing ConstraintScope for measures) was the missing keystone that both Task 1 and Task 3 needed.

The fix: reorder to fully sequential `Task 1 → Task 2 → Task 3 → Task 4 → Task 5`. Task 2 bridges the grammar change (Task 1) and the analysis change (Task 3).
