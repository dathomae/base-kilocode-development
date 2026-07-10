# Story210: CPL Concepts Conditional Examples

## Goal

Populate the `src/chronocone_planning_language/examples/` directory with dedicated `.cpl` example files that demonstrate Story 121 (CPL Concepts Parser — Conditional) and Story 122 (CPL Concepts Analyzer — Conditional) functionality. Examples must be provided both with and without errors, following the naming conventions and commenting style already established in the directory.

## References

- [Plan: Augment Examples](../plans/augment_examples_plan.md)
- [Story121: CPL Concepts Parser — Conditional](Story121_cpl-concepts-parser-conditional.md)
- [Story122: CPL Concepts Analyzer — Conditional](Story122_cpl-concepts-analyzer-conditional.md)
- [AST: cpl_concepts.go](../../src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go)
- [Test: cpl_concepts_conditional_test.go](../../src/chronocone_planning_language/src/main/go/parser/cpl_concepts_conditional_test.go)
- [Existing examples README](../../src/chronocone_planning_language/examples/README.md)

## Dependencies

- [Story121_cpl-concepts-parser-conditional.md](Story121_cpl-concepts-parser-conditional.md) — Done
- [Story122_cpl-concepts-analyzer-conditional.md](Story122_cpl-concepts-analyzer-conditional.md) — Implemented

## Dependent Stories

None

## Status: Done

## Tasks

1. Create valid conditional example files in `examples/without_errors/` — Completed
   a. Create `conditional_concept_basic.cpl` with minimal valid conditional mapping parent `Country` to child `State` values — Completed
   b. Create `conditional_concept_multiple_parent_values.cpl` with multiple parent values each mapping to different child sets — Completed
   c. Create `conditional_concept_default_values.cpl` with default unconditional child values alongside conditional overrides — Completed
   d. Create `conditional_concept_with_int_range.cpl` with conditional spec using Int range generator (`Year[2024] => Month=[Int[1, 3, 1]]`) — Completed
   e. Create `conditional_concept_with_concept_range.cpl` with conditional spec using Concept range generator (`Year[2024] => FirstHalf=[Quarter[Q1..Q2]]`) — Completed
   f. Create `conditional_concept_multiple_sections.cpl` with parent defined in one `concepts` section and conditional specs in a later section — Completed
   g. Verify each valid example parses and analyzes without errors — Completed

2. Create invalid parser-error example files in `examples/with_errors/` — Completed
   a. Create `invalid_conditional_concept_missing_arrow.cpl` — missing `=>` between parent value and child definition — Completed
   b. Create `invalid_conditional_concept_missing_parent_value.cpl` — empty brackets for parent value — Completed
   c. Create `invalid_conditional_concept_missing_equals.cpl` — missing `=` in child concept line — Completed
   d. Create `invalid_conditional_concept_empty_concept_line.cpl` — arrow with no concept line after it — Completed
   e. Create `invalid_conditional_concept_missing_close_bracket.cpl` — missing `]` closing bracket around parent value — Completed
   f. Verify each invalid example triggers the expected parser error — Completed

3. Create invalid analyzer-error example files in `examples/with_errors/` — Completed
   a. Create `invalid_conditional_concept_undefined_parent.cpl` — parent concept referenced in condition was never defined — Completed
   b. Create `invalid_conditional_concept_invalid_parent_value.cpl` — parent value in condition does not exist in the parent concept — Completed
   c. Create `invalid_conditional_concept_duplicate_default.cpl` — more than one default unconditional specification for the same child concept — Completed
   d. Create `invalid_conditional_concept_root_conditional.cpl` — conditional spec on a root concept (a concept with no parent) — Completed
   e. Verify each invalid example parses successfully but triggers the expected analyzer error — Completed

4. Update `examples/README.md` with entries for all new files — Completed
   a. Add entry for each valid example with title, code block snippet, "What it demonstrates" bullets, and story reference — Completed
   b. Add entry for each invalid example with expected error description — Completed

5. Run full CPL test suite regression check — Completed
   a. Run `go test ./parser/... -run Conditional` to verify no parser regressions — Completed
   b. Run `go test ./analysis/... -run Conditional` to verify no analyzer regressions — Completed
   c. Run `go test ./...` for full regression check — Completed
   d. Verify each file in `without_errors/` parses and analyzes without errors — Completed
   e. Verify each file in `with_errors/` produces the expected error (parse-time or analysis-time as appropriate) — Completed

## Test-First Development

This story follows the Logical TDD Lifecycle. For each example file created, verification against the existing parser and analyzer tests is performed immediately to confirm the example produces the expected behavior (successful parse/analysis for valid files, expected errors for invalid files). The test-and-implement cycle is integrated within each task rather than broken out as separate tasks.

## Constraints

- All example files must follow the established header-comment convention (title, description, story reference).
- Valid examples use tab indentation, `#` for inline comments, and minimal syntax.
- File naming uses `snake_case.cpl`; error examples are prefixed with `invalid_`.
- Valid examples go to `examples/without_errors/`; invalid examples go to `examples/with_errors/`.
- Examples must remain consistent with the current CPL grammar and analyzer implementation.
- Must not break existing concepts, before/after, or concept-range examples.

## Intent

Provide standalone `.cpl` example files that serve as living documentation for the conditional concept value specification syntax (`Parent[Value] => Child=[Values[...]]`), making it easy for users and maintainers to understand expected usage and common error cases without reading test code.

## Acceptance Criteria

1. Six valid `.cpl` files exist in `examples/without_errors/` covering basic, multiple-parent-values, default-values, int-range, concept-range, and multiple-sections conditional specs.
2. Five parser-error `.cpl` files exist in `examples/with_errors/` (prefixed `invalid_conditional_concept_`) testing missing arrow, missing parent value, missing equals, empty concept line, missing close bracket.
3. Four analyzer-error `.cpl` files exist in `examples/with_errors/` (prefixed `invalid_conditional_concept_`) testing undefined parent, invalid parent value, duplicate default, root conditional.
4. Every file follows the header-comment convention (title, description, story reference).
5. All valid examples parse and analyze without errors.
6. All invalid examples trigger the expected error (parse-time or analysis-time).
7. README updated with entries for all new files.
8. No regressions in the full CPL test suite.

## Notes

- The inline CPL strings in `cpl_concepts_conditional_test.go` can be used as the basis for standalone example files.
- The existing `examples/` directory already contains Concept Range and Before/After examples that establish the conventions to follow.

**Header comment convention (example):**

```cpl
# Conditional Basic Example
# Minimal valid conditional specification mapping parent values to child concept values.
# From Story121/Story122: CPL Concepts Parser/Analyzer — Conditional
```

**Syntax reference:**

```
ParentConcept[ParentValue] => ChildConcept=[Generator[...]]
```

Example: `Country[USA] => State=[Values[NY, CA, TX]]`

**Key Story 121 parser capabilities demonstrated:**

- `ValuesGenerator`, `IntRangeGenerator`, `ConceptRangeGenerator` in conditional position
- Multiple conditional specs for same/different child concepts
- Cross-section conditional specs (parent in section 1, condition in section 2)

**Key Story 122 analyzer validations demonstrated:**

- Parent concept must exist
- Parent value must exist in parent concept
- At most one default specification per child concept
- Conditional specs only valid for child concepts (not root)
- Parent-specific value resolution: `GetValuesForParent("USA")` returns condition-specific values
