# Story232: CPL Conversion Injection

## Goal

Implement automatic conversion injection into CIL expression trees as part of [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md). When the unit type checker (Story231) identifies unit mismatches between connector inputs, intermediate calculations, or IO spec output units, and a valid conversion pair exists to bridge the gap, inject a `CILConversionApplication` node into the CIL tree so the synthesizer produces code that performs the conversion at runtime. Works with Story231: the checker identifies gaps; this story fills them.

## References

- [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md)
- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — provides CIL expression types with unit tracking
- [Story229_cpl-compound-units.md](Story229_cpl-compound-units.md) — provides compound unit measures
- [Story230_cpl-conversions.md](finished-stories/Story230_cpl-conversions.md) — provides conversion pairs in CIL
- [Story231_cpl-unit-type-checking.md](Story231_cpl-unit-type-checking.md) — provides unit mismatch identification and `ComputeUnitExpression`
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — _Done_ — provides expression grammar, `BuildExpressionTree`, and `SynthesizeExpression`
- [Story235_unify-cilnode-cilexpression-hierarchies.md](finished-stories/Story235_unify-cilnode-cilexpression-hierarchies.md) — _Done_ — unifies CIL expression types into the `CILNode` hierarchy (transitive dependency through Story231)
- [Story229_cpl-compound-units.md](Story229_cpl-compound-units.md) — _Not Started_ — provides compound unit measures
- [Story230_cpl-conversions.md](finished-stories/Story230_cpl-conversions.md) — _Done_ — provides `CILConversionPair` and `CILMeasureConversionPair`
- [Story231_cpl-unit-type-checking.md](finished-stories/Story231_cpl-unit-type-checking.md) — _Done (code complete; NOT wired into pipeline)_ — provides `ComputeUnitExpression` and mismatch identification

## Dependent Stories

None

## Tasks

### Task Sizing

- **Task 1** (Define CILConversionApplication node): `[Small]` — single struct definition extending existing CIL expression types. Straightforward.
- **Task 2** (Implement InjectConversions algorithm): Medium — recursive bottom-up tree walk with unit computation and conversion lookup. Core algorithm of the story.
- **Task 3** (Update SynthesizeExpression for conversion nodes): `[Small]` — single function update to handle `CILConversionApplication` by emitting conversion function calls. Follows established pattern.
- **Task 4** (Wire injection into synthesis pipeline): `[Small]` — single integration point between verification (Story231) and code generation. Narrow scope.
- **Task 5** (Verification with injection examples): Medium — end-to-end tests requiring all preceding stories.

### Parallel Execution

- **Group A** (Tasks 1, 3): Task 1 defines the `CILConversionApplication` node type in `cil/` package; Task 3 updates `SynthesizeExpression` in `synthesis/` package to handle the new node. These are disjoint packages. Task 3 can be implemented with test stubs of the new node type while Task 1 proceeds in parallel.

### Execution Order

```
[Task 1 ‖ Task 3] → Task 2 → Task 4 → Task 5
```

