# Story225: CPL Refactoring — Phase 5 Address Fragility

## Goal

Restructure fragile implementation patterns: template-based code generation, conditional parsing state machine encapsulation, visitor pattern removal, and a stable analysis→synthesis contract.

## References

- [CPL Refactoring Plan](../plans/cpl_refactoring_plan.md)
- [Story224: Phase 4 Reduce Duplication](Story224_cpl-refactoring-phase4-reduce-duplication.md)

## Dependencies

- [Story224_cpl-refactoring-phase4-reduce-duplication.md](Story224_cpl-refactoring-phase4-reduce-duplication.md) — must be Done. Phase 4 introduces the shared operator registry and `IsScalar()`, which simplifies the codegen and visitor removal tasks.

## Dependent Stories

- [Story226_cpl-refactoring-phase6-test-hardening.md](Story226_cpl-refactoring-phase6-test-hardening.md) — Not Started

## Tasks

### Task Sizing

All tasks in this story are **not small** per the architect-for-story-planning sizing rules. None qualify for `[Small]` annotation:

- **Task 1a**: Touches 4+ packages, introduces a new `GeneratedProgram` struct, and requires prototyping `text/template` with byte-identical comparison. This spans multiple files and requires reasoning about distinct codegen patterns.
- **Task 1b** (conditional): Continues the template migration or implements the `GeneratedProgram` fallback across multiple packages. Either path touches 3+ source files with coordinated changes.
- **Task 2**: Introduces a new `ConditionalParseState` type with 4 methods across 2 new files and 2 existing files. Requires deep understanding of ANTLR parse-tree walk semantics and parenthesized sub-expression scope nesting.
- **Task 3**: Coordinates the `CILLiteral` code path with the postfix evaluator, removes 4 inter-referencing visitor methods, and rewrites 2+ visitor-specific test files.
- **Task 4**: Introduces a cross-cutting `SynthesisInput` struct in a new file, modifies 3 function signatures in `synthesize.go`, and updates tests. The struct threads through multiple codegen functions.

### Parallel Execution

- **Group A** (Tasks 2, 4): Task 2 modifies `src/main/go/parser/` files; Task 4 modifies `src/main/go/synthesis/` and potentially `src/main/go/analysis/` files. The file sets are completely disjoint and the tasks are semantically independent.

Task 3 is **not** included in a parallel group. The architect-for-story-planning recommends running Task 3 first so that visitor removal simplifies the codegen surface before Tasks 2 and 4 begin, even though the file sets are technically disjoint.

### Execution Order

```
[Task 3] → [Task 2 ‖ Task 4] → [Task 1a] → [Task 1b]
```

- Task 3 runs first: removing visitor methods simplifies the CIL transformer surface and eliminates code that Task 1a would otherwise audit.
- Tasks 2 and 4 run in parallel after Task 3 merges back.
- Task 1a runs after Tasks 2, 3, and 4 are merged: the `SynthesisInput` contract (Task 4) must be stable before template prototyping targets `GenerateMainFunction`, and visitor removal (Task 3) must be complete before auditing codegen.
- Task 1b runs after Task 1a: it depends on the go/no-go decision documented by Task 1a.

### Reintegration

**Group A (Tasks 2, 4):**
- Merge order: any order (fully disjoint packages, no shared infrastructure)
- Integration tests:
  - `go test ./src/main/go/parser/... -count=1` — verify conditional parsing behavior after Task 2
  - `go test ./src/main/go/synthesis/... -count=1` — verify synthesis contract after Task 4
  - `go test ./src/main/go/analysis/... -count=1` — verify analysis package if Task 4 touched it
  - `go test ./src/main/go/... -count=1 -timeout 120s` — full package run to catch cross-package issues after both worktrees are merged
- Watch for conflicts in:
  - `go.sum` (dependency resolution in separate worktrees may produce different hashes)
  - `src/main/go/analysis/cpl_analyzer.go` (Task 4 may modify this; Task 2 does not touch it, but concurrent work on definitions could create merge conflicts)

### Task 3: Remove visitor pattern vestige — Completed

**Goal**: Eliminate the legacy visitor methods from `CILAbstractTransformer`.

**Test-First Development**:

a. Verify the visitor path is only used for `CILLiteral` (Increment 1) metrics by tracing `TransformMetric` in `src/main/go/cil/cil_transformer_visitors.go`. Confirm that `src/main/go/cil/postfix_walk.go` already handles `*CILLiteral` (line 68–69) by pushing a `{Kind: "push", Value: n.Value}` op.

b. Write a test in `src/main/go/cil/cil_transformer_postfix_test.go` that exercises `TransformMetric` with a `CILLiteral`-only metric and asserts the generated function uses the postfix path (contains `EvaluatePostfixFloat64`) rather than the legacy `AllocateTensor` + fill-loop path. This test will initially fail because `TransformMetric` still dispatches to the visitor path for `CILLiteral`.

