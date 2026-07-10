## Talking About Time

Dealing with time is the most confusing and error prone element of business calculations due to its irregular nature. The reason is that time has complicated rules association with its relationships, but people want to think of it as if it was a continuous number range. To illustrate how incorrect that idea is one has only to think about what it means to subtract one thing from another. For example, if I think of two numbers x, y where x is less than y then the distance between x and y is `y-x` in every situation.
Time however, is not so neat. In the same situation say that x is 6 pm on some day and y is 6 pm on the following day what is the distance between them. On most days `y-x` is a duration of 24 hours, except sometimes its 23 hours and sometimes its 25 hours due to the effects of going on and off of daylight savings time. There are similar issues with the distance between the same date in different years due to leap years, the distance between the beginning of months and quarters because months aren't all the same length and the distance between movable holidays (e.g., Easter, Thanksgiving) in different years has the problem with year calculations compound by additional holiday specific rules. The business practice of thinking of time in terms of week number is an attempt to regularize the treatment of time, but because of its nature there are non-intuitive aspects of week numbers as normally implemented: the first week of of the year usually doesn't start on the first day of the year and some years have 52 week numbers and some have 53.

Chronocone, though a careful selection of concepts, units and types, makes it easy to think and deal with time in a manner that is consistent with it's often byzantine rules and makes it hard to deal with time incorrectly. A consequence is that a user who is operating from naive beliefs (e.g., a day is 24 hours long, a year is 365 days or 52 weeks long, Thanksgiving day is one year from the previous or subsequenct Thanksgiving day) may find that it is frustrating and difficult to build metric descriptions that are influenced by naive beliefs. This is a feature, not a bug.

There are three different, but interrelated, ideas that are fundamental in thinking about time as it applies to forecasting problems:

1. **Instant** — a set of constraints that match one or more UTC moments. Every parameter supplied to an `Instant` declaration is a constraint on matching moments. Any UTC moment that satisfies all supplied constraints matches that `Instant` declaration. A single `Instant` declaration may match exactly one moment, many moments, or an infinite number of moments. `Instant` has maximum resolution of one second and always uses UTC.
2. **Duration** — a concept or value set whose values define units used when measuring an `Interval`. `Duration` is **not a type**. It only provides concept values, such as `Second`, `Minute`, `Hour`, and `Day`, that may be used by `Interval` measures. Built-in and user-defined `Duration` concepts and units receive no special semantic treatment beyond normal concept, unit, measure, and conversion rules.
3. **Interval** — a time-related type that represents a span, range, or extent of time, usually described using `Instant` boundaries or constraints. The canonical parameter is `starts_at`, which accepts the starting `Instant` for an interval declaration, together with a unit reference.

The main lesson is that translation between different durations is sometimes a constant and sometimes not. A duration of a minute will always translate to 60 seconds. A duration of a day may translate to 23, 24 or 25 hours depending on the time span in question.

## Instant Constraints and Underspecified Instants

The arguments passed to `Instant` are **constraints**, not required components of a complete date-time tuple.

- A UTC moment matches an `Instant` declaration **if and only if** that moment satisfies all constraints supplied to the declaration.
- No `Instant` constraint is required merely because another constraint is supplied.
- Omitted constraints make the `Instant` less specific and broaden the set of matching moments.
- Because constraints can be sparse, an `Instant` can match many or infinitely many moments.

### Examples

- A fully specified `Instant(year=2025, month=March, day=14, hour=9, minute=30, second=0)` usually matches exactly one second-level UTC moment.
- An `Instant(month=March, day=14, hour=9, minute=30, second=0)` with no year matches that date and time in **every year**.
- An `Instant(hour=0, minute=0, second=0)` matches midnight of **every day**.
- An `Instant(day_of_week=Monday)` matches every moment during every Monday unless additional constraints narrow the match set.

### Anti-Example

The following is **not valid** because `Instant` does not accept a `precision` parameter:

```cpl
# INVALID — Instant has no precision parameter
types
    FiscalYear = Instant(precision=year)
```

## Duration Is Not A Type

`Duration` is never a type. It is a concept or value set whose values define units used when measuring an `Interval`.

Standard definitions should include the `Duration` concept and unit group used by canonical interval examples, but users may declare their own `Duration`-like concepts and units under normal CPL rules.

### Valid Example

```cpl
concepts
    Duration = [Values[Second, Minute, Hour, Day]]

units
    Duration = Duration

types
    Midnight = Instant(hour = 0, minute = 0, second = 0)

measures
    DurationStartingAtMidnight = Interval(starts_at = Midnight, unit = Duration)
```

Inline `Instant` is also valid:

```cpl
measures
    DurationStartingAtMidnight = Interval(starts_at = Instant(hour = 0, minute = 0, second = 0), unit = Duration)
```

A named `Interval`-derived type is also valid:

```cpl
concepts
    Duration = [Values[Second, Minute, Hour, Day]]

units
    Duration = Duration

types
    Midnight = Instant(hour = 0, minute = 0, second = 0)
    MidnightInterval = Interval(starts_at = Midnight, unit = Duration)

measures
    DurationStartingAtMidnight = MidnightInterval(unit = Duration)
```

