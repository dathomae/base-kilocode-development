# Concept Definitions Design

## Overview

Concepts are a fundamental building block of the Chronocone Planning Language (CPL). They define the dimensional structure of metrics and provide the values that index time series data. This document outlines the design for how concept definitions are captured, stored, and used throughout the CPL processing pipeline.

## Key Design Principle

**Concept definitions (both relationships AND values) are known at compile, analysis, synthesize, and runtime.** This enables static validation, optimization, and pre-computation that was not possible when concept values were only known at runtime.

## Core Rules for Concepts

### 1. Concept Sections

Concepts are defined within `concepts` sections. A CPL file may contain one or more `concepts` sections, and concept definitions may be spread across multiple files within the same compilation unit.

The general form of a concept section is:

```
concepts
    <list of concept definitions>
```

where each concept definition has the form:

```
<concept name> = [<list of generator patterns>]
```

```
concepts
    Country=[Values[USA, Canada, Mexico]]
      Region=[Values[North, South]]
```

Concepts may also be defined in the database (e.g., pre-defined taxonomies) and are treated as if they were defined in earlier concept sections in other files.

### 2. Concept Identity

A concept name in one concept section always refers to the same concept as the same name in a different concept section. Concept names are globally unique within a compilation unit and namespace.

### 3. Hierarchical Structure

Concepts form a taxonomic hierarchy within a namespace. This structure is essential for CPL analysis, aggregation, and spreading.

- Concepts can have parent-child relationships.
- A child concept "expands" its parent, providing more granular classification.
- Indentation in the `concepts` section explicitly defines this hierarchy.

### 4. Scoping via Indentation

CPL uses indentation to indicate scoping. A child concept is indented one level more than its immediate parent concept:

```cpl
concepts
    Duration=[Values[Year, Quarter, Month, Week, Day, Hour, Minute, Second]]
      WorkdayShift=[Values[Morning, Afternoon, Night]]  # Child of Duration
```

### 5. Generator-Based Value Definition

Every concept definition uses a **generator** that produces the set of values for that concept. The general form is:

```
<concept name> = <generator type>[<generator pattern>] [before <value>] [after <value>]
```

The `before` and `after` clauses are optional and used when extending existing concepts (see Section 6).

#### Supported Generator Types

**Values Generator**

- Pattern: `Values[<comma-separated list of values>]`
- Example: `Country=[Values[USA, Canada, Mexico]]`
- Produces: The ordered set of values specified in the list

**Int Generator**

- Pattern: `Int[<start>, <end>, <step>]` where start and end are integers, step is optional and defaults to 1
- Example: `DayNumbers=[Int[1, 31, 1]]`
- Produces: The inclusive set of integers from start to end, incrementing by step
- Constraints:
  - The end value is inclusive
  - The step value must be a positive integer
  - The start value must be less than or equal to the end value

**Concept Range Generator**

- Pattern: `<concept name>[<start concept value>..<end concept value>]`
- Example: `QuarterRange=[Quarter[Q1..Q4]]`
- Produces: All concept values from the named concept, from start to end (inclusive), based on the order of values in that concept
- Constraints:
  - The start concept value must appear earlier in the concept's order than the end concept value
  - Both values must exist in the referenced concept

**Multiple Generators**

A concept definition may include multiple generators. The relative order of values reflects the order of generators in the definition:

```cpl
concepts
    DayNumbers=[Int[4, 7, 1], Int[1, 3, 1]]
```

This produces values in order: 4, 5, 6, 7, 1, 2, 3.

### 6. Extending Concepts with Before/After Clauses

The `before` and `after` clauses enable inserting new values into an existing ordered concept without reordering existing values. This is useful when extending pre-defined concepts:

```cpl
concepts
    Duration=[Values[Year, Quarter, Month, Week, Day, Hour, Minute, Second]]

# Extending Duration to add Shift between Day and Hour
concepts
    Shift=[Values[Shift]] before Day after Hour
```

In this example:

