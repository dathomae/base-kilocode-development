# ERO004: Connectors Analysis Phase

## Title

Connectors Analysis Phase - Data Structures and Validation Logic

## Purpose/Goal

Define the analysis phase data structures and logic for connector definitions in CPL. This phase transforms parser-level connector definitions into analysis-level structures, validating all references to concepts and measures.

## Implementation Status

Implemented

**Related Epics:**

- [Epic004_cpl-connectors-implementation.md](../../../../../epics/Epic004_cpl-connectors-implementation.md)

## Context

### Situation

The parser attaches a structured [`ConnectorDeclaration`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:176) record to each [`ConnectorScope.Data`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:247) field before popping the scope. This record contains the connector name, dimension names, an ordered list of grain measure names, celltype measure name, and the source location for each. The analysis phase reads connector roles exclusively from this structured record, not from the [`Definitions`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:240) slice. This separation allows a dimension concept and a grain measure to share the same textual name without ambiguity.

The analysis phase validates that dimension names refer to existing concepts, that the celltype name refers to an existing measure, and that every grain name in the ordered grain list refers to an existing measure. Grain interval-family type validation is deferred to a post-derivation pass because it requires type derivation paths that are only available after all type definitions have been processed.

### Constraints

- The analysis phase must validate that dimension names refer to existing concepts.
- The analysis phase must validate that the celltype name refers to an existing measure.
- The analysis phase must validate that every grain name in the ordered grain list refers to an existing measure.
- The analysis phase must defer grain interval-family type validation until after type derivation paths have been computed.
- The analysis phase must produce structures usable by the CIL generator when connectors are referenced in metric calculations.
- The analysis phase must read connector information from the structured [`ConnectorDeclaration`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:176) record attached to [`ConnectorScope.Data`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:247), not by inferring roles from the [`Definitions`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:240) slice.

### Requirements

- Build a unified Connectors lookup map from structured parser declarations.
- Validate connector dimensions against the Concepts lookup map.
- Validate connector celltype against the Measures lookup map.
- Validate connector grain names against the Measures lookup map.
- Validate connector grain type against the Interval family after type derivation paths are computed.
- Provide the necessary data structures so that metric calculations can generate a `CILConnectorReferenceNode`.

## Evaluation Criteria

1. **Correctness** (High importance): The analysis phase must correctly validate all references in a connector declaration.
2. **Integration** (High importance): The analysis phase must integrate with the `AnalysisDefinitions` structure and the `WalkSymbolTableWithHandlers` pattern.

## Decision

**Chosen Option:** Implement validation logic and build an `AnalysisConnectorDefinition` for each connector.

**Rationale:**
The analysis phase reads a structured [`ConnectorDeclaration`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:176) from each [`ConnectorScope.Data`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:247), validates the declared references against the `Concepts` and `Measures` maps, and populates a `Connectors` map in the `AnalysisDefinitions` struct. A deferred pass then validates every grain measure type against the Interval family after type derivation paths have been computed.

### Connector Analysis

**Data Structures**

```go
// AnalysisConnectorDefinition contains the analyzed connector information.
// All references have been validated and resolved to their corresponding definitions.
type AnalysisConnectorDefinition struct {
    Name           string
    Dimensions     []*AnalysisConceptDefinition // Resolved concept references
    Grain          []*AnalysisMeasureDefinition // Resolved ordered slice of grain measure references
    CellType       *AnalysisMeasureDefinition   // Resolved measure reference for celltype
    SourceLocation SourceLocation               // For error reporting
}
```

**Enhanced AnalysisDefinitions**

```go
// AnalysisDefinitions contains all analyzed definitions for CIL generation.
type AnalysisDefinitions struct {
    Concepts   map[string]AnalysisConceptDefinition
    Units      map[string]AnalysisUnitDefinition
    Measures   map[string]AnalysisMeasureDefinition
    Connectors map[string]AnalysisConnectorDefinition
    CILTrees   map[string]CILAbstractNode
}
```

**Connectors Analysis Function**

[`AnalyzeConnectorsScope`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/connectors_analysis.go:173) is called by [`WalkSymbolTableWithHandlers`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:753) for each [`ConnectorScope`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:78). It takes three parameters: the [`ConnectorScope`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:78) symbol table, the analysis context, and the analyzer.

The function retrieves the structured connector declaration from [`table.Data`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:247), which the parser has set to a [`*symboltable.ConnectorDeclaration`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:176). It validates that the data is present and has the correct type, then extracts the connector name.

For each connector, the function performs the following steps in order:

1. **Duplicate detection.** It checks whether the connector name already exists in the analyzer-wide connectors map. If it does, the function returns an error.

2. **Celltype resolution.** It looks up the celltype measure name in the analyzer's measures map. If the measure is not found, the function returns an error naming the connector and the missing celltype measure.

