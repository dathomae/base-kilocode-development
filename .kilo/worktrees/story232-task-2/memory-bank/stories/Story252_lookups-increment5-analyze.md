# Story252: Lookups Increment 5 — Analyze

## Goal

Implement lookup reference resolution in spreader expressions, dimensional unit analysis for spreader-lookup combinations, and cell unit validation for lookup output units (Increment 5 of the lookups analysis phase). When the spreader analysis function encounters identifiers in expression definitions, it resolves them against the `AnalysisDefinitions.Lookups` map, determines the required dimensional form for the lookup in the spreader context, applies implicit aggregators/spreaders when dimensional alignment is needed, and validates celltype unit compatibility. No parser changes are required — lookup names appear as ordinary identifiers in spreader expressions.

This story is part of the lookups incremental implementation plan. See [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) and [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md) for the full plan.

## References

- [requirements](../specifications/requirements/requirements.md)
- [design](../specifications/design/design.md)
- [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md)
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md)
- [Source root: src/chronocone_planning_language](../../src/chronocone_planning_language/)

## Dependencies

- **Epic008 (Spreaders) at Increment 3 or above** — spreader analysis infrastructure including `AnalysisDefinitions.Spreaders` map, spreader expression resolution, and implicit aggregator resolution must exist
- [Story242_lookups-increment1-analyze.md](Story242_lookups-increment1-analyze.md) — Not Started (`AnalysisDefinitions.Lookups` map must be populated)

## Dependent Stories

- None — this story advances the lookups analysis phase. Subsequent lookups increments depend on this story for spreader-lookup integration.

## Tasks

### Task Sizing

All five tasks below are not annotated `[Small]` and route to `code-for-story-implementor`:

- **Task 1** modifies the spreader expression identifier resolution algorithm, changing priority order (Lookups added before Concepts). This requires understanding the existing multi-map resolution pipeline and precedence rules — more than a mechanical switch-case addition.
- **Task 2** creates a new validation function with multi-structure reasoning across lookup key dimensions, spreader dimensional contexts, and implicit transformation detection. A new file with new abstractions.
- **Task 3** extends the validation from Task 2 with compound unit composition rules. Requires understanding celltype unit algebra and unit compatibility semantics.
- **Task 4** modifies CIL node structures and records implicit transformation nodes. Requires understanding the CIL node hierarchy and the concept of implicit dimensional transformations.
- **Task 5** creates example CPL files across multiple subdirectories. Requires understanding the analysis pipeline's error output format and expected behavior for each example.

### Parallel Execution

None — all tasks have sequential dependencies. Each task builds on the output of the previous task.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

### Reintegration

None — no parallel groups.

### Task 1: Implement lookup reference resolution in spreader expressions — Not Started

Implement the identifier resolution algorithm that resolves lookup references in spreader expression definitions. When the spreader analysis function encounters identifiers in expressions, it checks `AnalysisDefinitions.Lookups` before `Concepts` to resolve ambiguous names (lookup names take precedence). The resolution follows this algorithm: (1) check `AnalysisDefinitions.Lookups[identifier]` — if found, it is a lookup reference; (2) check `AnalysisDefinitions.Metrics[identifier]` — if found, it is a metric reference; (3) check `AnalysisDefinitions.Concepts[identifier]` — if found, it is a concept reference; (4) if not found in any map, report error for undeclared reference; (5) if found in both Lookups and another map, report error for ambiguous reference.

   a. Write unit tests for lookup reference resolution: valid lookup reference in spreader expression, ambiguous lookup/metric name, ambiguous lookup/concept name, undeclared identifier — Not Started
   b. Verify tests fail (confirm test scaffolding correctly targets the new resolution behavior) — Not Started
   c. Add lookup name resolution to the spreader expression identifier resolution logic in `src/main/go/analysis/analyze_spreaders.go`: check `AnalysisDefinitions.Lookups` before `Concepts` — Not Started
   d. Report error when a name exists in both Lookups and another map (ambiguous reference) — Not Started
   e. Verify all tests pass: `go test ./src/main/go/analysis/... -run SpreaderLookup -count=1` — Not Started

**Files**: `src/main/go/analysis/analyze_spreaders.go` (modify), spreader lookup resolution test file (create)

**Acceptance criteria**:
- Identifiers in spreader expressions resolve against the Lookups map
- Lookup names take precedence over concept and metric names in the resolution order
- Ambiguous references (name in both Lookups and another map) produce clear error messages
- Undeclared references (name in no map) produce clear error messages

### Task 2: Implement dimensional form analysis for lookup references — Not Started

