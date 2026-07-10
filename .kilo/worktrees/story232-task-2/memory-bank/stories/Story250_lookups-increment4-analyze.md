# Story250: Lookups Increment 4 — Analyze

## Goal

Implement lookup reference resolution in aggregator expressions and dimensional unit validation for aggregator-lookup combinations (Increment 4 of the lookups analysis phase). When the aggregator analysis function encounters identifiers in expression definitions, it resolves them against the `AnalysisDefinitions.Lookups` map, validates dimensional compatibility between the lookup's key dimensions and the aggregator's dimensional context, and records resolved references in the aggregator's CIL node. No parser changes are required — lookup names appear as ordinary identifiers in aggregator expressions.

This story is part of the lookups incremental implementation plan. See [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) and [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md) for the full plan.

## References

- [requirements](../specifications/requirements/requirements.md)
- [design](../specifications/design/design.md)
- [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md)
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md)
- [Source root: src/chronocone_planning_language/src/main/go/](../../src/chronocone_planning_language/src/main/go/)

## Dependencies

- **Epic007 (Aggregators) at Increment 2 or above** — aggregator analysis infrastructure including `AnalysisDefinitions.Aggregators` map and aggregator expression resolution must exist
- [Story242_lookups-increment1-analyze.md](Story242_lookups-increment1-analyze.md) — Not Started (`AnalysisDefinitions.Lookups` map must be populated)

## Dependent Stories

- None — this story completes the lookups analysis phase. Subsequent increments (Increment 5 and beyond) depend on this story for lookup-aggregator integration.

## Tasks

### Task Sizing

All four tasks below are not annotated `[Small]`:

- **Task 1** modifies the aggregator expression identifier resolution algorithm, changing priority order (lookups before concepts). This requires understanding the existing resolution pipeline — more than a mechanical switch-case addition.
- **Task 2** creates a new validation function with complex multi-structure reasoning across lookup key dimensions, aggregator dimensional contexts, implicit transformations, and celltype unit compatibility.
- **Task 3** modifies CIL node structures and records implicit transformation nodes — requires understanding the CIL node hierarchy and the concept of implicit dimensional transformations.
- **Task 4** creates example CPL files and verifies they produce expected analysis results. The verification step requires understanding the analysis pipeline's error output format and expected behavior for each example.

All tasks route to `code-for-story-implementor`.

### Parallel Execution

None — all tasks have sequential dependencies. Each task builds on the output of the previous task.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

### Reintegration

None — no parallel groups.

### Task 1: Implement lookup reference resolution in aggregator expressions — Not Started

Implement the identifier resolution algorithm that resolves lookup references in aggregator expression definitions. When the aggregator analysis function encounters identifiers in expressions, it checks `AnalysisDefinitions.Lookups` before `Concepts` to resolve ambiguous names (lookup names take precedence). The resolution follows this algorithm: (1) check `AnalysisDefinitions.Lookups[identifier]` — if found, it is a lookup reference; (2) check `AnalysisDefinitions.Metrics[identifier]` — if found, it is a metric reference; (3) check `AnalysisDefinitions.Concepts[identifier]` — if found, it is a concept reference; (4) if not found in any map, report error for undeclared reference; (5) if found in both Lookups and another map, report error for ambiguous reference.

   a. Add lookup name resolution to the aggregator expression identifier resolution logic in `src/main/go/analysis/analyze_aggregators.go` — Not Started
   b. Search `AnalysisDefinitions.Lookups` before `Concepts` and `Metrics` (lookups take precedence) — Not Started
   c. Report error when a name exists in both Lookups and another map (ambiguous reference) — Not Started
   d. Write unit tests: valid lookup reference, non-existent identifier, ambiguous lookup/concept name — Not Started
   e. Verify tests pass: `go test ./src/main/go/analysis/... -run AggregatorLookup -count=1` — Not Started

**Files**: `src/main/go/analysis/analyze_aggregators.go` (modify), aggregator lookup test file (create)

**Acceptance criteria**:
- Identifiers in aggregator expressions resolve against the Lookups map
- Lookup names take precedence over concept and metric names
- Ambiguous references (name in both Lookups and another map) produce clear error messages
- Undeclared references (name in no map) produce clear error messages

### Task 2: Implement dimensional unit validation for lookup-aggregator combinations — Not Started

Create a validation function that checks dimensional and unit compatibility between a lookup referenced in an aggregator expression and the aggregator's dimensional context. The validation covers: (1) dimensional form compatibility — the lookup's key dimensions must be available in at least one of the aggregator's dimensional contexts (input or output); (2) implicit transformation detection — if the lookup's dimensions do not directly match, determine whether an implicit aggregator or spreader transformation is needed to align the lookup's dimensions with the aggregator context; (3) unit compatibility — the lookup's celltype units must compose correctly with other operands in the expression.

   a. Create `validateAggregatorLookupCompatibility` function in `src/main/go/analysis/aggregator_lookup_validation.go` — Not Started
   b. Check that the lookup's key dimensions are available in the aggregator's dimensional context — Not Started
   c. Determine if the lookup needs implicit dimensional transformation (aggregator or spreader applied to the lookup) — Not Started
   d. Verify lookup celltype units are compatible with the aggregation expression using Celltype compound unit composition — Not Started
   e. Write unit tests in `src/main/go/analysis/aggregator_lookup_validation_test.go`: compatible dimensions, incompatible dimensions, implicit transformation needed, unit incompatibility — Not Started
   f. Verify tests pass: `go test ./src/main/go/analysis/... -run AggregatorLookupValidation -count=1` — Not Started

