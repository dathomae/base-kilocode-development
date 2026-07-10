# Story130: Epic001 Story Updates Implementation

## Goal

Implement the Epic001 Update Plan by updating existing stories (Story114-Story126) to use the new generator-based syntax for concept declarations and creating new stories (Story128-Story129) for Increment 6 (Concept Range Generators). This story captures the work required to align all Epic001-related stories with the updated ERO001 generator syntax specification.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [Epic001 Update Plan](../epics/Epic001_update_plan.md)
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)
- [ERO001: Concepts Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md)
- [ERO001: Concepts Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO001_concepts-synthesis-phase.md)

## Dependencies

None - This is a documentation update task that can be performed independently of code implementation.

## Dependent Stories

- Story114_cpl-concepts-parser-basic.md - Will be updated
- Story115_cpl-concepts-analyzer-basic.md - Will be updated
- Story116_cpl-concepts-synthesizer-basic.md - Will be updated
- Story117_cpl-concepts-parser-hierarchy.md - Will be updated
- Story118_cpl-concepts-analyzer-hierarchy.md - Will be updated
- Story119_cpl-concepts-parser-multi-section.md - Will be updated
- Story120_cpl-concepts-analyzer-multi-section.md - Will be updated
- Story121_cpl-concepts-parser-conditional.md - Will be updated
- Story122_cpl-concepts-analyzer-conditional.md - Will be updated
- Story123_cpl-concepts-parser-range.md - Will be updated
- Story124_cpl-concepts-analyzer-range.md - Will be updated
- Story125_cpl-concepts-parser-before-after.md - Will be updated
- Story126_cpl-concepts-analyzer-before-after.md - Will be updated

## Tasks

### Phase 1: Update Existing Stories 114-116 (Increment 1 - Basic Concepts)

1. Update Story114: CPL Concepts Parser - Basic - Completed
   a. Update grammar rules to use new generator pattern syntax (`ConceptName=[Values[...]]`) - Completed
   b. Update token definitions: replace `LIST` with `VALUES`, add `LBRACKET`/`RBRACKET` - Completed
   c. Update ParserConceptDefinition struct to use `Generators []Generator` - Completed
   d. Update GeneratorType enum: rename `ListGenerator` to `ValuesGenerator` - Completed
   e. Update example CPL code in story - Completed
   f. Update acceptance criteria - Completed

2. Update Story115: CPL Concepts Analyzer - Basic - Completed
   a. Update data structures to use new Generator interface - Completed
   b. Update AnalysisConceptDefinition to use `Generators []Generator` - Completed
   c. Add ValuesGenerator struct definition - Completed
   d. Update example CPL code in story - Completed
   e. Update acceptance criteria - Completed

3. Update Story116: CPL Concepts Synthesizer - Basic - Completed
   a. Update example CPL code to use new syntax - Completed
   b. Verify no other changes needed for synthesizer - Completed

### Phase 2: Update Existing Stories 117-118 (Increment 2 - Hierarchy)

4. Update Story117: CPL Concepts Parser - Hierarchy - Completed
   a. Update conceptLine rule to use new generator pattern - Completed
   b. Update example CPL code in story - Completed
   c. Update acceptance criteria - Completed

5. Update Story118: CPL Concepts Analyzer - Hierarchy - Completed
   a. Update example CPL code to use new syntax - Completed
   b. Verify no other changes needed - Completed

### Phase 3: Update Existing Stories 119-120 (Increment 3 - Multi-Section)

6. Update Story119: CPL Concepts Parser - Multi-Section - Completed
   a. Update example CPL code in story - Completed
   b. Verify no other changes needed - Completed

7. Update Story120: CPL Concepts Analyzer - Multi-Section - Completed
   a. Update example CPL code in story - Completed
   b. Verify no other changes needed - Completed

### Phase 4: Update Existing Stories 121-122 (Increment 4 - Conditional)

8. Update Story121: CPL Concepts Parser - Conditional - Completed
   a. Update conditionalSpec rule to use new generator pattern - Completed
   b. Update example CPL code in story - Completed
   c. Update acceptance criteria - Completed

9. Update Story122: CPL Concepts Analyzer - Conditional - Completed
   a. Update example CPL code to use new syntax - Completed
   b. Verify no other changes needed - Completed

### Phase 5: Update Existing Stories 123-124 (Increment 5 - Int Range)

10. Update Story123: CPL Concepts Parser - Range - Completed
    a. Update syntax from `Range[start..end]` to `Int[start, end, step]` - Completed
    b. Update grammar rules: change `RANGE` token to `INT`, add `COMMA` token - Completed
    c. Update IntRangeGeneratorDef struct to use Start/End/Step fields - Completed
    d. Update example CPL code in story - Completed
    e. Update acceptance criteria - Completed

11. Update Story124: CPL Concepts Analyzer - Range - Completed
    a. Update IntRangeGenerator struct to use Start/End/Step - Completed
    b. Update ExpandRangeValues to use step parameter - Completed
    c. Update example CPL code in story - Completed
    d. Update acceptance criteria - Completed

### Phase 6: Create New Stories 128-129 (Increment 6 - Concept Range)

12. Create Story128: CPL Concepts Parser - Concept Range - Completed
    a. Create new story file with goal and references - Completed
    b. Add grammar rules for conceptPatternGenerator - Completed
    c. Add DOTDOT token definition - Completed
    d. Define ConceptRangeGeneratorDef struct - Completed
    e. Add example CPL code showing concept range syntax - Completed
    f. Add tasks for lexer, grammar, listener implementation - Completed
    g. Add acceptance criteria - Completed

