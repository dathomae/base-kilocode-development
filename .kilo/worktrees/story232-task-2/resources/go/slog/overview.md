# slog Overview

The `log/slog` package provides structured logging for Go programs. Unlike traditional logging libraries that format messages as plain strings, slog organizes log data as key-value pairs, making logs easier to search, filter, and analyze programmatically.

## Key Features

Structured logging separates the log message from its associated metadata. Each log entry contains a message, a severity level, and a collection of attributes. This structure enables automated processing and consistent formatting across different output destinations.

## The Logger Type

A [`Logger`](resources/go/slog/overview.md:17) is the primary interface for creating log entries. It provides methods for each severity level: `Debug`, `Info`, `Warn`, and `Error`. Each method accepts a message string followed by alternating keys and values that form the log attributes.

```go
logger.Info("request completed",
    "method", "GET",
    "path", "/api/users",
    "status", 200,
    "duration_ms", 42,
)
```

The output contains the message and all attributes in a structured format, typically JSON or a key-value text format.

## Severity Levels

slog defines four severity levels in increasing order of importance:

| Level | Usage                                                        |
| ----- | ------------------------------------------------------------ |
| Debug | Detailed information for debugging purposes                  |
| Info  | General operational information about program progress       |
| Warn  | Potentially harmful situations that warrant attention        |
| Error | Error conditions that prevented an operation from succeeding |

Each logger has a minimum level; messages below this threshold are discarded. This filtering happens before any formatting or output, making level checks efficient.

## Default Logger

The package provides a default logger accessible through top-level functions:

```go
slog.Info("application started")
slog.Error("connection failed", "error", err)
```

The default logger writes to standard error using the text handler. Production code typically creates a custom logger with specific configuration rather than relying on defaults.

## Creating a Custom Logger

Use [`slog.New`](resources/go/slog/overview.md:48) to create a logger with a specific handler:

```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
logger := slog.New(handler)
```

The [`HandlerOptions`](resources/go/slog/overview.md:52) struct configures the handler's behavior, including the minimum level and whether to include source code locations.

## Context Integration

slog integrates with Go's `context` package through methods like `InfoContext` and `DebugContext`. These methods extract a logger from the context if one was previously added, enabling request-scoped logging:

```go
// Add logger to context
ctx := slog.NewContext(ctx, logger)

// Retrieve and use logger from context
slog.InfoContext(ctx, "processing request")
```

This pattern ensures that logs include request-specific attributes that were added to the context logger.

## Performance Considerations

Loggers are safe for concurrent use by multiple goroutines. The handler handles synchronization, so you can share a single logger across your application.

For high-frequency logging, consider the potential overhead of attribute allocation. The package provides methods that accept pre-built attribute slices to reduce allocations in hot paths.
