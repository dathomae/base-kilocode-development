# Story242: Lookups Increment 1 — Analyze Phase

## Goal

Implement the analysis phase for basic lookup declarations without grain (Increment 1 of Epic006). This story creates `AnalyzeLookupScope`, `LookupInfo`, `CILLookupNode`, and wires them into the analyzer pipeline. The analysis phase walks the symbol table tree, extracts `ParserLookupDefinition` instances, resolves cross-references (concept names to `AnalysisConceptDefinition`, measure names to `AnalysisMeasureDefinition`), validates name uniqueness and reference existence, and generates one `CILLookupNode` per lookup.

## References

- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md)
- [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md)
- [Plan: Story242 — Lookups Increment 1 (Analyze)](../plans/Story242_lookups-increment1-analyze.md)

## Dependencies

- [Story241_lookups-increment1-parse.md](Story241_lookups-increment1-parse.md) — provides `ParserLookupDefinition` and `LookupScopeDefinition` from the symbol table (Not Started)
- Epic001 (concepts) — `AnalysisDefinitions.Concepts` for dimension resolution
- Epic003 (units/measures) — `AnalysisDefinitions.Measures` for celltype resolution

## Dependent Stories

- Story245 (lookups-increment2-grain) — adds grain handling to lookup analysis
- Story248 (lookups-increment3-metric-references) — uses `CILLookupNode` in metric input references

## Tasks

### Task Sizing

Task 1a (types + tests) is not annotated `[Small]` — it creates two new source files and a test file, introduces new structs and a CILNode interface implementation. The types are modular but the three-file scope exceeds the Small threshold.

Task 1b (wiring) is annotated `[Small]` — it modifies a single file (`analysis_definitions.go`), adds one map field following the existing pattern established by `Concepts`, `Types`, `Measures`, and `Connectors`. The change is purely additive and mechanical.

Tasks 2 through 4 are not annotated `[Small]` — each requires reasoning about multiple interacting structures: Task 2 implements a complex analysis function with symbol table traversal and cross-reference resolution; Task 3 touches the analyzer pipeline registration and requires integration test understanding; Task 4 creates novel CPL example files for a new construct type.

Task 1 was decomposed during review into Tasks 1a and 1b — the original plan's Task 1 had a self-contained preamble (create types + tests, independently compilable and testable) followed by infrastructure wiring (add field to AnalysisDefinitions). The preamble forms a natural task boundary; the split creates two well-focused tasks. The architect-for-story-planning decomposition review confirmed this natural seam.

### Parallel Execution

None — all tasks have sequential dependencies. Each task depends on the output of the preceding task: Task 1b needs Task 1a's types, Task 2 needs Task 1b's Lookups map, Task 3 needs Task 2's function, and Task 4 needs the full pipeline from Task 3.

### Execution Order

```
Task 1a → Task 1b → Task 2 → Task 3 → Task 4
```

### Reintegration

None — no parallel groups.

### Task 1a: Define LookupInfo and CILLookupNode types — Completed

Task 1a was created by decomposing the original Task 1. The original task's subtasks a, b, and d (create types, write tests) form a self-contained preamble: the new types reference only existing interfaces and structs (`CILNode`, `AnalysisConceptDefinition`, `AnalysisMeasureDefinition`, `SourceLocation`), compile independently, and can be verified before being wired into `AnalysisDefinitions`. The architect-for-story-planning decomposition review identified this natural seam.

Subtasks c and e from the original Task 1 (add field to AnalysisDefinitions, verify compile) were extracted into Task 1b.

   a. Create `src/main/go/analysis/lookup_info.go` with the `LookupInfo` struct — Not Started
      
      Fields (Increment 1 only):
      ```go
      type LookupInfo struct {
          Name              string
          Dimensions        []*AnalysisConceptDefinition   // Resolved concept references
          CelltypeMeasure   *AnalysisMeasureDefinition     // Resolved measure reference
          SourceLocation    SourceLocation
      }
      ```
      
      Follow the documentation convention established by `AnalysisConnectorDefinition` in `connectors_analysis.go`: a package-comment line, a doc comment explaining the type's purpose, field-level doc comments, and an example CPL declaration showing the resulting struct.

   b. Create `src/main/go/analysis/cil_lookup_node.go` with the `CILLookupNode` struct implementing `CILNode` — Not Started
      
      ```go
      type CILLookupNode struct {
          LookupName        string
          Dimensions        []string      // Resolved concept names (for CIL)
          Celltype          string        // Resolved measure name (for CIL)
          SourceLocation    SourceLocation
      }
      
      func (n *CILLookupNode) cilNode() {}
      ```
      
      Structure this file following the pattern of `cil/cil_connector.go`. The `cilNode()` marker method is the only method required to satisfy the `CILNode` interface.

   c. Write a unit test for `LookupInfo` struct initialization and `CILLookupNode` interface conformance — Not Started
      
      Create `src/main/go/analysis/lookup_types_test.go`. Verify:
      - A `LookupInfo` can be constructed with its Increment 1 fields (no grain fields)
      - A `CILLookupNode` satisfies the `cil.CILNode` interface (compile-time assertion with `var _ cil.CILNode = (*CILLookupNode)(nil)`)
      - Dimension slices are independent (modifying a caller slice does not affect the stored slice)
      - Source location is preserved

   d. Verify the package compiles: `go build ./src/main/go/analysis/...` — Not Started

