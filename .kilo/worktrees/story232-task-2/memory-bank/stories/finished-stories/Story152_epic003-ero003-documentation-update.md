# Story152: Epic003 and ERO003 Documentation Update

## Goal

Update Epic003, the ERO003 documents, and associated stories to clarify the separation between parsing and analysis phases, document the migration path from the old unit implementation to the new ERO003 design, and ensure consistency across all Epic003-related documentation.

## References

- [Epic003](../epics/Epic003_units-and-measures-implementation.md)
- [ERO003 Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md)
- [ERO003 Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md)
- [ERO003 Incremental Implementation](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md)
- [Documentation Update Plan](../epic003_ero003_documentation_update_plan.md)
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md)
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md)

## Dependencies

- [Story151_epic002-parsing-analysis-separation.md](finished-stories/Story151_epic002-parsing-analysis-separation.md) - Done

This story depends on Story151 which corrected the ERO002 documentation to properly separate parsing from analysis actions. This story applies similar corrections to ERO003 and Epic003 documentation.

## Dependent Stories

- [Story141_cpl-units-parser-basic.md](Story141_cpl-units-parser-basic.md) - Not Started
- [Story142_cpl-units-analyzer-basic.md](Story142_cpl-units-analyzer-basic.md) - Not Started
- [Story143_cpl-measures-parser-basic.md](Story143_cpl-measures-parser-basic.md) - Not Started
- [Story144_cpl-measures-analyzer-basic.md](Story144_cpl-measures-analyzer-basic.md) - Not Started
- [Story145_cpl-compound-units-parser.md](Story145_cpl-compound-units-parser.md) - Not Started
- [Story146_cpl-compound-units-analyzer.md](Story146_cpl-compound-units-analyzer.md) - Not Started
- [Story147_cpl-unit-conversions-parser.md](Story147_cpl-unit-conversions-parser.md) - Not Started
- [Story148_cpl-unit-conversions-analyzer.md](Story148_cpl-unit-conversions-analyzer.md) - Not Started
- [Story149_cpl-measure-conversions-parser.md](Story149_cpl-measure-conversions-parser.md) - Not Started
- [Story150_cpl-measure-conversions-analyzer.md](Story150_cpl-measure-conversions-analyzer.md) - Not Started

The stories implementing Epic003 (141-150) will benefit from the clarified documentation about phase separation and migration path.

## Tasks

1. Update ERO003 Analysis Phase Document - Completed
   a. Clarify the Historical Note at line 30 - replace confusing text about conflated concerns with clear statement that parsing phase ERO has been corrected - Completed
   b. Add cross-reference to postfix expression parsing pattern in the "Parsing vs Analysis Phase Boundary" section - Completed
   c. Verify all phase separation terminology is consistent - Completed

2. Update Epic003 Document - Completed
   a. Add "Migration from Old Unit Implementation" section after "Notes" - document current state (pre-Epic003), terminology clarification, migration steps, and backward compatibility - Completed
   b. Update Constraints section - add constraint about breaking change with old UnitDecl AST type - Completed
   c. Verify consistency with ERO003 documents - Completed

3. Update ERO003 Parsing Phase Document - Completed
   a. Add "Migration Note: Old vs New Unit Grammar" section after "Context" - document old vs new grammar differences - Completed
   b. Clarify ParserUnitDefinition Data Field Usage in Increment 1 section - explain how it attaches to symbol table - Completed
   c. Verify all symbol table-driven parsing references are consistent - Completed

4. Update ERO003 Incremental Implementation Document - Completed
   a. Add "Migration from Existing Implementation" section after "Context" - document how new design replaces old implementation - Completed
   b. Verify consistency with parsing and analysis phase documents - Completed

5. Update Stories 141-150 - Completed
   a. Update Story141 (Basic Units Parser) - add migration note and acceptance criteria about deprecating old UnitDecl - Completed
   b. Update Story142-150 - add phase separation note to each story referencing postfix expression and symbol table-driven parsing patterns - Completed
   c. Verify all stories have consistent terminology - Completed

## Test-First Development

This is a documentation-only story. The "tests" for this story are the verification checklist items that confirm all documentation updates are complete and consistent. Each task includes verification steps to ensure the documentation is correct and complete.

## Constraints

1. **Documentation Only**: No code changes will be made as part of this story
2. **Consistency**: All documents must use consistent terminology for phase separation (parsing vs analysis)
3. **Pattern References**: All relevant documents must reference the postfix expression parsing and symbol table-driven parsing patterns
4. **Breaking Change Documentation**: The migration from old to new unit implementation must be clearly documented as a breaking change
5. **Cross-References**: Documents should cross-reference each other where appropriate

## Intent

The intent of this story is to ensure that developers implementing Epic003 stories (141-150) have clear, consistent documentation that:

1. Understands the phase separation between parsing and analysis
2. Knows how to migrate from the old unit implementation to the new design
3. Follows the established patterns for postfix expression parsing and symbol table-driven parsing

## Acceptance Criteria

1. ERO003 Analysis Phase document no longer has confusing historical note about conflated concerns
2. Epic003 has a complete "Migration from Old Unit Implementation" section with:
   - Current state comparison table
   - Terminology clarification
   - Migration steps for story implementation
   - Backward compatibility notes
3. ERO003 Parsing Phase has migration note section comparing old vs new grammar
4. ERO003 Parsing Phase clarifies ParserUnitDefinition Data field usage
5. ERO003 Incremental Implementation has migration context section
6. Story141 has migration note and updated acceptance criteria
7. Stories 142-150 each have phase separation notes
8. All documents consistently use postfix expression pattern terminology
9. All documents consistently use symbol table-driven parsing terminology
10. All 14 files are updated and consistent with each other

## Verification Checklist

After completing all updates:

- [x] ERO003 Analysis Phase no longer has confusing historical note
- [x] Epic003 has "Migration from Old Unit Implementation" section
- [x] ERO003 Parsing Phase has migration note
- [x] ERO003 Incremental Implementation has migration context
- [x] Stories 141-150 have phase separation notes
- [x] Stories 141-150 reference migration from old implementation
- [x] All documents consistently use postfix expression pattern terminology
- [x] All documents consistently use symbol table-driven parsing terminology

## Notes

- This story affects 14 files total: 4 main documents (Epic003 + 3 ERO003 files) and 10 story files (Story141-150)
- The documentation updates in this story set the foundation for implementing Epic003
- Similar work was done in Story151 for Epic002/ERO002
