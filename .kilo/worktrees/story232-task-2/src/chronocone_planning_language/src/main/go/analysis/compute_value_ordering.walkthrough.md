# Story126 Task 1 Walkthrough: ComputeValueOrdering

## Overview

Task 1 implements [`ComputeValueOrdering`](concept_analysis.go:564), the core algorithm that computes the final ordered list of concept values when extensions with `before` and `after` insertion points are applied. This is part of [Story126: CPL Concepts Analyzer - Before/After](../../../memory-bank/stories/Story126_cpl-concepts-analyzer-before-after.md) and [Epic001: CPL Concepts Implementation](../../../memory-bank/epics/Epic001_cpl-concepts-implementation.md) — Increment 6.

The analyzer receives base values from an initial concept declaration and a sequence of insertion requests from concept extensions. [`ComputeValueOrdering`](concept_analysis.go:564) processes these requests in order, validates each insertion point, inserts new values at the computed positions, and returns a slice of [`AnalysisConceptValueDefinition`](concept_analysis.go:23) values with correct zero-based [`Position`](concept_analysis.go:29) indices.

## Key Types

### InsertionRequest

The [`InsertionRequest`](concept_analysis.go:538) struct pairs an [`ast.InsertionPoint`](../../../src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go:60) with the new values to insert. It separates positional information from the payload:

```go
type InsertionRequest struct {
    Position ast.InsertionPoint
    Values   []string
}
```

The [`Position`](concept_analysis.go:539) field holds either a [`BeforeValue`](../../../src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go:61) or an [`AfterValue`](../../../src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go:62), and the [`Values`](concept_analysis.go:540) field holds the strings to insert at that position. This separation lets the parser produce the positional anchor while the analyzer collects the values that belong there.

### Function Signature

```go
func ComputeValueOrdering(
    baseValues []string,
    insertions []InsertionRequest,
) ([]AnalysisConceptValueDefinition, error)
```

The function takes the ordered base values from the initial concept declaration and a slice of insertion requests from extensions. It returns the fully ordered value list or an error if any insertion point is invalid.

## Algorithm Walkthrough

The algorithm operates in three steps: copy the base values, process each insertion request in order, and convert the result into analysis-level definitions.

Consider the example from Story126:

```cpl
concepts
    Priority=[Values[Low, Medium, High]]

concepts
    Priority=[Values[Critical]] before Low
    Priority=[Values[Urgent]] after Medium
```

**Step 1: Initialize the ordered slice.**

The algorithm copies the base values into a working slice:

```
ordered = [Low, Medium, High]
```

**Step 2: Process insertion requests sequentially.**

Insertion 1 inserts `[Critical]` before `Low`. The algorithm calls [`findValue`](concept_analysis.go:613) to locate `Low` at index 0, sets `insertPos = 0`, and calls [`insertValues`](concept_analysis.go:623) to place `Critical` at that position:

```
ordered = [Critical, Low, Medium, High]
```

Insertion 2 inserts `[Urgent]` after `Medium`. [`findValue`](concept_analysis.go:613) locates `Medium` at index 2, sets `insertPos = 3` (one past the anchor), and inserts `Urgent`:

```
ordered = [Critical, Low, Medium, Urgent, High]
```

Because insertions mutate the working slice, subsequent requests can reference values added by earlier ones.

**Step 3: Build the result slice.**

The algorithm converts the final string slice into [`AnalysisConceptValueDefinition`](concept_analysis.go:23) values, assigning each entry a [`Position`](concept_analysis.go:29) equal to its index:

```
result = [
    {Name: "Critical", Position: 0},
    {Name: "Low",      Position: 1},
    {Name: "Medium",   Position: 2},
    {Name: "Urgent",   Position: 3},
    {Name: "High",     Position: 4},
]
```

## Helper Functions

### findValue

[`findValue`](concept_analysis.go:613) performs a linear scan over the ordered slice and returns the index of the target value, or `-1` and `false` if the value is not present:

```go
func findValue(ordered []string, target string) (int, bool)
```

This simple scan is sufficient because concept value lists are typically small, and correctness matters more than asymptotic complexity here. The function searches the current state of the ordered slice, so it naturally sees values inserted by prior requests.

### insertValues

[`insertValues`](concept_analysis.go:623) inserts a slice of strings into the ordered slice at a given position:

```go
func insertValues(ordered []string, pos int, values []string) []string
```

If the values slice is empty, the function returns the ordered slice unchanged. Otherwise it builds a new slice by concatenating the prefix before `pos`, the new values, and the suffix from `pos` onward. This preserves the relative order of existing values on both sides of the insertion point.

## Edge Cases

The implementation handles several edge cases explicitly:

**Empty insertions slice.** When no extensions are present, [`ComputeValueOrdering`](concept_analysis.go:564) copies the base values, skips the insertion loop, and returns the base values unchanged as [`AnalysisConceptValueDefinition`](concept_analysis.go:23) values.

