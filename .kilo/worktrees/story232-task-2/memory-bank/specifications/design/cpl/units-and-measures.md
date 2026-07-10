## Units and Measures Definition Design

Units and measures are intimately related in CPL.

In the expression "20 m/s", the "20" is data that is stored as some type, the "m/s" is the unit, and the combination of the two is the measure. In CPL concepts are used to define units, types and units are used to define measures.

Different units may be combined using `*`, `/` and `^` operations (the resulting units may not make business sense, but they're units nonetheless), but combining units using any other operations is an error unless some explicit unit conversion, possibly over multiple steps, is defined.

The reason for the intimate connection between units and measures is that a conversion between units automatically implies the same conversion for measures that have those units. This is the reason why the syntax for measure and unit declarations is nearly identical.

## Units Definitions

Units are defined in sections that begin with the keyword `units` followed by a set of units declarations that are indented one level. There may be more than one `units` section in a CPL file or group of CPL files.

There is one built-in unit named `Unitless`. A Unitless unit can combine with itself or any other unit using any operations.

Each unit declaration has the form:

```
<unit group>=<concept name>
   conversion
     pattern
       <expression>
     calculation
       <expression>
```

The `unit group` is a group of units that have values that are the same as the concept values for the given concept. For example, the declaration:

```
concepts
  Currency = [Values[USD, EURO, YEN]]

units
  Currency = Currency
```

Declares that there is a unit group named `Currency` that is associated with units named USD, EURO and YEN. It is important to note that the `unit group` is a CPL concept that is a shorthand for referring to a certain set of units. The units themselves are USD, EURO and YEN, not `Currency`. It is permissible (in fact it is likely to be common) for unit group names to be the same as type or measure group names.

Unit declarations may optionally accept one or more pattern/calculation clauses that define how to convert between units. For example, if we defined units that handled temperature conversions we might have something like:

```
concepts
    TemperatureScale = [Values[C, F]]

units
    TemperatureScale = TemperatureScale
        conversion
            pattern
                C(c) = F(f)
            calculation
                ((5 / 9)*(C/F))*(f-32*F)
```

declares that if CPL knows the values of a measure with units
of F and needs a measure with values of C, then it can use the calculation expression to convert between the two. The parameter in a pattern expression like `<unit name>(<identifier>)` defines an alias for the value of a measure that uses this unit that may be used in the calculation expression.

#### Temperature Conversion Formula Syntax

In the temperature conversion calculation `((5 / 9)*(C/F))*(f-32*F)`:

- `C` and `F` (uppercase) are **unit names** used in unit expressions to represent the units themselves
- `c` and `f` (lowercase) are **value aliases** that refer to the numeric values of measures using those units
- The expression `(C/F)` is a complex unit (i.e., a combination of other units). `(C/F)` is not unitless, C and F are different units and there is no way with the rules given to reduce it further. Under the rules of unit algebra it would be possible to multiply a measure with units of (C/F) by a measure with units of F to get a measure with units of C.
- The expression `f-32*F` means: take the value `f` (the Farenheit measure value), subtract `32*F` (32 degrees Farenheit expressed as a measure), resulting in a value that is 32 degrees less also with units of F.
- Multiple `(5/9)*(C/F)` (with units C/F) by (f-32\*F) with units F, resulting in a value with units C.

This syntax allows the conversion formula to be both mathematically correct and dimensionally consistent.

In this case the calculation uses the measure values associated with the F unit to say that 32 must be subtracted from the value of the measure with the F unit before the result of that subtraction is multiplied by 5/9.

Note that a conversion from F to C does not imply that a conversion from C to F is allowed, unless the pattern can be manipulated using `*`, `/` and `^`. operations to create a conversion from C to F. Because of the presence of the `+` term in this conversion, that's not possible in this case but there are some situations where it is possible. For example, imagine that we were converting between units of currency, thus:

```
concepts
    Currency = [Values[USD, EURO]]

units
    Currency = Currency
        conversion
            pattern
                USD(usd)/EURO(euro)
            calculation
                0.8
```

Also imagine that we can get 1.25 Euros for every 1 USD. Thus the number of `dollars/euro = 1/1.25 = 0.8`. Note also that unlike in the TemperatureScale example, the calculation expression doesn't use unit values (or unit group values) in the calculation expression, which means that the result has units of Unitless. Since Unitless will combine with any other unit, it is possible for conversion calculations to ignore units, the best practice is to explicitly include unit or unit group values in the calculation expression to make it possible to automatically detect subtle errors in conversions.

Now imagine that in some part of a metric calculation we have a measure where the desired measure has units of USD and the input measure has units of EURO. It is possible to manipulate the relationship `X USD = 27 EURO` to be `(X USD)/(27 EURO) = (X USD/27 EURO) Unitless`, where Unitless is a built-in unit. If there was a unit conversion pattern that had units of `USD/EURO` resulting in a Unitless measure, then it would usable for that calculation.

While naming low level units in conversion patterns and calculations is allowed, it could become very verbose. Imagine that there is a lookup that will provide the conversion values between any currency and any other currency (including the 1:1 conversion for a currency conversion to itself). In that case we might have the following definition:

```
concepts
   Currency = [Values[USD, EURO, YEN]]

units
   Currency = Currency
       conversion
           pattern
               Currency(c1)/Currency(c2)
           calculation
               conversion_lookup(c1, c2)
```

Using a unit group in a pattern or expression defines that expression for every unit in the unit group just as if the expression was written out for each unit in the unit group. As with naming individual units, naming a unit group in a pattern or expression defines an alias for the value of a measure that uses this unit that may be used in the calculation expression.

Calculation expressions may include any operations, arithmetic or boolean, that are valid for the types of the measures that use those units provided that those result in well-defined units (e.g., different clauses of a boolean statement can't produce different units). All `unit algebra`, that is operations/representation of the units themselves either in pattern expressions or implicitly in unit combinations must be expressible using `*`, `/` or `^`
operators. Combinations of units are preserved as a symbolic
formula. While they may be simplified or converted to other
units using conversion expressions combinations of units will
typically have no representation that is simpler than the
expression itself.

When units are defined in terms of other units the units or
unit groups on the right will be considered to be more
fundamental. Cyclical unit definitions are an error. While it
is possible that some sets of derived units might have no business meaning (and will generally be associated with mistakes in the formulation of metrics), unusual unit combinations are not, in and of themselves, errors. If there is no conversion path between units that are combined using other than `*`, `/` or `^` operations, it is an error.

## Compound Units

Units can be combined using `*`, `/`, and `^` operators to form compound units. These combinations are represented as symbolic algebraic expressions.

### Basic Compound Unit Examples

```
concepts
    Length = [Meter, Kilometer, Centimeter]
    Time = [Second, Minute, Hour]

units
    Length = Length
        conversion
            pattern
                Kilometer(km) / Meter(m)
            calculation
                1000
            pattern
                Centimeter(cm) / Meter(m)
            calculation
                0.01

    Time = Time
        conversion
            pattern
                Hour(h) / Second(s)
            calculation
                3600
            pattern
                Minute(m) / Second(s)
            calculation
                60

measures
    Speed = Number(unit=Length/Time)
    Acceleration = Number(unit=Length/Time^2)
    Area = Number(unit=Length^2)
    Volume = Number(unit=Length^3)
```

In this example:

- `Length/Time` represents speed (meters per second, kilometers per hour, etc.)
- `Length/Time^2` represents acceleration (meters per second squared)
- `Length^2` represents area (square meters, square kilometers)
- `Length^3` represents volume (cubic meters, cubic centimeters)

### Derived Units

Compound units can also be used to define derived units in terms of more fundamental units:

```
concepts
    Mass = [Kilogram, Gram]
    Length = [Meter]
    Time = [Second]

units
    Mass = Mass
        conversion
            pattern
                Gram(g) / Kilogram(kg)
            calculation
                0.001

    Length = Length
    Time = Time

    // Derived unit: Newton = kg * m / s^2
    Force = Mass * Length / Time^2

measures
    MassMeasure = Number(unit=Mass)
    LengthMeasure = Number(unit=Length)
    TimeMeasure = Number(unit=Time)
    ForceMeasure = Number(unit=Force)
```

The `Force` unit group is defined as a compound of `Mass * Length / Time^2`. When measures using these units are combined in calculations, the unit algebra is automatically tracked.

### Unit Simplification and Conversion

When compound units are involved in conversions, the unit algebra is preserved:

```
// Converting 100 km/h to m/s:
// 100 * (Kilometer / Hour) * (1000 Meter / 1 Kilometer) * (1 Hour / 3600 Second)
// = 100 * 1000 / 3600 * (Meter / Second)
// ≈ 27.78 m/s
```

Compound units are simplified when possible using the defined conversion relationships, but they retain their symbolic form to maintain dimensional correctness.

## Pattern Syntax Semantics

CPL uses a consistent convention where **more fundamental definitions appear on the right-hand side**, while **more derived definitions appear on the left-hand side**.

### Unit Definition Direction

In unit declarations:

```
units
    TemperatureScale = TemperatureScale
        conversion
            pattern
                C(c) = F(f)
            calculation
                ((5 / 9)*(C/F))*f-32*F
```

- The pattern `C(c) = F(f)` indicates we are converting **from F (right side) to C (left side)**
- F is the more fundamental unit in this conversion context
- C is the derived/target unit we want to obtain
- The calculation expression uses `f` (the F value) to compute the C value

### Measure Conversion Direction

In measure declarations:

```
measures
  Day = Interval(starts_at = Midnight, unit = Duration)
    conversion
      pattern
        Hour=Day(other_day)
      calculation
        when
          other_day = 03-09-2025 => 23
          other_day = 11-02-2025 => 25
          default => 24
        end when
```

- The pattern `Hour=Day(other_day)` indicates we are converting **from Day (right side) to Hour (left side)**
- Day is the source measure with its value captured in `other_day`
- Hour is the target measure we're computing
- `Duration` is a concept that supplies values for units; it is **not a type**

### Self-Referential Patterns

In some cases, a unit or unit group may appear on both sides of a pattern expression. This is valid when:

1. The relationship being defined is non-cyclical
2. The most fundamental unit is expressed as a compound of other units
3. The cyclic relationship is broken by the input data. If the input data is in meters, then that will effectively define the more fundamental unit.

For example:

```
units
    Length = Length
        conversion
            pattern
                Meter(m) / Kilometer(km)
            calculation
                0.001

```

Here `Length` (the unit group) appears in both definitions, but there's no cycle because:

- `Meter` is the fundamental unit in the `Length` group
- `Kilometer` is derived from `Meter`, not the other way around

This convention ensures that conversion paths can be traced from more derived units/measures to more fundamental ones.

## Measures Definitions

Measure definitions are very similar to unit definitions, except that they reference both types and units.

For example, in order to actually use temperature values as a cell type for actuals or forecasts in some time series, we would need to declare them as a measure, thus:

```
concepts
    TemperatureScale = [Values[C, F]]

units
    TemperatureScale = TemperatureScale
        conversion
            pattern
                C(c) = F(f)
            calculation
                ((5 / 9)*(C/F))*f-32*F

measures
    Temperature = Number(unit=TemperatureScale)
```

Measures are defined in sections that begin with the keyword `measures` followed by a set of measure declarations that are indented one level. There may be more than one `measures` section in a CPL file or group of CPL files.

The unit conversion pattern and calculation for units of C and F will automatically be defined for the Temperature measure that has those units.

A measure declaration has the form:

```
<measure group name> = <type name>(unit=<unit group name>)
    conversion
        pattern
            <expression>
        calculation
            <expression>
```

where a measure is defined for every unit in the associated unit group name.

Like units, measures may optionally define pattern and calculation clauses that define the conversion between measures. Since measures capture the idea of both an explicit value in combination with a unit, measure conversions make sense when there are conversions that aren't solely dependent on relative units. For example, in the case of a days to hours conversion that takes account of daylight savings time, the following measure conversion might be used:

```
concepts
  Duration=[Year, Quarter, Month, Week, Day, Hour, Minute, Second]

units
  Duration = Duration

measures
  Day = Interval(starts_at = Midnight, unit = Duration)
    conversion
      pattern
        Hour=Day(other_day)
      calculation
        when
          other_day = 03-09-2025 => 23
          other_day = 11-02-2025 => 25
          default => 24
        end when
```

There are a number of subtleties around handling time that are beyond the scope of this document. In this case the constants `03-09-2025` and `11-02-2025` refer to intervals that have the date constraint given and are Intervals that begin at midnight on the date given and end just prior to the midnight on the following day. Time is complex enough that it needs its own design document; see [Talking About Time](../../requirements/cpl/talking-about-time.md) for the corrected time model.

As with units conversions, the patterns may use any arithmetic operation that uses `*` and `/`, and the calculation may use any boolean or arithmetic operation that is defined for the type of the measure.

## Relationship Diagram

The following diagram illustrates the relationships between concepts, unit groups, units, types, measure groups, and measures:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CPL UNITS AND MEASURES                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────┐         defines          ┌──────────────────────┐     │
│  │   CONCEPT       │─────────────────────────▶│    UNIT GROUP        │     │
│  │  [TemperatureScale│                        │  [TemperatureScale]  │     │
│  │   Values[C,F]]  │                         │                      │     │
│  └─────────────────┘                         └──────────────────────┘     │
│           │                                            │                    │
│           │                                            │ contains           │
│           │                                            ▼                    │
│           │                                   ┌──────────────────────┐     │
│           │                                   │       UNITS          │     │
│           │                                   │   C       F          │     │
│           │                                   │   │       │          │     │
│           │                                   │   ▼       ▼          │     │
│           │                                   │ ┌───┐   ┌───┐        │     │
│           │                                   │ │unit│  │unit│        │     │
│           │                                   │ └───┘   └───┘        │     │
│           │                                   └──────────────────────┘     │
│           │                                            │                    │
│           │                                            │ used by            │
│           ▼                                            ▼                    │
│  ┌─────────────────┐         defines          ┌──────────────────────┐     │
│  │      TYPE       │─────────────────────────▶│   MEASURE GROUP      │     │
│  │    [Number]     │                          │    [Temperature]     │     │
│  └─────────────────┘                          │   unit=TemperatureScale  │ │
│                                               │                      │     │
│                                               │   Temperature =      │     │
│                                               │   Number(unit=C)     │     │
│                                               │   Number(unit=F)     │     │
│                                               └──────────────────────┘     │
│                                                                             │
│  CONVERSION FLOW:                                                           │
│                                                                             │
│   More Derived ◄───────────────────────────────────── More Fundamental      │
│                                                                             │
│        C  ◄──────────────────────────────  F                                │
│   (target unit)                         (source unit)                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Relationships

1. **Concept → Unit Group**: A concept defines the values that become units in a unit group
2. **Unit Group → Units**: The unit group contains individual units derived from the concept values
3. **Type + Unit Group → Measure Group**: A measure group combines a type with a unit group
4. **Measure Group → Measures**: Creates one measure for each unit in the unit group
5. **Conversions**: Flow from fundamental (right side) to derived (left side)

## Data Flow

The flow of units and measures definitions through the CPL processing pipeline follows the same pattern as concepts. For the overall data flow, see the [Data Flow diagram in Concept Definitions](concepts-definitions.md#data-flow).

At each processing phase, units and measures are handled as follows:

1. **Parser Phase**: Units and measures declarations are parsed from CPL source and stored in the symbol table with their associated patterns and calculations
2. **Analysis Phase**: Unit groups, measure groups, and conversion patterns are validated and assembled into lookup structures
3. **CIL Generation**: Units and measures are embedded in CIL definitions for use during synthesis
4. **Synthesis Phase**: Conversion calculations are incorporated into generated code
5. **Runtime**: Measures with units are evaluated using the defined conversion rules

### Cyclical Unit Definitions

```
concepts
    Length = [Meter, Kilometer]

units
    Length = Length
        conversion
            pattern
                Meter(m) / Kilometer(km)
            calculation
                1000 * km  // ERROR: Cyclical dependency
        conversion
            pattern
                Kilometer(km) / Meter(m)
            calculation
                m / 1000   // ERROR: Cyclical dependency
```

While this is, strictly speaking, a cyclical definition, it's not a problem in practice. Inputs will define the units of the input and therefore will effectively break cycles.

### More Specific Conversions Take Precedence

```
concepts
    Currency = [USD, EURO, YEN]

units
    Currency = Currency
        conversion
            pattern
                USD(usd) / EURO(euro)
            calculation
                0.85
        conversion
            pattern
                Currency(c1) / Currency(c2)
            calculation
                lookup(c1, c2)
```

When converting USD to EURO, both patterns could apply. The specific USD/EURO pattern conflicts with the general Currency/Currency pattern. In such cases the more specific pattern takes precedence, which means that the pattern that names specific units takes precedence over the pattern that uses a unit group. Thus, for conversions that have units of `USD/EURO` (directly or through unit algebra) the first conversion will used, all others will use the second conversion.

## Common Error Examples

The following examples illustrate common errors in units and measures definitions:

### Error 1: Invalid Unit Combinations/Operations

```
concepts
    Length = [Meter, Kilometer]
    Time = [Second, Hour]

measures
    Distance = Number(unit=Length)
    Duration = Number(unit=Time)

// Somewhere in a metric calculation:
// ERROR: Adding Length + Time is invalid
// speed = distance + duration  // Type error!
```

**Problem**: The `+` operation cannot combine measures with incompatible units. Unlike `*` and `/` which produce compound units, `+` requires identical units.

**Solution**: Ensure operations are dimensionally consistent:

```
// Valid: Different Length units can be added
// total_distance = 5 Kilometer + 500 Meter

// Valid: Length/Time produces speed
// speed = 100 Kilometer / 1 Hour
```

### Error 2: Conflicting Unit or Measure Conversions At The Same Level

```
concepts
    Currency = [USD, EURO, YEN]

units
    Currency = Currency
        conversion
            pattern
                Currency(c1)/Currency(c2)
            calculation
                0.85
        conversion
            pattern
                Currency(c1)=Currency(c2)
            calculation
                1.7
        conversion
            pattern
                Currency(c1) / Currency(c2)
            calculation
                lookup(c1, c2)
```

**Problem**: There are two or more definitions for currency conversion at the same level (in the example they're all at the unit group level, not mentioning specific units) that either directly or via valid unit algebra match to the same pattern.

**Solution**: Ensure that unit or measure conversions at the same level are always defined in a single conversion, possibly using conditional logic in the calculation clause.

## Time-Related Built-ins

Time related built-in concepts, types and units will be defined in a separate document. The examples in this document that relate to time should be considered illustrative of the syntax and semantics of the language and not definitive of the time-related concepts, types and units.

See also [Talking About Time](../../requirements/cpl/talking-about-time.md) for requirements discussion of temporal concepts.

## General Units and Measure Conversion Rules

1. Unitless can be combined with any other unit, including itself using arbitrary operations. However, since these are potentially prone to errors, the cpl compiler will issue a warning if it finds any such operations if a flag to the compiler has enabled them.
2. In the presence of multiple unit or measure conversions with equivalent patterns (that is, where after algebraic manipulation with `*` and `/` operators) where some are unit groups or measure groups and some are units or measures the more specific definition takes precedence. For example, the conversion that names a specific unit will be used in preference to the conversion that names a unit group.
3. In the presence of multiple unit or measure conversions with equivalent patterns that are at the same level (e.g., both are unit groups or both are units) it is an error.
4. If a measure level conversion and a unit level conversion could apply given the patterns specified, the more specific conversion takes precedence, which would be the measure level conversion.

### Conversion Precedence Hierarchy

When multiple conversions could apply to a given unit conversion scenario, the following precedence rules determine which conversion is used (highest to lowest):

| Precedence | Conversion Type             | Description                                                        |
| ---------- | --------------------------- | ------------------------------------------------------------------ |
| 1          | Specific Measure Conversion | A measure conversion naming specific units (e.g., `USD/EURO`)      |
| 2          | Specific Unit Conversion    | A unit conversion naming specific units (e.g., `USD/EURO`)         |
| 3          | Measure Group Conversion    | A measure conversion using unit groups (e.g., `Currency/Currency`) |
| 4          | Unit Group Conversion       | A unit conversion using unit groups (e.g., `Currency/Currency`)    |

**Key Points:**

- More specific definitions (naming actual units) always take precedence over general definitions (using unit groups)
- Measure-level conversions take precedence over unit-level conversions at the same specificity
- When multiple conversions exist at the same precedence level that could apply to the same pattern, this is an error

## Unit Conversion Path Resolution

When a direct conversion between two units is not defined, CPL can automatically resolve conversion paths through intermediate units.

### Multi-Step Conversion Example

Consider the following definitions:

```
concepts
    Length = [Meter, Kilometer, Mile]

units
    Length = Length
        conversion
            pattern
                Kilometer(km) / Meter(m)
            calculation
                1000
        conversion
            pattern
                Mile(mi) / Meter(m)
            calculation
                1609.34
```

In this scenario:

- A direct conversion from `Kilometer` to `Mile` is **not** defined
- However, CPL can resolve the path: `Kilometer → Meter → Mile`

When converting 5 Kilometers to Miles:

1. First convert: 5 km → 5,000 m (using km/m conversion)
2. Then convert: 5,000 m → 5,000/1609.34 mi ≈ 3.11 mi (using mi/m conversion)

### Path Resolution Rules

1. **Fundamental Unit Anchor**: Conversions are always expressed relative to a fundamental unit (appearing on the right side of patterns or appearing in input values). In the example above, `Meter` is the fundamental unit.

2. **Bidirectional Traversal**: Since `Kilometer/Meter` is defined, CPL can derive both directions provided that the conversions are expressed using only `*`, `/` and `^` operations:
   - `Kilometer → Meter` (forward)
   - `Meter → Kilometer` (reverse via algebraic manipulation)

3. **Compound Unit Paths**: Path resolution also works for compound units:

   ```
   // If Meter/Second and Kilometer/Hour are defined,
   // CPL can convert between them via:
   // m/s → (m→km)/(s→h) → km/h
   ```

4. **No Path Available**: If no conversion path exists between the source and target units, this is a compile-time error:

   ```
   // ERROR: No conversion path defined
   // Cannot convert between Temperature and Length
   ```

5. **Ambiguous Paths**: If multiple valid conversion paths exist with different results, this is an error:
   ```
   // If both paths exist:
   // A → B → C (result: 10)
   // A → D → C (result: 12)
   // This is an error - ambiguous resolution
   ```

### Best Practices

- Define conversions relative to a consistent fundamental unit within each dimension
- Avoid creating multiple paths between the same units with different conversion factors
- Use unit groups for consistent conversion logic across related units

## Related Documents

- [CPL Architecture - Introduction](../../architecture/cpl/introduction.md) - Overall CPL processing lifecycle
- [CPL Requirements - Concepts, Types, Units, Measures](../../requirements/cpl/concepts-types-units-measures-syntax.md) - Syntax requirements
- [CPL Architecture - Intermediate Language](../../architecture/cpl/analysis/intermediate_language.md) - CIL specification
- [Concept Definitions](concepts-definitions.md) - Concept definition design and data flow diagram