- Tasks 1 and 3 can run in parallel because they are in disjoint packages (`cil/` vs `synthesis/`) and Task 3 only needs the node type signature (available once Task 1's struct is defined, which is immediate).
- Task 2 depends on Task 1 (needs `CILConversionApplication` defined).
- Task 4 depends on Task 2 (needs injection algorithm) and Task 3 (needs synthesis support for injected nodes).
- Task 5 depends on all preceding tasks.

### Reintegration

**Group A (Tasks 1, 3):**
- Merge order: any order (disjoint packages — `cil/` vs `synthesis/`)
- Integration tests: `go build ./src/main/go/cil/... ./src/main/go/synthesis/...` — verify both packages compile together
- Watch for conflicts in: none expected

### Task 1: Define CILConversionApplication node type — Not Started [Small]
   a. Define `CILConversionApplication` struct implementing `CILExpression`:
      ```go
      type CILConversionApplication struct {
          Operand      CILExpression
          Conversion   CILConversionPair
          FromUnit     string
          ToUnit       string
      }
      ```
   b. Ensure the node type integrates cleanly with existing `CILExpression` interface — Not Started
   c. Write unit tests for node construction — Not Started
   d. Verify tests pass — Not Started

### Task 2: Implement InjectConversions algorithm — Not Started
   a. Write tests for `InjectConversions` covering:
      - Add/Subtract operand mismatch with available conversion
      - IO spec mismatch with available conversion
      - Multiple mismatches in one expression
      - No injection when units are consistent
      - Pass-through when mismatch has no conversion — Not Started
   b. Implement `InjectConversions(root CILExpression, conversions []CILConversionPair, ioSpecUnit string) (CILExpression, error)`:
      - Walk CIL expression tree bottom-up, computing unit expressions for each subtree — Not Started
      - For `CILBinaryOp Add/Subtract` with differing operand units: search conversion pairs and wrap mismatched operand in `CILConversionApplication` — Not Started
      - After processing all operations, compare final unit to IO spec target: wrap root in `CILConversionApplication` if conversion exists — Not Started
      - Return modified CIL tree with injected conversion nodes — Not Started
   c. Verify injection algorithm tests pass — Not Started

### Task 3: Update SynthesizeExpression for conversion application nodes — Not Started [Small]
   a. Write tests for `SynthesizeExpression` with `CILConversionApplication` nodes — Not Started
   b. Update `SynthesizeExpression` to handle `CILConversionApplication`: emit call to generated conversion function (e.g., `convert_USD_to_EUR(subExpression)`) — Not Started
   c. Verify conversion function signatures match those generated by Story230's synthesis — Not Started
   d. Verify synthesis tests pass — Not Started

### Task 4: Wire injection into synthesis pipeline — Not Started [Small]
   a. Add injection pass between verification (Story231) and code generation in the main synthesis pipeline — Not Started
   b. Ensure injected tree is passed to `SynthesizeExpression` instead of original tree — Not Started
   c. Write integration test verifying the full verify → inject → synthesize sequence — Not Started
   d. Verify integration test passes — Not Started

### Task 5: Verification with injection examples and regression — Not Started
   a. IO spec mismatch: connector produces EUR, metric computes EUR, IO spec requires USD, EUR↔USD conversion exists → verify synthesis emits `convert_EUR_to_USD(...)` — Not Started
   b. Add/Subtract mismatch: `Total_Revenue = Revenue_USD + Revenue_EUR`, USD↔EUR conversion exists → verify one operand wrapped, synthesis produces `rev_USD + convert_EUR_to_USD(rev_EUR)` — Not Started
   c. Multiple mismatches: expression with multiple add/subtract mismatches → verify each independently resolved — Not Started
   d. No unnecessary injections: expression with consistent units → verify CIL tree unchanged — Not Started
   e. Unfixable mismatch: units from different concepts without conversion → verify injection passes through (Story231 already reported error, synthesis fails gracefully) — Not Started
   f. Verify synthesis output compiles and produces correct results — Not Started
   g. Run full CPL test suite: `go test ./... -count=1` from `src/chronocone_planning_language/src/main/go/` — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Injection is a separate pass after verification in the synthesis pipeline
3. Only injects conversions when a conversion pair exists — never creates synthetic conversions
4. Preserves the original CIL tree structure except for wrapped subtrees
5. Injection is deterministic: given the same CIL tree and conversion pairs, always produces the same result
6. Go functions that can fail must return `(T, error)`
7. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Define the CIL node type that represents a conversion applied to a sub-expression
- **Task 2 Intent**: Build the core injection algorithm that walks CIL trees, identifies unit mismatches, and wraps mismatched sub-trees with conversion nodes
- **Task 3 Intent**: Update the synthesizer to emit conversion function calls for injected conversion nodes
- **Task 4 Intent**: Wire the injection pass into the main synthesis pipeline between verification and code generation
- **Task 5 Intent**: Verify conversions are injected correctly in realistic scenarios, never injected when unnecessary, and produce compilable output

## Acceptance Criteria

1. `CILConversionApplication` node type is defined and integrates with existing `CILExpression` interface
2. `InjectConversions` wraps mismatched `Add`/`Subtract` operands in `CILConversionApplication` when a conversion pair exists
3. `InjectConversions` wraps the entire expression root when final unit mismatches IO spec and a conversion exists
4. `SynthesizeExpression` emits valid Go code for `CILConversionApplication` nodes
5. No conversion nodes are added when units are consistent throughout the expression
6. Unfixable mismatches pass through unchanged (Story231 reports the error)
7. Synthesis output with injected conversions compiles and produces correct runtime results
8. Full CPL test suite passes

## Footer

### Open Questions

None

### Known Issues

None

### Additional Context

**Design foundation:**

- Operates on the same CIL expression trees with unit annotations built by Stories 228–230.
- Uses conversion pairs (`CILConversionPair`, `CILMeasureConversionPair`) from Story230.
- Is invoked after verification (Story231) identifies mismatches — the checker reports errors for unfixable mismatches (different unit groups, missing conversions); this story injects nodes for fixable ones.
- Conversion reversibility (Story230 algebraic reversal) means every forward pair has an inverse — injection can match in either direction.

**Explicitly out of scope:**
- Multi-step conversion path resolution (chaining Kilometer→Meter→Centimeter when Kilometer→Centimeter is needed but no direct conversion exists)
- Conversion optimization (choosing between multiple valid conversion paths based on cost or precision)
- Injecting conversions at analysis time (only during synthesis phase, on CIL)
