# Story218: Transition to ODS Verification

## Goal

With the ODS generator and verifier already implemented (Story219), transition all 17 `with_data` examples from JSON-based expected-output verification to ODS-based verification. This involves: creating a PDD script for example creation and regeneration, correcting all `expected_output.json` files with generator-computed values, generating `expected_output.ods` and `expected_output_schema.json` for every example, and validating end-to-end that the verifier passes for all examples and the full Go test suite passes.

## References

- [Plan: Transition to ODS Format](../plans/transition_to_csv_verification_plan.md)
- [Design: Example Generation and Verification](../specifications/design/cpl/example_generation_and_verification.md)
- [go-ods API Reference](../../resources/apis/go-ods/api-reference.md)
- [go-ods Version History](../../resources/apis/go-ods/version-history.md)

## Dependencies

- [Story219_ods-generator-and-verifier.md](Story219_ods-generator-and-verifier.md) — Not Started (provides the ODS generator and verifier tools used as prerequisites throughout this story)
- [Story207_correct-metrics-arithmetic-examples.md](Story207_correct-metrics-arithmetic-examples.md) — Done
- [Story214_cpl-metrics-conditional-examples.md](finished-stories/Story214_cpl-metrics-conditional-examples.md) — Done

Both dependency stories created the `with_data` examples that this story will generate ODS files for.

## Dependent Stories

None.

## Tasks

### Task Sizing

- **Small tasks**: Tasks 2–16 (per-example regeneration, at most 2 files each, established pattern, straightforward).
- **Medium tasks**: Task 1 (documentation/script task routed to `technical-writer-for-story-implementor`), Task 17 (cross-cutting test suite execution).

### Worktree Suitability

- **Group A (Tasks 2–16):** — each task touches files in a single example directory, no shared infrastructure, safe for subagents
- No `[Worktree]` groups needed.

### Parallel Execution

