# slog Groups

Groups organize related attributes into hierarchical structures. They are useful for namespacing attributes that share a common context, such as HTTP request details or database connection parameters.

## Creating Groups

Use [`slog.Group`](resources/go/slog/groups.md:5) to create a named collection of attributes:

```go
logger.Info("request completed",
    slog.Group("http",
        "method", "GET",
        "path", "/api/users",
        "status", 200,
    ),
    slog.Group("timing",
        "total_ms", 42,
        "db_ms", 15,
    ),
)
```

With `JSONHandler`, this produces:

```json
{
  "time": "2009-11-10T23:00:00.000000Z",
  "level": "INFO",
  "msg": "request completed",
  "http": {
    "method": "GET",
    "path": "/api/users",
    "status": 200
  },
  "timing": {
    "total_ms": 42,
    "db_ms": 15
  }
}
```

## Group Scoping

Groups can be nested to create deeper hierarchies:

```go
logger.Info("database query",
    slog.Group("db",
        "query", "SELECT * FROM users",
        slog.Group("connection",
            "host", "localhost",
            "port", 5432,
        ),
    ),
)
```

This nesting produces:

```json
{
  "db": {
    "query": "SELECT * FROM users",
    "connection": {
      "host": "localhost",
      "port": 5432
    }
  }
}
```

## Logger-bound Groups

Use [`Logger.WithGroup`](resources/go/slog/groups.md:38) to create a new logger that automatically groups all subsequent attributes:

```go
httpLogger := logger.WithGroup("http")
httpLogger.Info("request started",
    "method", "POST",
    "path", "/api/orders",
)
httpLogger.Info("request completed",
    "method", "POST",
    "path", "/api/orders",
    "status", 201,
)
```

All logs from `httpLogger` place their attributes under the `http` key:

```json
{"http": {"method": "POST", "path": "/api/orders"}}
{"http": {"method": "POST", "path": "/api/orders", "status": 201}}
```

This pattern is useful when creating request-scoped loggers that should consistently group all their output.

## Combining Groups and Attributes

Groups and regular attributes can be mixed freely:

```go
logger.Info("operation",
    "operation_id", "abc-123",
    slog.Group("input",
        "user_id", 42,
        "amount", 100.50,
    ),
    slog.Group("output",
        "success", true,
        "transaction_id", "tx-789",
    ),
    "duration_ms", 25,
)
```

This produces:

```json
{
  "operation_id": "abc-123",
  "input": {
    "user_id": 42,
    "amount": 100.5
  },
  "output": {
    "success": true,
    "transaction_id": "tx-789"
  },
  "duration_ms": 25
}
```

## Text Handler Behavior

With `TextHandler`, groups are represented by prefixing attribute keys with the group name and a dot:

```
time=2009-11-10T23:00:00.000000Z level=INFO msg="request completed" http.method=GET http.path=/api/users http.status=200
```

This flat format preserves the hierarchical information while remaining compatible with systems that expect key-value pairs rather than nested structures.

## Empty Groups

A group with no attributes is still included in the output. With `JSONHandler`, it appears as an empty object:

```go
logger.Info("event", slog.Group("metadata"))
// Output: {"time": "...", "level": "INFO", "msg": "event", "metadata": {}}
```

## Group Value Constructor

Attributes themselves can hold group values using [`slog.GroupValue`](resources/go/slog/groups.md:100):

```go
attrs := []slog.Attr{
    {
        Key: "request",
        Value: slog.GroupValue(
            slog.String("method", "GET"),
            slog.String("path", "/api/users"),
        ),
    },
}
logger.InfoAttrs("completed", attrs...)
```

This is equivalent to using `slog.Group` but provides more control when building attribute collections programmatically.
