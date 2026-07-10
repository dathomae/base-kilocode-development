# Story125: CPL Concepts Parser - Before/After

## Goal

Implement before/after clause support in the CPL parser for concept value ordering. This story adds syntax for inserting new values at specific positions when extending concepts, using `before` and `after` keywords. This enables new values to be inserted between existing values rather than only appending. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 6.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md) - Increment 6 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story124_cpl-concepts-analyzer-range.md](Story124_cpl-concepts-analyzer-range.md) - Not Started - Range must be in place

## Dependent Stories

- Story126_cpl-concepts-analyzer-before-after.md - Not Started - Analyzer before/after depends on parser

## Tasks

1. Update ChronoconePlanningLanguageLexer.g4 - Completed
   a. Add BEFORE token - Completed
   b. Add AFTER token - Completed

2. Update CPLConcepts.g4 for before/after - Completed
   a. Add conceptExtension rule - Completed
   b. Add insertionPoint rule - Completed
   c. Modify conceptDefinition to include conceptExtension - Completed

3. Define InsertionPoint struct - Completed
   a. Add BeforeValue field - Completed
   b. Add AfterValue field - Completed

4. Update SymbolDefinition for insertion point data - Completed
   a. Store InsertionPoint in SymbolDefinition.Data - Completed

5. Update ConceptsListener - Completed
   a. Handle conceptExtension parsing - Completed
   b. Parse insertion point (before/after value) - Completed
   c. Create InsertionPoint structure - Completed
   d. Distinguish between initial definition and extension - Completed

6. Implement validation - Completed
   a. Validate at least one of before/after is specified - Completed
   b. Validate before/after values refer to identifiers - Completed

7. Write unit tests for before/after parsing - Completed
   a. Test before clause - Completed
   b. Test after clause - Completed
   c. Test concept extension syntax - Completed
   d. Test error cases - Completed

## Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- At least one of before or after must be specified
- Before/after clauses only apply when adding new values to existing concepts
- Cannot reorder existing values - only insert new ones
- Follow the Logical TDD Lifecycle - write tests before implementation
- Must not break Increment 1-5 functionality

## Intent

The goal of this story is to enable concept extensions with specific insertion points. This allows developers to add new values to existing concepts at specific positions rather than only appending to the end. The initial declaration establishes the canonical order which cannot be changed.

## Acceptance Criteria

1. Parser can successfully parse before/after clauses: `Priority=[Values[Critical]] before Low`
2. InsertionPoint struct is defined with BeforeValue and AfterValue fields
3. SymbolDefinition.Data contains InsertionPoint field for concept extensions
4. ConceptsListener correctly parses concept extensions and stores data in SymbolDefinition.Data
5. Validation ensures at least one of before/after is specified
6. Concept extension syntax is distinguished from initial definition
7. All unit tests pass following BTDD practices
8. Increment 1-5 functionality still works

## Notes

**Increment 6 Capabilities:**

- Before clause: `Concept=[Values[NewValues]] before ExistingValue`
- After clause: `Concept=[Values[NewValues]] after ExistingValue`
- Concept extensions in multiple sections

**Important Rules:**

- Before/after clauses only apply when adding new values to an existing concept
- At least one of `before` or `after` must be specified
- The initial declaration establishes the relative order of all values
- Before/after clauses cannot alter the relative order of original values
- New values inserted between original values preserve the relative order of originals

**Increment 6 Limitations:**

- Cannot reorder existing values - only insert new ones
- No API integration (added in Increment 7)

**Example CPL code that should parse:**

```cpl
concepts
    Priority=[Values[Low, Medium, High]]

concepts
    # Insert Critical before Low
    Priority=[Values[Critical]] before Low

    # Insert Urgent after Medium
    Priority=[Values[Urgent]] after Medium
```

**Grammar Changes:**

```antlr
// Concept extension with before/after
conceptExtension
    : identifier EQ LBRACKET generatorPatternList RBRACKET insertionPoint
    ;

insertionPoint
    : BEFORE identifier                    // Insert before this existing value
    | AFTER identifier                     // Insert after this existing value
    ;
```

**InsertionPoint Structure:**

```go
type InsertionPoint struct {
    BeforeValue string  // Insert before this value (if set)
    AfterValue  string  // Insert after this value (if set)
}
```

**Concept Data Structure (stored in SymbolDefinition.Data):**

```go
// ConceptDefinitionData holds all concept-related data stored in SymbolDefinition.Data
type ConceptDefinitionData struct {
    Name                string
    GeneratorType       GeneratorType
    Values              []string
    RangeBounds         *RangeBounds
    IndentationLevel    int
    InsertionPoint      InsertionPoint   // NEW in Increment 6
    ConditionalSpecs    []ConditionalSpec
    DefaultValues       []string
    SourceLocation      SourceLocation
}
```

The concept data (including InsertionPoint) is stored in the symbol table via SymbolDefinition.Data during parsing, and the analyzer retrieves it to build the complete AST.

**Value Ordering Immutability:**

The initial declaration establishes the canonical order:

```cpl
concepts
    Priority=[Values[Low, Medium, High]]  # Canonical order: Low < Medium < High
```

Extensions insert new values without changing original order:

```cpl
concepts
    Priority=[Values[Critical]] before Low
    # Result: Critical, Low, Medium, High (original order preserved)
```

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
