# Story241: Lookups — Increment 1 (Parse)

## Goal

Implement the parsing phase for basic lookup declarations without grain (Increment 1 of Epic006). This story delivers the `CPLLookups.g4` grammar, listener integration for the lookups section, the `ParserLookupDefinition` data structure, and symbol table storage. The grammar is minimal — lookups have no expression body, only structural clauses with identifier references. The parser stores raw concept and measure names as strings; all cross-reference resolution is deferred to analysis (Story242). Lookup scopes in the symbol table are flat (no child scopes).

## References

- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md)
- [ERO006: Lookups Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO006_lookups-parsing-phase.md)
- [Symbol Table Design](../specifications/design/cpl/go-target/symbol-table-design.md)

## Dependencies

- Epic001 (concepts) must be complete — lookup dimensions reference concepts
- Epic003 (units/measures) must be complete — lookup celltypes are measures

## Dependent Stories

- Story242 (lookups analysis phase) depends on this story for parsed definitions

## Constraints

- Grammar must use `DIMENSIONS` (plural, section-level clause) which is distinct from the existing `DIMENSION` (singular) token used in metric input dimension references
- `LOOKUPS` (plural, section keyword) is distinct from `LOOKUP` (singular, to be added in Increment 3 for the `lookup` input kind)
- No cross-reference validation at the parser level — raw strings only
- `lookups` sections with zero lookup declarations produce no errors
- The `indent` token enforces readable indentation but does not define scope boundaries
- Lookup scopes in the symbol table are flat (no per-lookup child scopes); definitions are stored as `SymbolDefinition` entries in the `LookupsScope`'s `Definitions` list

## Intent

This story delivers the parsing phase for basic lookup declarations. Lookups represent timeless dimensional relationships (concepts mapped to a celltype measure) and are structurally simpler than metrics — they have no expression body, no inputs, and no calculate clause. The parser validates syntax and stores raw identifier references for later resolution.

## Tasks

### Task Sizing

Tasks 1, 2, and 5 are annotated `[Small]` — each touches at most 2 files, is purely additive or mechanically extends existing patterns, and addresses a single narrow concern. Tasks 3 and 4 are not annotated — Task 3 requires reasoning across the ANTLR listener interface, symbol table stack protocol, and scope/definition storage conventions; Task 4 integrates across 3 production files with compilation coupling. A sizing analysis was performed for each task against the rules in the stories template's Task Sizing section; Tasks 3 and 4 were identified as not small because they involve multiple interacting structures and coordinated multi-file changes.

### Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 creates `CPLLookups.g4` in `src/main/antlr/` and adds a lexer token. Task 2 creates `parser_lookup_definition.go` in `src/main/go/parser/` and adds scope/definition types in `src/main/go/symboltable/`. Disjoint directories, disjoint files, no compile-time dependency — the Go types defined in Task 2 do not depend on ANTLR-generated code. Each task compiles and tests independently.

**Architect note on plan deviation**: The original plan proposed Group A = (Tasks 2, 3) but the architect-for-story-planning analysis determined this cannot parallelize. Task 2 defines `ParserLookupDefinition`, `LookupsScope` (ScopeType), and `LookupScopeDefinition` (DefinitionType); Task 3's listener directly references all three types. Since both tasks modify files in the same Go package (`parser`), a compile-time dependency exists — Task 3 cannot compile in its worktree without Task 2's type definitions present. The corrected parallelism pairs Tasks 1 and 2 instead, which are fully disjoint.

### Execution Order

```
[Task 1 ‖ Task 2] → Task 3 → Task 4 → Task 5
```

- Task 1 and Task 2 run in parallel (Group A)
- Task 3 depends on Task 1 (ANTLR-generated listener base from grammar) and Task 2 (type definitions)
- Task 4 depends on Task 3 (integration needs the listener)
- Task 5 depends on all preceding tasks (examples need full pipeline)

### Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (completely disjoint directories — `src/main/antlr/` vs `src/main/go/parser/` + `src/main/go/symboltable/`, no shared infrastructure files)
- Integration tests: `go test ./src/main/go/symboltable/... ./src/main/go/parser/... -count=1 -timeout 60s` — verify both ParserLookupDefinition compiles and symbol table enum changes are valid after merging both worktree branches onto the story branch
- Watch for conflicts in: `go.sum` (ANTLR tooling dependency resolution may differ per worktree). No source files overlap between the two tasks.

---

### Task 1: Create CPLLookups.g4 grammar and add LOOKUPS lexer token — Completed [Small]

**Intent**: Define the ANTLR parser grammar for lookup declarations and register the `LOOKUPS` section keyword in the lexer. The grammar handles indentation, dimension lists, and celltype references using the same bracket and identifier tokens already defined in the lexer.

