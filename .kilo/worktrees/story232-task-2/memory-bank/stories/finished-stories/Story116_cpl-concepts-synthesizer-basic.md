# Story116: CPL Concepts Synthesizer - Basic

## Goal

Implement the basic synthesis phase for the Chronocone Planning Language (CPL) concepts feature. This story generates a placeholder Go CLI executable that demonstrates the concepts have been successfully analyzed. The synthesizer does not generate CIL elements for concepts themselves (concepts are used during analysis), but produces a working CLI as proof of the compilation pipeline. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 1.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO001_concepts-synthesis-phase.md) - Increment 1 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story115_cpl-concepts-analyzer-basic.md](Story115_cpl-concepts-analyzer-basic.md) - Done - Analyzer must produce AnalysisDefinitions

## Dependent Stories

- Story117_cpl-concepts-parser-hierarchy.md - Not Started - Hierarchy builds on basic synthesis

## Tasks

1. Add command line option for output file path - Completed
   a. Define `-o`/`--output` flag in CLI arguments - Completed
   b. Set default output path if not specified - Completed
   c. Validate output file path is writable - Completed

2. Modify CPLCompiler to orchestrate full pipeline - Completed
   a. Restructure Compile method to call Parse phase - Completed
   b. Add call to Analyze phase after Parse - Completed
   c. Add call to Synthesize phase after Analyze - Completed
   d. Handle errors from each phase appropriately - Completed

3. Implement Synthesize function - Completed
   a. Create output file - Completed
   b. Generate Go CLI boilerplate - Completed
   c. Handle file I/O errors - Completed

4. Implement generateGoCLI function - Completed
   a. Write package declaration - Completed
   b. Write imports (fmt, os) - Completed
   c. Write main function with placeholder - Completed
   d. Write usage message - Completed

5. Integrate Synthesize with CPLCompiler - Completed
   a. Add Synthesize method to compiler - Completed
   b. Pass output file path from CLI arguments - Completed
   c. Pass AnalysisDefinitions to Synthesize - Completed

6. Write unit tests for synthesis - Completed
   a. Test output file creation - Completed
   b. Test generated Go code is valid - Completed
   c. Test error handling - Completed

7. Write integration test - Completed
   a. Test full parse-analyze-synthesize pipeline - Completed
   b. Verify generated CLI can be compiled - Completed
   c. Test CLI argument parsing for output path - Completed
   d. Test pipeline error handling and propagation - Completed
   e. Test default output path behavior - Completed

8. Update CLI documentation - Completed
   a. Document `-o`/`--output` flag in help text - Completed
   b. Add usage examples showing output path option - Completed
   c. Document the full compilation pipeline phases - Completed

## Pipeline Phases

The CPL compiler executes the following phases in sequence:

1. **Parse Phase**: Reads CPL source file and populates symbol table
2. **Analyze Phase**: Walks symbol table to build AnalysisDefinitions
3. **Synthesize Phase**: Generates Go CLI executable from analysis results

Each phase depends on the successful completion of the previous phase. If any phase fails, compilation stops and errors are reported.

## Constraints

- Concept definitions do not generate CIL elements
- Synthesized code knows concept names and reads concept values
- Synthesized code has no knowledge of CIL structures
- Follow the Logical TDD Lifecycle - write tests before implementation
- Generated code must be valid Go that compiles

## Intent

The goal of this story is to complete the compilation pipeline by generating a working Go CLI executable. While concepts themselves don't generate CIL (they're used during analysis for validation), having a working synthesizer proves the entire parse-analyze-synthesize pipeline is functional. The placeholder CLI serves as a foundation for future increments that will generate actual executable code.

## Acceptance Criteria

1. Command line accepts `-o`/`--output` flag to specify output file path
2. Default output path is used when `-o` flag is not provided (e.g., `./output/cpl_cli.go`)
3. CPLCompiler orchestrates full pipeline: Parse → Analyze → Synthesize
4. Parse errors stop compilation before analysis phase
5. Analysis errors stop compilation before synthesis phase
6. Synthesize function creates output file successfully
7. Generated Go CLI code is valid and compiles
8. CLI displays placeholder message indicating successful analysis
9. All unit tests pass following BTDD practices
10. Integration test verifies end-to-end compilation pipeline
11. Integration test verifies CLI argument parsing for output path
12. Help text documents the `-o`/`--output` flag and pipeline phases

## Notes

**Increment 1 Synthesizer Characteristics:**

- Generates only a placeholder Go CLI that does nothing
- No CIL elements are generated (concepts don't produce CIL)
- No concept version checking (added in Increment 7)
- The CLI serves as proof that the parse-analyze pipeline worked

**Generated CLI Structure:**

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    // Placeholder CLI for Increment 1
    fmt.Println("CPL CLI - Increment 1 (Placeholder)")
    fmt.Println("Concepts analyzed successfully, no executable code generated yet.")
}
```

**Synthesizer Philosophy:**

Concept definitions do not generate CIL elements. The analyzer uses concept values to produce CIL with appropriate structure (array sizes, loop bounds). The synthesizer generates executable code from CIL - the synthesized code knows concept names and reads concept values at runtime, but has no knowledge of CIL structures.

**CLI Usage Examples:**

```bash
# Compile with default output path
./cpl_compiler input.cpl

# Compile with explicit output path
./cpl_compiler -o ./my_output/generated_cli.go input.cpl

# Compile with long-form output flag
./cpl_compiler --output ./my_output/generated_cli.go input.cpl

# Show help with all available options
./cpl_compiler --help
```

**Compiler Pipeline Flow:**

```
Input CPL File
      │
      ▼
┌─────────────┐
│ Parse Phase │───► Symbol Table
└─────────────┘          │
      │                  │
   Errors                ▼
      │          ┌───────────────┐
      │          │ Analyze Phase │───► AnalysisDefinitions
      │          └───────────────┘          │
      │              │                      │
      │           Errors                   ▼
      │              │              ┌───────────────┐
      │              │              │ Synthesize    │───► Go CLI File
      │              │              │ Phase         │
      │              │              └───────────────┘
      │              │                      │
      │              │                   Errors
      │              │                      │
      ▼              ▼                      ▼
   Compilation Failed                 Success
```

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
- [Synthesis Design](../specifications/design/cpl/go-synthesis-target/synthesis.md) - General synthesis phase design