c. Make `TransformMetric` route `*CILLiteral` through `generatePostfixComputeBody` (the postfix walker path) instead of `VisitCILLiteral`. Ensure the resulting generated code is semantically equivalent: a single-literal postfix table must still produce a tensor filled with the literal value.

d. Remove `VisitCILLiteral`, `VisitCILBinaryOp`, `VisitCILUnaryOp`, and `VisitCILInputRef` from `CILAbstractTransformer` in `src/main/go/cil/cil_transformer_visitors.go`. Remove any internal cross-references between these visitor methods.

e. Update or remove `src/main/go/cil/cil_transformer_visitor_literal_binary_test.go` and `src/main/go/cil/cil_transformer_visitor_unary_input_test.go` — these test files are visitor-specific. Either rewrite them to test the postfix path, or remove them if equivalent coverage already exists in `cil_transformer_postfix_test.go` and `cil_transformer_metric_test.go`.

f. Run `go test ./src/main/go/cil/... -count=1` — must pass.

**Acceptance Criteria**:
- All visitor methods removed from `CILAbstractTransformer`
- `TransformMetric` uses only the postfix walker for all node types, including `CILLiteral`
- All tests pass

### Task 2: Encapsulate conditional parsing state machine — Completed

**Goal**: Replace fragile inline state machine with a dedicated type.

**Test-First Development**:

a. Write dedicated unit tests in `src/main/go/parser/conditional_parse_state_test.go` that exercise the new `ConditionalParseState` type:
   - Expression count transitions (`0 → 1 → 2`)
   - Parenthesized sub-expressions (scope nesting — the new type must not conflate nested `TempScope` pushes with the outermost condition-to-body transition)
   - `savedConditionPostfix` capture and retrieval

b. Define a `ConditionalParseState` type in a new file `src/main/go/parser/conditional_parse_state.go` with clear methods:
   - `EnterCondition()`
   - `EnterBody()`
   - `IsComplete() bool`
   - `GetConditionPostfix() []symboltable.SymbolDefinition`

c. Replace the inline `conditionalParseState` struct and `ExitExpression` logic in `src/main/go/parser/parser_support.go` with calls to the new type. Update `src/main/go/parser/cpl_conditional_listener.go` to use the new type's methods when pushing states.

d. Run `go test ./src/main/go/parser/... -count=1` — must pass, including new unit tests.

**Acceptance Criteria**:
- `ConditionalParseState` type exists with clear method-based transitions
- Original `conditionalParseState` struct and `ExitExpression` handler replaced
- New unit tests cover all state transitions
- All existing tests pass

### Task 4: Define stable analysis→synthesis contract — Completed

**Goal**: Introduce a `SynthesisInput` struct to decouple analysis and synthesis.

**Test-First Development**:

a. Define `SynthesisInput` struct in a new file `src/main/go/synthesis/input.go` containing:
   - Connectors slice
   - Measures slice
   - Metrics slice
   - IO specification
   - CIL trees (unified from Phase 2)

b. Write tests in `src/main/go/synthesis/synthesize_test.go` that construct a `SynthesisInput` and pass it to the refactored `BuildTransformerFromAnalysis` and `GenerateMainFunction`. These tests will initially fail because the signatures have not been updated yet.

c. Update `BuildTransformerFromAnalysis` in `src/main/go/synthesis/synthesize.go` to accept `*analysis.AnalysisDefinitions` and return `*SynthesisInput` instead of mutating the transformer directly.

d. Update `GenerateMainFunction` in `src/main/go/synthesis/synthesize.go` to accept `*SynthesisInput` and extract data from it.

e. Update `BuildConceptValuesMap` to be a method on `SynthesisInput` (or a standalone function that accepts it) in `src/main/go/synthesis/synthesize.go`.

f. Document the implicit grain-name duplication contract: add a comment on `BuildConceptValuesMap` (or on the equivalent `SynthesisInput` method) explaining that connector grain measure names mirror their underlying concept values, and why this relationship exists. This addresses the P3 technical-debt item identified in Issue 17.

g. Run `go test ./src/main/go/synthesis/... ./src/main/go/analysis/... -count=1` — must pass.

**Acceptance Criteria**:
- `SynthesisInput` struct defined and used as the contract between analysis and synthesis
- `BuildTransformerFromAnalysis` returns `*SynthesisInput`
- `GenerateMainFunction` no longer couples to `AnalysisDefinitions` directly
- A comment on `BuildConceptValuesMap` (or the equivalent `SynthesisInput` method) documents the implicit grain-name duplication contract
- All tests pass

### Task 1a: Audit codegen and prototype template adoption — Completed