**Acceptance criteria**:
1. `CPLLookups.g4` compiles via ANTLR without grammar errors
2. The generated Go parser can tokenize a `lookups` keyword as the `LOOKUPS` token
3. ANTLR code generation produces valid Go files in `src/main/go/parser/antlr/`

   a. Create `src/main/antlr/CPLLookups.g4` — a `parser grammar` importing `ChronoconePlanningLanguage` with the following rules:
      - `lookupsSection: LOOKUPS lookupDeclaration*` — permits zero or more declarations
      - `lookupDeclaration: indent identifier dimensionsClause celltypeClause` — all clauses required
      - `dimensionsClause: indent DIMENSIONS LBRACKET identifier (COMMA identifier)* RBRACKET` — comma-separated concept names
      - `celltypeClause: indent CELLTYPE identifier` — a single measure name
      - Use `LBRACKET`/`RBRACKET` (not LSQUARE/RSQUARE) to match the existing lexer token names
      - Not Started

   b. Add the `LOOKUPS` section keyword token to `src/main/antlr/ChronoconePlanningLanguageLexer.g4`:
      - Insert `LOOKUPS: 'lookups';` adjacent to the existing `LOOKUP` token (line 36), with a comment noting it is the section keyword for lookups
      - Note: `DIMENSIONS` (line 28) and `CELLTYPE` (line 40) already exist in the lexer and do not need to be added
      - Not Started

   c. Write a grammar-level test: verify that running the ANTLR Gradle generation task (`./gradlew generateGrammarSource`) succeeds without errors and produces valid Go code in `src/main/go/parser/antlr/` — Not Started

   d. Verify the generated ANTLR Go package compiles: `go build ./src/main/go/parser/antlr/...` — Not Started

---

### Task 2: Define ParserLookupDefinition struct and add symbol table scope types — Completed [Small]

**Intent**: Create the data structure that carries parsed lookup information through the pipeline and extend the symbol table's type enumerations to support lookup scopes.

**Acceptance criteria**:
1. `ParserLookupDefinition` struct compiles with fields for name, dimension names, celltype name, and source location
2. `LookupsScope` appears in the `ScopeType` enum and its `String()` switch
3. `LookupScopeDefinition` appears in the `DefinitionType` enum and its `String()` switch
4. `go test ./src/main/go/symboltable/... && go test ./src/main/go/parser/...` passes

   a. Create `src/main/go/parser/parser_lookup_definition.go` with the following struct:
      ```go
      type ParserLookupDefinition struct {
          Name           string                      // Lookup identifier
          Dimensions     []string                    // Raw concept names (no grain in Inc 1)
          Celltype       string                      // Raw measure name
          SourceLocation symboltable.SourceLocation  // For error reporting
      }
      ```
      Include a package comment referencing the ERO006 parse phase design.
      — Not Started

   b. Add scope and definition types to `src/main/go/symboltable/symbol_table.go`:
      - Add `LookupsScope` to the `ScopeType` enum (inserted alphabetically/numerically after existing scope types, before `MeasuresScope` if following the existing ordering, or at the end after `DefaultBodyScope` — follow the existing convention within the file)
      - Add `LookupScopeDefinition` to the `DefinitionType` enum (inserted after `MetricScopeDefinition` to group with other section-level definition types)
      - Add corresponding `case` clauses to `ScopeType.String()` and `DefinitionType.String()` switch statements
      — Not Started

   c. Write a unit test in `src/main/go/parser/parser_lookup_definition_test.go`:
      - Verify that `ParserLookupDefinition` can be initialized with all fields
      - Verify the `SourceLocation` field correctly stores line and column information
      - Not Started

   d. Verify the parser and symboltable packages compile: `go build ./src/main/go/symboltable/... ./src/main/go/parser/...` — Not Started

---

### Task 3: Implement lookups listener methods — Completed

**Intent**: Implement the ANTLR listener callbacks for the lookups section. When the walker enters the lookups section, a `LookupsScope` is pushed onto the symbol table stack. When a lookup declaration is exited, identifiers are extracted from the parse tree, a `ParserLookupDefinition` is assembled, and it is stored as a `SymbolDefinition` in the `LookupsScope.Definitions` list.

