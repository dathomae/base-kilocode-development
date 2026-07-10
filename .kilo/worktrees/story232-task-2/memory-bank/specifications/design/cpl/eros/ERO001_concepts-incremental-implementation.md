# ERO001: Concepts Incremental Implementation

## Title

Concepts Incremental Implementation

## Purpose/Goal

Define the incremental approach for implementing concepts functionality across all CPL phases (parsing, analysis, synthesis, evaluation), detailing which subset of features is implemented in each increment and documenting the limitations of each intermediate state.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic001_cpl-concepts-implementation.md](../../../../epics/Epic001_cpl-concepts-implementation.md) - Epic for CPL concepts implementation

## Context

### Situation

The concepts-definitions.md document outlines comprehensive functionality for concept definitions in CPL, including:

- Concept sections with indentation-based scoping
- Tree-structured parent-child relationships
- Generator-based value definitions (Values, Int Range, and Concept Range generators)
- Before/after clauses for extending existing concepts
- Conditional value specifications based on parent values
- Multiple concept definitions across files
- Value ordering immutability

Implementing all of this functionality at once would be a significant undertaking and carries risks of complexity, bugs, and delayed feedback. An incremental approach allows for validated learning and progressive feature delivery.

### Constraints

- Each increment must produce a working system (though with limitations)
- Later increments must not break functionality from earlier increments
- The incremental evolution must maintain consistent data structures across phases
- Concept definitions must remain statically known at compile/analysis/synthesis time
- The approach must align with the existing CPL phase architecture (parse → analyze → synthesize)

### Requirements

- Support basic concept definition and usage in early increments
- Enable progressive enhancement of concept capabilities
- Document limitations at each increment clearly
- Ensure each increment provides value for some use cases
- Maintain backward compatibility within the incremental progression

### Assumptions

- Concept definitions will be primarily static within a compilation unit
- The most common use cases involve simple Values generators
- Before/after clauses and conditional specifications are advanced features used less frequently
- Multiple file concept definitions can be deferred to later increments

## Evaluation Criteria

1. **Risk Reduction** (High importance): Each increment should reduce implementation risk by validating assumptions
2. **Deliverable Value** (High importance): Each increment should provide usable functionality
3. **Complexity Management** (Medium importance): Increments should avoid overwhelming complexity
4. **Consistency** (High importance): Data structures should evolve consistently across all phases

## Options Considered

### Option 1: Big Bang Implementation

**Description:**
Implement all concept functionality in a single large effort.

**Pros:**

- Complete functionality available at once
- No intermediate limitations to document or work around
- Single design effort

**Cons:**

- High risk of bugs and design flaws
- Delayed feedback on core functionality
- Difficult to estimate and plan
- All-or-nothing delivery

### Option 2: Incremental Implementation by Feature

**Description:**
Implement concept features incrementally, adding capabilities across all phases in each increment.

**Pros:**

- Early validation of core concepts
- Reduced risk through iterative delivery
- Opportunity for feedback between increments
- Ability to course-correct based on learnings

**Cons:**

- Requires careful planning of increment boundaries
- Intermediate limitations must be clearly documented
- Some architectural decisions may need to accommodate future features

### Option 3: Incremental Implementation by Phase

**Description:**
Implement complete concept functionality in one phase at a time (e.g., complete parsing first, then analysis, then synthesis).

**Pros:**

- Completes each phase before moving to next
- Clear separation of concerns

**Cons:**

- No end-to-end functionality until final phase complete
- Risk of discovering phase integration issues late
- Less opportunity for holistic feedback

## Decision

**Chosen Option:** Option 2 - Incremental Implementation by Feature

**Rationale:**
The incremental-by-feature approach provides the best balance of risk reduction and early value delivery. By implementing a subset of features across all phases in each increment, we ensure end-to-end functionality is available early, enabling validation and feedback. This approach aligns with agile principles and allows for course correction based on learnings from each increment.

The specific increments are defined as follows:

### Increment 1: Basic Concept Definitions

**Concept Section Form:**

A concept section starts with the keyword `concepts` followed by a block of concept definitions. Each concept declaration has the form:

```cpl
concepts
    <concept name>=[Values[<comma separated list of values>]]
```

Example:

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
    Priority=[Values[Low, Medium, High]]