- `Shift` is inserted into the Duration hierarchy
- The relative order of existing Duration values is preserved
- The resulting order is: Year, Quarter, Month, Week, Day, Shift, Hour, Minute, Second

### 7. Multiple Concept Definitions

A concept may be defined multiple times across concept sections. When adding to concepts with multiple definitions:

- **Additions may only add to the values already present** - there is no way to delete existing values
- **Additions may not redefine existing values** (e.g., by reordering)
- **Tree structure violations are errors** - you cannot place an existing concept name under a different parent that is not its immediate parent

Example of valid extension:

```cpl
# File 1
concepts
    Country=[Values[USA, Canada, Mexico]]

# File 2 - Adding more countries
concepts
    Country=[Values[Brazil, Argentina]] # Extends Country with additional values
```

It is an error to define an existing concept value that already exists in the concept from a previous declaration. Thus, concepts may be extended but not redefined.

### 8. Value Ordering Immutability

When inserting new values using `before`/`after`:

- The relative order of values that existed before the addition remains unchanged
- New values are inserted at the specified position relative to existing values
- The resulting order preserves all original relative orderings

Example:

```cpl
# Original
concepts
    Priority=[Values[Low, Medium, High]]

# Extension
concepts
    Critical=[Values[Critical]] before High

# Resulting order: Low, Medium, Critical, High
# Original relative order (Low before Medium before High) is preserved
```

If no before declaration is included then the system may choose to insert the values in the list before any value that is already in the list. If no after declaration
is included then the system may choose to insert the values in the list after any value that is in the list. The resulting set of values for the concept will always result
in a set with a completely defined relative order.

### 9. Conditional Value Specifications

Child concept values can vary based on parent concept values using conditional specifications:

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
      Region=Country[USA] => [Values[New England, Mid Atlantic, South Atlantic, South]]
      Region=Country[Canada] => [Values[Ontario, Quebec, British Columbia]]
      Region=[Values[North, South]]  # Default for Mexico
```

The conditional syntax `<child> = <parent>[<value>] => [<generator>]` specifies which child values exist for each parent value. More specific definitions take precedence over broader definitions.

## Concept Structure

A concept has three properties:

1. **Name** - The identifier used to reference the concept
2. **Values** - An ordered set of zero or more values associated with the concept
3. **Relationships** - Parent-child relationships with other concepts

### Concept Values

Concept values are ordered, and the order is significant for operations that depend on it for calculations or display. Values may be:

- Explicitly listed via Values generator (e.g., `[Values[USA, Canada, Mexico]]`)
- Generated from an integer range via Int generator (e.g., `[Int[1, 10, 1]]`)
- Generated from a concept range via concept range generator (e.g., `[Quarter[Q1..Q4]]`)
- Extended using before/after clauses to insert values at specific positions
- Defined using multiple generators combined in a single definition

### Parent-Child Relationships

Each concept may have:

- Zero or one parent concepts
- Zero or more child concepts

The parent-child relationship means that for every value of the parent concept, there are zero or more values of the child concept. Child values for a given parent value may differ across parent values, and child values may be shared between parent values.

## Concept Definition Types

The following definition types capture concept information at different phases of processing:

### 1. Parser Concept Definition

During parsing, concept definitions are captured from the AST and stored in parser data structures:

- Raw concept name and generator pattern as parsed
- Parent-child relationships as declared via indentation
- Conditional value specifications (e.g., `Region = Country[USA] => [Values[...]]`)
- Before/after clauses for value ordering
- Source location information for error reporting

These definitions are pushed onto symbol table entries during parsing.

### 2. AnalysisDefinitions - Concept Definition

After parsing completes, the symbol table is walked to assemble the **Concepts lookup map** as part of AnalysisDefinitions:

- **ConceptInfo** - Core concept information including name and all values
- **ConceptHierarchy** - Parent-child relationship mappings
- **ConceptValueMapping** - Maps parent values to their associated child values
- **ValueOrdering** - Preserves the declared order of values, including insertions

The Concepts lookup map enables the analyzer to:

- Validate concept references in metrics and calculations
- Determine valid dimensional combinations
- Check aggregator and spreader compatibility
- Resolve concept values for CIL generation

### 3. CILDefinition - Concept References

In the Chronocone Intermediate Language (CIL), concepts are referenced in:

- **Dimension specifications** - Which concepts index a tensor
- **Value lists** - The ordered set of values for each dimension
- **Aggregation/spread matrices** - Transformation matrices for dimensional operations

Since concept values are known at synthesis time, the CIL can include:

- Pre-computed value lists
- Pre-built transformation matrices
- Optimized dimension orderings based on value counts

### 4. Synthesized Code - Concept Integration

In the generated executable code:

- Concept values are embedded as constants or initialization data
- Dimension orderings are fixed based on the known value sets
- Loop bounds and array dimensions are known at compile time of the synthesized code

## Data Flow

```
CPL Source
    ↓
