# Story198: CPL Concepts Concept Range Examples

## Goal

Populate the `src/chronocone_planning_language/examples/` directory with dedicated `.cpl` example files that demonstrate Story 128 (Concept Range parser) and Story 129 (Concept Range analyzer) functionality. Examples must be provided both with and without errors, following the naming conventions and commenting style already established in the directory.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [Plan: Concept Range Examples](../plans/concept_range_examples_plan.md)
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story128_cpl-concepts-parser-concept-range.md](finished-stories/Story128_cpl-concepts-parser-concept-range.md) - Done - Parser support for Concept Range generator syntax must be complete
- [Story129_cpl-concepts-analyzer-concept-range.md](finished-stories/Story129_cpl-concepts-analyzer-concept-range.md) - Done - Analyzer resolution for Concept Ranges must be complete

## Dependent Stories

None

## Tasks

1. Create valid example files in `examples/without_errors/` - Completed
   a. Create `concept_range_basic.cpl` - minimal valid Concept Range referencing a Values-based source concept - Completed
   b. Create `concept_range_multiple_from_source.cpl` - multiple concepts extracting different ranges from the same source concept - Completed
   c. Create `concept_range_combined_generators.cpl` - Concept Range mixed with Values and Int Range generators - Completed
   d. Create `concept_range_chained.cpl` - chained Concept Range references (A references B, B references C) - Completed
   e. Create `concept_range_from_int_source.cpl` - Concept Range where the source concept uses an Int Range generator - Completed
   f. Create `concept_range_full_concept.cpl` - range covering the entire source concept from first to last value - Completed
   g. Verify each valid example parses without syntax errors - Completed

2. Create invalid example files in `examples/with_errors/` - Completed
   a. Create `invalid_concept_range_circular_dependency.cpl` - circular dependency between two concepts - Completed
   b. Create `invalid_concept_range_self_reference.cpl` - a concept references itself via a Concept Range - Completed
   c. Create `invalid_concept_range_missing_source.cpl` - Concept Range references a non-existent concept - Completed
   d. Create `invalid_concept_range_missing_start_value.cpl` - start value not present in the referenced source concept - Completed
   e. Create `invalid_concept_range_missing_end_value.cpl` - end value not present in the referenced source concept - Completed
   f. Create `invalid_concept_range_start_after_end.cpl` - start value appears after end value in source ordering - Completed
   g. Create `invalid_concept_range_duplicate_value.cpl` - Concept Range generates a value that already exists from another generator - Completed
   h. Verify each invalid example triggers the expected analyzer/parser error - Completed

3. Run full regression test suite - Completed
   a. Run `go test ./parser/... -run ConceptRange` - Completed
   b. Run `go test ./analysis/... -run ConceptRange` - Completed
   c. Run `go test ./...` for full regression check - Completed

## Constraints

- File naming must use `snake_case.cpl`; prefix error examples with `invalid_`
- Valid examples go to `examples/without_errors/`; invalid examples go to `examples/with_errors/`
- Each file must include a header comment with title, one-line description, and story reference
- Body must use tab indentation and `#` for inline comments
- Examples must be minimal (only the syntax needed to demonstrate the point)
- Must not break existing Int Range, Values, or connector examples

## Intent

The goal is to provide standalone, documented CPL example files that illustrate the Concept Range feature for both learning and regression-testing purposes. Valid examples show the breadth of supported syntax, while invalid examples demonstrate the error cases the analyzer detects.

## Acceptance Criteria

1. Six valid `.cpl` files exist in `examples/without_errors/` covering basic, multiple-from-source, combined-generators, chained, int-source, and full-concept ranges
2. Seven invalid `.cpl` files exist in `examples/with_errors/` covering circular dependency, self-reference, missing source, missing start, missing end, start-after-end, and duplicate value
3. Every file follows the header-comment convention (title, description, story reference)
4. All valid examples parse without syntax errors
5. All invalid examples trigger the expected analyzer/parser errors
6. No regressions in the full CPL test suite

## Notes

**Header comment convention (example):**

```cpl
# Concept Range Basic Example
# Minimal valid Concept Range referencing a Values-based source concept.
# From Story128/Story129: CPL Concepts Parser/Analyzer - Concept Range
```

**Valid example content summary:**

| File                                     | Purpose                                                                                              |
| ---------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| `concept_range_basic.cpl`                | `Quarter=[Values[Q1, Q2, Q3, Q4]]` then `FirstHalf=[Quarter[Q1..Q2]]`                                |
| `concept_range_multiple_from_source.cpl` | Multiple ranges (`FirstHalf`, `SecondHalf`, `FirstQuarterOnly`) from one source                      |
| `concept_range_combined_generators.cpl`  | `Mixed=[Values[A, B], Int[1, 3, 1], Quarter[Q1..Q2]]`                                                |
| `concept_range_chained.cpl`              | `Month=[Int[1, 12, 1]]` → `Quarter=[Month[1..3], ...]` → `FirstQuarter=[Quarter[1..3]]`              |
| `concept_range_from_int_source.cpl`      | `Month=[Int[1, 12, 1]]` then `FirstSemester=[Month[1..6]]`                                           |
| `concept_range_full_concept.cpl`         | `Status=[Values[Active, Inactive, Pending, Archived]]` then `AllStatuses=[Status[Active..Archived]]` |

**Invalid example content summary:**

| File                                            | Expected Error                                      |
| ----------------------------------------------- | --------------------------------------------------- |
| `invalid_concept_range_circular_dependency.cpl` | Circular dependency detected in concept ranges      |
| `invalid_concept_range_self_reference.cpl`      | Circular dependency detected                        |
| `invalid_concept_range_missing_source.cpl`      | Unknown concept error                               |
| `invalid_concept_range_missing_start_value.cpl` | Start value 'Missing' not found in concept 'Source' |
| `invalid_concept_range_missing_end_value.cpl`   | End value 'Missing' not found in concept 'Source'   |
| `invalid_concept_range_start_after_end.cpl`     | Start value 'C' comes after end value 'A'           |
| `invalid_concept_range_duplicate_value.cpl`     | Duplicate value error                               |