```

**Features Implemented:**

- Simple Values generators: `<concept name>=[Values[<comma separated list of values>]]`
- Single-level concepts (no parent-child relationships)
- Single concept section per file
- Basic concept reference validation

**Limitations:**

- No Int Range generators
- No Concept Range generators
- No parent-child relationships
- No before/after clauses
- No conditional specifications
- No multiple file definitions
- No concept extension

**Value Provided:**

Basic dimensional concepts can be defined and used in metrics.

### Increment 2: Concept Hierarchy

**Features Added:**

- Parent-child relationships via indentation
- Tree structure validation
- Child concept value sets (non-conditional)
- Hierarchical concept references in metrics

Example:

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
      Region=[Values[North, South]]  # Child of Country (indented)
      City=[Values[NYC, LA]]         # Child of Country (indented)
        Neighborhood=[Values[Manhattan, Brooklyn]]  # Grandchild of Country
```

Examples of hierarchical concepts definition may be found in ../concepts-definitions.md

**Limitations:**

- No multiple concept declaration sections in the same file
- No conditional specifications
- No Int Range generators
- No Concept Range generators
- No before/after clauses

**Value Provided:**

Metrics can use hierarchical dimensions (e.g., Country → Region).

### Increment 3: Multiple Concept Declaration Sections in the Same File

**Features Added:**

- Multiple concept declaration sections within the same file
- Concept extension within the same compilation unit
- Duplicate value detection and prevention across sections
- Value merging and ordering across multiple declarations

Example:

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]

concepts
    Country=[Values[Brazil, Argentina]]  # Extends Country with additional values
```

Examples of hierarchical concepts definition may be found in ../concepts-definitions.md

**Limitations:**

- No Int Range generators
- No Concept Range generators
- No conditional specifications
- No before/after clauses for value ordering
- Multi-file handling is done by the CLI (concatenating files before parsing)

**Value Provided:**

Complex concept definitions can be built up from multiple declarations within a single file, enabling modular concept organization without requiring multi-file support in the parser.

**Note on Multi-File Support:** Multiple file handling is the responsibility of the top-level compiler CLI, which concatenates files together before passing them to the parser. The parser only needs to handle multiple concept declaration sections within the same (potentially concatenated) file.

### Increment 4: Conditional Specifications

**Features Added:**

- Conditional value specifications: `<parent concept name>[<parent value>] => <child concept name>=[Values[<comma separated list of values>]]`
- Value mapping based on parent context
- Default child values when no condition matches

**Example:**

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
      Country[USA] => Region=[Values[NW, SE, MID, SW, NE]]
      Country[Canada] => Region=[Values[QC, ON, BC]]
      Region=[Values[North, South]]  # Default for Mexico
```

The conditional expression before the child concept declaration will always be some concept declaration already made with a list of values after it. When the parent value is one of the declared values, the child concept will be declared with the values after the `=>`. When there is no conditional, the child concept(s) will have the values given when there is no more specific parent value declaration that is operative.

**Limitations:**

- No Int Range generators
- No Concept Range generators
- No before/after clauses for value ordering

**Value Provided:**

Child concept values that vary by parent (e.g., Regions differ by Country).

### Increment 5: Int Range Generators

**Features Added:**

- Int Range generators: `<concept name>=[Int[<start>, <end>, <step>]]`
- Value ordering is maintained from declaration order
- Note: Even without before/after clauses, value order is preserved from initial declaration

**Example:**

```cpl
concepts
    DayNumbers=[Int[1, 31, 1]]  # Generates 1, 2, 3, ..., 31
    FiscalPeriods=[Int[1, 12, 1]]  # Generates 1, 2, 3, ..., 12
```

**Limitations:**

- No Concept Range generators
- No before/after clauses for modifying value order after initial declaration
- No integration with external Concepts API
- Concepts are defined only within CPL files

**Value Provided:**

Numeric concepts (like days of month, fiscal periods, years) can be defined using integer ranges instead of listing all values.

### Increment 6: Concept Range Generators

**Features Added:**

- Concept Range generators: `<concept name>=[<source concept>[<start value>..<end value>]]`
- References a range of values from another concept
- Value ordering follows the referenced concept's order

**Example:**

```cpl
concepts
    Quarter=[Values[Q1, Q2, Q3, Q4]]
    FirstHalf=[Quarter[Q1..Q2]]  # References Q1, Q2 from Quarter
    FullYear=[Quarter[Q1..Q4]]   # References all quarters
```

**Limitations:**

- No before/after clauses for modifying value order after initial declaration
- No integration with external Concepts API
- Concepts are defined only within CPL files

**Value Provided:**

