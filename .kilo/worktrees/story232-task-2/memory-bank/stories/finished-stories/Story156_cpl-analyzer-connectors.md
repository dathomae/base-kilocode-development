# Story156: CPL Analyzer Connectors

## Goal

Implement the analysis phase for basic connector declarations. This validates that the referenced dimensions (concepts) and celltype (measure) exist in the symbol table. This story is part of [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md).

## References

- [ERO004: Connectors Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md)
- [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md)

## Dependencies

- [Story162_cpl-parser-connectors-basic.md](finished-stories/Story162_cpl-parser-connectors-basic.md) - Done

## Dependent Stories

- [Story157_cpl-synthesizer-connectors.md](Story157_cpl-synthesizer-connectors.md) - Not Started

## Tasks

1. Define `AnalysisConnectorDefinition` struct - Completed
   a. Phase A: Created struct definition with comprehensive docstrings - Completed
   b. Phase B: Wrote tests for struct creation and field access - Completed
   c. Phase C: Verified implementation - All 8 tests pass - Completed
   d. Phase D: Documentation - Comprehensive docstrings serve as documentation - Completed

   **Files Created:**
   - `src/chronocone_planning_language/src/main/go/analysis/connectors_analysis.go` - Struct definition
   - `src/chronocone_planning_language/src/main/go/analysis/connectors_analysis_test.go` - Test suite

   **Struct Fields:**
   - `Name` (string): The connector name
   - `Dimensions` ([]\*AnalysisConceptDefinition): Resolved concept references for dimensions
   - `CellType` (\*AnalysisMeasureDefinition): Resolved measure reference for celltype
   - `SourceLocation` (symboltable.SourceLocation): For error reporting

2. Update `AnalysisDefinitions` to include `Connectors` map and `CILTrees` map - Completed
   a. Phase A: Added `Connectors` and `CILTrees` fields to `AnalysisDefinitions` struct with comprehensive docstrings - Completed
   b. Phase B: Wrote tests for initialization and storage in `cpl_analyzer_connectors_test.go` - Completed
   c. Phase C: Updated `CPLAnalyzer` struct, `NewCPLAnalyzer()`, and `GetAnalysisDefinitions()` to include new maps - Completed
   d. Phase D: Documentation via comprehensive docstrings in struct definitions - Completed

   **Files Created/Modified:**
   - `src/chronocone_planning_language/src/main/go/analysis/cil_nodes.go` - Created: Defines `CILAbstractNode` interface and `CILConnectorReferenceNode`
   - `src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go` - Modified: Updated structs and methods
   - `src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer_connectors_test.go` - Created: Test suite for new functionality

   **All 15 new tests pass successfully.**

3. Implement `CILConnectorReferenceNode` and `CILAbstractNode` interface (basic structure) - Completed
   a. Phase A: Defined `CILAbstractNode` interface and `CILNodeType` constants - Completed
   b. Phase B: Created `CILConnectorReferenceNode` struct with comprehensive docstrings - Completed
   c. Phase C: Implemented `NodeType()`, `ConnectorName()`, and `NewCILConnectorReferenceNode()` methods - Completed
   d. Phase D: Verified all connector CIL tree tests pass - Completed

   **Files Created/Modified:**
   - `src/chronocone_planning_language/src/main/go/analysis/cil_nodes.go` - Contains CIL node definitions

4. Implement `AnalyzeConnectorsScope` function - Completed
   a. Phase A: Defined function signature with comprehensive docstrings - Completed
   b. Phase B: Wrote tests for dimension resolution, celltype resolution, duplicate detection, CIL tree generation - Completed
   c. Phase C: Implemented `AnalyzeConnectorsScope` logic including symbol validation and CIL tree creation - Completed
   d. Phase D: Verified all 12 connector scope tests pass - Completed

   **Test Coverage:**
   - Valid connectors with single/multiple/no dimensions
   - Undefined dimension and celltype error cases
   - Duplicate connector detection
   - CIL tree generation for single and multiple connectors

5. Register handler in `WalkSymbolTableWithHandlers` - Completed
   a. Phase A: Handler registered in `Analyze()` function in `cpl_analyzer.go` - Completed
   b. Phase B: Wrote integration test to verify handler invocation via `WalkSymbolTableWithHandlers` - Completed
   c. Phase C: Verified connector handler is called correctly during full analysis - Completed
   d. Phase D: All integration tests pass - Completed

   **Registration Location:**
   - `src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go` line 550

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The analyzer must correctly identify and report errors if a referenced dimension (concept) or celltype (measure) does not exist in the symbol table.
- The generated CIL tree must accurately represent the connector reference.

## Intent

- **Task 1**: To provide a structured representation of a connector after analysis.
- **Task 2**: To store the analyzed connectors and their corresponding CIL trees in the overall analysis definitions.
- **Task 3**: To represent a connector reference within the Compiler Intermediate Language (CIL) tree.
- **Task 4**: To perform the core semantic analysis of connectors, ensuring all references are valid and generating the CIL representation.
- **Task 5**: To integrate the connector analysis into the main symbol table walking process.

## Acceptance Criteria

- The analyzer successfully processes valid connector declarations.
- The analyzer reports appropriate errors for invalid connector declarations (e.g., missing dimensions or celltypes).
- The `AnalysisDefinitions` struct correctly contains the analyzed connectors and their CIL trees.
- All unit tests pass.
