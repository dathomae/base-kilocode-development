# Story163: CPL Units and Measures - Multiple Sections Support

## Goal

Implement support for multiple `units` and `measures` sections within the same CPL file, including parser support for multiple section declarations and analyzer cross-section duplicate detection. Each units section creates its own UnitsScope symbol table, and each measures section creates its own MeasuresScope symbol table. The analyzer detects duplicate unit or measure definitions across sections during the symbol table walk. This is part of [Epic003: CPL Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md) - Increment 2.

## References

- [Epic003: CPL Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md)
- [ERO003: Units and Measures Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md) - Increment 2 section
- [ERO003: Units and Measures Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md) - Increment 2 section
- [ERO003: Units and Measures Incremental Implementation](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md)
- [Units and Measures Design](../specifications/design/cpl/units-and-measures.md)

## Dependencies

- [Story153_cpl-units-parser-basic-reimplementation.md](finished-stories/Story153_cpl-units-parser-basic-reimplementation.md) - Done - Basic units parser with symbol table-driven parsing
- [Story142_cpl-units-analyzer-basic.md](finished-stories/Story142_cpl-units-analyzer-basic.md) - Done - Basic units analyzer with AnalyzeUnitScope
- [Story143_cpl-measures-parser-basic.md](finished-stories/Story143_cpl-measures-parser-basic.md) - Done - Basic measures parser with MeasuresScope symbol table
- [Story144_cpl-measures-analyzer-basic.md](finished-stories/Story144_cpl-measures-analyzer-basic.md) - Done - Basic measures analyzer with AnalyzeMeasureScope

## Dependent Stories

None

## Tasks

1. Verify parser supports multiple units sections - Completed
   a. Confirm unitsSection rule allows multiple occurrences via statement/document rules - Completed
   b. Confirm UnitsListener creates separate UnitsScope for each units section - Completed
   c. Verify scope push/pop behavior for each section entry/exit - Completed

   **Intent**: Determine if the existing parser already correctly handles multiple units sections.
   **Acceptance Criteria**: Clear understanding that parser creates sibling UnitsScope tables under GlobalScope for each section.

2. Write parser tests for multiple units sections - Completed
   a. Test file with two separate units sections - Completed
   b. Test file with three or more units sections - Completed
   c. Test that each section creates its own UnitsScope in symbol table - Completed
   d. Test unit definitions in different sections are independently parsed - Completed

   **Intent**: Verify the parser correctly handles multiple units sections and produces correct symbol table structure.
   **Acceptance Criteria**: All parser tests pass, demonstrating correct scoping and symbol table creation.

3. Verify analyzer handles multiple units sections - Completed
   a. Confirm EnterAnalyzeUnitScope is called for each UnitsScope during symbol table walk - Completed
   b. Confirm duplicate check (Units map) catches duplicates across sections - Completed
   c. Verify no special merge logic is required - Completed

   **Intent**: Verify existing analyzer logic correctly handles multiple units sections without code changes.
   **Acceptance Criteria**: Clear understanding that existing duplicate detection works across sections during WalkSymbolTableWithHandlers.

4. Write analyzer tests for units cross-section duplicate detection - Completed
   a. Test duplicate unit name in two different units sections (should error) - Completed
   b. Test unique unit names across multiple sections (should succeed) - Completed
   c. Test three sections with one duplicate (should error on duplicate) - Completed
   d. Test symbol table walk visits all UnitsScope tables - Completed

   **Intent**: Verify cross-section duplicate detection works correctly for unit definitions.
   **Acceptance Criteria**: All analyzer tests pass, demonstrating correct duplicate detection across units sections.

5. Verify parser supports multiple measures sections - Completed
   a. Confirm measuresSection rule allows multiple occurrences via statement/document rules - Completed
   b. Confirm MeasuresListener creates separate MeasuresScope for each measures section - Completed
   c. Verify scope push/pop behavior for each section entry/exit - Completed

   **Intent**: Determine if the existing parser already correctly handles multiple measures sections.
   **Acceptance Criteria**: Clear understanding that parser creates sibling MeasuresScope tables under GlobalScope for each section.

6. Write parser tests for multiple measures sections - Completed
   a. Test file with two separate measures sections - Completed
   b. Test file with three or more measures sections - Completed
   c. Test that each section creates its own MeasuresScope in symbol table - Completed
   d. Test measure definitions in different sections are independently parsed - Completed

   **Intent**: Verify the parser correctly handles multiple measures sections and produces correct symbol table structure.
   **Acceptance Criteria**: All parser tests pass, demonstrating correct scoping and symbol table creation.