### Task 1b: Add Lookups map to AnalysisDefinitions — Completed [Small]

This task was split from the original Task 1 (subtask c). It adds the `Lookups` field to `AnalysisDefinitions` and wires it through `GetAnalysisDefinitions`. The change is a mechanical addition following existing patterns — single file, single concern, no new abstractions.

   a. Add `Lookups map[string]*LookupInfo` field to the `AnalysisDefinitions` struct in `src/main/go/analysis/analysis_definitions.go` — Not Started
      
      Add the field after `Connectors` and before `Metrics`, with a doc comment following the convention of the existing map fields. The doc comment should state: "Lookups is a lookup map from lookup name to its analyzed definition. This map is built during the analysis phase by processing all lookup definitions from the symbol table."

   b. Initialize `Lookups` in `GetAnalysisDefinitions` in `src/main/go/analysis/cpl_analyzer.go` — Not Started
      
      Add `Lookups: a.lookups,` to the return value of `GetAnalysisDefinitions()`. Add a `lookups map[string]*LookupInfo` field to the `CPLAnalyzer` struct (after the `connectorAnalyzer` field) and initialize it in `NewCPLAnalyzer` with `make(map[string]*LookupInfo)`.

   c. Verify the package compiles: `go build ./src/main/go/analysis/...` — Not Started

### Task 2: Implement AnalyzeLookupScope function — Completed