[Parser] → Parser Concept Definitions → Symbol Table
    ↓
[Symbol Table Walk] → Concepts Lookup Map (AnalysisDefinitions)
    ↓
[Analyzer] → Validated Concept References in Data/Control Flow Graph
    ↓
[CIL Generation] → CILDefinition Trees with Concept Info
    ↓
[Synthesizer] → Generated Code with Embedded Concept Data
    ↓
[Runtime] → Executable using Known Concept Values
```

## Usage in Different Contexts

### Metrics

Metrics reference concepts to define their dimensions. The analyzer uses the Concepts lookup map to:

- Validate that referenced concepts exist
- Determine the tensor shape for the metric
- Verify compatibility between input and output metrics

### Aggregators

Aggregators transform data from more granular concepts to less granular ones. The analyzer uses concept relationships to:

- Verify the parent-child relationship exists
- Build the appropriate aggregation matrix
- Determine the resulting tensor shape

### Spreaders

Spreaders distribute data from less granular concepts to more granular ones. The analyzer uses concept relationships to:

- Verify the child-parent relationship exists
- Build the appropriate spreading matrix
- Validate weighting concepts if specified

### Lookups

Lookups map between different concept spaces. The analyzer uses concept definitions to:

- Validate source and target concepts exist
- Verify lookup keys match concept values
- Determine the lookup result dimensions

## Design Benefits

Knowing concept definitions at all phases provides several benefits:

1. **Early Error Detection** - Invalid concept references are caught during analysis, not at runtime
2. **Static Optimization** - Tensor shapes and loop bounds are known, enabling better code generation
3. **Pre-computation** - Transformation matrices can be pre-built during synthesis
4. **Type Safety** - Concept value compatibility can be verified statically
5. **Better Tooling** - IDE support can provide autocomplete for concept values

## Validation Rules

The following validation rules must be enforced during parsing and analysis:

1. **Name Uniqueness**: Concept names must be unique within their scope; duplicate definitions must be extensions, not redefinitions
2. **Tree Consistency**: A concept cannot be placed under a different parent than its original parent
3. **Generator Validity**:
   - Values generators must contain valid concept values
   - Int generators must have integer bounds where start <= end and step must be a positive integer
   - Concept range generators must reference existing concepts with valid start and end values that appear in the concept
4. **Before/After References**: Values specified in `before` and `after` clauses must exist in the target concept
5. **Conditional Validity**: Parent values in conditional specifications must exist in the parent concept
6. **Order Preservation**: Extensions must not change the relative order of existing values

## Future Considerations

This design assumes concept definitions are static within a CPL compilation unit. Future enhancements may include:

- Dynamic concept values loaded from external sources at synthesis time
- Concept versioning for evolving taxonomies
- Concept inheritance and composition patterns
- Cross-namespace concept references

## Related Documents

- [CPL Architecture - Introduction](../../architecture/cpl/introduction.md) - Overall CPL processing lifecycle
- [CPL Requirements - Concepts, Types, Units, Measures](../../requirements/cpl/concepts-types-units-measures-syntax.md) - Syntax requirements
- [CPL Architecture - Intermediate Language](../../architecture/cpl/analysis/intermediate_language.md) - CIL specification