**Goal**: Evaluate whether `text/template` is viable for generated Go program output, and make a go/no-go decision.

**Test-First Development**:

a. Audit current string-concatenation codegen in `src/main/go/synthesis/synthesize.go` (main function, setup, compute, cleanup, input loading, output writing), `src/main/go/connector/connector_loader.go`, and `src/main/go/io/generate_parquet.go`. Note that `src/main/go/cil/cil_transformer_visitors.go` has already been removed by Task 3.

b. Choose a single representative codegen target for prototyping (e.g., `generateMainFunctionSignature` or the main function body generation in `synthesize.go`). Write a comparison test that captures the output of the current `strings.Builder` approach as the expected golden output.

c. Design a `GeneratedProgram` struct in a new file `src/main/go/synthesis/generated_program.go` that holds program parts (imports, variables, functions, main) instead of building strings incrementally.

d. Prototype one template for the chosen target using `text/template`. Implement the template so it produces output from the `GeneratedProgram` struct.

e. Run the comparison test — the template output must be byte-identical to the golden output.

f. If the prototype succeeds with acceptable complexity, document the go decision and the migration plan for Task 1b. If `text/template` proves too unwieldy for the current string-heavy codegen (e.g., excessive escaping, loss of deterministic ordering, unreadable templates), document the no-go decision and commit the `GeneratedProgram` struct as the intermediate improvement.

g. Run `go test ./src/main/go/synthesis/... ./src/main/go/connector/... ./src/main/go/io/... -count=1` — must pass.

**Acceptance Criteria**:
- Go/no-go decision documented in story notes or code comments
- `GeneratedProgram` struct implemented
- Prototype template produces byte-identical output for the chosen target (verified by comparison test)
- All tests pass

### Task 1b: Migrate codegen or commit fallback — Completed

**Goal**: Complete the codegen restructuring based on the Task 1a decision.

**Test-First Development**:

a. **If Task 1a decided GO (templates):**
   - Write template-backed tests for each remaining codegen target before migrating it.
   - Migrate all remaining codegen in `src/main/go/synthesis/synthesize.go`, `src/main/go/connector/connector_loader.go`, and any other identified files to use `text/template` incrementally.
   - Ensure each migrated target passes its byte-identical comparison test.

   **If Task 1a decided NO-GO (fallback):**
   - Refactor existing `strings.Builder` codegen in `src/main/go/synthesis/synthesize.go`, `src/main/go/connector/connector_loader.go`, and `src/main/go/io/generate_parquet.go` to use the `GeneratedProgram` struct as an intermediate improvement.
   - Write tests for the `GeneratedProgram` serialization methods.
   - Ensure generated output remains byte-identical to the pre-refactoring version.

b. Run `go test ./src/main/go/synthesis/... ./src/main/go/cil/... ./src/main/go/connector/... ./src/main/go/io/... -count=1` — must pass.

**Acceptance Criteria**:
- Either `text/template` adopted for all codegen, or `GeneratedProgram` struct is used as the structured fallback
- Generated code output is byte-identical to before (verified by golden-file or comparison tests)
- All tests pass

## Constraints

1. Template-based codegen must produce byte-identical output or the decision must be documented
2. Conditional state machine must preserve exact parse tree behavior
3. Visitor removal must not break `CILLiteral` metric synthesis
4. `SynthesisInput` must be a stable contract — future analysis changes should not require synthesis changes
5. The grain-name duplication contract (Issue 17) must be documented with a code comment that explains the relationship between connector grain measure names and underlying concept values
6. All existing tests must pass after each task

## Intent

Replace fragile, tightly coupled patterns with robust, encapsulated, and well-contracted designs.

## Acceptance Criteria

1. Code generation is either template-based or uses a structured `GeneratedProgram` type
2. Conditional parsing state machine is encapsulated in a dedicated type with unit tests
3. Visitor methods removed; only postfix walker remains
4. `SynthesisInput` struct decouples analysis from synthesis
5. Grain-name duplication contract is documented on `BuildConceptValuesMap` or the equivalent `SynthesisInput` method
6. `go test ./src/main/go/... -count=1` passes with zero failures

## Notes

- Task 1a is the most exploratory. If `text/template` proves too unwieldy for the current string-heavy codegen, the fallback is the `GeneratedProgram` struct. The go/no-go decision is recorded in this story so the story-implementor does not need to make the design call mid-task.
- Task 3 depends on the postfix walker being capable of handling `CILLiteral` trees. `src/main/go/cil/postfix_walk.go` already handles `*CILLiteral` (line 68–69), so this precondition is satisfied.
- Task 4 may reveal that `BuildConceptValuesMap` grain duplication logic (Issue 17) is better documented once it lives inside the contract struct.
- Task 3 runs first in the execution order so that visitor removal simplifies the CIL transformer surface before Task 1a audits codegen.
