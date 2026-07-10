# Connectors

This document specifies requirements for connectors used to integrate ChronoCone with external systems and data sources.

## Connector Syntax

Connectors follow the general pattern for other parts of plan specifications. Connectors are defined in one or more `connectors` sections, followed by one or more connector declarations.

A `connectors` section is a syntactic container. It groups connector declarations and establishes a scope boundary, but it does not itself carry semantic meaning beyond grouping. The compiler validates connector declarations individually; the presence or absence of multiple `connectors` sections does not change the semantics of the declarations within them.

Discussion: I'm hunting for a general term for the specification of concepts, types, metrics and so on. One form is captured as CPL, but there may also be definitions pulled from the database, or in theory at least, other sources. I'm using "plan specification" as that term here, but if we have other ideas we should talk about it.

A connector declaration consists of

```
  <connector name>
    dimensions [<list of concept names>]
    grain <measure name>[, <measure name>...]
    celltype <measure name>
```

The `dimensions` clause lists concept names. The `grain` clause references one or more measure names, separated by commas; the order is preserved. Every grain measure's type must be `Interval` or derived from `Interval`. The `celltype` clause references a measure name.

For example, the following connector declaration

```
concepts
    Package = [Values[Each, Pallet]]
    Country = [Values[USA, Canada, Mexico]]
    Product = [Values[Widget, Gizmo, Thingy]]

    # The possible Price dimension values are 1 through 10.
    Price = [Int[1,10,1]]

    Duration = [Values[Second, Minute, Hour, Day, Week]]


types
    Count = Number(precision=0)
    MondayMidnight = Instant(hour = 0, minute = 0, second = 0, day_of_week = Monday)
    WeekInterval = Interval(starts_at = MondayMidnight, unit = Duration)

units
    Package = Package
    Duration = Duration

measures
    Package = Count(unit=Package)
    Week = WeekInterval(unit=Duration)

connectors
    demand_forecast
        dimensions [Country, Product, Price]
        grain Week # Refers to the measure named Week
        celltype Package
```

says that there is a connector named demand_forecast that may be referenced by that name in other parts of a plan specification. The dimensions are concept names. The grain clause references the measure named `Week`, whose type is derived from `Interval`. The celltype clause references the measure named `Package`. Every grain measure's type must be `Interval` or derived from `Interval`. Connector input metadata and the IO specification determine which `Duration`-derived unit is used for a particular run.

A connector may declare multiple grain measures:

```
connectors
    sales_report
        dimensions [Region, Product]
        grain Yearly, Quarterly
        celltype Revenue
```

A connector declaration describes the shape of an input, but doesn't
specify its source, its format, or how to get the information. Those elements are specified in the _IO Specification_ that is passed to the CPL compiler.

## Connector Data as Metric Computations

Connector data loaded from external sources (for example, Parquet files) is registered as `MetricComputation` objects on the `ChronoconeComputationController`. During synthesis, the generated compute function pre-registers every connector input as a `MetricComputation` instance before the postfix evaluator runs. This registration ensures that `controller.GetMetricComputation(connectorName)` returns a valid handle at runtime, allowing connector inputs to participate in metric calculations identically to computed metrics.

==== Begin Discussion ===
This form of a connector deviates from some of our previous discussions, in several ways:

1. Originally connectors had information about connecting to a source that was captured in the combination of a driver and a path (plus potentially some other ancillary information related to authentication and such). I've relegated that to the IO specification (plus maybe some other stuff). With that adjustment I think that calling this a 'connector' is a bit of a stretch. Should we call this 'inputs' and 'input declarations' instead of connectors?

2. We've messed with different forms of the connector/input spec, I think that this is more in line with the general form of things that we're settling on. In places where we need such things (e.g., cell types for aggregators/spreaders) we should keep the notation consistent.
   ==== End Discussion ===
