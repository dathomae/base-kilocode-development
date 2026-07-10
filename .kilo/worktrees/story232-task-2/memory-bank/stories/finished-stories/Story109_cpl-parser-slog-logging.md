# Story 109: CPL Parser - slog Logging Implementation

## Goal

Implement structured logging using Go's `log/slog` package for the Chronocone Planning Language parser with the following goals:

1. Generic entry/exit logging for all ANTLR listener callbacks at debug level
2. Entry/exit logging for major phase functions (Parse, etc.) at info level with configurable log level

## References

- [CPL Architecture - Introduction](specifications/architecture/cpl/introduction.md) - Overview of four-phase CPL lifecycle
- [slog Overview](../../resources/go/slog/overview.md) - Go's structured logging package
- [Story103_cpl-parser-units-and-dimensions.md](finished-stories/Story103_cpl-parser-units-and-dimensions.md) - Reference story for CPL parser tasks
- [Story104_cpl-parser-connectors.md](Story104_cpl-parser-connectors.md) - Dependent story that may leverage logging

## Dependencies

### Prerequisites

- [Story103_cpl-parser-units-and-dimensions.md](finished-stories/Story103_cpl-parser-units-and-dimensions.md) - Completed (requires parser environment, AST types, and listener infrastructure)

### Dependents