Create a validation function that determines the required dimensional form for a lookup referenced in a spreader expression and checks whether the lookup's current dimensions match. When the lookup's key dimensions do not directly match the spreader's dimensional context, determine what implicit transformations (aggregators or spreaders) can align the lookup's dimensions. This is the same mechanism used for metric references in spreaders from Epic008 Increment 3.

   a. Write unit tests for dimensional validation: lookup dimensions match spreader context, lookup dimensions need aggregation (implicit aggregator), lookup dimensions need spreading (implicit spreader), incompatible dimensions — Not Started
   b. Verify tests fail (confirm test scaffolding correctly targets the new validation logic) — Not Started
   c. Create `src/main/go/analysis/spreader_lookup_validation.go` with a `validateSpreaderLookupDimensions` function — Not Started
   d. Implement the required-dimensional-form derivation from the spreader's dimensional context — Not Started
   e. Implement implicit transformation detection: determine if an implicit aggregator or spreader can align the lookup's dimensions to the required form — Not Started
   f. Verify all tests pass: `go test ./src/main/go/analysis/... -run SpreaderLookupDimensions -count=1` — Not Started

**Files**: `src/main/go/analysis/spreader_lookup_validation.go` (create), `src/main/go/analysis/spreader_lookup_validation_test.go` (create)

**Acceptance criteria**:
- Lookup key dimensions are validated against the spreader's required dimensional form
- Mismatched dimensions trigger implicit transformation detection (aggregator or spreader)
- Incompatible dimensions (no valid transformation possible) produce clear error messages
- Dimensional form derivation reuses or parallels the existing mechanism for metric-in-spreader analysis from Epic008

### Task 3: Implement cell unit validation for lookup-spreader combinations — Not Started

Extend the spreader-lookup validation to check celltype unit compatibility. The lookup's celltype units must be dimensionally compatible with the spreader expression's dimensional algebra. Apply compound unit composition rules for the specific spreader expression context. For example, if the spreader distributes `Amount * WeightFactor`, the lookup's celltype units must be compatible with the multiplication context.

   a. Write unit tests for cell unit validation: compatible celltype units, incompatible celltype units, compound unit edge cases (unit composition involving the spreader expression's operands) — Not Started
   b. Verify tests fail (confirm test scaffolding correctly targets the new unit validation logic) — Not Started
   c. Extend `src/main/go/analysis/spreader_lookup_validation.go` with a celltype unit compatibility check using compound unit composition rules — Not Started
   d. Verify all tests pass: `go test ./src/main/go/analysis/... -run SpreaderLookupUnit -count=1` — Not Started

**Files**: `src/main/go/analysis/spreader_lookup_validation.go` (modify — extends Task 2's file), `src/main/go/analysis/spreader_lookup_validation_test.go` (modify — adds unit tests)

**Acceptance criteria**:
- Lookup celltype units are validated against the spreader expression's dimensional algebra
- Compatible unit compositions (e.g., matching units, valid compound unit multiplication) pass validation
- Incompatible unit compositions produce clear error messages
- Compound unit edge cases (numerator/denominator alignment, power expressions) are handled correctly

### Task 4: Record resolved references in CIL spreader nodes — Not Started

Store resolved lookup references in the spreader's CIL node structure (as defined by Epic008 Increment 3's `CILSpreadNode` or equivalent). For each lookup reference resolved in the spreader expression, record the reference metadata in the CIL node. When the validation functions from Tasks 2 and 3 determine that an implicit dimensional transformation is needed, record the corresponding CIL transformation node (implicit aggregator or spreader) so the synthesizer can generate the correct runtime code.

   a. Write unit tests: verify CIL node includes lookup reference metadata (lookup name, key dimensions), verify implicit transformation nodes are recorded when dimensional alignment is needed — Not Started
   b. Verify tests fail (confirm test scaffolding correctly targets the new CIL node fields) — Not Started
   c. Store resolved lookup references in the spreader's CIL node (per Epic008's CIL spreader node structure) — Not Started
   d. Record implicit transformation nodes (CIL aggregator or spreader nodes) when dimensional alignment requires them — Not Started
   e. Verify all tests pass: `go test ./src/main/go/analysis/... -run SpreaderLookupCIL -count=1` — Not Started

