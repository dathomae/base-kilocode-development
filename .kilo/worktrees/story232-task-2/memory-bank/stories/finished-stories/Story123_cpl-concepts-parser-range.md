# Story123: CPL Concepts Parser - Range

## Goal

Implement Int range generator support in the CPL parser. This story adds the Int generator as an alternative to Values for defining numeric concept values using syntax like `ConceptName=[Int[Start, End, Step]]`. This is useful for concepts like months (1-12), fiscal periods, years, etc. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 5.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md) - Increment 5 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- None - This story builds on the existing basic concept parser infrastructure (Values generator, symbol table scopes).

## Dependent Stories

- Story124_cpl-concepts-analyzer-range.md - Not Started - Analyzer range depends on parser
- Story121_cpl-concepts-parser-conditional.md - Not Started - Conditional specs may reference Int range generators once both are implemented

## Tasks

> **story-implementor Handoff Directive**: This is a coding task story. Per [`scripts/story-implementor.pdd.script.md`](../../../scripts/story-implementor.pdd.script.md), the story-implementor MUST:
>
> - Hand off to the **Code** agent via a **NEW AGENT TASK**, passing `story_context` (`story_file`, `current_task`, `activity_type: code`) and the current Logical TDD Lifecycle phase (A–D).
> - Ensure the Code agent returns results via a **NEW AGENT TASK** with the `story_context` section unchanged.
> - After each task completes, hand off to the **code-reviewer** via a **NEW AGENT TASK** with unchanged `story_context`.
> - Receive review results back via a **NEW AGENT TASK** with unchanged `story_context`.
> - **Checkpoint and WAIT for explicit user approval** before proceeding to the next task.

1. Update ChronoconePlanningLanguageLexer.g4 - Completed
   a. Add INT keyword token - Completed
   b. INTEGER token already exists - Completed
   c. COMMA token already exists - Completed

2. Update CPLConcepts.g4 for Int range generator - Completed
   a. Add intRangeGenerator rule - Completed
   b. Add intRangeParams rule - Completed
   c. Modify generator rule to include intRangeGenerator - Completed

3. Define IntRangeGeneratorDef struct - Completed
   a. Add Start field (int) - Completed
   b. Add End field (int) - Completed
   c. Add Step field (int, defaults to 1) - Completed

4. Enhance GeneratorType enumeration - Completed
   a. Add IntRangeGenerator constant - Completed

5. Update parser to capture Int range data - Completed
   a. Add IntGeneratorScope symbol table scope type - Completed
   b. Store Int range parameters in symbol table during parsing - Completed

6. Update ConceptsListener - Completed
   a. Handle intRangeGenerator parsing - Completed
   b. Parse integer parameters from intRangeParams - Completed
   c. Create IntRangeGeneratorDef structure - Completed

7. Implement validation - Completed
   a. Validate range bounds are integers - Completed
   b. Validate start < end when step > 0, start > end when step < 0 - Completed
   c. Validate step != 0 - Completed
   d. Validate non-negative bounds - Completed (enforced by INTEGER lexer token)

8. Write unit tests for Int range parsing - Completed
   a. Test basic Int range generator - Completed
   b. Test Int range with custom step - Completed
   c. Test invalid range bounds detection - Completed
   d. Test error cases - Completed

## Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- Range bounds must be non-negative integers with start < end (when step > 0)
- Step must be non-zero; defaults to 1 if not specified
- Must support Int range in direct concept definitions only (conditional specs not yet implemented)
- Follow the Logical TDD Lifecycle - write tests before implementation
- Must not break existing basic concept functionality (Values generator, multi-section, hierarchy)

## Intent

The goal of this story is to enable Int range generators for defining numeric concept value ranges. This provides a more concise way to declare numeric concepts compared to listing all values explicitly with Values, and supports custom step values for generating sequences like even numbers.

## Acceptance Criteria

1. Parser can successfully parse Int range generator syntax: `Month=[Int[1, 12, 1]]`
2. IntRangeGeneratorDef struct is defined with Start, End, and Step fields
3. GeneratorType includes IntRangeGenerator constant
4. Parser captures Int range parameters in the symbol table for analysis phase consumption
5. Int range bounds validation ensures start < end (when step > 0), step != 0, and non-negative bounds
6. Step defaults to 1 when not specified
7. All unit tests pass following BTDD practices
8. Existing basic concept functionality still works

## Notes

**Increment 5 Capabilities:**

- Int range generator: `Concept=[Int[Start, End, Step]]`
- Int range bounds validation with step support

**Increment 5 Limitations:**

- No conditional specifications (added in Increment 4 / Story121)
- No before/after clauses (added in Increment 7 / Story125)
- No API integration (added in Increment 8)

**Example CPL code that should parse:**

```cpl
concepts
    # Int range generator for months
    Month=[Int[1, 12, 1]]

    # Int range generator for fiscal periods
    FiscalPeriod=[Int[1, 13, 1]]

    # Int range with custom step (even numbers)
    EvenNumbers=[Int[2, 10, 2]]

    # Multiple generators combined
    Mixed=[Values[A, B], Int[1, 3, 1]]
```

**Enhanced GeneratorType:**

```go
type GeneratorType int

const (
    ValuesGenerator GeneratorType = iota
    IntRangeGenerator              // NEW in Increment 5
)
```

**Symbol Table Storage for Int Range Generators:**

The parser stores Int range generator parameters as child scopes under a `ConceptScope`, following the existing pattern for `ValuesPatternGenerator`. The analyzer walks the symbol table to build the complete AST.

Example symbol table structure for a concept with an Int range generator:

```
ConceptsScope [concepts section]
└── ConceptScope [Month]
    └── IntGeneratorScope [Int range params]
        ├── SymbolDefinition [1] (Start)
        ├── SymbolDefinition [12] (End)
        └── SymbolDefinition [1] (Step)
```

**IntRangeGeneratorDef Structure:**

```go
type IntRangeGeneratorDef struct {
    Start int
    End   int
    Step  int  // Defaults to 1 if not specified
}
```

The `IntRangeGeneratorDef` is created by the parser and passed to the analyzer for AST construction.

**Grammar Changes:**

```antlr
// Int range generator: ConceptName=[Int[Start, End, Step]]
intRangeGenerator
    : INT LBRACKET intRangeParams RBRACKET
    ;

intRangeParams
    : INTEGER COMMA INTEGER (COMMA INTEGER)?
    ;
```

**Validation:**

- Int range bounds must be integers
- Int range bounds must satisfy start < end (when step > 0)
- Step must be non-zero
- Range bounds must be non-negative
- Error messages include source location

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
