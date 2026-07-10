# Story166: Chronocone Time Concepts Documentation Consistency

## Goal

Update active Chronocone requirements, design, ERO, epic, index, and active story documentation so the time model is described consistently across the project. The story implements the approved consistency plan in [`chronocone_time_concepts_plan.md`](../../plans/chronocone_time_concepts_plan.md) and uses the target model in which [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) is constraint-based, [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48) is the canonical starting instant parameter, [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) is not a type, and metric or connector grain references a measure name whose type is [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22) or derived from [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22).

## References

- [`chronocone_time_concepts_plan.md`](../../plans/chronocone_time_concepts_plan.md) - Source plan being implemented by this story.
- [`talking-about-time.md`](../specifications/requirements/cpl/talking-about-time.md) - Source-of-truth requirements document for time concepts.
- [`concepts-types-units-measures-syntax.md`](../specifications/requirements/cpl/concepts-types-units-measures-syntax.md) - CPL requirements for concepts, types, units, and measures.
- [`connectors.md`](../specifications/requirements/cpl/connectors.md) - CPL connector requirements.
- [`io_specs.md`](../specifications/requirements/cpl/io_specs.md) - CPL input/output specification requirements.
- [`introduction.md`](../specifications/requirements/cpl/introduction.md) - CPL requirements introduction.
- [`types-definitions.md`](../specifications/design/cpl/types-definitions.md) - CPL type design definitions.
- [`units-and-measures.md`](../specifications/design/cpl/units-and-measures.md) - CPL unit and measure design definitions.
- [`concepts-definitions.md`](../specifications/design/cpl/concepts-definitions.md) - CPL concept design definitions.
- [`introduction.md`](../specifications/architecture/cpl/introduction.md) - CPL architecture introduction.
- [`Epic002_cpl-types-implementation.md`](../epics/Epic002_cpl-types-implementation.md) - Type implementation epic.
- [`Epic004_cpl-connectors-implementation.md`](../epics/Epic004_cpl-connectors-implementation.md) - Connector implementation epic.
- [`toc.md`](../toc.md) - Active story table of contents.

## Dependencies

None - This story is a documentation consistency story derived from the approved plan in [`chronocone_time_concepts_plan.md`](../../plans/chronocone_time_concepts_plan.md). It does not depend on completion of the active implementation stories that it updates.

## Dependent Stories

The following active stories should use the corrected documentation when they are implemented:

- [`Story137_cpl-types-parser-instant.md`](../Story137_cpl-types-parser-instant.md) - Not Started
- [`Story138_cpl-types-analyzer-instant.md`](../Story138_cpl-types-analyzer-instant.md) - Not Started
- [`Story139_cpl-types-parser-interval.md`](../Story139_cpl-types-parser-interval.md) - Not Started
- [`Story140_cpl-types-analyzer-interval.md`](../Story140_cpl-types-analyzer-interval.md) - Not Started
- [`Story158_cpl-parser-connectors-time.md`](../Story158_cpl-parser-connectors-time.md) - Not Started
- [`Story159_cpl-analyzer-connectors-time.md`](../Story159_cpl-analyzer-connectors-time.md) - Not Started
- [`Story160_cpl-synthesizer-connectors-time.md`](../Story160_cpl-synthesizer-connectors-time.md) - Not Started

## Tasks

1. Re-run targeted searches across active requirements, design, ERO, epic, index, and story documents - Completed
   a. Search active documentation for old time terminology, including precision-based [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10), [`PrecisionLevel`](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md:472), old [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22) `start` and `duration` parameters, and wording that treats [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) as a type - Completed
   b. Exclude files under [`finished-stories`](../finished-stories/) from the update scope unless an active index row repeats the stale wording - Completed
   c. Record the active files that require edits before beginning document updates - Completed

2. Update [`talking-about-time.md`](../specifications/requirements/cpl/talking-about-time.md) as the source-of-truth requirements document - Completed
   a. Replace precision-based [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) wording with explicit constraint matching semantics - Completed
   b. Add or revise underspecified-instants wording, including examples for fully specified moments, recurring annual moments, recurring daily midnight, and day-of-week constraints - Completed
   c. State that [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) has maximum one-second resolution and always uses UTC - Completed
   d. State that [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) is not a type and only supplies concept values and units used to measure intervals - Completed
   e. Add valid examples and anti-examples using [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48), normal units, and grain measure-name semantics - Completed
   f. Verify the document states that metric, connector, and IO-spec grain references are measure names whose resolved type is [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22) or derived from [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22) - Completed

