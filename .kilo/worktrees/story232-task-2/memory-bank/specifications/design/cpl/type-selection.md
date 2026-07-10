# Type Selection for Numeric Calculations

## Overview

This document will contain the detailed design for numeric type selection in CPL metric calculations. The design is deferred and will address two key concerns:

1. **Fixed-point arithmetic** for types with specified precision (e.g., `Number(precision=2)`)
2. **Overflow detection and arbitrary-precision fallback** using the metric dependency DAG

## Topics to Address

### Fixed-Point Representation

When precision is specified in a type declaration (e.g., `Number(precision=2)`), values should use an integer fixed-point type where the integer value is the actual value multiplied by the scale factor (10^precision). This avoids floating-point rounding errors that are unacceptable for financial and planning calculations.

### Overflow Detection via DAG Analysis

Once the metric dependency DAG is assembled, the range of possible values can be statically analyzed to determine whether native integer types (int32, int64) suffice or whether arbitrary-precision arithmetic via Go's `math/big` package is required. When inputs fall within the safe range, native operations proceed. When outside, `math/big` is used in the generated code.

### Related Patterns

- [Fixed-Point Arithmetic and Integer Overflow](../patterns/fixed-point-arithmetic.md) — Patterns document capturing background on fixed-point arithmetic and overflow handling

## Status

**Deferred** — Detailed design will be created after the initial metrics implementation increments are complete. The initial increments use `float64` for all numeric values.
