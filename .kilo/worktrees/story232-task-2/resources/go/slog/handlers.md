# slog Handlers

Handlers determine how log entries are formatted and where they are written. The `log/slog` package provides two built-in handlers, and you can implement custom handlers for specialized output formats or destinations.

## Built-in Handlers

### TextHandler

The [`TextHandler`](resources/go/slog/handlers.md:7) formats log entries as key-value pairs in a human-readable text format. It is the default handler when using the package-level logging functions.

```go
handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
logger := slog.New(handler)
```

Output format:

```
time=2009-11-10T23:00:00.000000Z level=INFO msg="request completed" method=GET path=/api/users
```

### JSONHandler

The [`JSONHandler`](resources/go/slog/handlers.md:20) formats log entries as JSON objects. This format is preferred for production environments where logs are processed by aggregation systems.

```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
    AddSource: true,
})
logger := slog.New(handler)
```

Output format:

```json
{
  "time": "2009-11-10T23:00:00.000000Z",
  "level": "INFO",
  "source": {
    "function": "main.processRequest",
    "file": "main.go",
    "line": 42
  },
  "msg": "request completed",
  "method": "GET",
  "path": "/api/users"
}
```

## Handler Options

The [`HandlerOptions`](resources/go/slog/handlers.md:33) struct configures handler behavior:

| Field       | Type      | Description                                                |
| ----------- | --------- | ---------------------------------------------------------- |
| Level       | `Leveler` | Minimum level to log; entries below this level are dropped |
| AddSource   | `bool`    | Include source file and line number in log entries         |
| ReplaceAttr | `func`    | Transform attributes before they are logged                |

### Level Configuration

The `Level` field accepts any type implementing the [`Leveler`](resources/go/slog/handlers.md:44) interface. The simplest approach uses the predefined level constants:

```go
slog.LevelDebug  // -4
slog.LevelInfo   // 0
slog.LevelWarn   // 4
slog.LevelError  // 8
```

For dynamic level changes, use [`LevelVar`](resources/go/slog/handlers.md:52):

```go
var logLevel = new(slog.LevelVar)
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: logLevel,
})

// Change level at runtime
logLevel.Set(slog.LevelDebug)
```

### Source Information

Setting `AddSource: true` includes the file path, line number, and function name in each log entry. This is useful for debugging but adds overhead and increases log size.

### Attribute Replacement

The `ReplaceAttr` function transforms attributes before output. Use it to rename keys, mask sensitive data, or filter attributes:

```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
        // Redact sensitive fields
        if a.Key == "password" {
            return slog.String(a.Key, "[REDACTED]")
        }
        // Rename fields for consistency
        if a.Key == "req_id" {
            a.Key = "requestId"
        }
        return a
    },
})
```

## Handler Interface

Custom handlers implement the [`Handler`](resources/go/slog/handlers.md:79) interface:

```go
type Handler interface {
    Enabled(context.Context, Level) bool
    Handle(context.Context, Record) error
    WithAttrs(attrs []Attr) Handler
    WithGroup(name string) Handler
}
```

| Method    | Purpose                                                  |
| --------- | -------------------------------------------------------- |
| Enabled   | Report whether the handler handles records at the level  |
| Handle    | Format and output a log record                           |
| WithAttrs | Return a new handler with additional attributes          |
| WithGroup | Return a new handler that groups attributes under a name |

## Creating a Custom Handler

Here is a simple custom handler that writes to a database:

```go
type DBHandler struct {
    db     *sql.DB
    level  slog.Level
    attrs  []slog.Attr
    groups []string
}

func (h *DBHandler) Enabled(_ context.Context, level slog.Level) bool {
    return level >= h.level
}

func (h *DBHandler) Handle(ctx context.Context, r slog.Record) error {
    // Collect all attributes
    attrs := make(map[string]any)
    r.Attrs(func(a slog.Attr) bool {
        attrs[a.Key] = a.Value.Any()
        return true
    })

    // Insert into database
    _, err := h.db.ExecContext(ctx,
        "INSERT INTO logs (time, level, msg, attrs) VALUES (?, ?, ?, ?)",
        r.Time, r.Level, r.Message, attrs,
    )
    return err
}

func (h *DBHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    h2 := *h
    h2.attrs = append(h2.attrs, attrs...)
    return &h2
}

func (h *DBHandler) WithGroup(name string) slog.Handler {
    h2 := *h
    h2.groups = append(h2.groups, name)
    return &h2
}
```

## Handler Composition

Use [`slog.MultiHandler`](resources/go/slog/handlers.md:139) to send logs to multiple destinations:

```go
fileHandler := slog.NewJSONHandler(logFile, nil)
consoleHandler := slog.NewTextHandler(os.Stderr, nil)

logger := slog.New(slog.MultiHandler(fileHandler, consoleHandler))
```

This writes every log entry to both handlers, enabling simultaneous file and console output.