**Files**: spreader CIL node file (modify, per Epic008's CIL node structure under `src/main/go/cil/`), CIL lookup recording test file (create)

**Acceptance criteria**:
- Resolved lookup references are recorded in the spreader's CIL node with complete metadata (lookup name, key dimensions, celltype)
- Implicit transformation nodes are recorded when dimensional alignment is needed
- CIL node structure includes all metadata required by the synthesizer for lookup resolution at runtime

### Task 5: Create parse_and_analyze_only examples — Not Started

Create example CPL files demonstrating spreader expressions that reference lookups. Provide both valid examples (correct dimensional compatibility and unit composition) and error examples (incompatible dimensions, implicit transformation failures, unit mismatches). Verify that all examples produce the expected analysis results when processed by the CPL analyzer.

   a. In `src/main/go/examples/parse_and_analyze_only/without_errors/`: add CPL files with spreaders referencing lookups in their expressions — Not Started
   b. In `src/main/go/examples/parse_and_analyze_only/with_errors/`: add CPL files with incompatible dimensions, implicit transformation failures, and unit mismatches — Not Started
   c. Verify all examples produce expected analysis results: `go test ./src/main/go/analysis/... -run Examples -count=1` — Not Started

**Files**: multiple CPL example files in `src/main/go/examples/parse_and_analyze_only/without_errors/` and `src/main/go/examples/parse_and_analyze_only/with_errors/` (create)

**Acceptance criteria**:
- Valid examples parse and analyze without errors
- Error examples produce the expected error messages for each error condition (dimensional mismatch, implicit transformation failure, unit incompatibility)
- All examples are documented with comments explaining the expected behavior

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- No parser changes are required for this increment — lookup names are ordinary identifiers in spreader expressions
- The spreader analysis infrastructure (`AnalysisDefinitions.Spreaders` map, spreader expression resolution, implicit aggregator resolution) must exist from Epic008 Increment 3 before this story can execute
- The `AnalysisDefinitions.Lookups` map must be populated from Story242 before this story can execute
- Resolution order: Lookups → Metrics → Concepts (lookups take precedence per Increment 4's established rule)
- Dimensional unit validation must reuse or parallel the existing mechanism for metric-in-spreader analysis
- The exact CIL spreader node structure depends on Epic008's design; Task 4 should adapt to whatever CIL node type Epic008 defines for spreaders
- All error messages must be clear and actionable, identifying the specific dimension, unit, or transformation that failed

## Intent

### Task 1 Intent

Enable the spreader analysis function to recognize lookup references in expression definitions. The identification of a lookup reference must happen during the analysis phase because the parser cannot distinguish lookup names from concept or metric names — all are identifiers. The analysis phase resolves this ambiguity by consulting the Lookups map and enforcing a precedence order.

### Task 2 Intent

Ensure that when a lookup is used in a spreader expression, the lookup's dimensional structure (key dimensions) is valid for the spreader's dimensional transformation context. This mirrors the dimensional analysis already performed for metric references in spreaders. When the lookup's dimensions do not directly match, detect what implicit transformations (aggregators or spreaders) are needed to align them.

### Task 3 Intent

Guard against unit-level incompatibilities between a lookup's celltype and the spreader expression's dimensional algebra. The spreader may multiply or divide values from the lookup with other operands — the lookup's units must compose correctly in that context. This prevents runtime errors from unit mismatches in the generated code.

### Task 4 Intent

Provide the synthesizer with the metadata it needs to generate correct runtime code. The CIL node for each spreader must record which lookups it references, the lookup's dimensional and unit characteristics, and what implicit transformations (aggregators or spreaders) are needed to align dimensional forms before the lookup can participate in the spreader expression.

### Task 5 Intent

Demonstrate the feature end-to-end with concrete CPL examples. The examples serve as both documentation and automated regression tests, ensuring that future changes to the analysis phase do not break lookup resolution, dimensional validation, unit validation, or CIL recording behavior.

## Acceptance Criteria

1. Spreader expression identifiers resolve against `AnalysisDefinitions.Lookups` map
2. Ambiguous lookup/metric and lookup/concept references produce clear errors
3. Dimensional form analysis determines required lookup dimensions in spreader context
4. Implicit aggregators/spreaders are generated for lookups when dimensional alignment is needed
5. Cell unit validation ensures lookup celltype units are compatible with spreader expression
6. Resolved references (including implicit transformations) are recorded in CIL spreader nodes
7. Clear error messages for incompatible dimensions, unit mismatches, and unresolvable transformations

## Notes

- The implementation relies on Epic008's spreader analysis infrastructure existing at Increment 3 or above. If the spreader analysis function structure changes during Epic008 implementation, the exact insertion points for lookup resolution may need adjustment.
- Tasks 2 and 3 both modify `spreader_lookup_validation.go` — Task 2 creates the file with dimensional validation; Task 3 extends it with cell unit validation. The sequential execution ensures both layers of validation are in place before Task 4 records the results in CIL nodes.
- The CIL node paths in Task 4 are indicative. The technical writer should align with the actual CIL spreader node structure and file naming as defined by Epic008 Increment 3.
