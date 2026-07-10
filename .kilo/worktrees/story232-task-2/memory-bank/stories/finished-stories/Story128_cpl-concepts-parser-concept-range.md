# Story128: CPL Concepts Parser - Concept Range

## Goal

Implement Concept Range generator support in the CPL parser. This story adds the Concept Range generator as an alternative to Values and Int Range for defining concept values by referencing a range of values from another concept using syntax like `FirstHalf=[Quarter[Q1..Q2]]`. This enables composition and reuse of existing concept definitions. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 6.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md) - Increment 6 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story123_cpl-concepts-parser-range.md](Story123_cpl-concepts-parser-range.md) - Not Started - Int Range parser must be complete

## Dependent Stories

- Story129_cpl-concepts-analyzer-concept-range.md - Not Started - Concept Range analyzer depends on parser

## Tasks

1. Update ChronoconePlanningLanguageLexer.g4 - Completed
   a. Add DOTDOT token (..) - Completed

2. Update CPLConcepts.g4 for Concept Range generator - Completed
   a. Add conceptPatternGenerator rule - Completed
   b. Add conceptRange rule - Completed
   c. Modify generatorPattern rule to include conceptPatternGenerator - Completed

3. Define ConceptRangeGeneratorDef struct - Completed
   a. Add ConceptName field (string) - Completed
   b. Add StartValue field (string) - Completed
   c. Add EndValue field (string) - Completed

4. Enhance GeneratorType enumeration - Completed
   a. Add ConceptRangeGenerator constant - Completed

5. Update SymbolDefinition for Concept Range data - Completed
   a. Store ConceptRangeGeneratorDef in SymbolDefinition.Data - Completed

6. Update ConceptsListener - Completed
   a. Handle conceptPatternGenerator parsing - Completed
   b. Extract concept name, start value, and end value - Completed
   c. Create ConceptRangeGeneratorDef structure - Completed

7. Implement validation - Completed
   a. Validate concept name is a valid identifier - Completed
   b. Validate start and end values are valid identifiers - Completed
   c. Report errors with source locations - Completed

8. Write unit tests for Concept Range parsing - Completed
   a. Test basic Concept Range generator - Completed
   b. Test Concept Range with Values and Int Range generators combined - Completed
   c. Test Concept Range in conditional specs - Completed (Skipped: depends on Story121)
   d. Test invalid concept name detection - Completed
   e. Test error cases - Completed

## Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- Concept Range references another concept by name; the referenced concept must exist (validated in analysis phase)
- Start and end values are identifiers that must exist in the referenced concept (validated in analysis phase)
- Must support Concept Range in both direct definitions and conditional specs
- Follow the Logical TDD Lifecycle - write tests before implementation
- Must not break Increment 1-5 functionality

## Intent

The goal of this story is to enable Concept Range generators for referencing value ranges from other concepts. This provides a way to compose concepts by extracting subsets of values from existing concepts, promoting reuse and reducing duplication. For example, a "FirstHalf" concept can reference the first two quarters from a "Quarter" concept.

## Acceptance Criteria

1. Parser can successfully parse Concept Range generator syntax: `ConceptName=[SourceConcept[StartValue..EndValue]]`
2. ConceptRangeGeneratorDef struct is defined with ConceptName, StartValue, and EndValue fields
3. GeneratorType includes ConceptRangeGenerator constant
4. SymbolDefinition.Data contains Generators field that can hold ConceptRangeGeneratorDef
5. ConceptsListener correctly builds Concept Range specifications and stores in SymbolDefinition.Data
6. Concept Range generator works in conditional specifications
7. All unit tests pass following BTDD practices
8. Increment 1-5 functionality still works

## Notes

**Increment 6 Capabilities:**

- Concept Range generator: `Concept=[SourceConcept[StartValue..EndValue]]`
- Concept Range in conditional specs: `Parent[Value] => Child=[SourceConcept[Start..End]]`
- Combined generators: `Concept=[Values[A, B], Int[1, 3, 1], Quarter[Q1..Q2]]`

**Increment 6 Limitations:**

- No before/after clauses (added in Increment 7)
- No API integration (added in Increment 8)
- Referenced concept existence validated in analysis phase, not parsing

**Example CPL code that should parse:**

```cpl
concepts
    # Values generator for quarters
    Quarter=[Values[Q1, Q2, Q3, Q4]]

    # Concept range generator - references Q1 and Q2 from Quarter
    FirstHalf=[Quarter[Q1..Q2]]

    # Concept range generator - references Q3 and Q4 from Quarter
    SecondHalf=[Quarter[Q3..Q4]]

    # Combined generators
    MixedConcept=[Values[A, B], Int[1, 3, 1], Quarter[Q1..Q2]]

    # Concept range in conditional spec
    FiscalYear=[Values[FY2023, FY2024]]
      ReportingPeriod=[Quarter[Q1..Q4]]

    FiscalYear[FY2023] => SpecialPeriod=[Quarter[Q1..Q2]]
```

**Enhanced GeneratorType:**

```go
type GeneratorType int

const (
    ValuesGenerator GeneratorType = iota
    IntRangeGenerator
    ConceptRangeGenerator      // NEW in Increment 6
)
```

**ConceptRangeGeneratorDef Structure:**

```go
// ConceptRangeGeneratorDef represents Concept[StartValue..EndValue] in the AST.
// This is a data structure that captures the parsed syntax only.
// Value resolution happens during analysis phase after all concepts are defined.
type ConceptRangeGeneratorDef struct {
    ConceptName string   // Name of the source concept to reference
    StartValue  string   // Start value (inclusive)
    EndValue    string   // End value (inclusive)
}
```

**Concept Data Structure (stored in SymbolDefinition.Data):**

```go
// ConceptDefinitionData holds all concept-related data stored in SymbolDefinition.Data
type ConceptDefinitionData struct {
    Name                string
    Generators          []interface{}    // Can contain ValuesGeneratorDef, IntRangeGeneratorDef, or ConceptRangeGeneratorDef
    IndentationLevel    int
    ConditionalSpecs    []ConditionalSpec
    SourceLocation      SourceLocation
}
```

The concept data (including ConceptRangeGeneratorDef in the Generators slice) is stored in the symbol table via SymbolDefinition.Data during parsing, and the analyzer retrieves it to build the complete AST.

**Grammar Changes:**

```antlr
// Concept range generator: ConceptName=[SourceConcept[StartValue..EndValue]]
conceptPatternGenerator
    : identifier LBRACKET conceptRange RBRACKET
    ;

// Concept range: StartValue..EndValue
conceptRange
    : identifier DOTDOT identifier
    ;

// Generator pattern includes Values, Int, and Concept Range
generatorPattern
    : valuesPatternGenerator
    | intPatternGenerator
    | conceptPatternGenerator    // NEW in Increment 6
    ;
```

**Lexer Token:**

```antlr
// NEW: Concept Range generator token
DOTDOT           : '..';
```

**Validation:**

- Concept name must be a valid identifier
- Start and end values must be valid identifiers
- Full validation of concept range references (concept existence, value existence) happens in analysis phase
- Error messages include source location

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
