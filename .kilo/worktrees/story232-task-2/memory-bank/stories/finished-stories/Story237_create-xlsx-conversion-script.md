# Story237: Create XLSX Conversion PDD Script and Convert Test Driver

## Goal

Create the reusable PDD script `scripts/convert-example-to-xlsx.pdd.script.md` that automates per-example conversion from JSON/ODS verification to XLSX verification. Also convert `src/main/go/cli/compiler/cpl_with_data_test.go` to use `xls_verify` instead of `parquet_verify`. The test conversion will temporarily cause test failures until the example conversions in Stories 238–239 are completed.

## References

- [Test Conversion Plan](../plans/test-conversion-plan.md)
- [Story219_ods-generator-and-verifier.md](../finished-stories/Story219_ods-generator-and-verifier.md)

## Dependencies

- [Story219_ods-generator-and-verifier.md](../finished-stories/Story219_ods-generator-and-verifier.md) — Done
  The XLSX generator CLI (`expected_xls_generator/main/main.go`) and XLSX verifier CLI (`xls_verify/main.go`) must exist and function correctly.
- `regenerate_data.sh` and `parquet_gen.go` already updated to use `data_*.json` as the source of truth for input data.

## Dependent Stories

- [Story238_convert-arithmetic-examples-to-xlsx.md](Story238_convert-arithmetic-examples-to-xlsx.md) — Not Started
- [Story239_convert-conditional-grain-examples-to-xlsx.md](Story239_convert-conditional-grain-examples-to-xlsx.md) — Not Started

## Tasks

### Task Sizing

Tasks 1 and 2 are single-file tasks (one script file, one Go test file) with straightforward, well-defined changes. Task 3 is a compilation verification.

### Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 creates a script file; Task 2 modifies a Go test file. Disjoint files, no shared dependencies.

### Execution Order

```
[Task 1 ‖ Task 2] → Task 3
```

### Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (fully disjoint files, no shared infrastructure)
- Integration tests: `cd src/chronocone_planning_language && go test ./src/main/go/cli/compiler/... -count=1 -run TestWithDataExamples_ExpectedOutputExists -timeout 60s` — verify the test file compiles with the new XLSX path
- Watch for conflicts in: none expected

### Task 1: Create PDD script for per-example XLSX conversion — Completed [Small]

**Intent**: Provide a reusable, parameterized script that the story-implementor can invoke once per example directory to perform all per-example file operations.

a. Write `scripts/convert-example-to-xlsx.pdd.script.md` covering:
   - **Parameter validation**: confirm `example_dir` exists and `example_base_dir/core.cpl` exists
   - **Regenerate parquet from `data_*.json`**: invoke `regenerate_data.sh <name> <example_base_dir>` from `src/chronocone_planning_language`
   - **Generate expected_output.xlsx**: invoke `go run src/main/go/cli/expected_xls_generator/main/main.go -example=<example_dir>` from `src/chronocone_planning_language`
   - **Delete redundant files**: `expected_output.json`, `expected_output.ods`, `parquet_gen_spec_*.json` (skip if missing)
   - **Verify conversion**: run `compile_and_run.sh <name>` and report mismatched cells if verification fails (do NOT attempt fixes)
   - **Report results**: example name, files deleted, XLSX generation status, verification status, mismatched cells if any
   — Not Started

**Acceptance Criteria**:
- `scripts/convert-example-to-xlsx.pdd.script.md` exists and contains all six script operations.
- Script parameters (`example_dir`, `example_base_dir`) are clearly documented with validation rules.
- The script explicitly states it must NOT attempt to fix pre-existing data discrepancies.

### Task 2: Convert test driver to XLSX verification — Completed [Small]

**Intent**: Update the Go test driver to look for `.xlsx` files, remove all skip lists, switch from `parquet_verify` to `xls_verify`, and remove the now-obsolete JSON parsability test.

a. In `src/main/go/cli/compiler/cpl_with_data_test.go`:
   - Change `discoverWithDataExamples()` to detect `.xlsx` instead of `.json` (~line 119)
   - Remove the entire `skippedExamples` map and `t.Skip` block from `TestWithDataExamples` (~lines 162–181)
   - Remove the entire `skippedExamples` map and `t.Skip` block from `TestWithDataExamples_CompileOnly` (~lines 318–331)
   - Change the verification step from `parquet_verify` to `xls_verify` with arguments `-parquet` and `-xlsx` (~lines 216–222)
   - Update `TestWithDataExamples_ExpectedOutputExists` error message to reference `.xlsx` (~lines 257–274)
   - Remove `TestWithDataExamples_JSONParsable` entirely (~lines 289–307)
   — Not Started

**Acceptance Criteria**:
- No `skippedExamples` maps remain in the test file.
- `TestWithDataExamples_JSONParsable` is removed.
- `discoverWithDataExamples` resolves `hasExpected` based on `.xlsx` presence.
- Verification command uses `xls_verify/main.go` with `-xlsx=` argument.

### Task 3: Verify test compiles and script is well-formed — Completed

**Intent**: Ensure the modified test file is syntactically valid and the PDD script is internally consistent.

a. Compile `cpl_with_data_test.go` to verify no syntax errors:
   `cd src/chronocone_planning_language && go test -c ./src/main/go/cli/compiler/...`
   — Not Started
b. Verify the PDD script follows project conventions and references correct file paths — Not Started

**Acceptance Criteria**:
- `go test -c ./src/main/go/cli/compiler/...` succeeds without compilation errors.
- PDD script references correct paths for `regenerate_data.sh`, `expected_xls_generator`, `compile_and_run.sh`, and the example directory structure.

## Constraints

- The test conversion in Task 2 will temporarily cause `TestWithDataExamples` to fail for examples that have not yet been converted to XLSX. This is expected and will be resolved by Stories 238 and 239.
- File paths in the PDD script must be relative to the project root or clearly parameterized.
- The PDD script must not attempt to fix pre-existing data discrepancies; it should only report them.

## Notes

- This story supersedes the ODS-based approach outlined in [Story218_transition-to-ods-verification.md](Story218_transition-to-ods-verification.md). The project is moving directly from JSON verification to XLSX verification, skipping ODS as an intermediate format.
