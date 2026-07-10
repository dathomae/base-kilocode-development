# Fixed-Point Arithmetic and Integer Overflow Pattern

## Context

The CPL (Chronocone Planning Language) supports type declarations with precision constraints (e.g., `Number(precision=2)` for currency values). When precision is specified, values must be represented without floating-point rounding errors. Financial and planning calculations require exact decimal arithmetic, which floating-point (`float64`) cannot guarantee.

The current implementation uses `float64` for all numeric values, but the design calls for fixed-point integer types when precision is specified, with overflow detection and fallback to arbitrary-precision arithmetic.

## Problem

Three distinct challenges arise when moving from `float64` to precise numeric types:

1. **Rounding errors**: Floating-point cannot exactly represent many decimal values (e.g., 0.1 is a repeating fraction in binary). Accumulated rounding errors in multi-step calculations can produce incorrect planning outputs.
2. **Range constraints**: Native integer types (`int32`, `int64`) have fixed ranges. Calculations on large or deeply nested values may exceed these ranges.
3. **Overflow silence**: Go does not check for integer overflow at compile time or runtime. Overflow results in silent wraparound — incorrect values with no error indication.

## Solution

### Fixed-Point Representation via Integer Scaling

A fixed-point number stores a value as an integer that is implicitly multiplied by a fixed scaling factor. For example, the value 1.23 is stored as the integer 123 with a scaling factor of 1/100 (or equivalently, with 2 fractional digits).

The scaling factor is determined by the type's precision: `Number(precision=n)` implies a scaling factor of 10^n, meaning values are stored as integers representing the actual value × 10^n.

In binary fixed-point, the scaling factor is a power of 2, which improves efficiency on binary hardware. However, for financial applications where decimal rounding rules apply, decimal scaling factors (powers of 10) are preferred for correctness.

### Fixed-Point Arithmetic Operations

**Addition and Subtraction**: Both operands must share the same scaling factor. If they differ, one must be rescaled before the operation. The result inherits the common scaling factor:

```
a (scale S) + b (scale S) = (a + b) (scale S)
```

**Multiplication**: The result's scaling factor is the product of the input scaling factors (or, equivalently, the result has the sum of the input fractional digits). The raw product must be divided by the scaling factor to normalize:

```
a (scale Sa) × b (scale Sb) = (a × b) / (Sa × Sb) at the desired output scale
```

This division introduces a rounding decision. For financial calculations, rounding must follow prescribed rules (e.g., round-half-even for bankers' rounding).

**Division**: The result's scaling factor is the dividend scale divided by the divisor scale. To preserve precision, the dividend is typically scaled up before division:

```
a (scale Sa) ÷ b (scale Sb): scale a up, then divide → (a × factor) / b at the desired output scale
```

### Overflow Detection in Go

Go's integers silently wrap on overflow. For planning calculations this is unacceptable — overflow must be detected and handled explicitly. The pattern is:

1. **Pre-check before operation**: Before adding or multiplying, check whether the result would exceed the type's bounds. For addition of two positive `int64` values:

```go
func Add64(left, right int64) (int64, error) {
    if right > 0 && left > math.MaxInt64-right {
        return 0, ErrOverflow
    }
    if right < 0 && left < math.MinInt64-right {
        return 0, ErrOverflow
    }
    return left + right, nil
}
```

2. **Post-operation detection**: After an operation, check whether the sign or magnitude indicates overflow. This is less reliable than pre-checks.

### Arbitrary-Precision Fallback via math/big

When values exceed the range of native integer types, Go's [`math/big`](https://pkg.go.dev/math/big) package provides arbitrary-precision integer arithmetic. The trade-off is performance: `math/big` operations are significantly slower than native integer operations.

```go
func SafeMul64(a, b int64) (*big.Int, error) {
    result := new(big.Int).Mul(big.NewInt(a), big.NewInt(b))
    if result.IsInt64() {
        return result, nil // Fits in int64
    }
    return result, nil // Requires big.Int for storage
}
```

### Static Range Analysis via the Metric Dependency DAG

Rather than checking for overflow at runtime on every operation, the CPL synthesis phase can perform static analysis on the metric dependency DAG. Given the known ranges of input values (from connector dimension cardinalities and IO specifications), the analyzer can determine the worst-case range for each metric's output value and select the appropriate numeric type:

- **Safe range**: If the maximum possible value fits in `int64`, emit native integer operations with overflow pre-checks.
- **Overflow possible**: If the static analysis shows potential overflow, emit `math/big` operations for the affected metric calculations.
- **Mixed**: Emit native operations with runtime overflow detection that falls back to `math/big` only when overflow actually occurs.

## Consequences

### Benefits

| Aspect                    | Fixed-Point (`float64`)                  | Fixed-Point (Integer with Scale)           |
| ------------------------- | ---------------------------------------- | ------------------------------------------ |
| **Precision**             | Approximate — rounding errors accumulate | Exact — no rounding for represented values |
| **Financial correctness** | Unacceptable for currency                | Correct, with proper rounding rules        |
| **Performance**           | Fast, hardware-supported                 | Fast for native int; slower for `math/big` |
| **Overflow safety**       | Infinity/NaN propagation                 | Explicit detection required                |
| **Range**                 | ~1.8×10^308                              | int64: ~9.2×10^18; `math/big`: unlimited   |

### Trade-offs

- **Complexity**: Fixed-point arithmetic requires explicit scale management in every operation.
- **Performance cliff**: Falling back to `math/big` introduces a significant performance penalty. The static analysis should minimize occurrences.
- **Rounding rules**: Division and rescaling require explicit rounding decisions. These must match the financial domain's requirements.
- **Mixed precision**: When operands have different scales, the code must handle rescaling, which can introduce rounding at unexpected points.

## When to Use

Use this pattern when:

- Numeric types have specified precision (e.g., `Number(precision=2)` for currency)
- Floating-point rounding errors are unacceptable
- Calculations involve financial values, planning quantities, or other exact-decimal domains
- The system can statically analyze value ranges to optimize between native and arbitrary-precision arithmetic

Do not use this pattern when:

- The type has no precision constraint (unconstrained `Number` can use `float64`)
- Approximate values are acceptable (e.g., statistical aggregates, machine learning features)
- Performance is paramount and exactness is not required

## Related Patterns

- [Postfix Expression Parsing](postfix-expression-parsing.md) — The expression parsing pattern that feeds operands and operators to the analysis phase, where type selection decisions are applied.
- [Symbol Table-Driven Parsing](symbol-table-driven-parsing.md) — The overarching pattern for CPL parsing; type information flows through the symbol table to the analysis phase.

## References

- [Fixed-point arithmetic — Wikipedia](https://en.wikipedia.org/wiki/Fixed-point_arithmetic) — Comprehensive overview of fixed-point representation, operations, and applications.
- [Integer Overflows in Golang — rene's blog](https://blog.rene.sh/blog/2020/06/22/int-overflow/) — Practical treatment of overflow detection strategies in Go, including pre-check patterns and `math/big` usage.
- [Type Selection for Numeric Calculations](../design/cpl/type-selection.md) — Deferred design document for CPL numeric type selection.
- [Metrics Design](../design/cpl/metrics-design.md) — The metrics design document; the "Deferred: Numeric Type Selection" section references this pattern.
