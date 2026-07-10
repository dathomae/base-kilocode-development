# ERO002: Types Synthesis Phase

## Title

Types Synthesis Phase - No Direct Synthesizer Impact

## Purpose/Goal

Document that type definitions do not require any work in the synthesis phase. The synthesizer uses CIL produced by the analyzer to generate code, but types themselves do not generate CIL elements. The analyzer uses type definitions for validation during analysis; the synthesized code is unaware of type metadata.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic001_cpl-concepts-implementation.md](../../../../../epics/Epic001_cpl-concepts-implementation.md) - Epic for CPL concepts implementation

## Context

### Situation

The synthesis phase generates executable code from CIL (Chronocone Intermediate Language) elements produced by the analyzer. The analyzer uses type definitions to:

- Validate metric type compatibility
- Ensure type conversions follow the rules (only to less restrictive types)
- Validate measure and unit declarations
- Determine appropriate data representations for generated code

### Constraints

- Type definitions do not generate CIL elements
- Synthesized code uses Go types that correspond to CPL types, but has no runtime type metadata
- Synthesized code structure depends on type decisions made during analysis
- Type checking and validation happens entirely in the analysis phase

### Requirements

- No synthesizer changes required for types implementation
- Synthesizer continues to generate code from CIL elements produced by the analyzer
- Generated Go code uses appropriate types based on analysis decisions

### Assumptions

- Analysis phase has validated all type definitions and references
- CIL captures the computation chain with appropriate type information for code generation
- Synthesized code uses Go types that reflect CPL type decisions made at analysis time
- Type validation errors are caught during analysis, not synthesis

## Data Structures by Increment

### Increment 1: Number Type with Type Capture

**Overview**

Type definitions do not generate CIL elements and therefore require no synthesizer changes. The analyzer uses type information to validate the CPL program during analysis. Types affect code generation only indirectly when used by other CPL constructs (e.g., metrics).

**Synthesizer Changes**

None. The synthesizer does not need to be modified for types support.

### Increment 2: String and Boolean Types

**Overview**

String and Boolean types are validated during analysis but do not generate CIL elements.

**Synthesizer Changes**

None.

### Increment 3: Multiple Type Declaration Sections

**Overview**

Multiple type sections are handled entirely within the analysis phase. The synthesizer is unaware of how types were organized in the source CPL.

**Synthesizer Changes**

None.

### Increment 4: Instant Type

**Overview**

Instant types are validated during analysis but do not generate CIL elements.

**Synthesizer Changes**

None.

### Increment 5: Interval Type

**Overview**

Interval types are validated during analysis but do not generate CIL elements.

**Synthesizer Changes**

None.

## Options Considered

### Option 1: Generate Type Metadata Code

**Description:**
Generate Go code that includes runtime type metadata for CPL types.

**Pros:**

- Runtime type information available
- Could support dynamic type inspection

**Cons:**

- Type metadata is used by the analyzer, not the synthesized code at runtime
- The synthesized code only needs to use appropriate Go types
- Unnecessary complexity and runtime overhead

**Decision:** Rejected. Type definitions are used by the analyzer for validation; the synthesized code only needs to use appropriate Go types based on analysis decisions.

### Option 2: Types Known to Synthesized Code via Go Types

**Description:**
Type definitions do not generate CIL elements. The analyzer uses type information to validate the CPL program and generate appropriate CIL. The synthesizer generates Go code from CIL, using Go types that correspond to CPL types. The synthesized code has no knowledge of CPL type metadata, only the Go types determined during analysis.

**Pros:**

- Clean separation: analyzer deals with type definitions, synthesizer deals with CIL
- No runtime overhead from type metadata
- Type errors caught during analysis, not at runtime

**Cons:**

- No runtime type introspection (not needed for current use cases)

**Decision:** Selected. The analyzer validates type definitions and references. The synthesizer generates Go code using appropriate types. There is no synthesizer work required for types.

## Decision

**Chosen Option:** No synthesizer changes required for types

**Rationale:**
Types define constraints and relationships that are used by the analyzer to validate CPL programs. The analyzer produces CIL with appropriate type information for code generation. The synthesizer generates executable Go code from CIL—the synthesized code uses Go types (`float64`, `string`, `bool`, `time.Time`, software based extended types, etc.) that correspond to CPL types, but has no knowledge of CPL type metadata or constraint information.

## Implications

### Positive Implications

- Clean separation between analysis (type definitions) and synthesis (CIL → Go)
- No synthesizer complexity added for types
- No runtime overhead from type metadata
- Type errors caught early during analysis

### Negative Implications

- No runtime type introspection (not needed for current use cases)
- Cannot dynamically adapt to type definition changes at runtime

### Code Generation Strategy

Type definitions do not require synthesizer changes in any increment. Types are used by the analyzer for validation only; they do not generate CIL elements and therefore do not affect synthesis.

**Note:** All type-related work happens in the analysis phase. Type definitions, constraint validation, and cross-reference checking are all handled during analysis. The synthesizer's work with types will come later as part of the metrics implementation ERO, where type information will inform code generation for metric evaluations.

## Notes

**Related Documents:**

- [Types Incremental Implementation](../../eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Types Definitions Design](../../types-definitions.md) - Full type design
- [Synthesis Design](../synthesis.md) - General synthesis phase design
