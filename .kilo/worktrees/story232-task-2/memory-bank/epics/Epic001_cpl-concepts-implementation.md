# Epic001: CPL Concepts Implementation

## Title

CPL Concepts Implementation

## Ultimate Goal

Implement the complete concepts functionality for the compiler for the Chronocone Planning Language (CPL) across all processing phases. Concepts define the dimensional structure of metrics and provide the values that index time series data. This epic delivers incremental feature implementation from basic Values generators through full API integration with version checking. The compiler operation is broken into three phases: parsing, analysis, and synthesis.

## References

- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md) - Comprehensive design for concept functionality
- [CPL Architecture - Introduction](../specifications/architecture/cpl/introduction.md) - Overall CPL processing lifecycle
- [ERO001: Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md) - Phase-specific parsing structures
- [ERO001: Concepts Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md) - Phase-specific analysis structures
- [ERO001: Concepts Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO001_concepts-synthesis-phase.md) - Phase-specific synthesis structures

## Dependencies

None - This is a foundational epic for CPL concept functionality.

## Stories

| Story                                                                                                         | Status      | Description                                                                                       | ERO                                                                                                                            |
| ------------------------------------------------------------------------------------------------------------- | ----------- | ------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| [Story114_cpl-concepts-parser-basic.md](../stories/Story114_cpl-concepts-parser-basic.md)                     | Not Started | Grammar changes for concepts section, Values generator support, ParserConceptDefinition structure | [ERO001_concepts-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)        |
| [Story115_cpl-concepts-analyzer-basic.md](../stories/Story115_cpl-concepts-analyzer-basic.md)                 | Not Started | Basic AnalysisConceptDefinition, Concepts lookup map, duplicate detection                         | [ERO001_concepts-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md)           |
| [Story116_cpl-concepts-synthesizer-basic.md](../stories/Story116_cpl-concepts-synthesizer-basic.md)           | Not Started | Placeholder Go CLI generation                                                                     | [ERO001_concepts-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO001_concepts-synthesis-phase.md) |
| [Story119_cpl-concepts-parser-multi-section.md](../stories/Story119_cpl-concepts-parser-multi-section.md)     | Not Started | Support for multiple concepts sections in the same file                                           | [ERO001_concepts-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)        |
| [Story120_cpl-concepts-analyzer-multi-section.md](../stories/Story120_cpl-concepts-analyzer-multi-section.md) | Not Started | Cross-section duplicate detection                                                                 | [ERO001_concepts-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md)           |
| [Story155_cpl-parser-multiple-files.md](../stories/Story155_cpl-parser-multiple-files.md)                     | Not Started | Support for parsing multiple CPL files as a single compilation unit                               | [ERO001_concepts-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)        |
| [Story117_cpl-concepts-parser-hierarchy.md](../stories/Story117_cpl-concepts-parser-hierarchy.md)             | Not Started | Indentation handling with INDENT/DEDENT tokens, parent-child relationship parsing                 | [ERO001_concepts-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)        |
| [Story118_cpl-concepts-analyzer-hierarchy.md](../stories/Story118_cpl-concepts-analyzer-hierarchy.md)         | Not Started | Parent-child relationship building, ConceptHierarchy structure, tree validation                   | [ERO001_concepts-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md)           |
| [Story121_cpl-concepts-parser-conditional.md](../stories/Story121_cpl-concepts-parser-conditional.md)         | Not Started | Conditional specification grammar (Parent[Value] => Child)                                        | [ERO001_concepts-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)        |
| [Story122_cpl-concepts-analyzer-conditional.md](../stories/Story122_cpl-concepts-analyzer-conditional.md)     | Not Started | Conditional value resolution, parent-specific child values                                        | [ERO001_concepts-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md)           |
| [Story123_cpl-concepts-parser-range.md](../stories/Story123_cpl-concepts-parser-range.md)                     | Not Started | Int Range generator grammar (Int[start, end, step])                                               | [ERO001_concepts-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)        |
| [Story124_cpl-concepts-analyzer-range.md](../stories/Story124_cpl-concepts-analyzer-range.md)                 | Not Started | Int Range value expansion                                                                         | [ERO001_concepts-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md)           |
| [Story128_cpl-concepts-parser-concept-range.md](../stories/Story128_cpl-concepts-parser-concept-range.md)     | Not Started | Concept Range generator grammar (SourceConcept[StartValue..EndValue])                             | [ERO001_concepts-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)        |
| [Story129_cpl-concepts-analyzer-concept-range.md](../stories/Story129_cpl-concepts-analyzer-concept-range.md) | Not Started | Concept Range value resolution from referenced concepts                                           | [ERO001_concepts-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md)           |
| [Story125_cpl-concepts-parser-before-after.md](../stories/Story125_cpl-concepts-parser-before-after.md)       | Not Started | Before/after clause grammar for value ordering                                                    | [ERO001_concepts-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)        |
| [Story126_cpl-concepts-analyzer-before-after.md](../stories/Story126_cpl-concepts-analyzer-before-after.md)   | Not Started | Value ordering computation with insertion points                                                  | [ERO001_concepts-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md)           |