- **Group A** (Tasks 2–16): Each task touches files in exactly one example directory. Directories are disjoint, no semantic dependencies among them.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3 ‖ Task 4 ‖ Task 5 ‖ Task 6 ‖ Task 7 ‖ Task 8 ‖ Task 9 ‖ Task 10 ‖ Task 11 ‖ Task 12 ‖ Task 13 ‖ Task 14 ‖ Task 15 ‖ Task 16] → Task 17
```

### Worktree Reintegration

None — no worktree groups. All parallel tasks are subagent-safe with fully disjoint file sets.

### Task 1: Create PDD script for example creation and regeneration — Not Started

Create `scripts/create_example.pdd.script.md` with pseudo-executable, low-level instructions for creating a new `with_data` example and for regenerating artifacts in an existing example. The script must include:

- Invocation parameters (`category_name`, `example_name`).
- Step-by-step commands to create the example subdirectory, CPL file, `iospec.json`, parquet generation specs, input data JSON files.
- Commands to generate `.parquet` files from specs.
- Command to generate `expected_output.ods` using the ODS generator.
- Command to compile, run, and verify via `compile_and_run.sh`.
- A **regeneration** section for existing examples: steps to re-run the ODS generator, correct `expected_output.json` if evaluator values differ, and re-verify.
- Manual verification instructions (open in LibreOffice Calc, confirm input sheets, branch formulas, final values).
- Commit instructions.
- A files-produced table.

- a. Draft the PDD script following the plan's step 7 template — Not Started
- b. Add the regeneration section for existing examples — Not Started
- c. Review script for completeness and unambiguous command invocations — Not Started
- d. Validate script by mentally walking through one example creation and one regeneration — Not Started

### Task 2: Example — multiplication_precedence — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_arithmetic/multiplication_precedence` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 3: Example — parenthesized — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_arithmetic/parenthesized` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 4: Example — short_form_addition — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_arithmetic/short_form_addition` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 5: Example — complex_expression — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_arithmetic/complex_expression` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 6: Example — subtraction_division — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_arithmetic/subtraction_division` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 7: Example — unary_negation — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_arithmetic/unary_negation` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 8: Example — conditional_if_else_arithmetic — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_conditionals/conditional_if_else_arithmetic` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 9: Example — conditional_if_elseif_else — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_conditionals/conditional_if_elseif_else` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 10: Example — conditional_logical_operators — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_conditionals/conditional_logical_operators` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 11: Example — conditional_nested_arithmetic — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_conditionals/conditional_nested_arithmetic` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 12: Example — conditional_short_form — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_conditionals/conditional_short_form` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 13: Example — conditional_tautological_if_only — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_conditionals/conditional_tautological_if_only` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 14: Example — conditional_truthiness — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_conditionals/conditional_truthiness` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 15: Example — conditional_unary_negation — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_conditionals/conditional_unary_negation` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 16: Example — conditional_when_default — Not Started [Small]

Follow the PDD script regeneration procedure for the `metrics_conditionals/conditional_when_default` example.

- a. Run the ODS generator to produce `expected_output.ods` — Not Started
- b. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- c. Run `compile_and_run.sh` to confirm `ods_verify` passes — Not Started

### Task 17: Run full test suite — Not Started

Execute all relevant tests to confirm the transition is complete and nothing is broken.

- a. Run `ods_verify` unit tests (`go test ./src/main/go/cli/ods_verify/...`) — Not Started
- b. Run `parquet_verify` tests (`go test ./src/main/go/cli/parquet_verify/...`) — Not Started
- c. Run compiler integration tests (`go test ./src/main/go/cli/compiler/...`) — Not Started
- d. Summarize results and flag any failures — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For this transition story, testing takes the form of acceptance tests at each step:

- **Task 1**: Validate the PDD script by mentally walking through one example creation and one regeneration (task 1d).
- **Tasks 2–16**: Each example task uses `compile_and_run.sh` as its acceptance test — the script compiles, runs, and invokes `ods_verify` (built by Story219) to confirm the generated ODS matches `output.parquet`. Generator-computed values are the authoritative expected results.
- **Task 17**: The full test suite run (`ods_verify`, `parquet_verify`, `compiler`) serves as the final acceptance test for the entire transition.

## Constraints

- **Backward compatibility**: `parquet_verify.go` and `cpl_with_data_test.go` must remain unchanged during this story.
- **Scope**: All 17 `with_data` examples (both `metrics_arithmetic` and `metrics_conditionals`).
- **PDD script validation**: The PDD script must be used to regenerate every existing example. If regeneration reveals problems in the script, the script must be corrected before proceeding with the remaining examples.
- **Generator and verifier are prerequisites**: Built by Story219. This story does not modify generator/verifier code. See [Story219_ods-generator-and-verifier.md](Story219_ods-generator-and-verifier.md).
- **ODS library**: `github.com/AlexJarrah/go-ods` v1.0.7 is already in `go.mod` (added by Story219).

## Intent

- **Task 1**: Capture the new example creation and regeneration workflow as a repeatable PDD script.
- **Tasks 2–16**: Apply the regeneration procedure to all 17 examples (correct `expected_output.json`, generate `.ods` and schema files, confirm `ods_verify` passes), confirming the PDD script works at scale.
- **Task 17**: Confirm the entire transition works end-to-end and nothing is regressed.

## Acceptance Criteria

1. The ODS generator and verifier (built by Story219) are functional and verified — `compile_and_run.sh` uses `ods_verify` for ODS-based verification.
2. All 17 `expected_output.json` files contain correct values and `parquet_verify` tests pass.
3. `expected_output.ods` and `expected_output_schema.json` exist for all 17 examples.
4. `compile_and_run.sh` passes for all 17 examples (compilation, execution, and `ods_verify`).
5. `create_example.pdd.script.md` exists, contains unambiguous instructions for both creation and regeneration, and has been validated by regenerating all 17 examples.
6. All Go tests (`ods_verify`, `parquet_verify`, `compiler`) pass.

## Open Questions

None — all design decisions resolved in the plan. The v2.0.0 → v1.0.7 version decision has been finalized: v1.0.7 was chosen because v2.0.0 is an experimental rewrite not yet production-ready. See [go-ods Version History](../../resources/apis/go-ods/version-history.md).

## Known Issues

- Many existing `expected_output.json` files currently contain incorrect values. These will be corrected during Tasks 2–16.
- The `.ods` verification utility reads only `final` flag rows (static values). Branch rows contain formulas that can only be evaluated when the file is opened in LibreOffice Calc — this is by design.

## Additional Context

- The plan refers to this as "Transition Expected-Output to ODS Format (Human-Verifiable Spreadsheets)". The `.ods` format was chosen over CSV because cross-sheet references work, it produces a single file per example, and it opens natively in LibreOffice Calc.
- The ODS generator and verifier are implemented by Story219. This story uses them as prerequisite tools; it does not modify generator/verifier code.
- The two validation examples (`simple_addition` and `conditional_comparison_operators`) with user inspection stops were validated as part of Story219 and are not repeated here.
