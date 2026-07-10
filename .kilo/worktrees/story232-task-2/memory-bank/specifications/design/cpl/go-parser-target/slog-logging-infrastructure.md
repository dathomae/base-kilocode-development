# CPL Parser - slog Logging Infrastructure

## Overview

The Chronocone Planning Language (CPL) parser implements structured logging using Go's standard [`log/slog`](../../../../resources/go/slog/overview.md) package. This infrastructure provides observability into parser execution through configurable entry/exit logging at two distinct levels: major phase functions and ANTLR listener callbacks.

## Design Goals

1. **Generic entry/exit logging** for all ANTLR listener callbacks at the debug level
2. **Entry/exit logging** for major phase functions (such as Parse) at the info level with configurable log level
3. **Separation of concerns** - Business logic remains independent of logging infrastructure
4. **Flexibility** - Configurable log levels and output destinations

## Architecture

### High-Level Component Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Parser Entry Points                              │
│  ┌─────────────────┐          ┌─────────────────────────────────────┐  │
│  │   Parse(input)  │─────────▶│ ParseWithLogLevel(input, config)    │  │
│  │  (convenience)  │          │   - nil config validation           │  │
│  │  Info level,    │          │   - handler creation                │  │
│  │  stdout output  │          │   - logger initialization           │  │
│  └─────────────────┘          │   - entry/exit logging              │  │
│                               └─────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       Logging Infrastructure                             │
│  ┌─────────────────────────────┐      ┌──────────────────────────────┐ │
│  │   ParseLoggingConfig        │      │   slog.Logger                │ │
│  │   ─────────────────────     │      │   ─────────────              │ │
│  │   LogLevel   slog.Level     │─────▶│   Structured logging with    │ │
│  │   OutputFile string         │      │   level filtering            │ │
│  │                             │      │                              │ │
│  │   Output Rules:             │      │   Log Level Support:         │ │
│  │   - "": stdout              │      │   - Debug                    │ │
│  │   - path: file (append)     │      │   - Info                     │ │
│  │                             │      │   - Warn                     │ │
│  │   File Permissions:         │      │   - Error                    │ │
│  │   - Directories: 0755       │      │                              │ │
│  │   - Files: 0644             │      │   Output Format:             │ │
│  │                             │      │   - Text handler (default)   │ │
│  │   Lifecycle Management:     │      │   - time=..., level=...,     │ │
│  │   - Explicit close (no      │      │     msg=..., key=value...    │ │
│  │     defer)                  │      │                              │ │
│  └─────────────────────────────┘      └──────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    ANTLR Listener Composition                            │
│                                                                          │
│   BaseChronoconePlanningLanguageListener (ANTLR generated)               │
│       ▲                                                                  │
│       │ embeds                                                           │
│   ┌───┴───────────────────────┐                                          │
│   │  CPLListener              │                                          │
│   │  ─────────────            │  Business logic only, no logging          │
│   │  (business logic)         │  concerns. Implements grammar-specific   │
│   │                           │  parsing for units, dimensions, etc.     │
│   └────┬──────────────────────┘                                          │
│        │ wrapped by (decorator pattern)                                  │
│   ┌────┴──────────────────────┐                                          │
│   │  LoggingCPLListener       │  Adds debug logging via                  │
│   │  ─────────────────        │  EnterEveryRule/ExitEveryRule,           │
│   │  (logging decorator)      │  delegates all calls to inner listener   │
│   │                           │                                          │
│   │  enter_rule rule=<Name>   │  Debug level - all grammar rules         │
│   │  exit_rule rule=<Name>    │  Debug level - all grammar rules         │
│   └───────────────────────────┘                                          │
│                                                                          │
│   Pattern: Decorator (composition) rather than inheritance               │
└─────────────────────────────────────────────────────────────────────────┘
```

## Key Components

### ParseLoggingConfig

The configuration struct that controls logging behavior:

```go
type ParseLoggingConfig struct {
    LogLevel   slog.Level  // Minimum level to log (Debug, Info, Warn, Error)
    OutputFile string      // Empty for stdout, path for file output
}
```

**Output Destination Rules:**

- Empty string (`""`): Logs written to stdout using `slog.NewTextHandler(os.Stdout, ...)`
- File path: Logs written to specified file with append semantics
  - Parent directories created if needed (`os.MkdirAll` with 0755)
  - File opened with `os.O_APPEND|os.O_CREATE|os.O_WRONLY` (permissions 0644)
  - File explicitly closed after parsing (no defer)

### Two-Function API

The parser exposes two entry points to balance simplicity with flexibility:

1. **`Parse(input string)`** - Convenience function
   - Default log level: Info
   - Default output: stdout
   - Suitable for most use cases

2. **`ParseWithLogLevel(input string, config *ParseLoggingConfig)`** - Full control
   - Configurable log level
   - Configurable output destination
   - Returns error if config is nil
   - Use when debugging or logging to files

### LoggingCPLListener (Decorator Pattern)

The `LoggingCPLListener` wraps another `antlr.ChronoconePlanningLanguageListener` to add debug logging without modifying the underlying business logic:

```go
type LoggingCPLListener struct {
    inner  antlr.ChronoconePlanningLanguageListener
    logger *slog.Logger
}
```

**Key Methods:**

- `EnterEveryRule(ctx)` - Logs `enter_rule rule=<RuleName>` at debug level, delegates to inner
- `ExitEveryRule(ctx)` - Logs `exit_rule rule=<RuleName>` at debug level, delegates to inner
- All other listener methods delegate directly to inner listener

**Why Decorator Pattern:**

- **Single Responsibility**: `CPLListener` handles business logic only
- **Testability**: Business logic can be tested without logging infrastructure
- **Composition over Inheritance**: Follows Go idioms
- **Flexibility**: Easy to add/remove logging layer

## Log Levels and Output

### Major Phase Functions (Info Level and Above)

Logged by `ParseWithLogLevel`:

```
time=2024-01-15T10:30:00.000Z level=INFO msg=parse_entry input_length=256
time=2024-01-15T10:30:00.005Z level=INFO msg=parse_exit duration_ms=5 error=null
```

**Attributes:**

- `input_length`: Length of input string in characters
- `duration_ms`: Parse duration in milliseconds
- `error`: Error message if parsing failed, "null" otherwise

### Listener Callbacks (Debug Level)

Logged by `LoggingCPLListener` via `EnterEveryRule` and `ExitEveryRule`:

```
time=2024-01-15T10:30:00.001Z level=DEBUG msg=enter_rule rule=Document
time=2024-01-15T10:30:00.002Z level=DEBUG msg=enter_rule rule=Statement
time=2024-01-15T10:30:00.003Z level=DEBUG msg=exit_rule rule=Statement
time=2024-01-15T10:30:00.004Z level=DEBUG msg=exit_rule rule=Document
```

**Attributes:**

- `rule`: Name of the grammar rule being entered or exited

### Level Filtering

- **Debug level**: Shows all messages (both phase functions and listener callbacks)
- **Info level**: Shows phase function entry/exit only (listener callbacks filtered)
- **Warn/Error levels**: Show warnings and errors only

## CLI Integration

The CPL compiler accepts logging configuration via command-line flags:

```bash
# Default: info level, stdout
./cpl_compiler input.cpl

