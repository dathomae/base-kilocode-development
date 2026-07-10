# Story121: CPL Concepts Parser - Conditional

## Goal

Implement conditional concept value specifications in the CPL parser. This story adds support for syntax like `Parent[Value] => Child=[Values[...]]` which defines child concept values that are conditional on specific parent values. This enables parent-specific child value sets. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 4.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md) - Increment 4 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story120_cpl-concepts-analyzer-multi-section.md](Story120_cpl-concepts-analyzer-multi-section.md) - Not Started - Multi-section must be in place

## Dependent Stories

- Story122_cpl-concepts-analyzer-conditional.md - Not Started - Analyzer conditional depends on parser

## Tasks

1. Update ChronoconePlanningLanguageLexer.g4 - Completed
   a. Add ARROW token (=>) - Completed
   b. Verify LBRACKET, RBRACKET exist - Completed

2. Update CPLConcepts.g4 for conditional specs - Completed
   a. Add conditionalSpec rule - Completed
   b. Modify conceptDefinition to include conditionalSpec - Completed
   c. Define conditionalSpec syntax: Parent[Value] => conceptLine - Completed

3. Define ConditionalSpec struct - Completed
   a. Add ParentValue field - Completed
   b. Add GeneratorType field - Completed
   c. Add Values slice field - Completed
   d. Add RangeBounds field (for future Range support) - Completed

4. Update SymbolDefinition for conditional data - Completed
   a. Store ConditionalSpecs in SymbolDefinition.Data - Completed (via ConceptDefinitionData)
   b. Store DefaultValues in SymbolDefinition.Data - Completed (via ConceptDefinitionData)

5. Update ConceptsListener - Completed
   a. Handle conditionalSpec parsing - Completed
   b. Extract parent value from brackets - Completed
   c. Build ConditionalSpec structures - Completed
   d. Handle default (unconditional) specifications - Completed

6. Write unit tests for conditional parsing - Completed
   a. Test simple conditional spec - Completed (19 tests PASS)
   b. Test multiple conditions for same child - Completed
   c. Test default value specification - Completed
   d. Test error cases - Completed

## Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- At most one default specification per concept (no parent value qualifier)
- Conditional specs only valid for child concepts (not root)
- Follow the Logical TDD Lifecycle - write tests before implementation
- Must not break Increment 1-3 functionality

## Intent

The goal of this story is to enable conditional value specifications for child concepts. This allows child concepts to have different values depending on which parent value is selected. For example, the Region concept might have different values for USA vs Canada.

## Acceptance Criteria

1. Parser can successfully parse conditional specifications like `Country[USA] => Region=[Values[North, South, East, West]]`
2. ConditionalSpec struct is defined with all required fields
3. SymbolDefinition.Data contains ConditionalSpecs and DefaultValues for concept definitions
4. ConceptsListener correctly builds conditional specifications and stores them in SymbolDefinition.Data
5. Default (unconditional) specifications are supported
6. Validation ensures at most one default per concept
7. All unit tests pass following BTDD practices
8. Increment 1-3 functionality still works

## Notes

**Increment 4 Capabilities:**

- Conditional value specifications: `Parent[Value] => Child=[Values[...]]`
- Multiple conditions per child concept
- Default (fallback) values when no condition matches

**Increment 4 Limitations:**

- No Range generators in conditions (added in Increment 5)
- No before/after clauses (added in Increment 6)
- Parent value references not validated during parsing (full validation in analysis)
- No API integration (added in Increment 7)

**Example CPL code that should parse:**

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
      Region=[Values[North, South, East, West]]

    Country[USA] => State=[Values[NY, CA, TX]]
    Country[Canada] => Province=[Values[Ontario, Quebec]]

    # Default values when no condition matches
    Product=[Values[Widget, Gadget]]
      Variant=[Values[Standard, Premium]]
```

**ConditionalSpec Structure:**

```go
type ConditionalSpec struct {
    ParentValue      string         // Value of parent concept
    GeneratorType    GeneratorType
    Values           []string
    RangeBounds      *RangeBounds   // For future Range support
}
```

**Concept Data Structure (stored in SymbolDefinition.Data):**

```go
// ConceptDefinitionData holds all concept-related data stored in SymbolDefinition.Data
type ConceptDefinitionData struct {
    Name                string
    GeneratorType       GeneratorType
    Values              []string
    IndentationLevel    int
    ConditionalSpecs    []ConditionalSpec  // NEW in Increment 4
    DefaultValues       []string           // NEW in Increment 4
    SourceLocation      SourceLocation
}
```

The concept data is stored in the symbol table via SymbolDefinition.Data during parsing:

```go
conceptData := ConceptDefinitionData{
    Name:           name,
    GeneratorType:  generatorType,
    Values:         values,
    IndentationLevel: currentTable.IndentationLevel,
    // ConditionalSpecs and DefaultValues populated when parsing conditional specs
}

// Store in symbol table
err = currentTable.Define(&name, ConceptDefinition, location)
if err != nil {
    listener.AddError(err)
}
// The analyzer retrieves SymbolDefinition.Data and works with ConceptDefinitionData
```

**Validation During Parsing:**

- Parent value in condition must be a valid identifier
- At most one default specification (no parent value qualifier)
- Conditional specs are syntactically validated; semantic validation (parent value exists) happens in analysis

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
