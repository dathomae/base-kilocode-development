# Chronocone Planning Language (CPL) Parser

A Go-based parser for the Chronocone Planning Language using ANTLR4.

## Overview

This project provides parsing capabilities for CPL files, which define types, concepts, and other structures for the Chronocone planning system.

For a complete directory listing, see [`toc.md`](./toc.md).

## Project Structure

### Source Organization

The source directories (`src/main/go/ast`, `src/main/go/parser`) contain both source files and their corresponding test files (co-located following Go best practices). Each grammar file has corresponding source and test files:

| Grammar File                    | AST Source   | Parser Listener   | AST Tests         | Parser Tests         |
| ------------------------------- | ------------ | ----------------- | ----------------- | -------------------- |
| `ChronoconePlanningLanguage.g4` | `cpl_ast.go` | `cpl_listener.go` | `cpl_ast_test.go` | `cpl_parser_test.go` |

```
src/
├── main/
│   ├── antlr/          # ANTLR4 grammar files (.g4)
│   │   ├── ChronoconePlanningLanguage.g4    # Root grammar
│   │   ├── ChronoconePlanningLanguageLexer.g4
│   │   ├── CPLConcepts.g4                   # Concepts section grammar
│   │   └── CPLTypes.g4                      # Types section grammar
│   └── go/
│       ├── ast/        # AST type definitions and their tests (co-located)
│       │   ├── cpl_ast.go
│       │   └── cpl_ast_test.go
│       └── parser/     # Parser implementation and tests (co-located)
│           ├── antlr/  # ANTLR-generated Go code (committed to version control)
│           ├── cpl_parser.go
│           ├── cpl_parser_test.go
│           ├── cpl_parser_examples_test.go
│           ├── cpl_parser_logging_test.go
│           ├── cpl_listener.go
│           ├── cpl_concepts_listener.go
│           ├── cpl_types_listener.go
│           └── ...
└── examples/           # Sample CPL files
```

## Build Instructions

### Prerequisites

- Java 11+ (to run ANTLR4)
- Go 1.21+
- Gradle 8.x
- ANTLR4 4.13+ (for grammar compilation)

### Building

**Important:** Always use Gradle for building and testing this project. Do not use Go tools (e.g., `go build`, `go test`) directly, as they will not properly regenerate ANTLR grammar files or handle project-specific build requirements.

```bash
# Delete the build artifacts, including the antlr generate source files.
./gradlew clean

# Build the parser code
./gradlew build

# Run tests. If needed, the code will be rebuilt before tests are run.
./gradlew test
```

**Note:** When running final tests, always use `./gradlew test` to ensure the complete test suite runs with proper dependencies and generated code.

## Usage

```go
import (
    "fmt"
    "github.com/chronocone/chronocone_planning_language/src/main/go/parser"
)

func main() {
    input := `types
        Currency = Number(precision=2)

    concepts
        Country = [Values[USA, Canada, Mexico]]`

    doc, err := parser.Parse(input)
    if err != nil {
        // handle error
    }

    fmt.Printf("Parsed %d types and %d concepts\n", len(doc.Types), len(doc.Concepts))
}
```

## Compiling CPL to Go

The CPL parser includes a synthesizer that generates executable Go programs from parsed CPL files.

### Generating Code

After parsing a CPL file, the synthesizer produces a `package main` Go program:

```go
// Example: generated/cmd/random_data_no_time/main.go
package main

import (
    "github.com/chronocone/chronocone_planning_language/src/main/go/calculation"
    "github.com/chronocone/chronocone_planning_language/src/main/go/connector"
    cplio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// ... generated code ...

func main() {
    // Load connectors, run computation, print result
}
```

### Running Generated Programs

Use the provided shell script to compile CPL source files and run the generated program:

```bash
./shell_scripts/compile_and_run.sh random_data_no_time
```

Where `random_data_no_time` is the name of a subdirectory under `examples/with_data/` containing `iospec.json`, `.cpl` files, and `.parquet` files.

This script:

1. Resolves its own directory and changes to the module root
2. Validates the example directory and required files
3. Runs the CPL compiler to generate a Go program
4. Executes the generated program with `go run`

See [`shell_scripts/README.md`](./shell_scripts/README.md) for full details.

## Examples

See the `examples/` directory for sample CPL files.

## Architectural Decisions

### Committing Generated ANTLR Code

Unlike typical code generation scenarios where generated files are excluded from version control, the ANTLR-generated Go files in `src/main/go/parser/antlr/` are intentionally committed to the repository. This decision was made because:

1. **Go Module Distribution**: The parser is designed to be consumed as a Go module via `go get`. Consumers need access to the generated lexer and parser without having to install Java or run ANTLR themselves.

2. **Zero-Dependency Consumption**: By committing the generated code, users of the parser library can simply import and use it without any build-time dependencies beyond the standard Go toolchain.

3. **Reproducible Builds**: The generated code is deterministic. When the grammar changes, the Gradle build regenerates the files, which are then committed as part of the normal development workflow.

### Test File Co-location

Test files are co-located with the source code they test (e.g., `cpl_ast.go` has its tests in `cpl_ast_test.go` in the same directory). This follows Go community best practices:

1. **Discoverability**: Tests are immediately visible alongside the code they test
2. **Package Structure**: Tests use `package xxx_test` for black-box testing while having access to the same package
3. **Simplified Navigation**: No need to navigate between separate `src/test` hierarchies

### Regenerating the Parser

When modifying the grammar file (`src/main/antlr/ChronoconePlanningLanguage.g4`), regenerate the Go code with:

```bash
./gradlew generateGrammarSource
```

Then commit the updated files in `src/main/go/parser/antlr/` along with your grammar changes.

## Agentic AI Directive

> **For AI Agents:** When modifying the file organization in this directory (adding, moving, or renaming directories), you **must** update the [`toc.md`](./toc.md) file to reflect the changed structure. Keeping the table of contents synchronized with the actual directory structure ensures accurate navigation for all users.

## License

See the project root LICENSE file.
