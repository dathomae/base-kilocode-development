# Story266: Spreaders Increment 1 — Synthesize (Epic008)

## Goal

Implement the synthesis (code generation) phase for spreader declarations: CIL tree walking for `CILSpreadNode`, spreading matrix emission as Go preamble declarations, `ExpandDimension` integration (from tensorops built in Story254), manual spreader test code generation, and end-to-end spreader examples with CPL + IO spec + parquet + expected output. This is Increment 1 of Epic008 (Spreaders), delivering the synthesis phase for basic spreaders (equal-proportion distribution). The spreader walker handles `CILSpreadNode` by emitting calls to `tensorops.ExpandDimension`; spreading matrices are emitted as Go struct literals in the synthesis preamble before metric computation functions.

## References

- [Plan: Story266](../../memory-bank/plans/Story266_spreaders-increment1-synthesize.md)
- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md)
- [ERO008: Spreaders Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md)
- [Epic008: CPL Spreaders Implementation](../epics/Epic008_cpl-spreaders-implementation.md)
- [Story254: Aggregators Increment 1 — Matrix Multiplication](Story254_aggregators-increment1-matrix-mult.md) — `tensorops.ExpandDimension` and `tensorops.AggregationMatrix` types
- [Story257: Aggregators Increment 2 — Synthesize](Story257_aggregators-increment2-synthesize.md) — CIL tree walker infrastructure and preamble generation pattern

## Dependencies

- Story265 (spreaders-increment1-analyze) — analysis must produce `CILSpreadNode` with pre-built `SpreadingMatrix` in CIL trees
- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ExpandDimension` function is the core synthesis primitive
- Story257 (aggregators-increment2-synthesize) — CIL tree walker infrastructure and preamble generation pattern already exist; spreader synthesis follows the same patterns

## Dependent Stories

- Story for Increment 2 (Spreader Integration into Metric DAG) — consumes the spreader synthesis pipeline built here
- Story for Increment 3 (Additional Spreader Function Types) — extends walker for new spreader node variants

## Tasks

### Task Sizing

- **Task 1** is `[Small]` — it creates a single new source file (`cil_spread_walker.go`) and its test file. The method handles one CIL node type (`CILSpreadNode`) with two straightforward cases (scalar source vs non-empty source). Follows the `WalkCILAggregateNode` pattern from Story257; the example code in the plan is nearly copy-paste ready. Purely additive; modifies no existing files.
- **Task 2** is `[Small]` — it adds one function (`EmitSpreadingMatrix`) to the synthesis preamble generator and creates a corresponding test file. The function emits a Go var declaration following the existing `EmitAggregationMatrix` pattern from Story257. At most 2 files; addresses a single concern.
- **Task 3** is not `[Small]` — it introduces new test code generation infrastructure (new file with new structs) that must produce correct Go test functions with tensor construction, `ExpandDimension` calls, and shape/value assertions. Lays down new code generation infrastructure rather than adding to an existing pattern. Multiple subtask steps spanning generation logic and verification.
- **Task 4** is `[Small]` — it creates a single example subdirectory following the existing `with_data/` template structure. Creates CPL, IO spec, parquet, expected output, and test files within one directory tree. Single concern (end-to-end example); follows established patterns mechanically.

### Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 modifies `preamble_generator.go` in the synthesis package; Task 3 creates a new test generation file. Disjoint file sets. Neither task depends on the other's code output at compile time — Task 2 emits Go preamble variables, Task 3 generates Go test functions that use the same naming conventions but do not import each other.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4
```

- Task 1 must complete first (the walker defines the contract for emitted `ExpandDimension` calls, establishing source dimension indexing, new dimension insertion points, and error handling format).
- Tasks 2 and 3 run in parallel (Group A) — preamble generator and test code generator are independent concerns.
- Task 4 depends on all preceding tasks (examples need the full pipeline: walker → preamble → test generation).

### Reintegration

