# Story236: Expression Pipeline Integration Test and Minor Cleanup

## Goal

Deliver the missing end-to-end integration test for the Story228 expression infrastructure (parse → analyze → convert → synthesize). Also fix two minor code hygiene issues flagged in the Story228 code review: add a grammar coupling comment in `ExitIdentifier` and remove the unused `operatorTypeForDefinitionType` function.

This story depends on Story235 because the integration test's conversion function and `SynthesizeExpression` both target the unified `CILNode` types.

## References

- [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md)
- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — provides expression grammar, `BuildExpressionTree`, CIL expression types, `SynthesizeExpression`
- [Story235_unify-cilnode-cilexpression-hierarchies.md](Story235_unify-cilnode-cilexpression-hierarchies.md) — provides unified `CILNode` types that this story's integration test targets
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — _Done_ — provides the expression pipeline components being tested
- [Story235_unify-cilnode-cilexpression-hierarchies.md](Story235_unify-cilnode-cilexpression-hierarchies.md) — _Not Started_ — the integration test targets unified `CILNode` types and `OperatorType` enum

## Dependent Stories

None

## Tasks

### Task Sizing

- **Tasks 1 and 2** are `[Small]`: each modifies a single file with a narrow, straightforward change (comment only, or dead-code removal).
- **Task 3** is Medium: creates a new test file with multiple test cases, a conversion helper, and a parse helper; then extracts the conversion helper to production code.
- **Task 4** is verification: runs the full test suite.

### Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 modifies `parser/cpl_document_listener.go`; Task 2 modifies `parser/parser_expression_types.go`. These are disjoint files in the same package with no semantic dependency.

### Execution Order

```
[Task 1 ‖ Task 2] → Task 3 → Task 4
```

- Tasks 1 and 2 can run in parallel because they touch disjoint files.
- Task 3 depends on Story235 being Done (needs unified `CILNode` types and `OperatorType` enum).
- Task 4 depends on Task 3.

### Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (disjoint files in the same package)
- Integration tests: `go test ./src/main/go/parser/... -count=1 -timeout 60s` — verify parser package compiles and passes with both changes
- Watch for conflicts in: none expected

### Task 1: Add ExitIdentifier grammar coupling comment — Completed
   a. Add a comment block inside `ExitIdentifier` — Completed
   b. Verify file compiles — Completed

### Task 2: Remove dead code operatorTypeForDefinitionType — Completed
   a. Remove `operatorTypeForDefinitionType` function — Completed
   b. Remove the `"fmt"` import if no longer used — Completed
   c. Verify package compiles and tests pass — Completed

### Task 3: Create end-to-end expression pipeline integration test — Completed
   a. Write parse helper — Completed
   b. Write conversion helper — Completed
   c. Write end-to-end pipeline tests — Completed
   d. Verify all pipeline tests pass — Completed
   e. Move `convertToCILExpression` to `analysis/expression_to_cil.go` — Completed
   f. Update test to import production function — Completed
   g. Verify tests pass after extraction — Completed

### Task 4: Run full test suite — Completed
   a. Run full test suite — Completed
   b. Fix any regressions — Completed (none found)

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. `ExpressionToCILNode` is a public function in the `analysis/` package — it is the contract Story229/230 will use to convert analysis expression trees into CIL nodes during measure analysis
3. `DerivedUnit` and `AnnotatedUnit` are left empty in the conversion helper; unit derivation belongs to Story229/Story231
4. Every emitted Go expression must parse successfully via `go/parser.ParseExpr`
5. Go functions that can fail must return `(T, error)`
6. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Document the grammar coupling in `ExitIdentifier` so future grammar changes do not silently break the arithmetic factor detection
- **Task 2 Intent**: Remove unused dead code to keep the parser package clean
- **Task 3 Intent**: Verify the full parse → analyze → convert → synthesize pipeline works end-to-end for all expression forms, and establish `ExpressionToCILNode` as the public conversion contract
- **Task 4 Intent**: Confirm no regressions anywhere in the CPL codebase after the integration test and cleanup changes

## Acceptance Criteria

1. `ExitIdentifier` has a comment documenting the grammar coupling to `arithmeticFactor` in `ChronoconePlanningLanguage.g4`
2. `operatorTypeForDefinitionType` function and its `"fmt"` import are removed
3. Integration test passes: parse → `BuildExpressionTree` → `ExpressionToCILNode` → `SynthesizeExpression` for all expression forms (unit ref, multiply, divide, power, add, subtract, nested)
4. Every emitted Go expression parses successfully via `go/parser.ParseExpr`
5. `ExpressionToCILNode` exists as production code in `analysis/expression_to_cil.go`
6. Full CPL test suite passes

## Footer

### Open Questions

None

### Known Issues

None

### Additional Context

**Conversion mapping (analysis.Expression → cil.CILNode):**

| `analysis.Expression` | Unified `cil.CILNode` type |
|---|---|
| `*analysis.UnitReference` | `*cil.CILUnitRef{UnitGroup: ref.Name}` |
| `*analysis.NumberNode` | `*cil.CILNumber{Value: n.Value}` |
| `*analysis.MultiplicationNode` | `*cil.CILBinaryOp{Op: cil.CILMultiply, Left: ..., Right: ...}` |
| `*analysis.DivisionNode` | `*cil.CILBinaryOp{Op: cil.CILDivide, Left: ..., Right: ...}` |
| `*analysis.PowerNode` | `*cil.CILPower{Base: ..., Exponent: p.Exponent}` |
| `*analysis.AdditionNode` | `*cil.CILBinaryOp{Op: cil.CILPlus, Left: ..., Right: ...}` |
| `*analysis.SubtractionNode` | `*cil.CILBinaryOp{Op: cil.CILMinus, Left: ..., Right: ...}` |

**Files created:**
- `src/main/go/synthesis/expression_pipeline_test.go`
- `src/main/go/analysis/expression_to_cil.go`

**Files modified:**
- `src/main/go/parser/cpl_document_listener.go` (comment only)
- `src/main/go/parser/parser_expression_types.go` (remove dead code)
