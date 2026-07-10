# Story240: Delete ODS-Related Code and Resources

## Goal

Once all with_data examples are converted to XLSX (Stories 238 and 239), delete all ODS-related CLI code, tests, utilities, and API documentation since ODS is no longer used as a verification format.

## References

- [Test Conversion Plan](../plans/test-conversion-plan.md)
- [Story238_convert-arithmetic-examples-to-xlsx.md](finished-stories/Story238_convert-arithmetic-examples-to-xlsx.md) — Done
- [Story239_convert-conditional-grain-examples-to-xlsx.md](finished-stories/Story239_convert-conditional-grain-examples-to-xlsx.md) — Done

## Dependencies

- [Story238_convert-arithmetic-examples-to-xlsx.md](finished-stories/Story238_convert-arithmetic-examples-to-xlsx.md) — Done
- [Story239_convert-conditional-grain-examples-to-xlsx.md](finished-stories/Story239_convert-conditional-grain-examples-to-xlsx.md) — Done
  Both must complete before this story can begin to ensure no example still relies on ODS files.

## Dependent Stories

None

## Tasks

### Task Sizing

Tasks 1–4 are small deletions of single directories or files. Task 5 is a compilation and test verification that touches no new files.

### Parallel Execution

- **Group C** (Tasks 1, 2, 3, 4): All four tasks delete independent directories/files. No shared files or compile dependencies between them (the Go compiler will resolve the absence of these packages at build time).

### Execution Order

```
[Task 1 ‖ Task 2 ‖ Task 3 ‖ Task 4] → Task 5
```

### Reintegration

**Group C (Tasks 1–4):**
- Merge order: any order (fully disjoint deletions)
- Integration tests: `cd src/chronocone_planning_language && go build ./src/main/go/... && go test ./src/main/go/cli/compiler/... -count=1 -timeout 300s`
- Watch for conflicts in: `go.mod` / `go.sum` (if any ODS dependencies are removed, module files may change)

### Task 1: Delete expected_ods_generator CLI — Completed [Small]

**Intent**: Remove the ODS expected-output generator since XLSX generation replaces it.

a. Delete directory `src/main/go/cli/expected_ods_generator/` and all files within it (8 files) — Completed

**Acceptance Criteria**:
- `src/main/go/cli/expected_ods_generator/` does not exist.

### Task 2: Delete ods_utils CLI — Completed [Small]

**Intent**: Remove the ODS utility package that was only used by ODS generator and verifier.

a. Delete directory `src/main/go/cli/ods_utils/` and the file `ods_utils.go` within it — Completed

**Acceptance Criteria**:
- `src/main/go/cli/ods_utils/` does not exist.

### Task 3: Delete ods_verify CLI — Completed [Small]

**Intent**: Remove the ODS verifier since xls_verify replaces it.

a. Delete directory `src/main/go/cli/ods_verify/` and both files within it (`main.go`, `main_test.go`) — Completed

**Acceptance Criteria**:
- `src/main/go/cli/ods_verify/` does not exist.

### Task 4: Delete go-ods API docs and update toc — Completed [Small]

**Intent**: Remove documentation for a library that is no longer used in the project.

a. Delete directory `resources/apis/go-ods/` and all files within it (`api-reference.md`, `toc.md`, `version-history.md`) — Completed

b. Update `resources/apis/toc.md`:
   - Remove the `go-ods` entry
   - Update the `excelize-v2` entry to drop the migration mention
   — Completed

**Acceptance Criteria**:
- `resources/apis/go-ods/` does not exist.
- `resources/apis/toc.md` contains no `go-ods` entry.
- The `excelize-v2` entry reads: "excelize v2 library for reading, writing, and manipulating Excel (XLSX/XLSM) files in pure Go" (no migration mention).

### Task 5: Verify project compiles and tests pass — Completed

**Intent**: Ensure no remaining references to deleted ODS packages break the build or tests.

a. Run `cd src/chronocone_planning_language && go build ./src/main/go/...` to ensure no broken references to deleted packages — Completed

b. Run `go test ./src/main/go/cli/compiler/... -count=1 -timeout 300s` to confirm all with_data tests still pass — Completed

c. Search for any remaining references to `ods_utils`, `expected_ods_generator`, or `ods_verify` in the codebase and remove or update them — Completed

**Acceptance Criteria**:
- `go build ./src/main/go/...` succeeds.
- All with_data tests pass.
- No remaining references to deleted ODS packages exist in the codebase.

## Constraints

- Do NOT delete any ODS dependencies from `go.mod` unless they are confirmed unused by other packages. The `go mod tidy` command can be used after deletion if appropriate.
- Do NOT delete `go-ods` from `go.mod` if other unrelated code still uses it.
- Ensure `go build` passes before considering this story complete.