# Debug level to stdout
./cpl_compiler -loglevel=debug input.cpl

# Debug level to file
./cpl_compiler -loglevel=debug -logfile=parser.log input.cpl

# Info level to file
./cpl_compiler -logfile=/var/log/cpl/parser.log input.cpl
```

**Flags:**

- `-loglevel`: Optional. Valid values: `debug`, `info`, `warn`, `error`. Default: `info`
- `-logfile`: Optional. Path to log file. Default: stdout

**Log Level Mapping:**

```go
var levelMap = map[string]slog.Level{
    "debug": slog.LevelDebug,
    "info":  slog.LevelInfo,
    "warn":  slog.LevelWarn,
    "error": slog.LevelError,
}
```

Invalid log level values produce an error message listing valid options.

## File Permissions

The logging infrastructure creates files and directories with specific permissions:

| Resource        | Permission                       | Octal | Rationale                                                                                       |
| --------------- | -------------------------------- | ----- | ----------------------------------------------------------------------------------------------- |
| Log directories | Owner: rwx, Group: rx, Other: rx | 0755  | Standard Unix directory permissions allowing traversal while restricting modifications to owner |
| Log files       | Owner: rw, Group: r, Other: r    | 0644  | Writable by owner, readable by log analysis tools that may run as different users               |

## Error Handling

All errors in the logging infrastructure are returned (not logged and swallowed):

- **Nil config**: `fmt.Errorf("logging config cannot be nil")`
- **Nil inner listener**: `fmt.Errorf("inner listener cannot be nil")`
- **Nil logger**: `fmt.Errorf("logger cannot be nil")`
- **Directory creation failure**: Wrapped and returned
- **File open failure**: Wrapped and returned

## Usage Patterns

### Programmatic Usage

```go
// Simple usage with defaults
doc, err := parser.Parse(input)

// Custom configuration
tmpDir := t.TempDir()
config := &parser.ParseLoggingConfig{
    LogLevel:   slog.LevelDebug,
    OutputFile: filepath.Join(tmpDir, "parser.log"),
}
doc, err := parser.ParseWithLogLevel(input, config)
```

### Testing

```go
func TestParseWithLogLevel_FileOutput(t *testing.T) {
    tmpDir := t.TempDir()
    logFile := filepath.Join(tmpDir, "test.log")

    config := &parser.ParseLoggingConfig{
        LogLevel:   slog.LevelInfo,
        OutputFile: logFile,
    }

    _, err := parser.ParseWithLogLevel("units\n\tunit = Count", config)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }

    content, err := os.ReadFile(logFile)
    if err != nil {
        t.Fatalf("Failed to read log file: %v", err)
    }

    if !strings.Contains(string(content), "parse_entry") {
        t.Error("Log file missing parse_entry message")
    }
}
```

## References

- [slog Overview](../../../../resources/go/slog/overview.md) - Go's structured logging package
- [slog Handlers](../../../../resources/go/slog/handlers.md) - Handler types and configuration
- [Story109: CPL Parser slog Logging Implementation](../../../../memory-bank/stories/Story109_cpl-parser-slog-logging.md) - Implementation story
