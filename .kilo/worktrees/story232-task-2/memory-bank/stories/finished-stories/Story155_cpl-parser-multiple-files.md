# Story155: CPL Parser Multiple Files Support

## Goal

Implement support for parsing multiple CPL files as a single compilation unit. This allows users to split their CPL definitions (concepts, types, units, measures, connectors, metrics) across multiple files for better organization. This story is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md).

## References

- [CPL Architecture - Introduction](../specifications/architecture/cpl/introduction.md)
- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)

## Dependencies

None

## Dependent Stories

- [Epic004_cpl-connectors-implementation.md](../epics/Epic004_cpl-connectors-implementation.md) - Not Started (Increment 2 depends on this)

## Tasks

1. Update CPL Parser CLI/API to accept multiple file paths - Completed
   a. Write tests for the CLI/API entry point to handle multiple file paths and directory inputs - Completed
   b. Modify the main parsing entry point to accept a list of file paths or a directory - Completed
   c. Implement file discovery logic for directories (e.g., find all `.cpl` files) - Completed

2. Implement multi-file parsing logic with source tracking - Completed
   a. Write tests for parsing multiple files and verifying source location tracking (file name, line, column) - Completed
   b. Implement the parsing strategy (either concatenating files with a source map or parsing individually and merging ASTs/Symbol Tables) - Completed
   c. Ensure the chosen strategy accurately tracks the origin file for each parsed token/node - Completed

3. Update error reporting to include file names - Completed
   a. Write tests for error reporting to ensure correct file names, lines, and columns are included in error messages - Completed
   b. Modify the error reporting mechanism to extract and display the file name from the source location tracking - Completed

4. Integration testing for multi-file compilation - Completed
   a. Write integration tests for parsing multiple files with valid cross-references - Completed
   b. Write integration tests for duplicate definitions across different files - Completed
   c. Implement any necessary fixes discovered during integration testing - Completed

5. Create standard definitions directory and initial CPL file - Completed
   a. Create the directory `src/chronocone_planning_language/standard_definitions/` - Completed
   b. Create a `.cpl` file containing the standard derived type definition `Currency`, derived from `Number` with `precision=2` - Completed
   c. Write tests verifying that standard definitions are automatically loaded and parsed as part of every CPL compilation - Completed
   d. Implement the mechanism to automatically include standard definition files in every compilation - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The parsing strategy must accurately preserve source locations (file name, line, column) for error reporting.
- The solution must be compatible with the existing ANTLR-based parser infrastructure.

## Intent

- **Task 1**: To allow users to specify multiple input files or directories containing CPL files via the CLI or API.
- **Task 2**: To process multiple files as a single compilation unit while maintaining accurate source tracking for each element.
- **Task 3**: To provide clear and actionable error messages that point users to the exact file and location of an issue.
- **Task 4**: To ensure the entire multi-file parsing system works correctly end-to-end, including cross-references and error handling.
- **Task 5**: To establish a local standard definitions mechanism that provides a baseline `Currency` type definition for every CPL compilation, ensuring consistency and reducing duplication across user-defined CPL files.

## Acceptance Criteria

- The CPL parser can accept a list of individual `.cpl` files as input.
- The CPL parser can accept a directory as input and automatically discover all `.cpl` files within it.
- Definitions spread across multiple files are successfully parsed as a single compilation unit.
- Syntax and semantic errors correctly identify the file name, line number, and column number where the error occurred.
- Duplicate definitions across different files are detected and reported with accurate locations for both the original and duplicate definitions.
- The directory `src/chronocone_planning_language/standard_definitions/` exists and contains a `.cpl` file with the standard derived type definition `Currency`, derived from `Number` with `precision=2`.
- Every CPL compilation automatically loads and parses the standard definition file as part of the compilation unit.
- The `Currency` type from the standard file can be referenced from user-provided CPL files without explicit import statements.
- All unit and integration tests pass.

## Notes

- **Parsing Strategy**: The compiler CLI will handle multi-file definitions by either concatenating files before parsing (with appropriate source map tracking) or parsing them individually and merging the symbol tables. The exact approach will be determined during implementation based on ANTLR's capabilities for multi-file source tracking in Go.
- **Standard Definitions**: The standard definitions mechanism is an interim solution. In the future, these definitions will be read from the database rather than local `.cpl` files. The local file-based approach is sufficient for now and provides a clear migration path.
- This story is a prerequisite for Increment 2 of the Connectors implementation, as time-related definitions (Instant, Interval, duration concepts) will be used repeatedly and should be organized in separate files.