This task implements the core analysis logic for lookup declarations. The function follows the same architectural pattern as `AnalyzeConnectorsScope`: it is registered as a scope-specific handler in `WalkSymbolTableWithHandlers`, walks the symbol table's children for `LookupScopeDefinition` entries, extracts `ParserLookupDefinition` from `SymbolDefinition.Data`, validates all cross-references, and produces `LookupInfo` and `CILLookupNode` outputs. The architect reviewed this task during decomposition — despite 9 subtask steps, all steps flow linearly through a single coherent change (build the function, test it) with no self-contained preamble.

   a. Create `src/main/go/analysis/analyze_lookups.go` with package declaration and imports — Not Started
      
      Imports needed: `fmt`, `github.com/chronocone/chronocone_planning_language/src/main/go/cil`, `github.com/chronocone/chronocone_planning_language/src/main/go/symboltable`.
      
      Add a file-level doc comment describing the function's role in the analysis phase, following the pattern in `connectors_analysis.go`.

   b. Implement the `AnalyzeLookupScope` function skeleton — Not Started
      
      ```go
      func AnalyzeLookupScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
      ```
      
      The function signature matches `AnalysisFunc` (defined in `symbol_table_walk.go`). It receives a `LookupsScope` table (the scope containing all lookup declarations), traverses its children (each child is a `LookupScopeDefinition`), and processes each one.

   c. Write unit tests in `src/main/go/analysis/analyze_lookups_test.go` — Not Started
       
       Test cases (use `package analysis` for white-box testing; tests will fail until implementation is complete):
       - **Valid timeless lookup with single dimension**: Create a CPLAnalyzer with a concept and a measure pre-loaded, call `AnalyzeLookupScope` with a child table carrying a valid `ParserLookupDefinition`, verify `analyzer.lookups` contains the entry and `analyzer.metricAnalyzer.GetCILTrees()` has the key `"lookup:<name>"`.
       - **Valid timeless lookup with multiple dimensions**: Same as above with 2+ dimensions.
       - **Duplicate lookup name error**: Call `AnalyzeLookupScope` twice with the same lookup name; second call must error.
       - **Missing dimension concept error**: Call with a dimension name not in concepts; must error with "undefined concept dimension".
       - **Missing celltype measure error**: Call with a celltype name not in measures; must error with "undefined celltype measure".

   d. Implement symbol table traversal to find `LookupScopeDefinition` entries — Not Started
       
       Iterate `table.Children`. For each child, assert that `child.ScopeType` is the `LookupScopeDefinition` scope type (to be defined in Story241). Extract the `*symboltable.ParserLookupDefinition` from `child.Data` (also defined in Story241). If `child.Data` is nil or the wrong type, return an error with the source location.

   e. Implement name uniqueness check — Not Started
       
       Before processing each lookup, check if `analyzer.lookups[lookupName]` already exists. If so, return an error: `fmt.Errorf("duplicate lookup definition: %s", lookupName)` with the source location.

   f. Implement dimension concept resolution — Not Started
       
       For each dimension name in the `ParserLookupDefinition.DimensionNames` slice, look up in `analyzer.conceptAnalyzer.HasConcept(dimName)`. If not found, return: `fmt.Errorf("lookup %s references undefined concept dimension: %s", lookupName, dimName)`. If found, retrieve via `analyzer.conceptAnalyzer.GetConcept(dimName)` and append a pointer to the copy (use the same pattern as `AnalyzeConnectorsScope`: take a local copy, then `&copy`).

   g. Implement celltype measure resolution — Not Started
       
       Look up `ParserLookupDefinition.CellTypeName` in `analyzer.measures`. If not found, return: `fmt.Errorf("lookup %s references undefined celltype measure: %s", lookupName, cellTypeName)`. If found, store `&measureCopy`.

   h. Build `LookupInfo` and store in analyzer — Not Started
       
       ```go
       analyzer.lookups[lookupName] = &LookupInfo{
           Name:            lookupName,
           Dimensions:      resolvedDimensions,
           CelltypeMeasure: &resolvedCellType,
           SourceLocation:  sourceLocation,
       }
       ```

   i. Build `CILLookupNode` and store in CIL trees — Not Started
       
       Build the CIL tree key as `"lookup:" + lookupName` (to avoid collision with connector CIL tree keys). Populate `CILLookupNode` with resolved names (strings, not pointers): dimension concept names and celltype measure name.
       
       Store via `analyzer.metricAnalyzer.AddCILTree("lookup:"+lookupName, &cil.CILLookupNode{...})`.

   j. Verify all analyzer tests pass: `go test ./src/main/go/analysis/... -count=1` — Not Started

### Task 3: Register AnalyzeLookupScope in the analyzer pipeline — Completed

   a. Add a `LookupsScope` handler entry to the `entryHandlers` map in `cpl_analyzer.go`'s `Analyze` function — Not Started
      
      Add `symboltable.LookupsScope: AnalyzeLookupScope,` in the `entryHandlers` map (position it after the `ConnectorScope` entry, before `MetricsScope`, consistent with CPL document ordering).

   b. Verify `GetAnalysisDefinitions` returns the Lookups map — Not Started
      
      Confirm that `GetAnalysisDefinitions()` in `cpl_analyzer.go` includes the `Lookups` field (done in Task 1b). If an export helper is needed for `lookups`, add it following the connector analyzer pattern.

   c. Write an integration test — Not Started
      
      Create `src/main/go/analysis/analyze_lookups_integration_test.go` (in `package analysis_test`). The test should:
      - Parse a CPL file with one `lookups` section containing a valid lookup declaration (concepts and measures must be defined in the same CPL)
      - Run `analysis.Analyze` on the parsed symbol table
      - Verify `defs.Lookups` contains the expected entry with resolved concept and measure pointers
      - Verify `defs.CILTrees["lookup:<name>"]` is a `*cil.CILLookupNode` with correct dimension names and celltype

   d. Verify integration test passes: `go test ./src/main/go/analysis/... -count=1` — Not Started