3. Update active requirements documents that depend on [`talking-about-time.md`](../specifications/requirements/cpl/talking-about-time.md) - Completed
   a. Update [`concepts-types-units-measures-syntax.md`](../specifications/requirements/cpl/concepts-types-units-measures-syntax.md) so time examples use [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) as concept values for units and use [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22) for interval measures - Completed
   b. Update [`connectors.md`](../specifications/requirements/cpl/connectors.md) so connector grain is described as a measure name and old `Interval(precision=Week)` examples are replaced with [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48) and unit semantics - Completed
   c. Update [`io_specs.md`](../specifications/requirements/cpl/io_specs.md) so IO specifications bind runs to already-declared grain measures and unit choices without making [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) a type - Completed
   d. Update [`introduction.md`](../specifications/requirements/cpl/introduction.md) to remove stale precision, restriction, and special [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30)-type wording - Completed
   e. Verify all updated requirements documents remain consistent with the source-of-truth wording - Completed

4. Update active design and architecture documents - Completed
   a. Update [`types-definitions.md`](../specifications/design/cpl/types-definitions.md) so type design describes field-based [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) constraints and [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48) unit constraints - Completed
   b. Update [`units-and-measures.md`](../specifications/design/cpl/units-and-measures.md) so examples use [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48) and keep measure-level conversion discussion intact - Completed
   c. Review [`concepts-definitions.md`](../specifications/design/cpl/concepts-definitions.md) and remove any wording that implies [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) defines [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) structure or acts as a type - Completed
   d. Review [`introduction.md`](../specifications/architecture/cpl/introduction.md) and update any directly inconsistent ground-type or time-model wording - Completed
   e. Verify design and architecture wording matches the corrected requirements model - Completed

5. Update active ERO documents that guide parser, analyzer, unit, measure, and connector work - Completed
   a. Update [`ERO002_types-incremental-implementation.md`](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) to remove stale [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) precision and old [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22) `start` or `duration` examples - Completed
   b. Update [`ERO002_types-parsing-phase.md`](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) to describe parser capture for supplied [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) field constraints, [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48), and unit references - Completed
   c. Update [`ERO002_types-analysis-phase.md`](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) to replace [`PrecisionLevel`](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md:472), precision validation, and restriction validation with field constraint validation - Completed
   d. Update [`ERO003_units-measures-incremental-implementation.md`](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) and [`ERO003_units-measures-parsing-phase.md`](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md) examples to use [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48) - Completed
   e. Review [`ERO004_connectors-incremental-implementation.md`](../specifications/design/cpl/eros/ERO004_connectors-incremental-implementation.md), [`ERO004_connectors-analysis-phase.md`](../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md), and [`ERO002_types-synthesis-phase.md`](../specifications/design/cpl/go-synthesis-target/eros/ERO002_types-synthesis-phase.md) for stale time or grain semantics and update only where needed - Completed
   f. Verify ERO guidance clearly treats [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) through normal concept, unit, measure, and conversion rules - Completed

6. Update epics and story index descriptions - Completed
   a. Update [`Epic002_cpl-types-implementation.md`](../epics/Epic002_cpl-types-implementation.md) so the summary and Story137 through Story140 rows describe field-based [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) constraints and [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48) unit semantics - Completed
   b. Update [`Epic004_cpl-connectors-implementation.md`](../epics/Epic004_cpl-connectors-implementation.md) so connector grain is a measure name and unit selection comes from connector metadata and IO specifications - Completed
   c. Update [`toc.md`](../toc.md) active story descriptions for Story137 through Story140 and connector-time stories where needed - Completed
   d. Verify all planning summaries match the corrected documentation model - Completed

7. Update active stories for [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) and [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22) parsing and analysis - Completed
   a. Update [`Story137_cpl-types-parser-instant.md`](../Story137_cpl-types-parser-instant.md) to replace [`PrecisionLevel`](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md:472) and restriction tasks with parser tasks for year, month, day, hour, minute, second, day-of-week, and week-number constraints - Completed
   b. Update [`Story138_cpl-types-analyzer-instant.md`](../Story138_cpl-types-analyzer-instant.md) to replace precision and restriction validation with range validation, compatible field-combination validation, and underspecification acceptance criteria - Completed
   c. Update [`Story139_cpl-types-parser-interval.md`](../Story139_cpl-types-parser-interval.md) to replace `start` and `duration` constraints with [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48), inline or named [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) declarations, unit references, and standard [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) definitions - Completed
   d. Update [`Story140_cpl-types-analyzer-interval.md`](../Story140_cpl-types-analyzer-interval.md) to validate [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48) and unit references through normal definitions without built-in-vs-user-defined [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) special cases - Completed
   e. Verify each updated story integrates its verification criteria into the relevant task rather than adding separate test-only tasks - Completed

8. Update active connector time stories if needed - Completed
   a. Review [`Story158_cpl-parser-connectors-time.md`](../Story158_cpl-parser-connectors-time.md) and update grain wording so the parsed grain value is a measure name, not a type name or unit name - Completed
   b. Review [`Story159_cpl-analyzer-connectors-time.md`](../Story159_cpl-analyzer-connectors-time.md) and update validation context for connector metadata and IO-spec time-dimension selection - Completed
   c. Review [`Story160_cpl-synthesizer-connectors-time.md`](../Story160_cpl-synthesizer-connectors-time.md) and update time-column-to-grain wording to align with connector metadata, IO specifications, interval measures, and [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30)-derived units - Completed
   d. Verify connector-time stories do not conflict with the corrected grain model - Completed

