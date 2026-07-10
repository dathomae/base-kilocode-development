# Story127: ADR to ERO Rename

## Goal

Rename ADR (Architecture Decision Record) to ERO (Epic Restraining Order) across the entire ChronoCone project. The "Epic Restraining Order" terminology adds a humorous/aggressive tone while maintaining technical accuracy. EROs are temporary documents that capture design decisions specifically related to managing confusion from incremental implementation.

## References

- [ADR to ERO Rename Plan](../adr_to_ero_rename_plan.md)
- [Terms](../terms.md)
- [Concepts](../concepts.md)

## Dependencies

None - This is a documentation and file renaming task that can be executed independently.

## Dependent Stories

None

## Tasks

1. Update Memory Bank Terms and Concepts - Completed
   a. Update `memory-bank/terms.md` line 55: Change ADR term definition to ERO with new definition capturing the purpose of EROs - Completed
   b. Update `memory-bank/concepts.md` line 24: Change ADR concept to ERO with comprehensive definition - Completed
   c. Add new section to concepts.md after line 30 explaining why EROs are needed and their specific scope - Completed
   d. Verify ERO scope is clearly distinguished from general architecture/design decisions - Completed

2. Rename and Update Template File - Completed
   a. Rename `resources/templates/adr.md` to `resources/templates/ero.md` - Completed
   b. Update all internal ADR references to ERO throughout the template - Completed
   c. Update tone to reflect "Epic Restraining Order" spirit (e.g., "When to Issue an ERO" instead of "When to Write an ADR") - Completed
   d. Update file naming convention from ADRNNN to ERONNN - Completed
   e. Verify template file exists and contains correct content - Completed

3. Rename and Update Script File - Completed
   a. Rename `scripts/add-new-adr.pdd.script.md` to `scripts/add-new-ero.pdd.script.md` - Completed
   b. Update all internal ADR references to ERO throughout the script - Completed
   c. Update all parameter names from adr*\_ to ero*\_ - Completed
   d. Update file patterns from ADRNNN to ERONNN - Completed
   e. Update directory references from adrs/ to eros/ - Completed
   f. Verify script file exists and contains correct content - Completed

4. Rename ADR Directories and Files - Completed
   a. Rename `memory-bank/specifications/design/cpl/adrs/` to `eros/` - Completed
   b. Rename `memory-bank/specifications/design/cpl/go-analysis/adrs/` to `eros/` - Completed
   c. Rename `memory-bank/specifications/design/cpl/go-parser-target/adrs/` to `eros/` - Completed
   d. Rename `memory-bank/specifications/design/cpl/go-synthesis-target/adrs/` to `eros/` - Completed
   e. Rename ADR files to ERO files (ADR001** to ERO001**) in all 4 directories - Completed
   f. Update internal references in each ERO file to use new paths and ERO terminology - Completed
   g. Verify all directories and files renamed correctly - Completed

5. Update All TOC Files - Completed
   a. Update `resources/templates/toc.md` line 5: Change adr.md reference to ero.md - Completed
   b. Update `memory-bank/specifications/design/cpl/toc.md` line 11: Change ADR to ERO reference - Completed
   c. Update `memory-bank/specifications/design/cpl/eros/toc.md` (renamed from adrs/toc.md): Update headers and file links - Completed
   d. Update `memory-bank/specifications/design/cpl/go-analysis/eros/toc.md`: Update headers and file links - Completed
   e. Update `memory-bank/specifications/design/cpl/go-parser-target/eros/toc.md`: Update headers and file links - Completed
   f. Update `memory-bank/specifications/design/cpl/go-synthesis-target/eros/toc.md`: Update headers and file links - Completed
   g. Verify all TOC files display correctly with updated links - Completed

6. Update Epic File - Completed
   a. Update `memory-bank/epics/Epic001_cpl-concepts-implementation.md` lines 15-18: Update ADR references to ERO with new paths - Completed
   b. Update lines 26-40: Update Stories table ADR column links to ERO - Completed
   c. Update line 46: Change ADR001 reference to ERO001 - Completed
   d. Update lines 100-102: Update ADR references in acceptance criteria - Completed
   e. Verify all ADR references changed to ERO with correct paths - Completed