### Task 4: Create parse_and_analyze_only analysis-level examples — Completed

   a. Add CPL files that analyze successfully in `examples/parse_and_analyze_only/without_errors/` — Not Started
      
      Create example files demonstrating valid lookup declarations:
      - `lookup_single_dimension.cpl` — One lookup with one dimension concept and one celltype measure
      - `lookup_multiple_dimensions.cpl` — One lookup with three dimension concepts and one celltype measure
      - `lookup_multiple_lookups.cpl` — Two distinct lookups in the same `lookups` section, each with different dimensions and celltypes
      
      Each file must include the concepts and measures sections that the lookups reference.

   b. Add CPL files that produce analysis errors in `examples/parse_and_analyze_only/with_errors/` — Not Started
      
      Create example files demonstrating error cases:
      - `lookup_duplicate_name.cpl` — Two lookups with the same name in the same section
      - `lookup_undefined_dimension.cpl` — A lookup referencing a concept name not defined in the concepts section
      - `lookup_undefined_celltype.cpl` — A lookup referencing a measure name not defined in the measures section

   c. Verify all examples produce expected analysis results — Not Started
      
      For without_errors examples: verify `go test ./src/main/go/analysis/... -count=1` passes and includes the new examples (if the test framework auto-discovers them). For with_errors examples: verify the test framework reports analysis errors. If the existing `analysis_integration_test.go` does not auto-discover new examples, update the test expectations map.

## Test-First Development

This project follows the Logical TDD Lifecycle. Tests must be written before the corresponding implementation code. Each coding task integrates its test-and-implement cycle:

- **Task 1a**: Write the unit test (subtask c) before implementing the structs (subtasks a, b)
- **Task 2**: Write unit tests (subtask c) before implementing the analysis function logic (subtasks d-i). The file and function skeleton (subtasks a, b) may be created first to establish the signature, but all resolution and validation logic follows the tests.
- **Task 3**: Write the integration test (subtask c) before registering the handler (subtask a)
- **Task 4**: Example files are input data, not test code, so they may be created before verifying

## Constraints

- Lookup name uniqueness is enforced within the lookups section — no two lookups may share the same name
- All dimension concept references must resolve to existing entries in `AnalysisDefinitions.Concepts`
- The celltype measure reference must resolve to an existing entry in `AnalysisDefinitions.Measures`
- CIL trees for lookups are keyed with the prefix `"lookup:"` to avoid collision with connector CIL tree keys
- No grain handling exists in this increment — grain is added in Story245
- This story only processes lookup *declarations* — lookup *references* in metric inputs are handled in Story248

## Intent

Task 1a establishes the data structures that carry analysis results (`LookupInfo`) and their CIL representation (`CILLookupNode`). These types are purely additive and independently compilable, following the same pattern as `AnalysisConnectorDefinition` and `CILConnectorNode`.

Task 1b wires the new types into the container (`AnalysisDefinitions`) and the analyzer struct, making them available to the analysis and synthesis pipeline.

Task 2 delivers the core analysis logic: walking the symbol table tree, extracting parsed lookup data, validating all cross-references against the analyzer's concepts and measures maps, enforcing name uniqueness, and producing both the analysis-level and CIL-level outputs. The function follows the same architectural pattern as `AnalyzeConnectorsScope`.

Task 3 integrates the new analysis function into the main `Analyze` entry point, ensuring lookup analysis runs during the standard symbol table walk at the correct point (after connector analysis, before metric analysis).

Task 4 creates example CPL files that exercise the full parse-and-analyze pipeline for lookups, providing regression coverage and documentation of supported patterns and expected error messages.

## Acceptance Criteria

1. `LookupInfo` struct exists in `src/main/go/analysis/lookup_info.go` with fields: `Name`, `Dimensions []*AnalysisConceptDefinition`, `CelltypeMeasure *AnalysisMeasureDefinition`, `SourceLocation`
2. `CILLookupNode` struct exists in `src/main/go/analysis/cil_lookup_node.go` implementing the `CILNode` interface, with fields: `LookupName`, `Dimensions []string`, `Celltype string`, `SourceLocation`
3. `AnalysisDefinitions.Lookups map[string]*LookupInfo` is populated after analysis completes successfully
4. CIL trees for lookups are stored in `AnalysisDefinitions.CILTrees` keyed by `"lookup:<name>"`
5. `AnalyzeLookupScope` validates that all dimension concept names exist in `Concepts` map
6. `AnalyzeLookupScope` validates that the celltype measure name exists in `Measures` map
7. `AnalyzeLookupScope` rejects duplicate lookup names with an error message containing the lookup name
8. `AnalyzeLookupScope` is registered in `WalkSymbolTableWithHandlers` entry handlers in `cpl_analyzer.go`
9. Analysis errors are reported with source location information from the symbol table
10. No grain handling exists (grain is added in Increment 2 — Story245)
11. `go test ./src/main/go/analysis/...` passes with no regressions
12. `go build ./src/main/go/analysis/...` succeeds
