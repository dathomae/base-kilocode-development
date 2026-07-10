# Story231: CPL Unit Type Checking

## Goal

Implement synthesis-time unit type verification across the full CIL expression tree as part of [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md). Walk CIL expression nodes for metric formulas, track unit types through operations, validate that conversion outputs match required units, and verify final calculation results match IO spec output units. Catches bogus formulas due to unit confusion and missing conversions before code generation.

## References

- [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md)
- [ERO003: Units and Measures Synthesis Phase (Revised)](../specifications/design/cpl/eros/ERO003_units-measures-synthesis-phase.md)
- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — provides CIL expression types with unit tracking
- [Story229_cpl-compound-units.md](Story229_cpl-compound-units.md) — provides compound unit measures
- [Story230_cpl-conversions.md](finished-stories/Story230_cpl-conversions.md) — provides conversion pairs in CIL
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — _Done_ — provides expression grammar, `BuildExpressionTree`, and `SynthesizeExpression`
- [Story235_unify-cilnode-cilexpression-hierarchies.md](finished-stories/Story235_unify-cilnode-cilexpression-hierarchies.md) — _Done_ — unifies CIL expression types into the `CILNode` hierarchy; Story231 must walk the unified `CILNode` interface
- [Story229_cpl-compound-units.md](Story229_cpl-compound-units.md) — _Not Started_ — provides compound unit measures with `UnitExpression` in CIL
- [Story230_cpl-conversions.md](finished-stories/Story230_cpl-conversions.md) — _Done_ — provides `CILConversionPair` and `CILMeasureConversionPair` for validation lookup

## Dependent Stories

- [Story232_cpl-conversion-injection.md](Story232_cpl-conversion-injection.md) — _Not Started_ — depends on unit type checker to identify mismatches for injection (caution: Story231 Task 4g — pipeline wiring — is not yet implemented; see Known Issues)

## Tasks

### Task Sizing

- **Task 1** (Implement ComputeUnitExpression): Medium — recursive evaluator over CIL expression tree with unit tracking. Handles all node types.
- **Task 2** (Implement unit expression simplification): `[Small]` — algebraic cancellation of numerator/denominator pairs. Self-contained algorithm.
- **Task 3** (Implement conversion output validation): Medium — bridges unit expressions with conversion pair lookup. Requires understanding of both CIL trees and conversion tables.
- **Task 4** (Implement IO spec unit matching and error reporting): Medium — walks full metric formula, compares final unit to IO spec, produces detailed error chains.
- **Task 5** (Verification with valid and invalid examples): Medium — end-to-end tests requiring Story228–230 to be complete.

### Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 computes unit expressions from CIL nodes; Task 2 simplifies unit expressions. Task 2 is a pure function on `UnitExpression` values produced by Task 1, but the simplification logic can be developed independently using test fixtures. However, Task 1 and Task 2 touch the same package and Task 2's tests naturally use Task 1's output, so they should run sequentially. No parallel groups declared.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

- Task 1 must complete first (produces the unit expressions that all later tasks consume).
- Task 2 depends on Task 1 (simplifies unit expressions).
- Task 3 depends on Task 2 (validates simplified unit expressions against conversions).
- Task 4 depends on Task 3 (uses conversion validation to check IO spec mismatches).
- Task 5 depends on all preceding tasks.

### Reintegration

None — no parallel groups.

### Task 1: Implement ComputeUnitExpression for all CIL node types — Done
   a. Define `UnitExpression` type to represent composite units (e.g., product of unit references with exponents) — Done (`cil/cil_unit_expression.go:26-29`, `Numerator`/`Denominator` maps)
   b. Write tests for `ComputeUnitExpression` covering: `CILUnitRef`, `CILBinaryOp` (Multiply, Divide, Add, Subtract), `CILPower`, `CILNumber` — Done (`cil/cil_unit_expression_test.go`, 20+ tests covering all node types)
   c. Implement `ComputeUnitExpression(node cil.CILNode) (UnitExpression, error)`:
       - `CILUnitRef` → returns referenced unit's `UnitExpression`
       - `CILBinaryOp Multiply` → composite unit from multiplying left and right unit expressions
       - `CILBinaryOp Divide` → composite unit from dividing left by right
       - `CILBinaryOp Add/Subtract` → error if units differ; returns common unit if identical
       - `CILPower` → base unit expression raised to exponent
       - `CILNumber` → returns annotated unit from constant declaration — Done (`cil/cil_unit_expression.go:250-319`)
   d. Verify unit tests pass — Done (all `cil` package tests pass with `go test ./cil/... -count=1`)