13. Create Story129: CPL Concepts Analyzer - Concept Range - Completed
    a. Create new story file with goal and references - Completed
    b. Define ConceptRangeGenerator struct with GenerateValues method - Completed
    c. Add resolution algorithm description - Completed
    d. Add error handling for circular dependencies - Completed
    e. Add example CPL code - Completed
    f. Add tasks for implementation - Completed
    g. Add acceptance criteria - Completed

### Phase 7: Update Existing Stories 125-126 (Increment 7 - Before/After)

14. Update Story125: CPL Concepts Parser - Before/After - Completed
    a. Update syntax to use new generator wrapper - Completed
    b. Update conceptExtension grammar rule - Completed
    c. Update example CPL code from `List[...]` to `[Values[...]]` - Completed
    d. Update ParserConceptDefinition references - Completed
    e. Update acceptance criteria - Completed

15. Update Story126: CPL Concepts Analyzer - Before/After - Completed
    a. Update to work with new generator structures - Completed
    b. Update ComputeValueOrdering documentation - Completed
    c. Update example CPL code in story - Completed
    d. Update acceptance criteria - Completed

### Phase 8: Update Epic001 Document

16. Update Epic001 document with new syntax examples - Completed
    a. Update all syntax examples to use `=[Values[...]]` format - Completed
    b. Add Increment 6 section with Story128 and Story129 references - Completed
    c. Add Increment 7 section with Story125 and Story126 references - Completed
    d. Update Increment 5 description to reflect `Int[start, end, step]` - Completed
    e. Add new generators section explaining Values, Int, and Concept Range - Completed

### Phase 9: Final Verification

17. Verify all story updates are consistent - Completed
    a. Check syntax consistency across all updated stories - Completed
    b. Verify cross-references are correct - Completed
    c. Verify story dependencies are accurate - Completed

18. Update stories toc.md - Completed
    a. Add entries for new Story128 and Story129 - Completed
    b. Verify all updated stories maintain correct status - Completed

## Constraints

- All syntax changes must align with ERO001 generator syntax specification
- Updated stories must maintain the same scope and intent - only syntax changes
- New stories (128-129) must follow the same format as existing parser/analyzer stories
- All example CPL code must use the new generator syntax consistently
- Must preserve all existing task structure while updating content

## Intent

The goal of this story is to align all Epic001-related stories with the new generator-based syntax for concept declarations. The ERO001 documents have been updated to use a more flexible generator pattern (`ConceptName=[Values[...]]` instead of `ConceptName=List[...]`), and the stories that implement these increments must reflect the new syntax. Additionally, this story creates the missing Increment 6 stories (Concept Range Generators) that were not included in the original story set.

## Acceptance Criteria

1. Story114-Story126 are updated with the new generator syntax
2. Story128 is created with complete parser specification for Concept Range Generators
3. Story129 is created with complete analyzer specification for Concept Range Generators
4. All example CPL code in updated stories uses the new syntax:
   - `Country=[Values[USA, Canada, Mexico]]` (not `Country=List[USA, Canada, Mexico]`)
   - `DayNumbers=[Int[1, 31, 1]]` (not `DayNumbers=Range[1..31]`)
   - `FirstHalf=[Quarter[Q1..Q2]]` for concept ranges
5. Epic001 document is updated with new syntax examples and Increment 6/7 sections
6. stories/toc.md is updated with entries for Story128 and Story129
7. All cross-references between stories are accurate
8. All tasks in this implementation story are marked as completed

## Notes

### Syntax Change Summary

| Aspect              | Old Syntax                   | New Syntax                      |
| ------------------- | ---------------------------- | ------------------------------- |
| Declaration wrapper | `ConceptName=Generator[...]` | `ConceptName=[Generator[...]]`  |
| List generator      | `List[Value1, Value2]`       | `Values[Value1, Value2]`        |
| Int range generator | `Range[start..end]`          | `Int[start, end, step]`         |
| Concept range       | `Source[start..end]`         | `Source[start..end]` (same)     |
| Multiple generators | Not supported                | `=[Gen1[...], Gen2[...]]`       |
| Before/after        | Same syntax, new wrapper     | `=[Values[...]] before/after X` |

### Story-Increment Mapping (Final)

| Increment | Feature         | Stories       |
| --------- | --------------- | ------------- |
| 1         | Basic Concepts  | 114, 115, 116 |
| 2         | Hierarchy       | 117, 118      |
| 3         | Multi-Section   | 119, 120      |
| 4         | Conditional     | 121, 122      |
| 5         | Int Range       | 123, 124      |
| 6         | Concept Range   | 128, 129      |
| 7         | Before/After    | 125, 126      |
| 8         | API Integration | (Future)      |

### Execution Order for Updates

The updates should be made in the following order:

1. **Increment 1 Updates** (Stories 114-116)
2. **Increment 2 Updates** (Stories 117-118)
3. **Increment 3 Updates** (Stories 119-120)
4. **Increment 4 Updates** (Stories 121-122)
5. **Increment 5 Updates** (Stories 123-124)
6. **Increment 6 New Stories** (Stories 128-129)
7. **Increment 7 Updates** (Stories 125-126 updates)
8. **Epic001 Document Update** (syntax examples)

This order ensures dependencies are maintained and each increment builds on the updated previous increment.

### Related Documents

- [Epic001 Update Plan](../epics/Epic001_update_plan.md) - Detailed plan for all changes
- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
- [Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md) - Parser-specific ERO
- [Concepts Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md) - Analyzer-specific ERO
- [Concepts Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO001_concepts-synthesis-phase.md) - Synthesizer-specific ERO