Concepts can be defined by referencing value ranges from other concepts, enabling composition and reuse of existing concept definitions.

### Increment 7: Before/After Clauses for Value Ordering

**Features Added:**

- Before/after clauses for inserting values when extending concepts
- Syntax: `<existing concept name>=[Values[<new values>]] before <existing value>`
- Syntax: `<existing concept name>=[Values[<new values>]] after <existing value>`
- Value ordering preservation during extension

**Important Notes:**

- Before/after clauses only apply when adding new values to an existing concept
- At least one of `before` or `after` must be specified (both may be specified, but not neither)
- The initial declaration establishes the relative order of all values in that declaration
- Before/after clauses cannot alter the relative order of original values from the initial declaration
- New values inserted between original values change the absolute positions but preserve the relative order of the original values

**Example:**

```cpl
concepts
    Priority=[Values[Low, Medium, High]]

concepts
    Priority=[Values[Critical]] before Low
    Priority=[Values[Urgent]] after Medium
```

**Limitations:**

- Cannot reorder existing values - before/after only applies to new values being added
- No integration with external Concepts API
- Concepts are defined only within CPL files

**Value Provided:**

Custom value ordering when extending concepts, allowing new values to be inserted at specific positions rather than only appending.

### Increment 8: Concepts API Integration

**Features Added:**

- **Parsing Phase:** Before parsing begins, concepts are read from the Concepts API and integrated with CPL file definitions
- **Analysis Phase:** API concepts are merged with file-based concepts, with file definitions taking precedence for duplicates
- **Synthesis Phase:** Version checking ensures API concepts match the versions that existed when the synthesized code was created
- **Version Tracking:** Each concept from the API carries a version identifier for consistency checking

**API Integration Flow:**

1. Before parsing: Query Concepts API for relevant concepts
2. During parsing: API concepts are added to the symbol table as pre-populated definitions
3. During analysis: File-based concept definitions can extend or override API concepts
4. During synthesis: Version fingerprints are embedded in generated code for runtime validation

**Limitations:**

None - Full concept functionality including external API integration.

**Value Provided:**

- Concepts can be shared across multiple CPL programs via the Concepts API
- Centralized concept management with versioning
- Runtime consistency checking between compiled code and API concepts
- Enables distributed concept definitions across an organization

## Implications

### Positive Implications

- Reduced implementation risk through iterative delivery
- Early availability of core functionality
- Opportunity for feedback and course correction
- Clear milestones and progress tracking
- Each increment provides usable value

### Negative Implications / Trade-offs

- Requires discipline to maintain backward compatibility between increments
- Some architectural decisions must anticipate future features
- Intermediate limitations may constrain early use cases
- Documentation must clearly track which features are available in each increment
- Testing must cover both current increment features and ensure no regression of prior increments

### Impact on System Qualities

| Quality Attribute | Impact | Notes                                               |
| ----------------- | ------ | --------------------------------------------------- |
| Performance       | =      | No significant impact expected                      |
| Scalability       | =      | Incremental approach doesn't affect scalability     |
| Maintainability   | +      | Smaller increments easier to understand and modify  |
| Security          | =      | No security implications                            |
| Usability         | +      | Early availability of features improves usability   |
| Testability       | +      | Each increment can be thoroughly tested before next |

### Future Considerations

- Each increment should be considered complete before starting the next
- EROs for each phase should document the phase-specific data structures for each increment
- This ERO should be updated as increments are completed
- If significant learnings emerge during implementation, the increment plan may be adjusted

## Notes

**Related Documents:**

- [Concept Definitions Design](../concepts-definitions.md) - Comprehensive design for concept functionality
- [CPL Architecture - Introduction](../../architecture/cpl/introduction.md) - Overall CPL processing lifecycle
- Parsing Phase ERO: [ERO001_concepts-parsing-phase.md](../../go-parser-target/eros/ERO001_concepts-parsing-phase.md) - Phase-specific parsing structures
- Analysis Phase ERO: [ERO001_concepts-analysis-phase.md](../../go-analysis/eros/ERO001_concepts-analysis-phase.md) - Phase-specific analysis structures
- Synthesis Phase ERO: [ERO001_concepts-synthesis-phase.md](../../go-synthesis-target/eros/ERO001_concepts-synthesis-phase.md) - Phase-specific synthesis structures

**Phase-Specific EROs:**

Each phase (parsing, analysis, synthesis) has its own ERO documenting the data structures and functionality specific to that phase across all increments. These EROs should be consulted for detailed phase-specific designs.
