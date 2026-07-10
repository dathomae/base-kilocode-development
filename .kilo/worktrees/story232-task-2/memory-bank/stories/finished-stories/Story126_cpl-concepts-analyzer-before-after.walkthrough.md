# Task 2 Walkthrough: ConceptAnalyzer Before/After Support

## Purpose

This walkthrough documents the implementation of concept extension detection and merging in the CPL analyzer (Story126, Task 2). It enables the analyzer to recognize when a concept with the same name appears in a later `concepts` section with a `before` or `after` clause, merge the extension's values into the existing concept, and recompute the value ordering via [`ComputeValueOrdering()`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:564).

## Data Flow

### 1. Parser stores InsertionPoint (Story125)

The CPL listener's [`ExitConceptExtension`](src/chronocone_planning_language/src/main/go/parser/cpl_concepts_listener.go:460) builds an [`ast.ConceptDefinitionData`](src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go:46) containing the concept name and an [`ast.InsertionPoint`](src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go:60) populated with the `before`/`after` anchor value. This data is stored in the concept scope's `table.Data` field.

### 2. Analyzer extracts InsertionPoint (Phase C addition)

[`EnterAnalyzeConceptScope`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:458) now reads `table.Data` after creating the [`AnalysisConceptDefinition`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:263). If the data contains a `*ast.ConceptDefinitionData` with a non-zero `InsertionPoint`, it sets the `InsertionPoint` field on the analysis concept.

### 3. Generators populate values

During the symbol table walk, generator scope handlers ([`AnalyzeValueGeneratorScope`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:515), [`AnalyzeIntGeneratorScope`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:548)) call [`AddGenerator`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:413) which generates values and populates `ResolvedValues` on the concept. This happens in the analysis context stack before the concept scope exits.

### 4. AddOrMergeConcept decides: new, duplicate, or extension

[`ExitAnalyzeConceptScope`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:496) now calls [`AddOrMergeConcept`](src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:298) instead of `AddConcept`. The decision tree:

| Existing concept? | InsertionPoint non-zero? | Action                                                                                                                  |
| ----------------- | ------------------------ | ----------------------------------------------------------------------------------------------------------------------- |
| No                | No                       | Add as new concept (standard `AddConcept` behavior)                                                                     |
| No                | Yes                      | Error: extension with no base concept                                                                                   |
| Yes               | No                       | Error: duplicate definition                                                                                             |
| Yes               | Yes                      | **Merge**: call [`mergeConceptValues()`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:547) |

### 5. mergeConceptValues performs the merge

The [`mergeConceptValues`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:576) function:

1. **Validates** the `InsertionPoint` (exactly one of `BeforeValue`/`AfterValue` must be set)
2. **Extracts** new value names from the extension's already-resolved `ResolvedValues`
3. **Extracts** base value names from the existing concept's `ResolvedValues`
4. **Builds** an [`InsertionRequest`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:538) from the `InsertionPoint` and new values
5. **Calls** [`ComputeValueOrdering`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:578) which handles before/after insertion with validation
6. **Updates** the existing concept's `ResolvedValues`, rebuilds the `index` map
7. **Appends** the extension's `Generators` to the existing concept's `Generators` list

## Structural Changes

### New field on `AnalysisConceptDefinition`

```go
// InsertionPoint specifies where new values should be inserted relative to
// existing values for concept extensions. A zero-value InsertionPoint (both
// BeforeValue and AfterValue empty) indicates this is an initial concept
// definition, not an extension.
InsertionPoint ast.InsertionPoint
```

This field is at [line 272](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:272) of `concept_analysis.go`.

### New functions

| Function                                                                                                | Location                | Purpose                                       |
| ------------------------------------------------------------------------------------------------------- | ----------------------- | --------------------------------------------- |
| [`mergeConceptValues()`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:547) | concept_analysis.go:547 | Merges extension values into existing concept |
| [`AddOrMergeConcept()`](src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:298)      | cpl_analyzer.go:298     | Decision logic: add new, merge, or error      |

## Test Coverage

11 new test functions in [`concept_extension_merge_test.go`](src/chronocone_planning_language/src/main/go/analysis/concept_extension_merge_test.go):

