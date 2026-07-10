## Types Definition Design

Types are one of the fundamental building blocks of the Chronocone Planning Language (CPL) that provide part of the information about how to represent data in time series in the system (the others being units and measures).

This specification describes the design of the types definition in CPL.

A types section begins with the keyword `types` and is followed by a series of type definitions. Where each type definition is the declaration of a new type name based on an existing type with optional constraints. For example,
the following declaration:

```
types
    BrokerageCurrency = Number(precision=4)
    Currency = BrokerageCurrency(precision=2)
```

Declares two new types. BrokerageCurrency a built-in Number type with an added constraint that the precision (i.e., number of places to the right of the decimal point) is exactly 4. Currency is a new type derived from BrokerageCurrency that adds the constraint that the precision is exactly 2.

Strictly speaking, conversions between types isn't possible. However measures and units declarations may optionally specify calcuations that can
convert measure of one type into measures of a different type.

Chronocone defines a number of built-in types that are always present.

1. **Number** - Number is an unrestricted real number with units of Unitless that can accomodate values that are infinitely large with infinitely fine precision. Note that this is not the same as int or float as defined in many programming languages because those types have limitations on the magnitude and precision of the stored data. The possible constraints on Number are maximum value, minimum value, and precision (i.e., the number of places to the right of the decimal point).
2. **String** - Represents an unrestricted sequence of unicode code points of zero or more length. There is no upper bound on the length of a string. The possible constraints on a String are maximum length, minimum length, and a regular expression that the string of that type must match.
3. **Instant** - Represents one or more UTC moments in time. Every parameter supplied to an `Instant` declaration is a **constraint** on matching moments. Any UTC moment that satisfies all supplied constraints matches that `Instant` declaration. A single `Instant` declaration may match exactly one moment, many moments, or an infinite number of moments. `Instant` has maximum resolution of one second and always uses UTC. Constraints include year, month, day, hour, minute, second, day of week, and week number.
4. **Interval** - Represents a span, range, or extent of time, usually described using `Instant` boundaries or constraints together with a unit reference. The canonical parameter is `starts_at`, which accepts the starting `Instant` for an interval declaration. `Duration` is not a type; it only provides concept values used to define units for `Interval` measures.
5. **Boolean** - Represents a logical value of either true or false.

The syntax for declaring a types section is:

```
types
  <list of type declarations>
```

There may be zero or more type ty pes section in a CPL file or set of CPL files. Type declarations may also exist in the database and be pulled in by the cpl compiler at compilation time.

The syntax for declaring a type is:

```
<type name> = <base type name>(<constraints>)
```

The following constraints exist for each built-in type:

1. Number, which is number with unlimited size and infinite precision
   - maximum: The maximum value that the number may have. Must be larger than the minimum if a minimum constraint is specified or inherited from the base type.
   - minimum: The minimum value that the number may have. Must be smaller than the maximum if a maximum constraint is specified or inherited from the base type.
   - precision: The number of places to the right of the decimal point that a number has. Note that a number with a specified precision always has the specified digits to the right of the decimal place, even if the value is zero. Must be greater than or equal to zero, must be less than or equal to the precision of the base type.
2. String, which represents a possibly infinitely long sequence of Unicode points. The possible constraints on a String are:
   - maximum: The maximum number of Unicode points allowed in the string. Must be larger than the minimum if a minimum constraint is specified or inherited from the base type.
   - minimum: The minimum number of Unicode points allowed in the string. Must be smaller than the maximum if a maximum constraint is specified or inherited from the base type.
   - pattern: A regular expression that must match the content of the string for it to be valid. If a base type declares a pattern the specified pattern may only match strings that are also matched by the base type.
3. Instant, which represents one or more UTC moments in time. Every parameter supplied to an `Instant` declaration is a **constraint** on matching moments. Any UTC moment that satisfies all supplied constraints matches that `Instant` declaration. A single `Instant` declaration may match exactly one moment, many moments, or an infinite number of moments. `Instant` has maximum resolution of one second and always uses UTC. The possible constraints on an `Instant` are:
   - year: An integer year constraint (e.g., `year=2025`).
   - month: A month constraint (e.g., `month=January`).
   - day: A day-of-month constraint (e.g., `day=14`).
   - hour: An hour constraint (e.g., `hour=9`).
   - minute: A minute constraint (e.g., `minute=30`).
   - second: A second constraint (e.g., `second=0`).
   - day_of_week: A day-of-week constraint (e.g., `day_of_week=Monday`).
   - week_number: A week-number constraint (e.g., `week_number=1`).
   - Omitted fields are valid and mean underspecification, which broadens the match set.
4. Interval, which represents a span, range, or extent of time. The canonical parameter is `starts_at`, which accepts either a named `Instant`-derived type reference or an inline `Instant` constraint expression. An `Interval` also references a unit through normal unit declarations. `Duration` is not a type; it only provides concept values used to define units for `Interval` measures.
5. Boolean, which is a true/false value. Booleans have a unit of Unitless.

Note that, particulary for time related types, there is a strong connection between the types, units and measure and a dependency on certain built-in concepts that will be specified in the section on units and measures definitions.

By themselves types have no units associated with them, they are simply a declaration about some properties of how values are represented.

## Related Documents

- [CPL Architecture - Introduction](../../architecture/cpl/introduction.md) - Overall CPL processing lifecycle
- [CPL Requirements - Concepts, Types, Units, Measures](../../requirements/cpl/concepts-types-units-measures-syntax.md) - Syntax requirements
- [CPL Architecture - Intermediate Language](../../architecture/cpl/analysis/intermediate_language.md) - CIL specification