**Acceptance criteria**:
1. Entering a `lookupsSection` pushes a `LookupsScope` child scope onto the symbol table stack
2. Exiting a `lookupsSection` pops the scope
3. Exiting a `lookupDeclaration` extracts name, dimension names, and celltype name from the context
4. The assembled `ParserLookupDefinition` is stored in a `SymbolDefinition` with `DefinitionType = LookupScopeDefinition`
5. Zero lookup declarations produce an empty `Definitions` list with no errors
6. Listener unit tests pass, verifying correct symbol table tree structure

   a. Create `src/main/go/parser/cpl_lookups_listener.go` — define methods on the existing `*CPLListener` receiver (following the same pattern as `cpl_metrics_listener.go`, `cpl_connectors_listener.go`, etc.):
      - Package comment linking to this story and the `CPLLookups.g4` grammar file
      - Import `symboltable` and the ANTLR-generated `antlr` package
      — Not Started

   b. Implement `EnterLookupsSection` / `ExitLookupsSection`:
      - `EnterLookupsSection`: peek the current symbol table, call `AddChild(symboltable.LookupsScope)`, push the new scope onto `l.symbolStack`. Add error handling via `l.AddError` if stack operations fail (follow the pattern in `EnterMetricsSection` in `cpl_metrics_listener.go`). Call `l.logEnter`.
      - `ExitLookupsSection`: pop the symbol stack. Call `l.logExit`.
      - Not Started

   c. Implement `ExitLookupDeclaration`:
      - Peek the current scope (should be `LookupsScope`)
      - Extract the lookup name from `ctx.Identifier()` — report an error if the identifier is missing
      - Extract dimension names: iterate `ctx.DimensionsClause().AllIdentifier()` (ANTLR generates an accessor for the `dimensionsClause` sub-rule), collect each identifier's text
      - Extract celltype name: get the identifier from `ctx.CelltypeClause().Identifier()`
      - Create a `ParserLookupDefinition` with the extracted values and a source location derived from the name identifier token
      - Create a `symboltable.SymbolDefinition` with `DefinitionType: symboltable.LookupScopeDefinition`, `Data` set to the `ParserLookupDefinition`, and `Location` set from the identifier token
      - Append the `SymbolDefinition` to `currentScope.Definitions`
      - Handle nil/missing sub-contexts gracefully (report errors for missing dimensions or celltype clauses)
      - Not Started

   d. Implement a helper method `extractDimensionNames(ctx *antlr.DimensionsClauseContext) []string`:
      - Use `ctx.AllIdentifier()` to get all identifiers in the dimensions clause
      - Map to string names, returning an empty slice if no identifiers are present
      - Not Started

   e. Write listener unit tests in `src/main/go/parser/cpl_lookups_listener_test.go`:
      - Test: parsing a CPL document with a single lookup declaration produces one `SymbolDefinition` in `LookupsScope.Definitions` with correct name, dimension names, celltype name, and `DefinitionType == LookupScopeDefinition`
      - Test: parsing a CPL document with multiple lookup declarations produces multiple definitions, ordered correctly
      - Test: parsing an empty `lookups` section (section keyword with no declarations) produces a `LookupsScope` with an empty `Definitions` list and no errors
      - Each test parses via `parser.Parse(input)` and inspects the resulting symbol table tree by finding the `LookupsScope` child of the root table
      - Not Started

   f. Verify all listener tests pass: `go test ./src/main/go/parser/... -run Lookups -count=1 -v` — Not Started

---

### Task 4: Integrate lookups into the root grammar and main parser pipeline — Completed

**Intent**: Wire the `CPLLookups` grammar into the root `ChronoconePlanningLanguage` grammar so that `lookups` sections are recognized during parsing, and ensure the ANTLR code generation produces the necessary listener base.

**Acceptance criteria**:
1. `ChronoconePlanningLanguage.g4` imports `CPLLookups` and includes `lookupsSection` in the `statement` rule
2. `./gradlew generateGrammarSource` succeeds and produces generated code that includes lookups-related listener methods
3. Parsing a CPL file with a `lookups` section populates the symbol table with `LookupScopeDefinition` entries in `LookupsScope.Definitions`
4. Parsing a CPL file with malformed lookups syntax (missing `dimensions` or `celltype` clause) produces parse errors

   a. Add `import CPLLookups;` to the root grammar file `src/main/antlr/ChronoconePlanningLanguage.g4`:
      - Insert `CPLLookups` into the existing import line: `import CPLConcepts, CPLMeasures, CPLTypes, CPLUnits, CPLConnectors, CPLMetrics, CPLLookups;`
      — Not Started

   b. Add `| lookupsSection` to the `statement` rule in `src/main/antlr/ChronoconePlanningLanguage.g4`:
      - Append to the existing alternatives: `statement: conceptsSection | typesSection | unitsSection | measuresSection | connectorsSection | metricsSection | lookupsSection;`
      — Not Started

   c. Regenerate ANTLR Go code:
      - Run `./gradlew generateGrammarSource` from `source_root` (`src/chronocone_planning_language`)
      - Verify the generated `chronoconeplanninglanguage_listener.go` and `chronoconeplanninglanguage_base_listener.go` now include `EnterLookupsSection`/`ExitLookupsSection` and `EnterLookupDeclaration`/`ExitLookupDeclaration` methods
      — Not Started

   d. Verify the full parser pipeline:
      - Write an integration test (or extend existing parser tests) that calls `parser.Parse()` with a CPL string containing a valid `lookups` section with one lookup declaration
      - Assert the root symbol table contains a `LookupsScope` child, and that child's `Definitions` list contains one `SymbolDefinition` with `DefinitionType == LookupScopeDefinition` and `Data` of type `*parser.ParserLookupDefinition`
      - Command: `go test ./src/main/go/parser/... -run "Lookups|Lookup" -count=1 -v`
      — Not Started

   e. Verify parser error handling:
      - Test parsing a CPL string with a `lookups` section containing a lookup declaration that is missing the `dimensions` clause — expect a parse error
      - Test parsing a CPL string with a lookup declaration missing the `celltype` clause — expect a parse error
      - Command: same as (d) — include error-case test functions
      — Not Started

