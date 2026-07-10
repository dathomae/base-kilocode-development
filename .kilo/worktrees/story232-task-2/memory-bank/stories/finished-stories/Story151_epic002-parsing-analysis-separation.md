# Story151: Epic002 Parsing/Analysis Separation Correction

## Goal

Correct the ERO002 documentation and Epic002 story files to properly separate parsing actions from analysis actions, aligning with the established patterns in the codebase. This story addresses architectural confusion where semantic validation was incorrectly placed in the parsing phase and handler patterns were inconsistent with existing code.

## References

- [Epic002_parsing_analysis_separation_plan.md](../epics/Epic002_parsing_analysis_separation_plan.md) - The correction plan this story implements
- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Epic for CPL types implementation (to be updated)
- [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) - Types parsing phase documentation (to be corrected)
- [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) - Types analysis phase documentation (to be corrected)
- [memory-bank/specifications/patterns/symbol-table-driven-parsing.md](../specifications/patterns/symbol-table-driven-parsing.md) - Pattern for parser structure
- [Story131_cpl-types-parser-number.md](Story131_cpl-types-parser-number.md) - Types parser story (to be updated)
- [Story132_cpl-types-analyzer-number.md](Story132_cpl-types-analyzer-number.md) - Types analyzer story (to be updated)

## Dependencies

None - This is a documentation correction story that establishes correct patterns before implementation continues.

## Dependent Stories

- [Story131_cpl-types-parser-number.md](Story131_cpl-types-parser-number.md) - Will use corrected patterns
- [Story132_cpl-types-analyzer-number.md](Story132_cpl-types-analyzer-number.md) - Will use corrected patterns
- Stories 133-140 - All types implementation stories depend on corrected documentation

## Tasks

1. Add TypesScope to Symbol Table - Not Started
   a. Add `TypesScope` constant to `ScopeType` enum in `symbol_table.go` after `ValueGeneratorScope` - Not Started
   b. Update `String()` method to include TypesScope case - Not Started
   c. Verify the change compiles and integrates with existing code - Not Started

   **Intent**: Establish the symbol table scope type for types section parsing.
   **Acceptance Criteria**: TypesScope is available for use in parser and analysis phases.

2. Correct ERO002_types-parsing-phase.md - Not Started
   a. Remove semantic validation from "Validation During Parsing" sections (constraint value checking, duplicate detection) - Not Started
   b. Update "Types Listener" section to show methods on unified `CPLListener` struct instead of separate `TypesListener` - Not Started
   c. Rename section to "Types Listener Methods" referencing `cpl_types_listener.go` file pattern - Not Started
   d. Simplify `ParserTypeDefinition` to raw data storage only, remove `GetNumberConstraints()` methods - Not Started
   e. Update documentation to show syntactic validation only (type name syntax, base type keyword, constraint format) - Not Started

   **Intent**: Align parsing phase documentation with established pattern: syntactic validation only, raw data storage, unified listener.
   **Acceptance Criteria**: Documentation shows correct pattern matching cpl_concepts_listener.go structure.

3. Correct ERO002_types-analysis-phase.md - Not Started
   a. Change handler pattern from `TypeAnalyzer` struct with methods to function-based handlers - Not Started
   b. Add `EnterAnalyzeTypesScope` function signature matching existing analysis pattern - Not Started
   c. Add `ExitAnalyzeTypesScope` function signature matching existing analysis pattern - Not Started
   d. Add constraint validation steps in `CreateConstraints` (duplicate constraint names, value type validation) - Not Started
   e. Add cross-section duplicate detection during symbol table walk - Not Started
   f. Update "Integration with WalkSymbolTable" section to show function-based handler registration - Not Started

   **Intent**: Align analysis phase documentation with established function-based handler pattern.
   **Acceptance Criteria**: Documentation shows `EnterAnalyzeTypesScope` and `ExitAnalyzeTypesScope` functions integrated with CPLAnalyzer.

4. Update Epic002_cpl-types-implementation.md - Not Started
   a. Add note about TypesScope needing to be added to symbol table before implementation - Not Started
   b. Update story descriptions to reflect parser = syntactic only, analyzer = semantic validation - Not Started
   c. Verify story ordering has parser stories before analyzer stories - Not Started
   d. Add cross-reference to corrected ERO002 documents - Not Started

   **Intent**: Ensure epic documentation reflects corrected separation of concerns.
   **Acceptance Criteria**: Epic documentation correctly describes parsing vs analysis responsibilities.

5. Update Implementation Stories 131-140 - Not Started
   a. Update Story131: Change Task 7 to show `cpl_types_listener.go` with methods on `CPLListener`, remove semantic validation from Task 8 - Not Started
   b. Update Story132: Change Task 5 to function-based handlers, add semantic validation to Task 3 - Not Started
   c. Update Stories 133, 135, 137, 139 (parser stories): Remove semantic validation tasks - Not Started
   d. Update Stories 134, 136, 138, 140 (analyzer stories): Ensure semantic validation is included - Not Started
   e. Verify all stories reference correct patterns (unified listener, function-based handlers) - Not Started

   **Intent**: Align all types implementation stories with corrected patterns.
   **Acceptance Criteria**: All stories 131-140 follow consistent parsing/analysis separation.

6. Verify CPLTypes.g4 Grammar Documentation - Not Started
   a. Verify grammar documentation shows CPLTypes.g4 as separate grammar file - Not Started
   b. Verify grammar documentation shows import by main grammar - Not Started
   c. Verify grammar follows established ANTLR patterns - Not Started
   d. Update any references to patterns or structures that have changed - Not Started

   **Intent**: Ensure grammar documentation is consistent with corrected architecture.
   **Acceptance Criteria**: Grammar documentation references correct patterns and structures.

## Notes

- This is a documentation-only story; no code implementation required
- The corrections align Epic002 types implementation with patterns established in Epic001 (concepts)
- Key pattern corrections:
  - Listener: Unified `CPLListener` with methods in separate files (not separate structs)
  - Validation: Syntactic only in parser, semantic in analysis
  - Handlers: Function-based (`EnterAnalyzeTypesScope`) not method-based
  - Constraint extraction: In analysis phase (`CreateConstraints`) not parser
  - Duplicate detection: Analysis-level during symbol table walk
- Follow Logical TDD Lifecycle principles: documentation corrections establish the "structural definition" for subsequent implementation stories