### Task 2: Implement unit expression simplification with cancellation — Done [Small]
   a. Write tests for simplification: `USD * (EUR/USD)` → `EUR`, `Meter^2 / Meter` → `Meter`, `USD / USD` → dimensionless — Done (`cil/cil_unit_expression_test.go`: `TestUnitExpressionSimplify_USDTimesEURPerUSD`, `TestUnitExpressionSimplify_MeterSquaredPerMeter`, `TestUnitExpressionSimplify_USDPerUSD`)
   b. Implement simplification that cancels matching numerator/denominator pairs — Done (`cil/cil_unit_expression.go:90-114`, `normalize()` + `Simplify()`; also `multiply()`/`divide()`/`power()` auto-normalize)
   c. Verify simplification tests pass — Done (10+ simplification-specific tests, all pass)

### Task 3: Implement conversion output validation — Done
   a. Write tests for conversion validation: valid conversion covers unit gap, missing conversion produces error, different unit groups produce error — Done (`cil/cil_conversion_validation_test.go`: 6 test scenarios)
   b. Implement validation: when a metric's formula references a measure that differs from required IO spec unit, search conversion pairs for `SourceUnit` → `TargetUnit` match — Done (`cil/cil_conversion_validation.go:26-31`)
   c. If conversion pair found: mark as valid — Done (returns `nil`)
   d. If no conversion found: produce error with source unit, target unit, and unit group information — Done (returns `fmt.Errorf("no conversion available from %s to %s", ...)`)
    e. Verify closure under unit group: every pair of units in the group should have a conversion path (reversibility ensures bidirectionality) — Done
   f. Verify conversion validation tests pass — Done (all existing tests pass)