**Group A (Tasks 2, 3):**
- Merge order: Task 2 first (preamble emission is infrastructure — the `EmitSpreadingMatrix` function signature is referenced by integration tests), then Task 3
- Integration tests: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` — verify that preamble declarations and test code generation compile together in the synthesis package
- Watch for conflicts in: `preamble_generator.go` (modified by Task 2; Task 3's generated test code may reference exported symbols added by Task 2); `go.sum` (dependency resolution in separate worktrees may produce different hashes)

### Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, write tests before implementing the code. The test-and-implement cycle is integrated within each task in the correct order:

- **Task 1**: Write walker tests (subtask e) before the walker implementation passes them.
- **Task 2**: Write preamble tests (subtask d) before `EmitSpreadingMatrix` produces valid output.
- **Task 3**: Write test generation tests (subtask c) before the generator implementation passes them.
- **Task 4**: Example test verifies full end-to-end pipeline behavior.

Tests are NOT broken out as separate tasks. See `memory-bank/concepts.md` for details on the Logical TDD Lifecycle.

### Constraints

- The `tensorops.ExpandDimension` function and `tensorops.AggregationMatrix` type must exist (built in Story254)
- The `CILSpreadNode` type with pre-built spreading matrices must exist in the CIL node hierarchy (built in Story265)
- Spreading matrices are pre-built from concept value sets at compile time and emitted as Go struct literals
- Spreaders expand tensor dimensionality by exactly one per spreader
- For equal-proportion distribution, spread-then-aggregate yields the original values (provides natural verification)
- All generated code must compile under `source_root` (`src/chronocone_planning_language`)
- Generated code follows the existing pattern: `check()` helper for error handling, `ChronoconeTensor[float64]` for tensor types
- Spreaders are applied manually in test code — not yet integrated into metric computation (Increment 2)
- The `from []` (scalar source) edge case must be handled: `srcDim=0, newDim=0`

### Task 1: Implement WalkCILSpreadNode in CIL tree walker — Not Started [Small]

Add a `WalkCILSpreadNode` handler for the CIL tree walker (`src/main/go/synthesis/cil_spread_walker.go`) that walks a `CILSpreadNode` and emits Go code calling `tensorops.ExpandDimension`. The handler must: walk the source CIL subtree (which may reference connector data or upstream computation results), determine the source dimension index and new dimension insertion point, emit the `ExpandDimension` call with the correct variable names and spreading matrix parameter, emit error handling after the call, and register the result variable for use by subsequent operations. Wire the `CILSpreadNode` case into the main walker dispatch.

   a. Create `src/main/go/synthesis/cil_spread_walker.go` — define the `WalkCILSpreadNode` method on the CIL tree walker struct — Not Started
   b. Handle `FromEmpty` (scalar source) case: walk the source CIL subtree, emit `tensorops.ExpandDimension(sourceVar, spreadingMatrix_X, 0, 0)` with the source dimension and new dimension both at index 0 — Not Started
   c. Handle non-empty `from` case: walk the source CIL subtree, compute correct `srcDim` index from `node.FromDimensions[node.FromDimensions[0]]` and `newDim` index from `node.ToDimensions` insertion point, emit `tensorops.ExpandDimension(sourceVar, spreadingMatrix_X, srcDim, newDim)` — Not Started
   d. After the `ExpandDimension` call, emit error-checking block: `if err != nil { return nil, fmt.Errorf("spread(%s) failed: %w", err) }` — Not Started
   e. Write unit tests in `src/main/go/synthesis/cil_spread_walker_test.go`: verify generated Go code contains correct `tensorops.ExpandDimension` call with correct parameters for (1) single-dimension expansion (`from [Product] to [Country, Product]` → `srcDim=<Product_idx>, newDim=<Country_insert>`), (2) scalar spreading (`from [] to [Country]` → `srcDim=0, newDim=0`), and (3) multi-concept `from` (`from [Country, Product] to [Country, Product, Region]`). Also verify error-handling code emission and correct spreading matrix variable naming (`spreadingMatrix_<concept>` sanitized). Tests must fail initially before implementation passes them — Not Started
   f. Verify walker tests pass: `go test ./src/main/go/synthesis/... -run Spread -count=1` — Not Started

**Files**: `src/main/go/synthesis/cil_spread_walker.go` (create), `src/main/go/synthesis/cil_spread_walker_test.go` (create)

**Intent**: The walker must transform a `CILSpreadNode` into executable Go code that calls the dimension-expanding matrix multiplication infrastructure (`tensorops.ExpandDimension`) built in Story254. The generated code follows existing patterns from Story257's `WalkCILAggregateNode`: variable naming via `freshVar`, error propagation via `fmt.Errorf` wrapping, and tensor variable registration for downstream use.

**Acceptance criteria**:
- `CILSpreadNode` with known source tensor, dimension mappings, and spreading matrix produces correct `tensorops.ExpandDimension` call in generated code
- Scalar spreading (`from []`) emits `srcDim=0, newDim=0`
- Non-empty `from` case computes correct `srcDim` and `newDim` from `FromDimensions`/`ToDimensions`
- Generated code includes error handling: `if err != nil { return nil, fmt.Errorf("spread(%s) failed: %w", err) }`
- Spreading matrix variable name uses sanitized concept name (e.g., `spreadingMatrix_Country`)
- Result variable is registered for use by subsequent operations

### Task 2: Implement EmitSpreadingMatrix preamble generation — Not Started [Small]

Add an `EmitSpreadingMatrix` function to the synthesis preamble generator (`src/main/go/synthesis/preamble_generator.go`) that writes a `var spreadingMatrix_X = &tensorops.AggregationMatrix{...}` Go declaration for a pre-built spreading matrix. The generated code must include: the matrix variable name (derived from the spreader's new dimension), `Rows` and `Columns` counts, `ColLabels` as a `[]string` literal, and `Data` as a `[][]float64` literal. Follow the existing preamble generation pattern (e.g., `EmitAggregationMatrix` from Story257).

   a. Add `EmitSpreadingMatrix` function to the preamble generator — signature: `func (g *PreambleGenerator) EmitSpreadingMatrix(buf io.Writer, name string, matrix *AggregationMatrix) error` — writes the `var spreadingMatrix_X = &tensorops.AggregationMatrix{...}` declaration — Not Started
   b. Emit `ColLabels` as a `[]string` literal (e.g., `[]string{"USA", "Canada"}`) and `Data` as a `[][]float64` literal (e.g., `[][]float64{{0.5, 0.5}}`) using the matrix's column labels and data fields — Not Started
   c. Handle the empty matrix edge case (0 columns): emit a valid `AggregationMatrix` with `Columns: 0` and empty `ColLabels`/`Data` slices, or produce a clear error — Not Started
   d. Write unit tests in `src/main/go/synthesis/preamble_generator_test.go` (or a dedicated spreading test file): verify generated preamble contains correct matrix data for a 1×2 equal-distribution matrix, a 3×1 size matrix, and an edge case with 0 columns. Verify the emitted Go var name matches the expected `spreadingMatrix_X` pattern. Tests must fail initially before implementation passes them — Not Started
   e. Verify preamble tests pass: `go test ./src/main/go/synthesis/... -run SpreadingMatrix -count=1` — Not Started

**Files**: `src/main/go/synthesis/preamble_generator.go` (create or modify), `src/main/go/synthesis/preamble_generator_test.go` (create)

**Intent**: The preamble generator must produce valid Go var declarations for `tensorops.AggregationMatrix` so the synthesis preamble can declare pre-built spreading matrices as package-level variables. Matrix values come from concept value sets and are materialized at compile time. This follows the same pattern as aggregation matrix emission in Story257.

**Acceptance criteria**:
- `EmitSpreadingMatrix` emits valid Go `var spreadingMatrix_X = &tensorops.AggregationMatrix{...}` literals with correct column labels and data
- `ColLabels` and `Data` arrays reflect the matrix's column labels and data values
- Generated code compiles (verified via test that checks output against expected format)
- Data matrix dimensions match Rows × Columns
- Empty matrix edge case (0 columns) is handled gracefully

### Task 3: Create manual spreader test code generation — Not Started

Implement the code generation path that produces a standalone Go test function for manual spreader application. The generated test function must: construct a source tensor (using `chronocone_compute.NewChronoconeTensor[float64]`), call `tensorops.ExpandDimension` with the correct spreading matrix and dimension parameters, and verify the output shape matches the expected expanded dimensions. The generated test code should be compilable and executable as a standalone Go test.

   a. Create `src/main/go/synthesis/spreader_test_gen.go` — implement the code generation logic that produces a complete `func TestSpreader_<name>(t *testing.T)` function. The generated function creates a source tensor from known values and shape, calls `tensorops.ExpandDimension`, and asserts output shape correctness — Not Started
   b. Generated test must include: source tensor construction with `NewChronoconeTensor[float64](data, shape)`, the `ExpandDimension` call with correct `srcDim`/`newDim` parameters, shape verification with `assert.Equal`, and import block for required packages (`chronocone_compute`, `tensorops`, `testing`) — Not Started
   c. Write unit tests in `src/main/go/synthesis/spreader_test_gen_test.go`: verify the full generated test function for equal-proportion distribution (`from [Product] to [Country, Product]`) produces correct tensor construction and `ExpandDimension` call. Verify the generated code compiles (by checking output against expected format). Also verify scalar spreading case (`from [] to [Country]`). Tests must fail initially before implementation passes them — Not Started
   d. Verify test generation tests pass: `go test ./src/main/go/synthesis/... -run SpreaderTestGen -count=1` — Not Started

**Files**: `src/main/go/synthesis/spreader_test_gen.go` (create), `src/main/go/synthesis/spreader_test_gen_test.go` (create)

**Intent**: The test code generator must produce valid, compilable Go test functions that manually exercise the spreader's `ExpandDimension` call. This provides an independently verifiable test path for spreader correctness before full metric DAG integration (Increment 2). The generated test follows the same manual-test pattern as aggregator test generation in Story257.

**Acceptance criteria**:
- Generated test function correctly constructs a `ChronoconeTensor[float64]` with known data and shape
- Generated test calls `tensorops.ExpandDimension` with correct parameters (source tensor, spreading matrix, srcDim, newDim)
- Generated test verifies output shape matches expected expanded dimensions
- Generated Go code passes `go build` compilation check
- Both equal-distribution and scalar spreading cases produce correct generated code

### Task 4: Create end-to-end spreader examples — Not Started [Small]

Create an end-to-end `with_data/spreaders_basic/` example that validates the full spreader pipeline: CPL → parse → analyze → CIL → synthesize → execute → verify. The example uses a small equal-proportion spreader (e.g., Product={widget,gadget} → spread to Country={USA,Canada} × Product={widget,gadget} with equal 0.5/0.5 distribution). Create the CPL file with concepts, connector, and spreader declaration. Create IO specification, parquet test data, and expected output. Write a test that exercises the full pipeline and verifies results match ground truth.

   a. Create `examples/with_data/spreaders_basic/` directory with standard category structure (core.cpl, example subdirectory with .cpl, iospec, data, and expected output) — Not Started
   b. Write CPL file defining concepts (`Country = [USA, Canada]`, `Product = [widget, gadget]`), a connector with `Country × Product` grain, a connector for source data, and a spreader declaration: `from [Product] to [Country, Product] = Product * 1/size(Country) * Country` — Not Started
   c. Create IO specification JSON defining input connector and output spec for spread result — Not Started
   d. Generate parquet test data with known source values: Product={widget=100.0, gadget=200.0} — Not Started
   e. Create expected output (`.expected.xlsx` or `.expected.json`) with hand-computed ground-truth spread values: USA,widget=50.0, USA,gadget=100.0, Canada,widget=50.0, Canada,gadget=100.0 — Not Started
   f. Create a scalar spreading example (`from [] to [Country]` with source value 42.0 → USA=21.0, Canada=21.0) using a separate subdirectory or variant within `spreaders_basic/` — Not Started
   g. Verify full end-to-end pipeline: `go test ./examples/with_data/spreaders_basic/... -count=1` passes, confirming parse → analyze → synthesize → execute → output matches expected — Not Started

**Files**: `examples/with_data/spreaders_basic/core.cpl` (create), `examples/with_data/spreaders_basic/spreaders_basic/` subdirectory with example .cpl, iospec.json, parquet data, .expected.xlsx (create), Go test file(s) (create)

**Intent**: This example provides the first end-to-end spreader validation. The manual test pattern follows existing `with_data/` examples (metrics_arithmetic, conditional_grain). The spread-then-aggregate round-trip (with matching aggregator covering Country) provides natural verification: spreading with equal distribution to Country, then aggregating back, should recover the original Product values.

**Acceptance criteria**:
- CPL parses and analyzes without errors
- Synthesized Go code compiles and executes
- Manual spreading with `tensorops.ExpandDimension` produces correct expanded values matching ground truth
- Output tensor shape has one more dimension than the input tensor
- Scalar spreading example (`from []`) works correctly: rank-0 scalar expands to 1D tensor
- Spread-then-aggregate round-trip (with sum aggregator) recovers original values
- Example test passes: parse → analyze → synthesize → execute → verify against expected output

## Notes

- Spreaders are applied manually in test code in this increment — not yet integrated into automatic metric computation (that happens in Increment 2)
- Spreading matrix values from concept value sets are materialized at compile time (not loaded at runtime)
- The `with_data/spreaders_basic/` example follows the standard category structure: `core.cpl` with shared infrastructure, subdirectories for each example variant
- Spreaders are the first consumer of the `ExpandDimension` path built in Story254
- File paths in all task descriptions are relative to `source_root` (`src/chronocone_planning_language`)
- All test commands run from `source_root`: `cd src/chronocone_planning_language && go test ./src/main/go/synthesis/...`
- The `from []` (scalar source) case is semantically equivalent to "distribute this single value evenly across the new dimension"
- Equal-proportion spread followed by sum aggregation yields original values, providing a simple round-trip correctness check
- The preamble generator (`preamble_generator.go`) may need to be created if not already built by Story257's aggregator preamble work; if the file exists, add `EmitSpreadingMatrix` to it; if not, create it
