# Epic004: CPL Connectors Implementation

## Title

CPL Connectors Implementation

## Ultimate Goal

Implement the complete connectors functionality for the compiler for the Chronocone Planning Language (CPL) across all processing phases. Connectors define the shape of input data coming into the Chronocone system. The compiler operation is broken into three phases: parsing, analysis, and synthesis.

## References

- [Connectors Definitions Design](../specifications/design/cpl/connectors.md)
- [CPL Architecture - Introduction](../specifications/architecture/cpl/introduction.md)
- [ERO004: Connectors Incremental Implementation](../specifications/design/cpl/eros/ERO004_connectors-incremental-implementation.md)
- [ERO004: Connectors Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO004_connectors-parsing-phase.md)
- [ERO004: Connectors Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md)
- [ERO004: Connectors Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO004_connectors-synthesis-phase.md)

## Dependencies

- [Epic001_cpl-concepts-implementation.md](Epic001_cpl-concepts-implementation.md) - Concepts implementation must be complete as connectors reference concepts for dimensions. Specifically, Story119, Story120, and Story155 (multiple sections and files) are prerequisites for Increment 2.
- [Epic002_cpl-types-implementation.md](Epic002_cpl-types-implementation.md) - Types implementation must be complete as connectors reference types (Instant, Interval) for time handling in Increment 2.
- [Epic003_units-and-measures-implementation.md](Epic003_units-and-measures-implementation.md) - Units and measures implementation must be complete as connectors reference measures for celltype and grain.

## Stories

| Story File                                                                                                       | Status      | Description                                                                                                                                                                  | ERO                                                                                                                                                                                                                                                   |
| ---------------------------------------------------------------------------------------------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Story162_cpl-parser-connectors-basic.md](../stories/finished-stories/Story162_cpl-parser-connectors-basic.md)   | Done        | **Parser (Syntactic)**: Grammar changes for connectors section, basic connector declarations (dimensions, celltype).                                                         | [ERO004_connectors-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO004_connectors-parsing-phase.md)                                                                                                                           |
| [Story156_cpl-analyzer-connectors.md](../stories/finished-stories/Story156_cpl-analyzer-connectors.md)           | Done        | **Analyzer (Semantic)**: Validation that referenced dimensions (concepts) and celltype (measure) exist in the symbol table.                                                  | [ERO004_connectors-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md)                                                                                                                              |
| [Story157_cpl-synthesizer-connectors.md](../stories/Story157_cpl-synthesizer-connectors.md)                      | Not Started | **Synthesizer**: Generation of Go code that uses the `go-arrow` library to read data from a parquet file.                                                                    | [ERO004_connectors-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO004_connectors-synthesis-phase.md)                                                                                                                    |
| [Story161_parquet-test-utility.md](../stories/Story161_parquet-test-utility.md)                                  | Not Started | **Test Utility**: Create a test utility function and CLI to generate parquet files with random cell values based on dimension definitions.                                   | [ERO004_connectors-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO004_connectors-synthesis-phase.md)                                                                                                                    |
| [Story158_cpl-parser-connectors-time.md](../stories/finished-stories/Story158_cpl-parser-connectors-time.md)     | Done        | **Parser (Syntactic)**: Support for the mandatory `grain` keyword in connector declarations.                                                                                 | [ERO004_connectors-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO004_connectors-parsing-phase.md)                                                                                                                           |
| [Story159_cpl-analyzer-connectors-time.md](../stories/finished-stories/Story159_cpl-analyzer-connectors-time.md) | Done        | **Analyzer (Semantic)**: Validation that the mandatory grain refers to an existing measure, and that the measure's underlying type is `Interval` or derived from `Interval`. | [ERO004_connectors-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md)                                                                                                                              |
| [Story160_cpl-synthesizer-connectors-time.md](../stories/Story160_cpl-synthesizer-connectors-time.md)            | Not Started | **Synthesizer**: Generation of Go code that maps the parquet time column to the connector's grain.                                                                           | [ERO004_connectors-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO004_connectors-synthesis-phase.md)                                                                                                                    |
| [Story173_cpl-mandatory-connector-grain.md](../stories/Story173_cpl-mandatory-connector-grain.md)                | In Progress | Enforces mandatory connector `grain` across grammar, parser/listener behavior, analysis validation, examples, compiler tests, and documentation.                             | [ERO004_connectors-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO004_connectors-parsing-phase.md), [ERO004_connectors-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO004_connectors-analysis-phase.md) |

## Story Ordering and Dependencies

### Phase 1: Basic Connector Declaration and Parquet Reading

_Prerequisites: Epic001 (Concepts Implementation), Epic003 (Units/Measures)_

1. **Story162 - Parser - Connectors Basic**: Grammar for connectors section, basic connector declarations
2. **Story156 - Analyzer - Connectors**: Validation of dimensions and celltype
3. **Story161 - Parquet Test Utility**: Generate parquet files for testing
4. **Story157 - Synthesizer - Connectors**: Generation of Go code to read parquet files

### Phase 2: Time Handling

_Prerequisites: Story162, Story156, Story157 (Phase 1: Basic Connector Declaration), Story119, Story120, Story155 (Multiple sections and files), Epic002 (Instant/Interval types)_

4. **Story158 - Parser - Connectors Time**: Support for the mandatory `grain` keyword
5. **Story159 - Analyzer - Connectors Time**: Validation that the mandatory grain measure is an Interval type
6. **Story160 - Synthesizer - Connectors Time**: Mapping parquet time column to grain

## Constraints

1. Each increment must produce a working system with documented limitations
2. Later increments must not break functionality from earlier increments
3. All phases must align with the existing CPL phase architecture (parse → analyze → synthesize)
4. Follow the Logical TDD Lifecycle for all coding stories
5. Parser changes require corresponding ANTLR grammar updates
6. Analysis phase changes must maintain symbol table walking patterns
7. Connector declarations must include an explicit `grain` clause with one or more ordered grain measure names. All grain measures must be based on `Interval` or derived from `Interval`. Declaration order is preserved from parsing through analysis. Single-grain declarations remain supported as single-element grain lists.
8. IO output specifications represent grain as an ordered array of grain names.

## Acceptance Criteria

1. All parser stories implement the grammar features defined in ERO004 parsing phase
2. All analyzer stories implement the validation and structure building defined in ERO004 analysis phase
3. All synthesizer stories generate appropriate Go code for reading parquet files
4. Each increment provides usable functionality with clearly documented limitations
5. All stories include comprehensive unit tests following BTDD practices
6. Documentation is updated to reflect which features are available at each increment
7. IO specifications represent grain as an ordered array of grain names.
8. Full synthesis output generation for multiple grain dimensions is deferred to Story160; generated output-writing code may provide minimal or partial multi-grain handling.