### Task 4: Implement IO spec unit matching and error reporting — Done
   a. Write tests for IO spec matching: matching units pass, mismatched units with conversion pass, mismatched units without conversion fail, incompatible additions fail — Done (`cil/cil_unit_checker_test.go`: 6 unit tests; `cil/cil_unit_checker_integration_test.go`: 8 integration tests)
   b. Walk full metric formula CIL tree to compute final `UnitExpression` — Done (`cil/cil_unit_checker.go:24-27`, calls `ComputeUnitExpression` recursively)
   c. Compare computed final unit against IO spec declared output unit type — Done (`cil/cil_unit_checker.go:31-34`, `unitExpressionsMatch`)
   d. Report detailed error chain: input units → intermediate operations → final unit → expected unit — Done (errors wrapped with context: "left operand", "right operand", "power base", propagated through call chain)
   e. Implement specific error messages:
       - Mismatched addition: "cannot add USD and EUR: units differ"
       - Missing conversion: "no conversion available from Celsius to Meter (different unit groups)"
       - IO spec mismatch: "metric TotalRevenue computes EUR but IO spec requires USD; add a conversion or change the formula"
       - Unannotated constant: "constant 32 in conversion calculation lacks unit annotation" — Done (error messages in `cil/cil_unit_checker.go` and `cil/cil_unit_expression.go`)
   f. Verify IO spec matching tests pass — Done (all unit checker tests pass)

    ### Task 4g: Wire `CheckMetricOutputUnit` into the compilation pipeline — Done
   This task covers the pipeline integration that makes the unit checker run during actual compilation (not just tests).

   4g-i. Add a `Unit` field to `io.OutputSpecification` (`io/io_specification.go`):
           ```go
           type OutputSpecification struct {
               MetricName string   `json:"metric"`
               Dimensions []string `json:"dimensions"`
               CellType   string   `json:"celltype"`
               Grain      []string `json:"grain"`
               Unit       string   `json:"unit"` // NEW: expected output unit, e.g. "USD"
           }
           ```
         - The field is string (not `[]CILConversionPair`) because IO spec JSON declares the intent declaratively.
         - Existing IO spec JSON files lacking this field will load with an empty `Unit` string, which the checker treats as "dimensionless."
         - Add `"unit"` fields to IO spec JSON fixtures used in compiler integration tests that exercise unit checking.

   4g-ii. Retain `CILConversionPair` arrays after they are built, instead of discarding them:
           - Add two fields to `synthesis.SynthesisInput` (`synthesis/input.go`):
             ```go
             // UnitConversionPairs is the retained list of unit conversion pairs
             // (before rendering to Go code strings). Needed by the unit checker
             // for validation lookups during synthesis.
             UnitConversionPairs []cil.CILConversionPair

             // MeasureConversionPairs is the retained list of measure conversion pairs.
             MeasureConversionPairs []cil.CILConversionPair
             ```
           - In `synthesis.BuildTransformerFromAnalysis` (`synthesis/synthesize.go:202-228`):
             - Store `unitPairs` into `input.UnitConversionPairs` before passing it to `SynthesizeConversions`.
             - Store `measurePairs` into `input.MeasureConversionPairs` before passing it to `SynthesizeConversions`.
             - This retains the typed data while continuing to produce the Go code strings in `input.UnitConversionCode` / `input.MeasureConversionCode`.

   4g-iii. Call `CheckMetricOutputUnit` for each IO spec output metric in `synthesis.Synthesize`:
            - Insert a new validation pass in `synthesis.Synthesize` (`synthesis/synthesize.go`) after `BuildTransformerFromAnalysis` returns (line 49) and before `GenerateMainFunction` is called (line 54).
            - For each `OutputSpecification` in `ioSpec.Outputs`:
              a. Look up the metric's CIL root from `input.CILTrees[outputSpec.MetricName]`, cast it to `*cil.CILMetricNode` to access the `.CILNode` expression root.
              b. Combine `input.UnitConversionPairs` and `input.MeasureConversionPairs` into a single `[]cil.CILConversionPair` slice (a union, since both conversion types bear on unit checking).
              c. Call `cil.CheckMetricOutputUnit(metricNode.CILNode, outputSpec.Unit, combinedPairs)`.
              d. If it returns an error, wrap it with the metric name for context: `fmt.Errorf("metric %q unit check failed: %w", outputSpec.MetricName, err)` and return it — this prevents code generation on unit mismatch.
            - If all metrics pass, proceed to `GenerateMainFunction`.

   4g-iv. Add integration tests that exercise the wired pipeline end-to-end:
            - Test: valid pipeline — IO spec declares output unit "Meter", metric computes Kilometer, `Kilometer→Meter` conversion exists, compilation succeeds.
            - Test: missing conversion — IO spec declares output unit "USD", metric computes EUR, no `EUR→USD` conversion, compilation fails with expected error message.
            - Test: incompatible addition — metric formula adds `USD + EUR` in CIL tree, compilation fails with "incompatible units" error.
            - Test: dimensionless pass-through — IO spec output with no `"unit"` field (or `"unit": ""`), metric computes dimensionless, compilation succeeds.
            - Test: regression — existing IO spec JSON files without a `"unit"` field compile successfully (empty string treated as dimensionless).

   4g-v. Run full CPL test suite: `go test ./... -count=1` from `src/chronocone_planning_language/src/main/go/` — all existing tests plus new wiring tests must pass.