---

### Task 5: Create parse_and_analyze_only test examples — Completed [Small]

**Intent**: Provide example CPL files that exercise the lookups parser for both valid and invalid input. These examples serve as documentation and regression tests for the parsing phase.

**Acceptance criteria**:
1. Valid examples parse successfully with no errors
2. Error examples produce parse errors (not nil return)
3. Examples follow the existing `parse_and_analyze_only/` directory structure and naming conventions

   a. Create valid lookup example CPL files under `src/chronocone_planning_language/examples/parse_and_analyze_only/without_errors/`:
      - `lookups_single_dimension.cpl`: a lookups section with one lookup declaration having a single dimension (e.g., `SalesLookup dimensions[Store] celltype RevenueMeasure`)
      - `lookups_multiple_dimensions.cpl`: a lookups section with one lookup declaration having multiple comma-separated dimensions (e.g., `ProductLookup dimensions[Store, Product, Region] celltype SalesMeasure`)
      - `lookups_multiple_declarations.cpl`: a lookups section with two or more lookup declarations
      - `lookups_empty_section.cpl`: a lookups section keyword with no declarations (empty body) — should parse with no errors
      — Not Started

   b. Create error example CPL files under `src/chronocone_planning_language/examples/parse_and_analyze_only/with_errors/`:
      - `lookups_missing_dimensions.cpl`: a lookup declaration missing the `dimensions` clause entirely (e.g., `BadLookup celltype SomeMeasure`)
      - `lookups_missing_celltype.cpl`: a lookup declaration missing the `celltype` clause entirely (e.g., `BadLookup dimensions[Store]`)
      - `lookups_empty_dimension_list.cpl`: a lookup declaration with `dimensions[]` (empty bracket list) — should produce a parse error
      - `lookups_malformed_syntax.cpl`: a lookup declaration with invalid syntax (e.g., missing identifier, extra tokens)
      — Not Started

   c. Verify all with_errors examples produce parse errors and all without_errors examples parse successfully:
      - Write a Go test using `parser.Parse()` to parse each example file, asserting expected pass/fail outcome
      - Command: `go test ./src/main/go/parser/... -run "LookupsExamples|LookupExamples" -count=1 -v`
      — Not Started

## Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task (Tasks 1–4), tests are written before or alongside the implementation within the same task. Task 1 includes grammar-level validation (ANTLR generation success). Task 2 includes unit tests for struct initialization. Task 3 includes listener unit tests that verify symbol table tree structure. Task 4 includes integration tests for the full pipeline. Task 5 is a data-only task creating example directories — no implementation code.

## Notes

- The `LOOKUPS` lexer token should be inserted near the existing `LOOKUP` token (line 36 of `ChronoconePlanningLanguageLexer.g4`) with a comment distinguishing the section keyword from the future input kind keyword
- `DIMENSIONS` (plural, for the dimensions clause in lookups/connectors) already exists in the lexer at line 28 and is distinct from `DIMENSION` (singular, line 35, used in metric input declarations)
- `CELLTYPE` already exists at line 40
- The grammar uses `LBRACKET`/`RBRACKET` for bracket tokens, matching the existing lexer and the `CPLConnectors.g4` convention
- The listener methods are defined on the existing `*CPLListener` receiver, not on a separate listener struct — following the project convention where all section listener methods live on the single `CPLListener` type
- Lookup scopes are flat: there is no per-lookup child scope (unlike metrics, which use `MetricScope` child scopes). All lookup definitions are stored as `SymbolDefinition` entries in the `LookupsScope.Definitions` list
- ANTLR regeneration uses `./gradlew generateGrammarSource` from `source_root` (`src/chronocone_planning_language`). The Gradle task processes the root grammar, which imports sub-grammars — no separate build configuration is needed for `CPLLookups.g4`
- `source_root` for all build/test commands is `src/chronocone_planning_language`
