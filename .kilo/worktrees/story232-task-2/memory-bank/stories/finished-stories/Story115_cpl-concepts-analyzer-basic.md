# Story115: CPL Concepts Analyzer - Basic

## Goal

Implement the basic concepts analysis functionality for the Chronocone Planning Language (CPL) analyzer. This story builds the AnalysisConceptDefinition structure and the Concepts lookup map from parser-level concept definitions, enabling duplicate detection and providing the foundation for concept-based validation in metrics and other CPL elements. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 1.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md) - Increment 1 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story114_cpl-concepts-parser-basic.md](Story114_cpl-concepts-parser-basic.md) - Done - Parser populates symbol table with concept definitions

## Dependent Stories

- Story116_cpl-concepts-synthesizer-basic.md - Not Started - Synthesizer depends on analysis structures
- Story117_cpl-concepts-parser-hierarchy.md - Not Started - Hierarchy builds on basic concepts

## Tasks

1. Define AnalysisConceptDefinition struct - Completed
   a. Add ConceptName field - Completed
   b. Add Generators slice field ([]Generator) - Completed
   c. Add ResolvedValues slice field - Completed
   d. Add index map field - Completed

2. Define AnalysisConceptValueDefinition struct - Completed
   a. Add Name field - Completed
   b. Add Position field - Completed

3. Define Generator interface and ValuesGenerator struct - Completed
   a. Define Generator interface with GenerateValues method - Completed
   b. Add ValuesGenerator struct with Values []string field - Completed
   c. Document GenerateValues method for ValuesGenerator - Completed

4. Define AnalysisDefinitions struct - Completed
   a. Add Concepts map field - Completed

5. Implement WalkSymbolTableWithHandlers function - Completed
   a. Define AnalysisFunc type - Completed
   b. Implement recursive walk with handler dispatch - Completed
   c. Handle nil tables gracefully - Completed

6. Implement CPLAnalyzer struct - Completed
   a. Create analyzer with concepts map - Completed
   b. Add AnalyzeConceptsScope method - Completed
   - Walk symbol table to find concept definitions
   - Extract generator data from symbol table definitions
   - Convert symbol table generators to analysis generators
   - Generate resolved values from each generator
     c. Add GetAnalysisDefinitions method - Completed

7. Integrate with CPLCompiler - Completed
   a. Add Analyze method to compiler - Completed
   b. Register ConceptsScope handler - Completed
   c. Return AnalysisDefinitions from Analyze - Completed

8. Write unit tests for concepts analysis - Completed
   a. Test single concept analysis - Completed
   b. Test multiple concepts analysis - Completed
   c. Test duplicate concept detection - Completed
   d. Test value indexing - Completed
   e. Test multiple generators support - Completed

## Constraints

- Must work with parser-level concept definitions from the symbol table
- Must produce structures usable by the CIL generator
- Must validate all concept references
- Follow the Logical TDD Lifecycle - write tests before implementation
- Analysis errors must prevent CIL generation

## Intent

The goal of this story is to transform parser-level concept definitions into analysis-level structures optimized for validation and code generation. The analyzer builds a Concepts lookup map that will be used to validate metrics, aggregators, spreaders, and lookups that reference concepts.

## Acceptance Criteria

1. AnalysisConceptDefinition struct is defined with ConceptName, Generators slice, ResolvedValues, and index fields
2. Generator interface is defined with GenerateValues method signature
3. ValuesGenerator struct is defined with Values []string field
4. AnalysisConceptValueDefinition struct is defined with Name and Position
5. AnalysisDefinitions struct contains Concepts lookup map
6. WalkSymbolTableWithHandlers function correctly dispatches to scope handlers
7. CPLAnalyzer processes ConceptsScope tables and builds concept definitions
8. CPLAnalyzer correctly processes multiple generators from parser (Generators []Generator)
9. Duplicate concept definitions are detected and reported as errors
10. All unit tests pass following BTDD practices
11. ValueIndex map provides O(1) lookup of concept value positions

## Notes

**Increment 1 Limitations:**

- Flat concept namespace (no hierarchy) - added in Increment 2
- Cannot validate parent-child relationships - added in Increment 2
- No support for tensor shape computation from hierarchical dimensions - added in Increment 2
- No multiple concept declaration sections in the same file - added in Increment 3
- No conditional specifications - added in Increment 4
- No Int Range generators - added in Increment 5
- No before/after clauses for value ordering - added in Increment 7

**Symbol Table Walking Approach:**

The analysis phase uses a single WalkSymbolTableWithHandlers call with a map of scope-specific handlers. This approach:

- Centralizes tree-walking logic in the symboltable package
- Allows different scope types to be processed in a single tree walk
- Makes it easy to add new scope type handlers in future increments

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
- [Intermediate Language](../specifications/architecture/cpl/analysis/intermediate_language.md) - CIL specification