- [Story104_cpl-parser-connectors.md](Story104_cpl-parser-connectors.md) - Not Started (may use logging for debugging)
- [Story105_cpl-evaluator-cli.md](Story105_cpl-evaluator-cli.md) - Not Started (may use logging for debugging)

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Major Phase Functions                     │
│  ┌──────────────┐        ┌──────────────────────┐          │
│  │Parse function│───────▶│ParseWithLogLevel func│          │
│  └──────────────┘        └──────────────────────┘          │
└─────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    Logging Infrastructure                    │
│  ┌─────────────────────┐      ┌────────────────┐           │
│  │ParseLoggingConfig   │      │slog.Logger     │           │
│  │- LogLevel           │─────▶│                │           │
│  │- OutputFile         │      │                │           │
│  └─────────────────────┘      └────────────────┘           │
└─────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                 ANTLR Listener Composition                   │
│  ┌──────────────────────┐                                   │
│  │BaseChronoconePlanning│                                   │
│  │LanguageListener      │                                   │
│  │(ANTLR generated)      │                                   │
│  └──────────┬───────────┘                                   │
│             ▲                                               │
│             │ embeds                                        │
│  ┌──────────┴───────────┐                                   │
│  │CPLListener           │                                   │
│  │(business logic only) │                                   │
│  └──────────┬───────────┘                                   │
│             │ wrapped by                                    │
│  ┌──────────┴───────────┐                                   │
│  │LoggingCPLListener    │◀───── slog.Logger                │
│  │(adds debug logging)  │                                   │
│  └──────────────────────┘                                   │
└─────────────────────────────────────────────────────────────┘
```

## Tasks

### Task 1: Phase A - Structural Definition - Completed

**Intent:** Define the structural signatures for the logging infrastructure before writing tests or implementation.

**Files:**

- `src/main/go/parser/cpl_parser.go`
- `src/main/go/parser/logging_listener.go`

**Requirements:**

1. In `cpl_parser.go`, add the `ParseLoggingConfig` struct definition:

   ```go
   type ParseLoggingConfig struct {
       LogLevel   slog.Level
       OutputFile string  // If empty, logs to stdout; otherwise to specified file
   }
   ```

2. In `cpl_parser.go`, add the `ParseWithLogLevel` function signature:

   ```go
   func ParseWithLogLevel(input string, config *ParseLoggingConfig) (*ast.CPLDocument, error)
   ```

   - Add proper docstring describing the function
   - Implementation should be empty (`pass` equivalent in Go)

3. In `logging_listener.go`, create the `LoggingCPLListener` struct:

   ```go
   type LoggingCPLListener struct {
       inner  antlr.ChronoconePlanningLanguageListener
       logger *slog.Logger
   }
   ```

4. In `logging_listener.go`, add the constructor signature:

   ```go
   func NewLoggingCPLListener(inner antlr.ChronoconePlanningLanguageListener, logger *slog.Logger) (*LoggingCPLListener, error)
   ```

   - Add proper docstring describing the function
   - Implementation should return nil, nil for now

5. Ensure all imports are correct for compilation.

**Completion Criteria:**

- All structs and function signatures are defined
- Code compiles without errors (implementations can be empty/stubs)
- Docstrings describe the intended behavior

---

### Task 2: Phase B - Logical Tests - Completed

**Intent:** Write tests that define the expected behavior of the logging infrastructure before implementing the functionality.

**Files:**

- `src/test/go/logging/parser_logging_test.go` (new file)
- `src/test/go/logging/logging_listener_test.go` (new file)

**Requirements:**

1. Create the directory `src/test/go/logging/`

2. Create `parser_logging_test.go` with the following test functions (signatures only, bodies can be empty with `t.Skip("TODO: implement")`):

   ```go
   func TestParseLoggingConfig_ValidLogLevels(t *testing.T)
   // Table-driven test that all valid log level strings (debug, info, warn, error)
   // can be parsed correctly to their corresponding slog.Level values.

   func TestParseLoggingConfig_InvalidLogLevel(t *testing.T)
   // Test that an invalid log level string returns an appropriate error
   // with a helpful message listing valid values.

   func TestParseWithLogLevel_NilConfig(t *testing.T)
   // Test that passing a nil config to ParseWithLogLevel returns an error as expected.

   func TestParseWithLogLevel_FileOutput(t *testing.T)
   // Test that logs are written to a specified file path correctly, including:
   // - Creating the file if it doesn't exist
   // - Appending to the file if it already exists
   // - Writing the expected log format (entry and exit messages at the configured level)

   func TestParseWithLogLevel_StdoutOutput(t *testing.T)
   // Test that logs are written to stdout when no output file is specified.

   func TestParseWithLogLevel_LogLevelFiltering(t *testing.T)
   // Test that:
   // - At Info level, debug messages from the listener are not logged
   // - At Debug level, both info and debug messages are logged

   func TestParseWithLogLevel_DurationLogging(t *testing.T)
   // Test that parse_exit includes duration_ms in the log output.

   func TestParseWithLogLevel_ParseEntry(t *testing.T)
   // Test that parse_entry includes input_length in the log output.
   ```

3. Create `logging_listener_test.go` with the following test functions:

   ```go
   func TestLoggingCPLListener_NilInner(t *testing.T)
   // Test that constructor returns error when inner listener is nil.

   func TestLoggingCPLListener_NilLogger(t *testing.T)
   // Test that constructor returns error when logger is nil.

   func TestLoggingCPLListener_EnterEveryRule(t *testing.T)
   // Test that EnterEveryRule logs enter_rule message with rule name at debug level.

   func TestLoggingCPLListener_ExitEveryRule(t *testing.T)
   // Test that ExitEveryRule logs exit_rule message with rule name at debug level.

   func TestLoggingCPLListener_Delegation(t *testing.T)
   // Test that all methods delegate to the inner listener correctly.
   ```

4. Create `cpl_compiler_test.go` in `src/test/go/` with:

   ```go
   func TestCompiler_LogLevelFlag_Valid(t *testing.T)
   // Test that valid log level flags are accepted.

   func TestCompiler_LogLevelFlag_Invalid(t *testing.T)
   // Test that invalid log level flags produce an error with helpful message.

   func TestCompiler_LogFileFlag(t *testing.T)
   // Test that logfile flag is correctly passed to ParseWithLogLevel.
   ```

**Completion Criteria:**

- All test files compile (tests can be skipped with TODO)
- Test names clearly describe what behavior they will verify
- Test file structure follows Go testing conventions

---

### Task 3: Phase C - Implementation - ParseLoggingConfig and ParseWithLogLevel - Completed

**Intent:** Implement the logging configuration and parse function to satisfy the tests written in Task 2.

**File:** `src/main/go/parser/cpl_parser.go`

**Requirements:**

1. Add imports for `log/slog`, `os`, `path/filepath`, and `time`

2. Implement `ParseLoggingConfig` struct (already defined in Phase A)

3. Implement `ParseWithLogLevel(input string, config *ParseLoggingConfig) (*ast.CPLDocument, error)`:
   - Returns error if `config` is nil with message: `fmt.Errorf("logging config cannot be nil")`
   - Records start time for duration calculation
   - Creates handler based on `config.OutputFile`:
     - If empty: `slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: config.LogLevel})`
     - If set:
       - Ensure parent directory exists: `os.MkdirAll(filepath.Dir(config.OutputFile), 0755)`
       - Open file with append semantics: `os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)`
       - **Important:** The file must remain open during parsing; close explicitly after parsing completes (do not use defer)
   - Creates logger with the configured level
   - Logs entry at configured level: `parse_entry input_length=<len>`
   - Creates the base listener: `NewCPLListener()`
   - Creates the logging decorator wrapping the base listener (this will delegate to LoggingCPLListener once implemented)
   - Performs parsing using the decorated listener
   - Calculates duration
   - Logs exit at configured level: `parse_exit duration_ms=<ms> error=<err>`
   - Closes log file if one was opened
   - Returns result

4. Update `Parse(input string)` to call:
   ```go
   return ParseWithLogLevel(input, &ParseLoggingConfig{LogLevel: slog.LevelInfo})
   ```

**Completion Criteria:**

- All tests for ParseWithLogLevel pass
- Nil config returns appropriate error
- File creation and append behavior works correctly
- Stdout output works when OutputFile is empty
- Log level filtering works correctly
- Duration and input length are logged

---

### Task 4: Phase C - Implementation - LoggingCPLListener - Completed

**Intent:** Implement the decorator listener to satisfy the tests written in Task 2.

**File:** `src/main/go/parser/logging_listener.go`

**Requirements:**

1. Create the file with package declaration and imports:

   ```go
   package parser

   import (
       "fmt"
       "log/slog"
       antlr "github.com/antlr4-go/antlr/v4"
   )
   ```

2. Implement the `LoggingCPLListener` struct (already defined in Phase A):

   ```go
   type LoggingCPLListener struct {
       inner  antlr.ChronoconePlanningLanguageListener
       logger *slog.Logger
   }
   ```

3. Implement `NewLoggingCPLListener`:

   ```go
   func NewLoggingCPLListener(inner antlr.ChronoconePlanningLanguageListener, logger *slog.Logger) (*LoggingCPLListener, error) {
       if inner == nil {
           return nil, fmt.Errorf("inner listener cannot be nil")
       }
       if logger == nil {
           return nil, fmt.Errorf("logger cannot be nil")
       }
       return &LoggingCPLListener{inner: inner, logger: logger}, nil
   }
   ```

4. Implement `EnterEveryRule`:

   ```go
   func (l *LoggingCPLListener) EnterEveryRule(ctx antlr.ParserRuleContext) {
       l.logger.Debug("enter_rule", "rule", ctx.GetRuleName())
       l.inner.EnterEveryRule(ctx)
   }
   ```

5. Implement `ExitEveryRule`:

   ```go
   func (l *LoggingCPLListener) ExitEveryRule(ctx antlr.ParserRuleContext) {
       l.logger.Debug("exit_rule", "rule", ctx.GetRuleName())
       l.inner.ExitEveryRule(ctx)
   }
   ```

6. All other listener methods delegate directly to the inner listener.

**Completion Criteria:**

- All tests for LoggingCPLListener pass
- Nil inner listener returns error
- Nil logger returns error
- EnterEveryRule logs correctly at debug level
- ExitEveryRule logs correctly at debug level
- Delegation to inner listener works

---

### Task 5: Phase C - Implementation - Update Parse Function to Use Decorator - Completed

**Intent:** Integrate the logging listener into the parsing pipeline.

**File:** `src/main/go/parser/cpl_parser.go`

**Requirements:**

1. In `ParseWithLogLevel`, after creating the base `CPLListener`:

   ```go
   baseListener := NewCPLListener()
   loggedListener, err := NewLoggingCPLListener(baseListener, logger)
   if err != nil {
       return nil, fmt.Errorf("failed to create logging listener: %w", err)
   }
   ```

2. Register the `loggedListener` with the parser instead of the base listener:

   ```go
   antlrruntime.ParseTreeWalkerDefault.Walk(loggedListener, parser.Document())
   ```

3. Get the document from the base listener (not the decorator):
   ```go
   return baseListener.GetDocument(), nil
   ```

**Completion Criteria:**

- Decorator is properly applied in ParseWithLogLevel
- Logs are generated at the expected levels
- Parsing still works correctly with decorator

---

### Task 6: Phase C - Implementation - Update Compiler Entry Point - Completed

**Intent:** Expose logging configuration via CLI flags.

**File:** `src/main/go/cpl_compiler.go`

**Requirements:**

1. Add imports for `flag`, `log/slog`, and `strings` packages

2. Define command-line flags:
   - `-loglevel`: Optional flag to set logging level (valid values: debug, info, warn, error). Defaults to "info" if not specified
   - `-logfile`: Optional flag to specify log output file path. Defaults to stdout if not specified

3. Create a lookup map to convert string values to `slog.Level`:

   ```go
   var levelMap = map[string]slog.Level{
       "debug": slog.LevelDebug,
       "info":  slog.LevelInfo,
       "warn":  slog.LevelWarn,
       "error": slog.LevelError,
   }
   ```

4. Parse the log level flag value:
   - Convert to lowercase for case-insensitive comparison
   - Look up in levelMap
   - If not found, print error to stderr listing valid values and exit with non-zero status

5. Create `ParseLoggingConfig` with the parsed level and optional output file

6. Replace the call to `parser.Parse(string(content))` with `parser.ParseWithLogLevel(string(content), config)`

7. Handle any errors from `ParseWithLogLevel` (including nil config errors)

**Example usage after modification:**

```bash
# Default: info level, stdout output
./cpl_compiler input.cpl