## Grain of Metrics and Connectors

The grain of a metric or connector must be the **name of a measure** whose type is `Interval` or whose type is derived from `Interval`.

The grain measure's type resolution follows the general measure-declaration rule: a measure can be declared from a built-in type with constraints or from an already derived type with additional constraints.

Connector input metadata and the IO specification passed to the CPL compiler determine which `Duration`-derived unit is used for a particular run.

## Usage Of Time

Note: For the following discussion I'm going to assume that there are concepts declared for all of the common units, such as Hour having values of 1..25 (yes, 1..25 because of daylight savings time transitions, although we could potentially have different concepts/units for Hour, DaylightSavingsTimeHour or even SpringForwardDaylightSavingsTimeHour and FallBackDaylightSavingsTimeHour with conversion relationships. Getting all of the nit-picky details right is going to be loads of fun, although at least once they're figured out those things will just work in calculations.) Also assume that all time related concepts are also declared as units.

Now that we have a way to represent time, we need to understand how it can be used.

### Time Series Grain

The grain of a time series is a measure name whose resolved type is `Interval` or derived from `Interval`. Using the rules for concept value ordering discussed above we'll always be able to create a time ordered sequence of values for a time series. Under this concept the grain of a time series could be considered to be just another dimension with ordered values and there's a cell of the time series of length one for each combination of dimension values. Under that interpretation the time shifting operations can just be aggregators and spreaders where the 'dimensions' specified happen to be Interval types.

### Time Used As A Dimension

To Be Completed

Discussion: Since time related elements are concepts they can be used as dimensions in a metric or lookup. That particularly seems to make sense for period over period comparisons. Imaging that we have a metric with a daily grain that covers an entire year and some of the output metrics are quarter over quarter or week over week comparisons. Using the machinery we have in CPL that's pretty easy, we can have the Quarter or the Week as a pair of dimensions. It has to be a pair of dimensions, although that pair is the same concept named twice, because we're comparing two different dimensions (e.g., two Quarter or two Week sets of values). It's unlikely that the user really wants to see a comparison of every Quarter or Week, they're typically just going to want to see the comparison between adjacent periods. To handle that we could invent some kind of filtering syntax on metrics, but I think we'd be fine just allowing for dimensional comparisons in the business logic. For example, imagine that we have an inventory_actuals metric that spans multiple quarter and weeks that we want to get a period over period comparison for, using the aliasing concept (which hasn't yet been defined in the other docs):

```
metrics
    inventory_period_comparison of Count
        inputs
            metric inventory_actuals as inventory_actuals_previous_period
            metric inventory_actuals as inventory_actuals_current_period
            dimension previous_period of Interval
            dimension current_period of Interval
        calculate
            if previous_period == current_period - 1
                inventory_actuals_current_period - inventory_actuals_previous_period

aggregators
    from [Week] to [Quarter] = sum(Week)
    from [Quarter] to [Year] = sum(Week)
```

Since metrics have no inherent dimensions and intervals just another dimension whether this would result in a year over year comparison, a quarter over quarter comparison or a week over week comparison would depend on the dimension specified by the user, assuming that a set of aggregators and/or spreaders have been defined to allow for the translation from the grain (for time interpreted as a dimension) of the inputs to the dimensions of the outputs. The only constraint is that because the previous_period and current_period dimensions are called out as inputs the calcuator will know that the metric has two dimensions that must both be Intervals and that they must be the same concept since they are used together in a comparison operation (e.g., this is a variation of the 'dimensional units' analysis).

Discussion: In this perspective there are no actual 'shifters' as such, they're just aggregators and spreaders that happen to mention Interval related concepts. Also, when thinking about calculations it's important to distinguish operations on dimensions from those on data values. I think it makes sense to say that calculations need to be fully specified, so that we'll never have indeterminate values in a time series. However, I think that it's necessary to allow for the results of operation(s) on dimensions to not be specified, with the interpretation that an unspecified result means that the dimension values won't appear in the output for the metric. In the example above the outputs are only specified when `previous_period == current_period - 1`, which means that when the periods aren't contiguous the combination of the dimension values won't appear in the output. Another implication is that because concept values are discrete, we're not requiring the cells in a time series to necessarily be contiguous (ordered in time, but not contiguous). Thus if you wanted to create a metric that was a time series of all Thanksgiving days (one a year) you could do that. In the example above the `previous_period == current_period - 1` would only produce results for contiguous periods, so that if the output was a quarter over quarter comparison but Q2 was missing from the data there would be no output for Q2 over Q1 and we wouldn't produce output for Q3 over Q1. If, in that case, the user wanted to generate Q3 over Q1 results they couldn't just say `previous_period < current_period` because that would result in every period being compared to every period that had come before it. To handle that we'd have to have Interval operations defined for things like 'previous period' and 'next period' and then we could say something like `previous_period == previous(current_period)` (it's not surprise that there are time related operations needed that aren't appropriate for other types.)
