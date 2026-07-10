# ERO003: Units and Measures Synthesis Phase

## Title

Units and Measures Synthesis Phase - No Direct Synthesis Work

## Purpose/Goal

Document that units and measures definitions do not generate synthesis-phase work. Units and measures are analyzed and validated during the analysis phase, producing resolved structures (`AnalysisUnitDefinition`, `AnalysisMeasureDefinition`) with expression trees and conversion pairs. The synthesizer generates code from CIL elements produced by the analyzer — units and measures inform the CIL structure but do not require special synthesis handling.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic003: Units and Measures Implementation](../../../epics/Epic003_units-and-measures-implementation.md) - Epic tracking the implementation of units and measures

**Related Stories:**

- [Story153: CPL Units Parser - Basic Unit Definitions Reimplementation](../../../../stories/Story153_cpl-units-parser-basic-reimplementation.md) - Increment 1 parser (replaces superseded Story141)
- ~~Story141~~ through ~~Story150~~ - Superseded stories (old implementation incompatible with types system)
- **Note**: ERO003 documents that no synthesis-phase work is required for units and measures (all work is in parsing and analysis phases)

## Context

### Situation

The synthesis phase generates executable code from CIL (Chronocone Intermediate Language) elements produced by the analyzer. Units and measures are fully processed during the analysis phase:

- **Parsing phase**: Captures raw unit and measure definitions, pushing expression components to symbol table scopes in postfix order
- **Analysis phase**: Evaluates postfix sequences to build resolved expression trees, resolves unit/measure references, assembles conversion pairs
- **Synthesis phase**: Generates code from CIL — units and measures inform CIL structure (e.g., unit conversion logic, measure type handling) but do not require dedicated synthesis code

The synthesizer's role is to generate executable code from CIL elements. Units and measures contribute to the CIL structure through the analyzer's work, but the synthesizer does not need special handling for unit or measure definitions themselves.

### Constraints

- Units and measures are fully resolved during the analysis phase
- The synthesizer works from CIL elements, not raw analysis structures
- Unit conversion logic and measure type handling are captured in CIL by the analyzer
- No dedicated synthesis code is needed for units and measures

### Requirements

- Document that no synthesis work is required for units and measures
- Ensure each increment explicitly notes the lack of synthesis work
- Reference the parser and analysis ERO003 documents for context

### Assumptions

- The analysis phase produces complete CIL elements for units and measures
- The synthesizer generates code from CIL, not from analysis structures directly
- Future increments may add synthesis work if new unit/measure features require it

## Data Structures by Increment

### Increment 1: Basic Unit Definitions

**Synthesizer Changes**

No changes to the synthesizer. Unit definitions (`UnitName = ConceptName`) are validated during analysis and contribute to the CIL structure. The synthesizer generates code from CIL elements — unit definitions themselves do not generate synthesis code.

**Limitations**

- Synthesizer generates only placeholder code (if any) for Increment 1

### Increment 2: Basic Measure Definitions

**Synthesizer Changes**

No changes to the synthesizer. Measure definitions (`MeasureGroup = Number(unit=UnitGroup)`) are validated during analysis. The analyzer resolves unit group references and produces CIL elements reflecting the measure's type and unit association. The synthesizer generates code from those CIL elements.

**Limitations**

- Synthesizer generates only placeholder code (if any) for Increment 2

### Increment 3: Compound Units and Unit Algebra Tree

**Synthesizer Changes**

No changes to the synthesizer. Compound unit expressions (e.g., `Length/Time^2`) are parsed into postfix sequences and evaluated during analysis to build resolved expression trees. The analyzer produces CIL elements reflecting the resolved unit algebra. The synthesizer generates code from those CIL elements.

**Limitations**

- Synthesizer generates only placeholder code (if any) for Increment 3

### Increment 4: Unit Conversions

**Synthesizer Changes**

No changes to the synthesizer. Unit conversion clauses (pattern/calculation pairs) are assembled during analysis into `AnalysisConversionPair` structures with resolved expression trees. The analyzer produces CIL elements that capture the conversion logic. The synthesizer generates code from those CIL elements.

**Limitations**

- Synthesizer generates only placeholder code (if any) for Increment 4

### Increment 5: Measure Conversions

**Synthesizer Changes**

No changes to the synthesizer. Measure conversion clauses are assembled during analysis into `AnalysisMeasureConversionPair` structures with resolved pattern and calculation expressions. The analyzer produces CIL elements that capture the measure conversion logic. The synthesizer generates code from those CIL elements.

**Limitations**

- Synthesizer generates only placeholder code (if any) for Increment 5

## Options Considered

### Option 1: Generate Unit/Measure Definition Code

**Description:**
Generate Go code for unit and measure definitions (metadata, conversion tables) that would be available at runtime.

**Pros:**

- Unit/measure metadata available at runtime
- Could support dynamic unit/measure inspection

**Cons:**

- Unit and measure definitions are used by the analyzer, not the synthesized code at runtime
- The synthesized code only needs the computation logic derived from analysis
- Unnecessary complexity

**Decision:** Rejected. Unit and measure definitions are used by the analyzer; the synthesized code only needs the CIL elements that capture the computation logic.

### Option 2: Synthesizer Works from CIL

**Description:**
The analyzer processes unit and measure definitions to produce CIL elements. The synthesizer generates code from CIL. Units and measures do not require special synthesis handling.

**Pros:**

- Clean separation: analyzer handles unit/measure definitions, synthesizer handles CIL
- Synthesized code reflects the computation logic derived from analysis
- No duplication of unit/measure resolution logic in synthesis

**Cons:**

- Requires complete CIL representation for unit/measure features

**Decision:** Selected. The analyzer uses unit and measure definitions to produce CIL. The synthesizer generates code from CIL. No dedicated synthesis work is needed for units and measures.

## Decision

**Chosen Option:** Synthesizer Works from CIL

**Rationale:**
Units and measures define relationships and constraints that are used by the analyzer to validate CPL programs and generate appropriate CIL. The CIL captures the computation logic with unit conversions, measure type handling, etc. The synthesizer generates executable code from CIL — it does not need to process unit or measure definitions directly. This maintains a clean separation between analysis (unit/measure definitions) and synthesis (CIL code generation).

## Implications

### Positive Implications

- Clean separation between analysis (unit/measure definitions) and synthesis (CIL)
- No duplication of unit/measure resolution logic in synthesis
- Synthesized code reflects the computation logic derived from analysis

### Negative Implications

- Requires complete CIL representation for all unit/measure features
- Any unit/measure feature that affects code generation must be captured in CIL by the analyzer

### Code Generation Strategy

Units and measures do not generate synthesis code in Increments 1-5. The synthesizer generates code from CIL elements produced by the analyzer:

1. **Increment 1:** No synthesizer changes (basic unit definitions handled in analysis)
2. **Increment 2:** No synthesizer changes (basic measure definitions handled in analysis)
3. **Increment 3:** No synthesizer changes (compound unit expressions handled in analysis)
4. **Increment 4:** No synthesizer changes (unit conversions handled in analysis)
5. **Increment 5:** No synthesizer changes (measure conversions handled in analysis)

## Notes

**Related Documents:**

- [Units and Measures Parsing Phase](../go-parser-target/eros/ERO003_units-measures-parsing-phase.md) - Parser implementation for units and measures
- [Units and Measures Analysis Phase](../go-analysis/eros/ERO003_units-measures-analysis-phase.md) - Analysis implementation for units and measures
- [Units and Measures Incremental Implementation](../../eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Synthesis Design](../synthesis.md) - General synthesis phase design
