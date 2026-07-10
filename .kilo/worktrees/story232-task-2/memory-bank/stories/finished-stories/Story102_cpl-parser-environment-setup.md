# Story 102: CPL Parser Environment Setup

## Overview

Set up the Gradle/Java environment for ANTLR4 parser generation and create a "hello world" grammar that accepts a single integer as the only valid statement. This establishes the foundation for incremental CPL (Chronocone Planning Language) parser development.

## References

- [CPL Language Examples](../specifications/requirements/metrics-definition-examples-2a.md) - Reference examples of full CPL syntax
- [Story016_intellij-workspace-setup.md](Story016_intellij-workspace-setup.md) - Similar Gradle environment setup
- [ANTLR4 Go Target Documentation](https://github.com/antlr/antlr4/blob/master/doc/go-target.md)

## Dependencies

### Prerequisites

None - This is the foundational story for the CPL parser.

### Dependents

- [Story103_cpl-parser-units-and-dimensions.md](Story103_cpl-parser-units-and-dimensions.md) - Not Started
- [Story104_cpl-parser-connectors.md](Story104_cpl-parser-connectors.md) - Not Started
- [Story105_cpl-evaluator-cli.md](Story105_cpl-evaluator-cli.md) - Not Started
- [Story106_cpl-parser-metrics.md](Story106_cpl-parser-metrics.md) - Not Started

## Tasks

1. Create the CPL parser project structure - Completed
   a. Create `src/chronocone_planning_language/` directory as a VSCode workspace
   b. Create `src/main/antlr/` subdirectory for ANTLR4 grammar files (standard Gradle location)
   c. Create `src/main/go/` subdirectory for custom Go parser logic (standard Gradle Go plugin location)
   d. Create `examples/` subdirectory with sample CPL files
   e. Create `.gitignore` to exclude `build/` and Gradle wrapper from version control
   f. Create `README.md` with project overview and build instructions

2. Set up Gradle build configuration - Completed
   a. Create `build.gradle.kts` with ANTLR4 plugin configured for Go target
   b. Configure ANTLR4 to generate Go code into the default `build/generated-src/antlr/main/` directory
   c. Set up Gradle wrapper for consistent builds
   d. Configure Go plugin to compile generated and custom code
   e. Write test to verify Gradle build succeeds

3. Create the "hello world" grammar - Completed
   a. Create `src/main/antlr/HelloCPL.g4` with minimal grammar accepting a single integer
   b. Define lexer rule for INTEGER tokens
   c. Define parser rule `program` that accepts a single integer statement
   d. Configure grammar to generate Go code with visitor and listener interfaces
   e. Write test to verify grammar compiles without errors

4. Generate Go code from grammar - Completed
   a. Run Gradle task to generate Go lexer and parser from `HelloCPL.g4`
   b. Verify generated files appear in `build/generated-src/antlr/main/`
   c. Confirm generated files include lexer, parser, listener, and visitor interfaces
   d. Write test to verify generated Go code compiles

5. Initialize Go module - Completed
   a. Create `go.mod` in project root with appropriate module path
   b. Add ANTLR4 Go runtime as a dependency
   c. Configure module to reference generated code in `build/generated-src/antlr/main/`
   d. Write test to verify `go build` succeeds

6. Create basic parser wrapper - Completed
   a. Create `src/main/go/parser/hello_parser.go` with `Parse(input string) error` function
   b. Integrate ANTLR lexer and parser
   c. Implement basic error handling for invalid input
   d. Write unit tests for parsing valid and invalid integers

7. Add VSCode workspace configuration - Completed
   a. Create `.vscode/settings.json` with Go and ANTLR4 file associations
   b. Create `.vscode/tasks.json` with Gradle build and test tasks
   c. Create `.vscode/launch.json` for debugging Go parser code

8. Create starter example file - Completed
   a. Create `examples/hello.cpl` containing a single valid integer
   b. Create `examples/README.md` explaining the hello world example
   c. Write integration test that parses the example file

## Decisions

1. **Grammar naming**: The initial grammar is named `HelloCPL.g4` to indicate it is a minimal "hello world" version that will be extended in subsequent stories.

2. **Go module path**: The module path follows the pattern `github.com/chronocone/chronocone_planning_language` to align with the project structure.

3. **Generated code location**: ANTLR-generated Go code is placed in `src/main/go/parser/antlr/` and committed to version control. Unlike typical build artifacts, these files must be available for Go module distribution via `go get`. Consumers of the parser package need access to the generated lexer and parser without running ANTLR themselves. The Gradle build regenerates these files when the grammar changes, and the updated files are committed as part of normal development.

4. **Build tool**: Gradle with Kotlin DSL is used for the build system to match the existing IntelliJ plugin workspace pattern and provide consistent ANTLR4 integration.

5. **Parser structure**: A thin wrapper in `src/main/go/parser/` provides a clean API over the generated ANTLR code, allowing the implementation details to change without affecting consumers.

## Notes

- The ANTLR4 Go runtime is required as a dependency. It can be added via `go get github.com/antlr4-go/antlr/v4`.
- The "hello world" grammar intentionally accepts only a single integer to validate the entire tool chain without complexity.
- Future stories will replace `HelloCPL.g4` with the full `ChronoconePlanningLanguage.g4` grammar.
- The VSCode workspace integration ensures the parser project is easily accessible from the main development environment.