| Test                                                   | Coverage                                                      |
| ------------------------------------------------------ | ------------------------------------------------------------- |
| `TestMergeConceptValues_BeforeInsertion`               | 4 subcases: before first, middle, last value; multiple values |
| `TestMergeConceptValues_AfterInsertion`                | 4 subcases: after first, middle, last value; multiple values  |
| `TestAddOrMergeConcept_AddNewConcept`                  | New concept with zero InsertionPoint → added normally         |
| `TestAddOrMergeConcept_MultipleExtensions`             | Two consecutive extensions (before + after) on same concept   |
| `TestAddOrMergeConcept_DuplicateWithoutInsertionPoint` | Same name, no insertion → duplicate error                     |
| `TestAddOrMergeConcept_ExtensionWithNoExistingConcept` | Extension for non-existent concept → error                    |
| `TestAddOrMergeConcept_InvalidInsertionPoint`          | Both before+after or neither set → error                      |
| `TestAddOrMergeConcept_BeforeValueNotFound`            | Before/after anchor value not in concept → error              |
| `TestAddOrMergeConcept_OriginalOrderPreserved`         | Original values maintain relative order after insertions      |
| `TestAddOrMergeConcept_GeneratorsMerged`               | Extension generators appended to existing list                |
| `TestMergeConceptValues_EmptyExtensionValues`          | Valid anchor, empty values → no-op                            |

## Example Walkthrough

**Input CPL:**

```cpl
concepts
    Priority=[Values[Low, Medium, High]]

concepts
    Priority=[Values[Critical]] before Low
    Priority=[Values[Urgent]] after Medium
```

**Processing:**

1. **First section**: `EnterAnalyzeConceptScope` creates Priority with `InsertionPoint{}` (zero value). `AnalyzeValueGeneratorScope` adds Values[Low, Medium, High]. `ExitAnalyzeConceptScope` calls `AddOrMergeConcept` → no existing concept, zero InsertionPoint → added as new. `concepts["Priority"].ResolvedValues = [Low, Medium, High]`

2. **Second section, first line**: `EnterAnalyzeConceptScope` creates Priority with `InsertionPoint{BeforeValue:"Low"}` (extracted from `table.Data`). `AnalyzeValueGeneratorScope` adds Values[Critical]. `ExitAnalyzeConceptScope` calls `AddOrMergeConcept` → existing concept found, non-zero InsertionPoint → calls `mergeConceptValues`.

3. **mergeConceptValues**: baseValues = ["Low", "Medium", "High"], insertion = {BeforeValue:"Low", Values:["Critical"]}. `ComputeValueOrdering` computes [Critical, Low, Medium, High]. Existing concept updated with new ResolvedValues, rebuilt index, appended generator.

4. **Second section, second line**: Same process with `InsertionPoint{AfterValue:"Medium"}` and Values[Urgent]. `mergeConceptValues` computes against current state [Critical, Low, Medium, High] → [Critical, Low, Medium, Urgent, High].

**Final state:**

```
Priority.ResolvedValues = [
    {Name: "Critical", Position: 0},
    {Name: "Low",      Position: 1},
    {Name: "Medium",   Position: 2},
    {Name: "Urgent",   Position: 3},
    {Name: "High",     Position: 4},
]
```

## Post-Mortem

### What went well

- The BTDD approach (Phase A → B → C) caught the structural design early. The `InsertionPoint` field on `AnalysisConceptDefinition` was clearly the right place for extension metadata.
- Reusing `ComputeValueOrdering` (already tested and working from Task 1) kept the merge logic simple and correct.
- The decision tree in `AddOrMergeConcept` handles all three cases (new, duplicate, extension) exhaustively.

### What could be improved

- **ConceptRangeGenerator interaction**: Extensions with `ConceptRangeGenerator` values are not yet handled. These are deferred generators that `ResolveConceptRanges()` processes after all concepts are collected. Future work may need to re-run `mergeConceptValues` after range resolution, or defer extension merging until after `ResolveConceptRanges()` completes.
- **Extension ordering within a section**: Currently extensions are processed in the order they appear in the symbol table walk. If multiple extensions in the same section target the same anchor value, the last one wins. This is correct per spec but could benefit from explicit documentation.
- **`mergeConceptValues` is unexported**: For direct unit testing, the function is tested indirectly through `AddOrMergeConcept`. Exporting it as `MergeConceptValues` would allow more isolated testing.

### Files changed

| File                                                                                                                       | Change                                                                                                                 |
| -------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| [`concept_analysis.go`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go)                         | Added `InsertionPoint` field, `mergeConceptValues()`, updated `EnterAnalyzeConceptScope` and `ExitAnalyzeConceptScope` |
| [`cpl_analyzer.go`](src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go)                                 | Added `AddOrMergeConcept()` method                                                                                     |
| [`concept_extension_merge_test.go`](src/chronocone_planning_language/src/main/go/analysis/concept_extension_merge_test.go) | New file: 11 test functions                                                                                            |