**Empty base values with insertions.** If the base values slice is empty and an insertion request references an anchor value, [`findValue`](concept_analysis.go:613) cannot locate the anchor. The function returns an error because no anchor can exist in an empty list.

**Empty Values slice in InsertionRequest.** If an insertion request specifies a valid anchor but an empty [`Values`](concept_analysis.go:540) slice, [`insertValues`](concept_analysis.go:623) returns the ordered slice unchanged. The anchor is still validated, so requesting `before NonExistent` with no values still produces an error.

**Single-element base value list.** The algorithm works correctly when the base list contains exactly one value. Inserting before the sole element places the new value at index 0; inserting after it places the new value at index 1.

**Insertions referencing values added by prior insertions.** Because the working slice is updated after each request, a later insertion can use a value introduced by an earlier one as its anchor. This supports chaining extensions where each layer builds on the previous result.

## Error Conditions

[`ComputeValueOrdering`](concept_analysis.go:564) returns an error in three situations:

**Neither BeforeValue nor AfterValue is set.** If an [`InsertionRequest.Position`](concept_analysis.go:539) has empty [`BeforeValue`](../../../src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go:61) and [`AfterValue`](../../../src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go:62), the function returns an error indicating that the insertion must specify one direction.

**Both BeforeValue and AfterValue are set.** If both fields are non-empty, the function returns an error because the insertion point is ambiguous. The grammar prevents this at the syntax level, but the check defends against malformed programmatically constructed requests.

**Referenced anchor value not found.** If the anchor value specified in [`BeforeValue`](../../../src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go:61) or [`AfterValue`](../../../src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go:62) does not exist in the current ordered slice (including values added by prior insertions), [`findValue`](concept_analysis.go:613) returns `false` and the function reports which anchor was missing.

## Test Coverage

The test suite in [`compute_value_ordering_test.go`](compute_value_ordering_test.go) contains nine test functions with 27 table-driven sub-cases covering all scenarios:

- [`TestComputeValueOrdering_BeforeClause`](compute_value_ordering_test.go:16) — 4 cases: single before insertion, before first value, before middle value, before only value in single-element base.
- [`TestComputeValueOrdering_AfterClause`](compute_value_ordering_test.go:82) — 4 cases: single after insertion, after last value, after middle value, after only value in single-element base.
- [`TestComputeValueOrdering_MultipleExtensions`](compute_value_ordering_test.go:148) — 4 cases: apply before then after, apply multiple befores, multiple values in single insertion, insertion referencing value from prior insertion.
- [`TestComputeValueOrdering_InvalidBeforeValue`](compute_value_ordering_test.go:216) — 2 cases: before non-existent value in non-empty list, before value in empty base list.
- [`TestComputeValueOrdering_InvalidAfterValue`](compute_value_ordering_test.go:257) — 2 cases: after non-existent value in non-empty list, after value in empty base list.
- [`TestComputeValueOrdering_OriginalOrderPreserved`](compute_value_ordering_test.go:298) — 3 cases: preserve order with before insertion, with after insertion, and with multiple insertions.
- [`TestComputeValueOrdering_EmptyInsertions`](compute_value_ordering_test.go:356) — 2 cases: empty insertions with populated base values, empty insertions with single base value.
- [`TestComputeValueOrdering_EmptyBaseValues`](compute_value_ordering_test.go:399) — 2 cases: empty base values with before insertion, empty base values with after insertion.
- [`TestComputeValueOrdering_InvalidInsertionPoint`](compute_value_ordering_test.go:440) — 2 cases: both before and after set, neither before nor after set.
- [`TestComputeValueOrdering_EmptyValuesSlice`](compute_value_ordering_test.go:485) — 2 cases: empty values with before anchor, empty values with after anchor.

Each test verifies both the value names and their [`Position`](concept_analysis.go:29) fields, ensuring the analysis-phase index remains consistent with the computed order.

## Files Modified

- [`concept_analysis.go`](concept_analysis.go) — added [`InsertionRequest`](concept_analysis.go:538), [`ComputeValueOrdering`](concept_analysis.go:564), [`findValue`](concept_analysis.go:613), and [`insertValues`](concept_analysis.go:623)
- [`compute_value_ordering_test.go`](compute_value_ordering_test.go) — 9 test functions with 27 sub-cases

## Key Decisions

- The working slice is a `[]string` rather than `[]AnalysisConceptValueDefinition` during insertion processing. This avoids repeatedly rebuilding position indices until the final step, keeping the insertion logic simple.
- [`findValue`](concept_analysis.go:613) uses a linear scan because concept value lists are small; an O(n) search over a typically short slice is fast enough and easier to reason about than a map-based approach.
- [`insertValues`](concept_analysis.go:623) allocates a new slice for each insertion rather than modifying in place with `append`. This avoids subtle slice-sharing bugs and makes the operation straightforward to read.
- Position indices are assigned only after all insertions complete, so the final [`Position`](concept_analysis.go:29) values reflect the definitive order.