9. Perform a final search-and-consistency pass - Completed
   a. Search active documentation again for stale precision-based [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10), stale [`PrecisionLevel`](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md:472), old [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22) `start` or `duration` parameters, [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30)-as-type wording, and grain wording that does not resolve to measure names - Completed
   b. Correct remaining active-documentation conflicts or explicitly defer them with user approval - Completed
   c. Confirm that final active documentation satisfies every final consistency check in [`chronocone_time_concepts_plan.md`](../../plans/chronocone_time_concepts_plan.md) - Completed

## Test-First Development

This is a documentation consistency story. It does not introduce production code. Verification is still integrated into each documentation task: every update task includes a local consistency check, and the final task requires a full search-and-consistency pass. If implementation discovers that code changes are required, those changes must be captured in a separate story that follows the Logical TDD Lifecycle.

## Constraints

- Do not edit files under [`finished-stories`](../finished-stories/) unless the user explicitly expands the scope.
- Do not execute code implementation, tests, git operations, or end-of-task activities unless the user explicitly requests them.
- Preserve the corrected model from [`chronocone_time_concepts_plan.md`](../../plans/chronocone_time_concepts_plan.md): supplied [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) parameters are constraints; omitted parameters broaden the match set; [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) uses UTC with maximum one-second resolution; [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) is never a type; grain references measure names.
- Preserve the general measure-declaration rule that a measure can be declared from a built-in type with constraints or from an already derived type with additional constraints.
- When updating stories, keep test and verification work integrated into each functional story task rather than creating standalone test-only tasks.

## Intent

1. The first task ensures the implementor begins from a verified list of active documents instead of relying only on the initial plan inventory.
2. The second task establishes the source-of-truth time-concept wording before dependent documents are edited.
3. The third task propagates the corrected source-of-truth requirements into all affected active requirements documents.
4. The fourth task aligns design and architecture guidance with the corrected requirements model.
5. The fifth task prevents stale ERO guidance from directing future parser, analyzer, unit, measure, or connector implementation toward the old model.
6. The sixth task keeps high-level planning summaries and the active story index consistent with the corrected time model.
7. The seventh task updates the active type-parser and type-analyzer stories so future implementation work follows the corrected model.
8. The eighth task ensures connector-time implementation stories align with measure-name grain semantics and IO-spec unit selection.
9. The ninth task verifies that active documentation no longer contains contradictory time-model statements.

## Acceptance Criteria

- Active documentation consistently describes [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) as a UTC moment constraint set that may match one, many, or infinitely many moments.
- Active documentation does not describe [`Instant`](../specifications/requirements/cpl/talking-about-time.md:10) as precision-based and does not require a complete date-time tuple merely because one field is supplied.
- Active documentation uses [`Interval.starts_at`](../specifications/requirements/cpl/talking-about-time.md:48) as the canonical starting instant parameter.
- Active documentation treats [`Duration`](../specifications/requirements/cpl/talking-about-time.md:30) as concept values and units used for measuring intervals, not as a type.
- Active documentation states that metric and connector grain references are measure names whose resolved measure type is [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22) or derived from [`Interval`](../specifications/requirements/cpl/talking-about-time.md:22).
- The active stories [`Story137_cpl-types-parser-instant.md`](../Story137_cpl-types-parser-instant.md), [`Story138_cpl-types-analyzer-instant.md`](../Story138_cpl-types-analyzer-instant.md), [`Story139_cpl-types-parser-interval.md`](../Story139_cpl-types-parser-interval.md), [`Story140_cpl-types-analyzer-interval.md`](../Story140_cpl-types-analyzer-interval.md), [`Story158_cpl-parser-connectors-time.md`](../Story158_cpl-parser-connectors-time.md), [`Story159_cpl-analyzer-connectors-time.md`](../Story159_cpl-analyzer-connectors-time.md), and [`Story160_cpl-synthesizer-connectors-time.md`](../Story160_cpl-synthesizer-connectors-time.md) no longer contradict the corrected time model.
- The finished story table of contents in [`toc.md`](../toc.md) includes this story with status Done.
- The final consistency pass finds no unresolved active-documentation contradictions, or any remaining contradictions are explicitly documented as deferred with user approval.

## Open Questions

None.

## Known Issues

The source plan identifies many active documents with stale time-model wording. This story intentionally updates documentation and active story guidance only; it does not implement parser, analyzer, or synthesizer code changes.

## Additional Context

The story was created using the process defined by [`add-new-story.pdd.script.md`](../../scripts/add-new-story.pdd.script.md). The next story number was selected by scanning active and finished story filenames and assigning the next number after [`Story165_cpl-generated-code-compilation.md`](Story165_cpl-generated-code-compilation.md).

---