# Debug level to stdout
./cpl_compiler -loglevel=debug input.cpl

# Debug level to file
./cpl_compiler -loglevel=debug -logfile=parser.log input.cpl

# Info level to file
./cpl_compiler -logfile=/var/log/cpl/parser.log input.cpl
```

**Completion Criteria:**

- Valid log level flags are accepted (debug, info, warn, error)
- Invalid log level returns error with helpful message
- Default values work when flags not provided
- Integration with ParseWithLogLevel works correctly
- All compiler tests pass

---

### Task 7: Phase D - Integration and Backward Compatibility - Completed

**Intent:** Ensure backward compatibility with existing tests and verify end-to-end functionality.

**Files:**

- `src/test/go/parser/*_test.go`
- `src/test/go/logging/*_test.go`

**Requirements:**

1. Update existing parser tests if needed:
   - All calls to `Parse(input)` should continue to work unchanged
   - Add integration tests for `ParseWithLogLevel` in existing test files

2. Run all tests to verify:
   - All existing parser tests pass
   - All new logging tests pass
   - No regressions introduced

3. Create an integration test that exercises the full flow:
   ```go
   func TestParseWithLogLevel_FullIntegration(t *testing.T)
   // Test parsing a complete CPL file with debug logging enabled
   // Verify both parse_entry/parse_exit and enter_rule/exit_rule logs
   ```

**Completion Criteria:**

- All existing tests continue to pass
- All new tests pass
- No regressions in parser functionality
- Integration test verifies end-to-end logging

---

### Task 8: Phase D - Documentation and Walkthrough - Completed

**Intent:** Create documentation for the logging feature and a walkthrough of the implementation.

**Files:**

- `memory-bank/stories/walkthroughs/Story109_walkthrough.md` (new file)

**Requirements:**

1. Create a walkthrough document that describes:
   - The design decisions made
   - How the decorator pattern was implemented
   - How to use the logging functionality (API and CLI)
   - Any challenges encountered and how they were resolved

2. Include code examples showing:
   - Basic usage with default logging
   - Debug logging to stdout
   - Logging to file
   - Programmatic usage with ParseWithLogLevel

**Completion Criteria:**

- Walkthrough document is complete and accurate
- Code examples are tested and work correctly
- Post-mortem section is added with lessons learned

## Logical TDD Lifecycle Summary

This story follows the **Logical TDD Lifecycle** as defined in `memory-bank/concepts.md`:

### Phase A: Structural Definition

- Define all structs and function signatures
- No implementation logic - only structure
- Ensures the API contract is clear before testing

### Phase B: Logical Tests

- Write all test functions with clear names and docstrings
- Tests define expected behavior
- Tests are skipped initially ("TODO: implement")

### Phase C: Implementation

- Implement functionality to satisfy each test
- Work through tests one by one
- Ensure all tests pass before moving to next component

### Phase D: Documentation

- Verify backward compatibility
- Create walkthrough documentation
- Capture lessons learned

## Constraints

1. **Log Level Requirements**: Must support debug, info, warn, and error levels per slog standard levels
2. **Output Destinations**: Must support both stdout and file output with append semantics
3. **File Permissions**: Use 0755 for directories, 0644 for log files
4. **Error Handling**: All errors must be properly wrapped and returned, not logged and swallowed
5. **Resource Management**: Log files must be explicitly closed after parsing (do not use defer)
6. **Decorator Pattern**: Must use composition (decorator pattern) rather than inheritance for the logging listener
7. **Backward Compatibility**: The existing `Parse(input string)` API must remain unchanged
8. **Test Isolation**: All file-based tests must use temporary directories via `t.TempDir()`

## Design Decisions

### 1. Generic Listener Logging via EnterEveryRule/ExitEveryRule

The ANTLR-generated base listener provides `EnterEveryRule` and `ExitEveryRule` which are called for every grammar rule. By creating a `LoggingCPLListener` that implements these methods, we get automatic entry/exit logging for all listener callbacks without modifying individual methods.

### 2. Decorator Pattern for Listener Composition

**Chosen Approach: Decorator Pattern**

```
BaseChronoconePlanningLanguageListener (ANTLR generated)
    ↑ (CPLListener embeds base)
CPLListener (business logic only)
    ← (wrapped by)
LoggingCPLListener (adds debug logging, delegates to inner)
```

The `LoggingCPLListener` wraps another listener using composition:

```go
type LoggingCPLListener struct {
    inner  antlr.ChronoconePlanningLanguageListener
    logger *slog.Logger
}

func (l *LoggingCPLListener) EnterEveryRule(ctx antlr.ParserRuleContext) {
    l.logger.Debug("enter_rule", "rule", getRuleName(ctx))
    l.inner.EnterEveryRule(ctx)
}
```

**Benefits of this approach:**

- Keeps `CPLListener` independent of logging concerns (single responsibility)
- Follows Go's preference for composition over inheritance
- Allows `CPLListener` to be tested without logging infrastructure
- Enables flexible layering of cross-cutting concerns

**Rejected Alternative: Inheritance Hierarchy**

```
BaseChronoconePlanningLanguageListener (ANTLR generated)
    ↑
LoggingCPLListener (adds debug logging)
    ↑
CPLListener (business logic)
```

This approach was rejected because it tightly couples business logic to logging infrastructure and makes testing more difficult.

### 3. Two-Function API

- `Parse(input string)` - Simple API with default Info level, stdout output
- `ParseWithLogLevel(input string, config *ParseLoggingConfig)` - Full control over logging level and output destination. Returns error if config is nil.

### 4. Nil Config Handling

`ParseWithLogLevel` requires a valid non-nil config. Passing nil will return an error. The caller is responsible for providing a valid `ParseLoggingConfig`.

### 5. Output Destination and File Handling

The `OutputFile` field in `ParseLoggingConfig` controls where logs are written:

- Empty string (`""`): Logs go to stdout
- File path: Logs go to the specified file with append semantics (created if doesn't exist, appended if exists)

`ParseWithLogLevel` manages the full lifecycle of the log file:

- Ensures parent directory exists using `os.MkdirAll()` with 0755 permissions (creates all necessary parent directories)
  - If `OutputFile` is just a filename without a path (e.g., "parser.log"), `filepath.Dir()` returns "." which refers to the current directory, and `os.MkdirAll(".")` will succeed without creating any new directories
- Opens file with `os.O_APPEND|os.O_CREATE|os.O_WRONLY` permissions (0644)
- Closes file explicitly after parsing completes (do not use defer)
- Returns error if directory cannot be created or file cannot be opened

**Permission Rationale:**

- Directory permissions (0755): Allow owner full access, group/others read+execute. This is the standard Unix directory permission that allows directory traversal while restricting modifications to the owner.
- File permissions (0644): Allow owner read+write, group/others read-only. This ensures log files are writable by the owner while remaining readable for log analysis tools that may run as different users.

## Intent

The purpose of this story is to add observability to the CPL parser through structured logging. The logging infrastructure serves two audiences:

1. **End users** running the compiler CLI: They need high-level visibility into parser progress (entry/exit with timing) at info level.

2. **Developers** debugging parser issues: They need detailed trace of every rule entry and exit at debug level to understand parser behavior.

The design uses the decorator pattern to keep business logic (CPLListener) independent of logging concerns, following Go's preference for composition over inheritance. This separation allows the listener to be tested without logging infrastructure and enables flexible layering of cross-cutting concerns.

## Acceptance Criteria

1. The `ParseLoggingConfig` struct exists with `LogLevel` and `OutputFile` fields
2. `ParseWithLogLevel` returns an error when passed a nil config
3. `ParseWithLogLevel` creates parent directories as needed when OutputFile is specified
4. `ParseWithLogLevel` opens log files with append semantics (creates if not exists, appends if exists)
5. `ParseWithLogLevel` logs `parse_entry` at the configured level with input length
6. `ParseWithLogLevel` logs `parse_exit` at the configured level with duration and error status
7. `Parse` provides a simple API defaulting to Info level and stdout output
8. `LoggingCPLListener` wraps another listener and adds debug logging via EnterEveryRule/ExitEveryRule
9. `LoggingCPLListener` constructor returns errors for nil inner listener or nil logger
10. The compiler CLI accepts `-loglevel` and `-logfile` flags
11. Invalid log level values produce a helpful error message listing valid options
12. Debug level shows enter_rule/exit_rule messages; Info level does not
13. All new code has accompanying unit tests written BEFORE implementation
14. Existing tests continue to pass

## File Changes Summary

| File                                           | Change Type | Description                                                                                           |
| ---------------------------------------------- | ----------- | ----------------------------------------------------------------------------------------------------- |
| `src/main/go/parser/cpl_parser.go`             | Modify      | Add ParseLoggingConfig struct, ParseWithLogLevel function, and update Parse to use default Info level |
| `src/main/go/parser/logging_listener.go`       | New         | Decorator listener for generic logging via EnterEveryRule/ExitEveryRule                               |
| `src/main/go/parser/cpl_listener.go`           | No change   | No modifications needed - decorator pattern keeps business logic independent                          |
| `src/main/go/cpl_compiler.go`                  | Modify      | Add loglevel and logfile flags, use ParseWithLogLevel                                                 |
| `src/test/go/logging/parser_logging_test.go`   | New         | Unit tests for parser logging functionality in new logging test directory                             |
| `src/test/go/logging/logging_listener_test.go` | New         | Unit tests for LoggingCPLListener                                                                     |
| `src/test/go/cpl_compiler_test.go`             | New         | Unit tests for compiler CLI logging flags                                                             |
| `src/test/go/parser/*_test.go`                 | Modify      | Add integration tests for ParseWithLogLevel                                                           |

## Example Usage

### Default Parse (Info level, stdout)

```go
doc, err := parser.Parse(input)
```

### With Debug Level to Stdout

```go
config := &parser.ParseLoggingConfig{
    LogLevel: slog.LevelDebug,
}
doc, err := parser.ParseWithLogLevel(input, config)
```

### With Debug Level to File

```go
config := &parser.ParseLoggingConfig{
    LogLevel:   slog.LevelDebug,
    OutputFile: "/var/log/cpl/parser.log",
}
doc, err := parser.ParseWithLogLevel(input, config)
```

### CLI Examples

```bash
# Default: info level, stdout output
./cpl_compiler input.cpl

# Debug level to stdout
./cpl_compiler -loglevel=debug input.cpl

# Debug level to file
./cpl_compiler -loglevel=debug -logfile=parser.log input.cpl

# Info level to file
./cpl_compiler -logfile=/var/log/cpl/parser.log input.cpl
```

## Expected Log Output Examples

### Debug Level Output

```
time=2024-01-15T10:30:00.000Z level=INFO msg=parse_entry input_length=256
time=2024-01-15T10:30:00.001Z level=DEBUG msg=enter_rule rule=Document
time=2024-01-15T10:30:00.002Z level=DEBUG msg=enter_rule rule=Statement
time=2024-01-15T10:30:00.003Z level=DEBUG msg=exit_rule rule=Statement
time=2024-01-15T10:30:00.004Z level=DEBUG msg=exit_rule rule=Document
time=2024-01-15T10:30:00.005Z level=INFO msg=parse_exit duration_ms=5 error=null
```

### Info Level Output (debug messages filtered)

```
time=2024-01-15T10:30:00.000Z level=INFO msg=parse_entry input_length=256
time=2024-01-15T10:30:00.005Z level=INFO msg=parse_exit duration_ms=5 error=null
```

## Notes

- The `LoggingCPLListener` uses ANTLR's `EnterEveryRule` and `ExitEveryRule` callbacks to provide generic logging for all grammar rules without modifying individual listener methods.
- Log output format follows slog's default text handler: `time=2024-01-15T10:30:00.000Z level=INFO msg=parse_entry input_length=256`
- File handling explicitly closes files after parsing to avoid resource leaks; defer is not used to ensure immediate cleanup
- The decorator pattern was chosen over inheritance to maintain single responsibility and enable independent testing of business logic
- Following the Logical TDD Lifecycle, tests are written BEFORE implementation (Phase B before Phase C)
