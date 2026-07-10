# Lookups

This document specifies requirements for lookups, which provide "side data" to metric calculations — data that maps input dimensions (from concepts or time grains) to values. Lookups are used for items such as conversion factors, weighting tables, and dimension mappings.

## Lookup Syntax

Lookups follow the general pattern for other parts of plan specifications. Lookups are defined in one or more `lookups` sections, followed by one or more lookup declarations.

A `lookups` section is a syntactic container. It groups lookup declarations and establishes a scope boundary, but it does not itself carry semantic meaning beyond grouping. The compiler validates lookup declarations individually; the presence or absence of multiple `lookups` sections does not change the semantics of the declarations within them.

A lookup declaration consists of

```
  <lookup name>
    dimensions [<list of concept names>]
    [grain <measure name>[, <measure name>...]]
    celltype <measure name>
```

The `dimensions` clause lists the concept names that serve as keys into the lookup — analogous to the key columns of a database table. The `grain` clause is optional; when present, it references one or more measure names, separated by commas, whose types must be `Interval` or derived from `Interval`. The `celltype` clause references a single measure name that defines the type and units of the values returned by the lookup.

When a lookup has no `grain` clause, it is interpreted as having no time dimension — the data returned for a given set of concept dimension values does not vary with time. When a `grain` clause is present, the lookup behaves like a slowly-changing reference table whose values may differ at different points in time.

### Time Values in Lookup Keys

When a lookup declares a `grain`, the grain dimensions function as part of the lookup key alongside the concept dimensions. The Chronocone calculation framework uses the time coordinate of the requesting cell to match against the grain dimensions to select the appropriate row from the lookup.

For example, the following lookup declaration defines a full currency conversion matrix — given a source currency and a target currency, the lookup returns the conversion factor between them:

```
concepts
    Currency = [Values[USD, EURO, YEN]]

units
    Currency = Currency

types
    ConversionFactor = Number(precision=6)

measures
    ConversionRate = ConversionFactor(unit=Currency/Currency)

lookups
    currency_conversion
        dimensions [Currency, Currency]
        celltype ConversionRate
```

The two `Currency` entries in the `dimensions` clause are interpreted as (source\_currency, target\_currency) by convention: the first dimension listed is the source, the second is the target. The `celltype` measure `ConversionRate` has units of `Currency/Currency` — for any given row in the lookup, the unit pair composes to a specific compound unit such as `EURO/USD`. The compound units allow the calculation framework to verify that multiplying a value denominated in the source currency by this conversion rate correctly produces a value in the target currency.

A lookup without a `grain` (a "timeless" lookup) might define a product category mapping — a genuinely unitless example:

```
types
    CategoryName = String

measures
    CategoryLabel = CategoryName(unit=Unitless)

lookups
    product_category
        dimensions [Product]
        celltype CategoryLabel
```

where `CategoryLabel` is a `String` measure. Since `String` has `Unitless` units, this lookup requires no compound unit analysis.

A temporal lookup with multiple grain measures might define an exchange rate table keyed by currency pair and time at two levels of granularity:

```
concepts
    Duration = [Values[Second, Minute, Hour, Day, Week]]

types
    MondayMidnight = Instant(hour = 0, minute = 0, second = 0, day_of_week = Monday)
    WeekInterval = Interval(starts_at = MondayMidnight, unit = Duration)
    MonthInterval = Interval(starts_at = MondayMidnight, unit = Duration)

units
    Duration = Duration

measures
    Week = WeekInterval(unit=Duration)
    Month = MonthInterval(unit=Duration)

lookups
    exchange_rate
        dimensions [Currency, Currency]
        grain Week, Month
        celltype ConversionRate
```

This lookup can satisfy requests at either a weekly or monthly grain, using the time coordinate of the requesting cell to select the appropriate row.

### Constraints

1. Every concept name in the `dimensions` clause must be a valid concept name in the compilation unit.
2. When a `grain` clause is present, every grain measure's type must be `Interval` or derived from `Interval`.
3. The measure name in the `celltype` clause must be a valid measure in the compilation unit.

## Lookup IO Specification

A lookup declaration describes the shape of the lookup data (its dimensions, optional grain, and cell type), but does not specify its source, format, or how to obtain the information. Those elements are specified in the _Lookup IO Specification_ that is passed to the CPL compiler.

The Lookup IO Specification uses its own schema, separate from the connector IO specification. In practice this schema will closely resemble the connector IO spec, but keeping them independent allows each to evolve as use cases demand without forcing a premature unification. They can be merged later if experience shows the schemas are effectively identical.

## Lookup Data as Metric Computations

Lookup data loaded from external sources (for example, static files or slowly-changing reference databases) is registered as `MetricComputation` objects on the `ChronoconeComputationController`. During synthesis, the generated compute function pre-registers every lookup input as a `MetricComputation` instance before the postfix evaluator runs. This registration ensures that `controller.GetMetricComputation(lookupName)` returns a valid handle at runtime, allowing lookup data to participate in metric calculations in the same manner as connector inputs and computed metrics.

The runtime representation of a lookup is a tensor indexed by its dimensions (concept dimensions plus any grain dimensions), with cell values of the type specified by `celltype`. For lookups without a grain, the tensor has no time axis — it is purely dimensionally indexed.

A request for a lookup value with a key (dimensional combination) that does not exist in the source data is a runtime error. There is no default value mechanism; the caller is responsible for ensuring that the lookup covers the required key space.