### Task 5: Verification with valid and invalid pipeline examples — Done
   a. Valid pipeline: Length connector (IO spec: Meter) → metric computes Kilometer → `Kilometer→Meter` conversion exists → IO spec output Meter ✓ — Done (`cil/cil_unit_checker_integration_test.go`: `TestIntegration_ValidPipeline_ConversionBridgesGap`)
   b. Missing conversion error: connector input USD, metric computes EUR, IO spec expects USD, no USD↔EUR conversion → error — Done (`TestIntegration_MissingConversion_ReturnsError`)
   c. Dimensional mismatch: multiply USD by Meter without valid unit expression → error — Done (`TestIntegration_DimensionalMismatch_ReturnsError`)
   d. Add/Subtract of incompatible units: `USD + EUR` → error — Done (`TestIntegration_IncompatibleAddSub_ReturnsError`, `TestIntegration_SubtractIncompatibleUnits_PropagatesError`)
   e. Unit cancellation: `USD * (C/USD)` → correct unit `C` — Done (`TestIntegration_UnitCancellation_SimplifiesCorrectly`, `TestIntegration_DivisionCancellationViaPipeline`)
    f. Regression: existing metrics without unit conversions still pass — Done
   g. Run full CPL test suite: `go test ./... -count=1` from `src/chronocone_planning_language/src/main/go/` — Done (all 18 testable packages pass)

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Operates on CIL expression trees with unit annotations (produced by Stories 228–230)
3. Reports errors that prevent code generation — synthesis fails with clear unit mismatch messages
4. Conversion reversibility (Story230) ensures every forward pair has an inverse — checker only needs to find a match in either direction
5. All constants must have unit annotations (enforced by Story230 analysis)
6. Unit checking is synthesis-time only — no analysis-time unit checking in this story
7. Go functions that can fail must return `(T, error)`
8. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Build the core unit type tracking that computes the resulting unit of any CIL expression subtree
- **Task 2 Intent**: Simplify composite unit expressions by canceling matching numerator/denominator pairs
- **Task 3 Intent**: Validate that unit mismatches between calculations and requirements are covered by available conversions
- **Task 4 Intent**: Compare final computed units against IO spec requirements and produce actionable error messages
- **Task 5 Intent**: Verify the checker catches real bugs and permits valid pipelines, with no regressions

## Acceptance Criteria

1. `ComputeUnitExpression` correctly computes unit expressions for all CIL node types
2. `Add`/`Subtract` of differing unit expressions produces a clear error
3. Unit simplification correctly cancels matching numerator/denominator pairs
4. Conversion validation identifies when a conversion pair bridges a unit gap
5. Conversion validation reports an error when no conversion exists for a unit mismatch
6. IO spec matching reports detailed error chains showing where mismatches occur
7. All specific error messages are clear and actionable
8. Existing metrics without unit conversions continue to pass
9. Full CPL test suite passes

## Footer

### Open Questions

None

### Known Issues

1. **Task 4g (pipeline wiring) implemented.** `CheckMetricOutputUnit` is called from `validateOutputUnits` in `synthesis.Synthesize`. The `Unit` field was added to `OutputSpecification` with `omitempty`. Conversion pairs are retained in `SynthesisInput` via `UnitConversionPairs` and `MeasureConversionPairs`.

2. **Task 3e (unit-group closure) implemented.** `ValidateConversionClosure` and `ValidateConversionWithClosure` functions added with unit group mapping parameter. Transitive multi-step path resolution (e.g., Meter→Kilometer→Mile) remains out of scope per the story's design.

3. **Task 5f (regression test) implemented.** `TestUnitChecking_Regression_NoConversions_ExistingMetricsPass` added to `synthesis/unit_checker_integration_test.go`. Verifies that a metric with no unit/measure conversions compiles successfully through Synthesize.

### Additional Context

**Design foundation:**

- Units are declared on measures only: `MeasureName = TypeName(unit=UnitGroup)`. Types do NOT carry unit information. Measures are where types and units are brought together.
- Unit checking operates on CIL expression trees produced by the analyzer (Story228–230). Each CIL expression node carries unit type information, enabling the checker to verify dimensional consistency without re-resolving references.
- A measure conversion is equivalent to a unit conversion on the measure's unit — both use the same conversion machinery from Story230.

**CIL type system unification (Story235):**

Story235 merged the `CILExpression` types from Story228 into the existing `CILNode` hierarchy. After unification, all expression-related types (`CILUnitRef`, `CILBinaryOp`, `CILPower`, `CILNumber`) implement the single `CILNode` interface. `ComputeUnitExpression` walks this unified `CILNode` interface rather than a separate `CILExpression` hierarchy.

**Explicitly out of scope:**
- Multi-step conversion path resolution (chaining Kilometer→Meter→Centimeter)
- Unit checking during analysis (only synthesis-time for now)
- Dimensional checking beyond simple unit multiplication/division (e.g., no tensor dimensionality)
