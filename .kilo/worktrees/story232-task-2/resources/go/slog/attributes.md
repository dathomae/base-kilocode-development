# slog Attributes and Values

Attributes are the fundamental building blocks of structured logging in slog. Each attribute consists of a string key and a typed value, forming key-value pairs that provide context for log messages.

## Creating Attributes

The [`slog.Attr`](resources/go/slog/attributes.md:5) type represents a single attribute. The simplest way to create attributes is using the constructor functions:

```go
slog.String("method", "GET")
slog.Int("status", 200)
slog.Bool("cached", true)
slog.Duration("elapsed", time.Second)
slog.Time("created", time.Now())
slog.Any("data", myStruct)
```

Each constructor produces an attribute with the appropriate type preserved in the [`slog.Value`](resources/go/slog/attributes.md:15).

## The Value Type

The [`slog.Value`](resources/go/slog/attributes.md:17) type holds attribute data with one of several kinds:

| Kind      | Go Type                | Constructor     |
| --------- | ---------------------- | --------------- |
| Bool      | `bool`                 | `BoolValue`     |
| Duration  | `time.Duration`        | `DurationValue` |
| Float64   | `float64`              | `Float64Value`  |
| Int64     | `int64`                | `Int64Value`    |
| String    | `string`               | `StringValue`   |
| Time      | `time.Time`            | `TimeValue`     |
| Uint64    | `uint64`               | `Uint64Value`   |
| Any       | `any` (interface)      | `AnyValue`      |
| Group     | `[]Attr`               | `GroupValue`    |
| LogValuer | Implements `LogValuer` | (any)           |

## Automatic Type Resolution

When using logger methods with alternating key-value arguments, slog automatically resolves types:

```go
logger.Info("request",
    "count", 42,           // becomes Int64
    "ratio", 3.14,         // becomes Float64
    "name", "example",     // becomes String
    "ok", true,            // becomes Bool
)
```

The resolution rules handle common types as follows:

- `string` → `KindString`
- `int`, `int8`, `int16`, `int32`, `int64` → `KindInt64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr` → `KindUint64`
- `float64` → `KindFloat64` (note: `float32` is converted to `float64`)
- `bool` → `KindBool`
- `time.Time` → `KindTime`
- `time.Duration` → `KindDuration`
- `[]byte` → base64-encoded string
- `error` → error message string via `Error()` method
- structs, maps, slices → `KindAny` (passed through for handlers to process)

## The LogValuer Interface

Types implementing [`LogValuer`](resources/go/slog/attributes.md:56) can customize their log representation:

```go
type LogValuer interface {
    LogValue() Value
}
```

This is useful for controlling how complex types appear in logs:

```go
type UserID int64

func (id UserID) LogValue() slog.Value {
    return slog.StringValue(fmt.Sprintf("user_%d", id))
}

userID := UserID(12345)
logger.Info("login", "user", userID)  // logs: "user": "user_12345"
```

Implementing `LogValuer` also prevents accidental logging of sensitive fields by explicitly choosing what to expose.

## Logging Errors

The `error` type receives special handling. When an error is passed as an attribute value, slog extracts the error message:

```go
err := fmt.Errorf("connection refused")
logger.Error("request failed", "error", err)
// logs: "error": "connection refused"
```

For structured error types that implement `LogValuer`, the custom representation takes precedence over the error message extraction.

## Attribute Collections

Logger methods accept attributes in several forms:

**Alternating key-value pairs:**

```go
logger.Info("msg", "key1", val1, "key2", val2)
```

**Pre-built attributes:**

```go
attrs := []slog.Attr{
    slog.String("key1", val1),
    slog.Int("key2", val2),
}
logger.InfoAttrs("msg", attrs...)
```

The `InfoAttrs`, `DebugAttrs`, `WarnAttrs`, and `ErrorAttrs` methods accept only `slog.Attr` values, providing a type-safe alternative when attributes are pre-built.

## Pre-formatting Attributes

For expensive computations that should only run when the log level is enabled, use the `LogValuer` interface or guard the call with `Enabled`:

```go
if logger.Enabled(ctx, slog.LevelDebug) {
    expensiveData := computeExpensiveData()
    logger.Debug("details", "data", expensiveData)
}
```

This pattern avoids unnecessary work when the log level would discard the entry anyway.