7. Update Story Files - Completed
   a. Update `memory-bank/stories/Story114_cpl-concepts-parser-basic.md`: Update ADR references to ERO - Completed
   b. Update `memory-bank/stories/Story115_cpl-concepts-analyzer-basic.md`: Update ADR references to ERO - Completed
   c. Update `memory-bank/stories/Story116_cpl-concepts-synthesizer-basic.md`: Update ADR references to ERO - Completed
   d. Update `memory-bank/stories/Story117_cpl-concepts-parser-hierarchy.md`: Update ADR references to ERO - Completed
   e. Update `memory-bank/stories/Story118_cpl-concepts-analyzer-hierarchy.md`: Update ADR references to ERO - Completed
   f. Update `memory-bank/stories/Story119_cpl-concepts-parser-multi-section.md`: Update ADR references to ERO - Completed
   g. Update `memory-bank/stories/Story120_cpl-concepts-analyzer-multi-section.md`: Update ADR references to ERO - Completed
   h. Update `memory-bank/stories/Story121_cpl-concepts-parser-conditional.md`: Update ADR references to ERO - Completed
   i. Update `memory-bank/stories/Story122_cpl-concepts-analyzer-conditional.md`: Update ADR references to ERO - Completed
   j. Update `memory-bank/stories/Story123_cpl-concepts-parser-range.md`: Update ADR references to ERO - Completed
   k. Update `memory-bank/stories/Story124_cpl-concepts-analyzer-range.md`: Update ADR references to ERO - Completed
   l. Update `memory-bank/stories/Story125_cpl-concepts-parser-before-after.md`: Update ADR references to ERO - Completed
   m. Update `memory-bank/stories/Story126_cpl-concepts-analyzer-before-after.md`: Update ADR references to ERO - Completed
   n. Verify all 13 story files updated with correct ERO paths - Completed

8. Update add-new-epic Script and Other Referencing Files - Completed
   a. Update `scripts/add-new-epic.pdd.script.md`: Change associated_adr to associated_ero, update patterns from ADR to ERO - Completed
   b. Update `resources/templates/epics.md`: Update ADR references to ERO, update examples - Completed
   c. Update `memory-bank/specifications/design/cpl/go-parser-target/symbol-table-design.md` line 37: Update ADR reference to ERO - Completed
   d. Verify all scripts and templates updated with ERO terminology - Completed

## Test-First Development

This story involves file renaming and content updates rather than code implementation. The verification approach is:

1. **Structural Definition (Phase A)**: Define the expected final state of all renamed files and directories before making changes
2. **Logical Tests (Phase B)**: Create verification checklists to confirm:
   - No files with "adr" in the name exist (except in archive directories)
   - No directories named "adrs" exist
   - All ERO files are accessible via new paths
   - All TOC files display correctly with updated links
3. **Implementation (Phase C)**: Execute the rename operations and content updates
4. **Verification**: Run the verification checklist to confirm all changes completed successfully

## Constraints

- Tasks must be executed in the specified order to maintain consistency
- This is a large-scale rename operation affecting many files - careful execution is required
- All file references must be updated to prevent broken links
- Directory structure remains the same, only directory names change (adrs → eros)
- The three-digit numbering scheme is maintained (ERO001, ERO002, etc.)
- If issues are discovered during execution, stop immediately and document the issue

## Intent

The intent of this story is to:

1. Rename all ADR (Architecture Decision Record) artifacts to ERO (Epic Restraining Order)
2. Update the tone of documentation to reflect the more aggressive/humorous nature of "Epic Restraining Order"
3. Ensure all internal references, links, and paths are updated to prevent broken references
4. Maintain the same organizational structure while updating terminology

## Acceptance Criteria

- [x] No files with "adr" in the name exist (except in archive directories)
- [x] No directories named "adrs" exist
- [x] Search for "ADR" returns no results in markdown files (except in adr_to_ero_rename_plan.md which is the plan document)
- [x] Search for "Architecture Design Record" returns no results
- [x] All ERO files are accessible via new paths (memory-bank/specifications/design/cpl/eros/, etc.)
- [x] All TOC files display correctly with updated links
- [x] Template file `resources/templates/ero.md` exists and contains correct ERO content
- [x] Script file `scripts/add-new-ero.pdd.script.md` exists and contains correct ERO content
- [x] Epic file references updated ERO paths correctly
- [x] All 13+ story files reference updated ERO paths correctly
- [x] memory-bank/terms.md contains updated ERO definition
- [x] memory-bank/concepts.md contains updated ERO concept definition

## Notes

- This is a documentation and file organization task, not a coding task
- The rename operation affects 4 directories, 4 ADR files, 6+ TOC files, 1 epic file, 13+ story files, 1 template, and 2 scripts
- The "Epic Restraining Order" terminology should be used consistently throughout all updated content
- EROs are temporary documents that capture design decisions specifically related to managing confusion from incremental implementation
- Rollback plan: If issues are discovered, stop execution, document the issue, and restore from git before the rename operation if needed