7. Verify analyzer handles multiple measures sections - Completed
   a. Confirm EnterAnalyzeMeasureScope is called for each MeasuresScope during symbol table walk - Completed
   b. Confirm duplicate check (Measures map) catches duplicates across sections - Completed
   c. Verify no special merge logic is required - Completed

   **Intent**: Verify existing analyzer logic correctly handles multiple measures sections without code changes.
   **Acceptance Criteria**: Clear understanding that existing duplicate detection works across sections during WalkSymbolTableWithHandlers.

8. Write analyzer tests for measures cross-section duplicate detection - Completed
   a. Test duplicate measure name in two different measures sections (should error) - Completed
   b. Test unique measure names across multiple sections (should succeed) - Completed
   c. Test three sections with one duplicate (should error on duplicate) - Completed
   d. Test symbol table walk visits all MeasuresScope tables - Completed

   **Intent**: Verify cross-section duplicate detection works correctly for measure definitions.
   **Acceptance Criteria**: All analyzer tests pass, demonstrating correct duplicate detection across measures sections.

9. Write integration tests for mixed units and measures sections - Completed
   a. Test file with multiple units sections and multiple measures sections together - Completed
   b. Test that units and measures scopes do not interfere with each other - Completed
   c. Test cross-references from measures to units across different sections - Completed

   **Intent**: Verify correct behavior when both units and measures have multiple sections in the same file.
   **Acceptance Criteria**: Integration tests pass, demonstrating independent handling of units and measures sections.

## Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- Must use existing WalkSymbolTableWithHandlers infrastructure for analysis
- Must not break Increment 1 functionality for units or measures
- Cross-section duplicate detection happens in analysis phase, not parsing
- Follow the Logical TDD Lifecycle - write tests before implementation
- Each task must include its own test-and-implement cycle

## Intent

The goal of this story is to enable multiple `units` and `measures` sections in a single CPL file. This allows developers to organize their unit and measure definitions more flexibly, potentially grouping related definitions together or separating base definitions from extensions. The parser creates separate UnitsScope and MeasuresScope tables for each section, and the analyzer handles duplicate detection across all tables of the same scope type during the symbol table walk.

## Acceptance Criteria

1. Parser can successfully parse multiple units sections in the same file
2. Each units section creates its own UnitsScope symbol table as a sibling under GlobalScope
3. Parser can successfully parse multiple measures sections in the same file
4. Each measures section creates its own MeasuresScope symbol table as a sibling under GlobalScope
5. Duplicate unit definitions across multiple sections are detected as errors during analysis
6. Duplicate measure definitions across multiple sections are detected as errors during analysis
7. Valid multi-section scenarios (different names in each section) work correctly for both units and measures
8. Cross-references from measures to units work correctly across multiple sections
9. All unit tests pass following BTDD practices
10. Increment 1 functionality for units and measures still works

## Notes

**Increment 2 Capabilities:**

- Multiple units sections in the same file
- Multiple measures sections in the same file
- Cross-section duplicate detection for both units and measures
- Support for valid multi-section organization

**Increment 2 Limitations:**

- No compound unit expressions (added in later increment)
- No unit conversions (added in later increment)
- No measure conversions (added in later increment)
- No API integration

**Example CPL code that should parse:**

```cpl
units
    meter
    kilometer = 1000 * meter

measures
    Length : meter

units
    second
    minute = 60 * second

measures
    Time : second
```

**Symbol Table Structure:**

```
GlobalScope (root)
├── UnitsScope (first section)
│   ├── SymbolDefinition: meter
│   └── SymbolDefinition: kilometer
├── MeasuresScope (first section)
│   └── SymbolDefinition: Length
├── UnitsScope (second section)
│   ├── SymbolDefinition: second
│   └── SymbolDefinition: minute
└── MeasuresScope (second section)
    └── SymbolDefinition: Time
```

**Design Decision:**

No special merging logic during parsing or analysis. Each `units` section creates a new UnitsScope table, and each `measures` section creates a new MeasuresScope table. These tables are siblings under the GlobalScope. The symbol table walk naturally visits all scope tables of the same type, and the existing duplicate checks in the analyzer handlers catch cross-section duplicates.

**Related Documents:**

- [Units and Measures Incremental Implementation](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