**Files**: `src/main/go/analysis/aggregator_lookup_validation.go` (create), `src/main/go/analysis/aggregator_lookup_validation_test.go` (create)

**Acceptance criteria**:
- Lookup key dimensions are validated against aggregator dimensional context
- Incompatible dimensions produce clear error messages
- Implicit dimensional transformations on lookups are detected
- Lookup celltype units are validated against aggregation expression units
- Unit incompatibilities produce clear error messages

### Task 3: Record resolved lookup references in CIL aggregator nodes — Not Started

Store resolved lookup references in the aggregator's CIL node structure (as defined by Epic007 Increment 2's `CILAggregateNode` or equivalent). For each lookup reference resolved in the aggregator expression, record the reference metadata in the CIL node. When the validation function from Task 2 determines that an implicit dimensional transformation is needed, record the corresponding CIL transformation node (implicit aggregator or spreader) so the synthesizer can generate the correct runtime code.

   a. Store resolved lookup references in the aggregator's CIL node structure per Epic007's CIL node design — Not Started
   b. Record any implicit transformation nodes (CIL aggregator or spreader nodes) needed for dimensional alignment — Not Started
   c. Write unit tests: verify CIL node structure includes lookup reference metadata and implicit transformation nodes — Not Started
   d. Verify tests pass: `go test ./src/main/go/analysis/... -run AggregatorLookupCIL -count=1` — Not Started

**Files**: aggregator CIL node file (modify, per Epic007's structure), CIL lookup test file (create)

**Acceptance criteria**:
- Resolved lookup references are recorded in the aggregator's CIL node
- Implicit transformation nodes are recorded when dimensional alignment is needed
- CIL node structure includes all metadata required by the synthesizer for lookup resolution

### Task 4: Create parse_and_analyze_only examples — Not Started

Create example CPL files demonstrating aggregator expressions that reference lookups. Provide both valid examples (correct dimensional compatibility) and error examples (incompatible dimensions, ambiguous lookup/concept names, unit mismatches). Verify that all examples produce the expected analysis results when processed by the CPL analyzer.

   a. In `src/main/go/examples/parse_and_analyze_only/without_errors/`: add CPL files with aggregators referencing lookups in their expressions — Not Started
   b. In `src/main/go/examples/parse_and_analyze_only/with_errors/`: add CPL files with incompatible dimensions, ambiguous lookup/concept names, and unit mismatches — Not Started
   c. Verify all examples produce expected analysis results: `go test ./src/main/go/analysis/... -run Examples -count=1` — Not Started

**Files**: multiple CPL example files in `src/main/go/examples/parse_and_analyze_only/without_errors/` and `src/main/go/examples/parse_and_analyze_only/with_errors/` (create)

**Acceptance criteria**:
- Valid examples parse and analyze without errors
- Error examples produce the expected error messages for each error condition
- All examples are documented with comments explaining the expected behavior

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- No parser changes are required for this increment — lookup names are ordinary identifiers in aggregator expressions
- The aggregator analysis infrastructure (`AnalysisDefinitions.Aggregators` map, aggregator expression resolution) must exist from Epic007 Increment 2 before this story can execute
- The `AnalysisDefinitions.Lookups` map must be populated from Story242 before this story can execute
- Resolution order: Lookups > Metrics > Concepts (lookups take precedence)
- Dimensional unit validation must produce clear, actionable error messages

## Intent

### Task 1 Intent

Enable the aggregator analysis function to recognize lookup references in expression definitions. The identification of a lookup reference must happen during the analysis phase because the parser cannot distinguish lookup names from concept names — both are identifiers. The analysis phase resolves this ambiguity by consulting the Lookups map and enforcing a precedence order.

### Task 2 Intent

Ensure that when a lookup is used in an aggregator expression, the lookup's dimensional structure (key dimensions and celltype units) is valid for the aggregator's dimensional context. This prevents runtime errors from dimensional mismatches and ensures the synthesizer has complete information about the data transformation pipeline.

### Task 3 Intent

Provide the synthesizer with the metadata it needs to generate correct runtime code. The CIL node for each aggregator must record which lookups it references and what implicit transformations (aggregators or spreaders) are needed to align dimensional forms.

### Task 4 Intent

Demonstrate the feature end-to-end with concrete CPL examples. The examples serve as both documentation and automated regression tests, ensuring that future changes to the analysis phase do not break lookup resolution or dimensional validation.

## Acceptance Criteria

1. Aggregator expression identifiers resolve against `AnalysisDefinitions.Lookups` map
2. Lookup names take precedence over concept names when both exist (ambiguous reference is an error)
3. Lookup key dimensions are validated against aggregator dimensional context
4. Implicit dimensional transformations on lookups are detected and recorded
5. Lookup celltype units are validated against aggregation expression units
6. Resolved lookup references are recorded in the aggregator's CIL node
7. Clear error messages for incompatible dimensions and unit mismatches
8. Example CPL files demonstrate both valid and error cases