**Note on Increment 8 (API Integration):** Stories for Increment 8 (API concept loading, API concept integration with version extraction, and synthesizer version checking) will be created after the concept read API is defined. This integration requires the external API specification to be finalized first.

## Story Ordering and Dependencies

Stories in this epic follow a strict incremental sequence based on the ERO001 concepts incremental implementation plan:

### Phase 1: Basic Concepts (Must be completed first)

_Prerequisites: None_

1. **Story114 - Parser - Basic**: Grammar and Values generator support (Increment 1)
2. **Story115 - Analyzer - Basic**: Concepts lookup map and duplicate detection (Increment 1)
3. **Story116 - Synthesizer - Basic**: Placeholder CLI generation (Increment 1)

### Phase 2: Multiple Sections and Files

_Prerequisites: Story114, Story115, Story116 (Phase 1: Basic Concepts)_

4. **Story119 - Parser - Multi-Section**: Multiple concepts sections support (Increment 2)
5. **Story120 - Analyzer - Multi-Section**: Cross-section duplicate detection (Increment 2)
6. **Story155 - Parser - Multiple Files**: Support for parsing multiple CPL files (Increment 2)

### Phase 3: Hierarchy Support

_Prerequisites: Story119, Story120, Story155 (Phase 2: Multiple Sections and Files)_

7. **Story117 - Parser - Hierarchy**: Indentation handling and INDENT/DEDENT tokens (Increment 3)
8. **Story118 - Analyzer - Hierarchy**: Parent-child relationships and tree validation (Increment 3)

### Phase 4: Conditional Specifications

_Prerequisites: Story117, Story118 (Phase 3: Hierarchy Support)_

8. **Story121 - Parser - Conditional**: Conditional value specifications (Increment 4)
9. **Story122 - Analyzer - Conditional**: Parent-specific child value resolution (Increment 4)

### Phase 5: Int Range Generators

_Prerequisites: Story121, Story122 (Phase 4: Conditional Specifications)_

10. **Story123 - Parser - Int Range**: Int Range generator support (Increment 5)
11. **Story124 - Analyzer - Int Range**: Int Range value expansion (Increment 5)

### Phase 6: Concept Range Generators

_Prerequisites: Story123, Story124 (Phase 5: Int Range Generators)_

12. **Story128 - Parser - Concept Range**: Concept Range generator support (Increment 6)
13. **Story129 - Analyzer - Concept Range**: Concept Range value resolution from referenced concepts (Increment 6)

### Phase 7: Value Ordering

_Prerequisites: Story128, Story129 (Phase 6: Concept Range Generators)_