3. **Dimension resolution.** It iterates over the dimension names from the structured declaration and looks each one up in the analyzer's concepts map. If any dimension concept is not found, the function returns an error naming the connector and the missing dimension. The resolved concept definitions are collected in a slice.

4. **Grain resolution.** The structured declaration provides an ordered list of grain measure names (`GrainNames`). The function rejects an empty grain list and returns an error. For each grain measure name in the declared order, the function looks it up in the analyzer's measures map. If any grain measure is not found, the function returns an error naming the connector and the missing grain measure. A single-grain declaration becomes a single-element grain slice; the declaration order is preserved in the analyzed representation. Grain interval-family type validation is intentionally skipped at this stage.

5. **Definition creation.** The function assembles an [`AnalysisConnectorDefinition`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/connectors_analysis.go:37) from the resolved references, storing the grain measures as an ordered slice, and stores it in the analyzer's connectors map.

6. **CIL tree creation.** For each valid connector, the function creates a one-node CIL tree consisting of a [`CILConnectorReferenceNode`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/cil_nodes.go) and stores it in the analyzer's CILTrees map under the connector name.

The function does not reconstruct connector roles from the [`Definitions`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:240) slice. Because all roles are read from explicit fields in the structured declaration, a dimension concept and a grain measure may legally share the same textual name.

**Deferred Grain Validation**

[`ValidateConnectorGrainTypes`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/connectors_analysis.go:273) is a deferred validation pass that runs after [`ComputeDerivationPaths()`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:295) in [`Analyze()`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:725). It iterates over all connectors in the analyzer's connectors map and examines each connector's `Grain` slice. Connectors without an explicit grain are skipped.

Every grain measure in the slice must pass validation: a grain measure is accepted when its type name is the built-in [`Interval`](../../../../requirements/cpl/talking-about-time.md:12) type. A grain measure is also accepted when its type name names a declared type whose derivation path terminates in [`Interval`](../../../../requirements/cpl/talking-about-time.md:12), which indicates that the declared type ultimately derives from the built-in interval type. All other grain measure types are rejected, and the error names the connector, the grain measure index, the grain measure, and the grain measure type.

**Integration with WalkSymbolTableWithHandlers**

The [`Analyze`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:725) function registers [`AnalyzeConnectorsScope`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/connectors_analysis.go:173) as the entry handler for [`ConnectorScope`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:78) in the handler map passed to [`WalkSymbolTableWithHandlers`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:753). After the walk completes, [`Analyze`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:725) calls [`ComputeDerivationPaths`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:295), then validates interval type cross-references and unit concept references, and finally calls [`ValidateConnectorGrainTypes`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/connectors_analysis.go:273) to validate connector grain types.

**CIL Generation Process**

During connector analysis, the analyzer creates a one-node CIL tree for each connector. The tree consists of a single [`CILConnectorReferenceNode`](../../../../../../../../src/chronocone_planning_language/src/main/go/analysis/cil_nodes.go) whose payload is the connector name. This tree is stored in the analyzer's CILTrees map under the connector name. It serves as the root for connector references in metrics and aggregators during synthesis.

**Limitations:**

- The analysis phase only validates the references within the CPL file. It does not validate the connector against the external data source (e.g., parquet file schema). That validation occurs during synthesis or execution.
- Grain interval-family type validation is deferred until after all type definitions have been processed and derivation paths have been computed. This means that a CPL file with an invalid grain type will pass the first-pass connector analysis and fail only during the deferred validation step.
- Synthesis multi-grain output generation is not described in this document; it is deferred to [`Story160_cpl-synthesizer-connectors-time.md`](../../../../../../stories/Story160_cpl-synthesizer-connectors-time.md).

**Value Provided:**

The analysis phase ensures that connector declarations are semantically valid and makes their resolved definitions available for metric calculations to generate `CILConnectorReferenceNode`s.

## Implications

### Positive Implications

- Semantic validation ensures that only valid CPL programs proceed to synthesis.
- Clean integration with the existing `WalkSymbolTableWithHandlers` pattern.
- Clear separation between connector declaration (which builds the definition) and connector usage (which generates the CIL node).
- The structured [`ConnectorDeclaration`](../../../../../../../../src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:176) record eliminates ambiguity when a dimension concept and a grain measure share the same name, because each role is read from an explicit field rather than inferred from identifier position.

### Negative Implications / Trade-offs

- Grain interval-family type validation is deferred until after type derivation paths are computed. This means that connector analysis produces a seemingly valid connector definition before the deferred pass rejects it for an invalid grain type. The error is still reported during analysis, but it occurs in a separate pass after the initial connector scope walk.

## Notes

**Related Documents:**

- [CPL Requirements - Connectors](../../../../requirements/cpl/connectors.md)
- [CIL Representation](../cil-representation.md)
- [Symbol Table Design](../../go-parser-target/symbol-table-design.md)