14. **Story125 - Parser - Before/After**: Before/after clause grammar (Increment 7)
15. **Story126 - Analyzer - Before/After**: Value ordering computation (Increment 7)

### Phase 8: API Integration (Final phase)

_Prerequisites: Story125, Story126 (Phase 7: Value Ordering)_

**Note:** Stories for Phase 8 will be created after the concept read API is defined. This phase includes:

- **Parser - API Integration**: API concept loading (Increment 8)
- **Analyzer - API Integration**: API concept integration and version extraction (Increment 8)
- **Synthesizer - API Integration**: Version checking in generated CLI (Increment 8)

Each phase builds upon the previous phases. Within each increment, parser stories generally precede analyzer stories.

## Generators

Concept values are defined using generators that specify how to produce the set of valid values for a concept. The following generators are supported:

### Values Generator

The Values generator explicitly lists all values for a concept:

```cpl
Country=[Values[USA, Canada, Mexico]]
Priority=[Values[Low, Medium, High]]
```

### Int Range Generator

The Int Range generator produces integer values from a start to end value with a specified step:

```cpl
DayNumbers=[Int[1, 31, 1]]
EvenNumbers=[Int[2, 10, 2]]
```

### Concept Range Generator

The Concept Range generator extracts a range of values from another concept:

```cpl
Quarter=[Values[Q1, Q2, Q3, Q4]]
FirstHalf=[Quarter[Q1..Q2]]
SecondHalf=[Quarter[Q3..Q4]]
```

### Combined Generators

Multiple generators can be combined to define a concept's values:

```cpl
MixedConcept=[Values[A, B], Int[1, 3, 1], Quarter[Q1..Q2]]
```

### Before/After Clauses

When extending concepts across multiple sections, before/after clauses specify where to insert new values:

```cpl
concepts
    Priority=[Values[Low, Medium, High]]

concepts
    Priority=[Values[Critical]] before Low
    Priority=[Values[Urgent]] after Medium
```

## Constraints

1. Each increment must produce a working system with documented limitations
2. Later increments must not break functionality from earlier increments
3. Concept definitions must remain statically known at compile/analysis/synthesis time
4. All phases must align with the existing CPL phase architecture (parse → analyze → synthesize)
5. Follow the Logical TDD Lifecycle for all coding stories
6. Parser changes require corresponding ANTLR grammar updates
7. Analysis phase changes must maintain symbol table walking patterns. Note that 'analysis' in this context refers to the CPL analysis phase, not work planning analysis or question analysis. If there is any lack of clarity of what 'analysis' means in any particular context ask the user to clarify the request before proceeding with any action.

## Acceptance Criteria

1. All parser stories implement the grammar features defined in ERO001 parsing phase
2. All analyzer stories implement the validation and structure building defined in ERO001 analysis phase
3. Synthesizer stories generate appropriate CLI code (placeholder or version-checking as specified)
4. Each increment provides usable functionality with clearly documented limitations
5. Concept definitions are validated at analysis time, not deferred to runtime
6. Version checking detects concept definition changes that would invalidate generated code (Increment 8)
7. All stories include comprehensive unit tests following BTDD practices
8. Documentation is updated to reflect which features are available at each increment

## Notes

**Implementation Details:**

- Multi-file concept definitions are handled by the compiler CLI concatenating files before parsing, not by the parser directly
- Before/after clauses are designed to only insert new values, not reorder existing values (preserves value ordering immutability)
- The synthesizer does not generate concept metadata code for runtime inspection; concept values are used at analysis time to generate CIL (see [Calculation/Operations Framework In Chronocone](../specifications/architecture/cpl/analysis/intermediate_language.md) for details about the CIL intermediate representation. This knowlege will only be required for the analyzer and synthesizer stories)
- CLI version check failures result in an error message and exit

**Related Epics:**

- Future epics for metrics, aggregators, and spreaders will build on the concept infrastructure defined here
