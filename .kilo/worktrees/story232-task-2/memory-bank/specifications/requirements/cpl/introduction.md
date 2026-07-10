# Chronocone Planning Language (CPL) Concepts

## Introduction

The Chronocone System is a time series oriented forecasting and planning system
intended to be used by businesses for high level planning and forecasting operations. Chronocone is "high level" because it doesn't attempt to manage
individual resources (e.g., people, vehicles) or determine when they'll be available - that's the domain of scheduling systems (although the outputs of Chronocone plans may be used to drive scheduling systems).

At the core of the Chronocone System is the Chronocone Planning Language (CPL), which is a domain-specific language (DSL) designed for defining metrics, data sources, calculations, and dimensional relationships in planning and forecasting systems.

CPL and the Chronocone system are fundamentally organized around **metrics** which are represented as a group of time series indexed by various dimensions, such as product, region or warehouse where each dimension may have a number of different values. For example, `product` might have values such as `widgets`, `gadgets` and `gizmos`, and `region` might have values such as `Central America` or `West Coast`. Conceptually, a metric consists of a time series for each combination of values of the dimensions, although in practice Chronocone will only materialize the time series that are actually needed to generate a particular result.

The rest of this document will describe, at a high level, the elements of the Chronocone method for describing the elements of a metric, how they are created from input data, how intermediate and output metrics are calculated and how outputs are made available to other systems or user inspection.

Chronocone defines the concept of a namespace, the exact definition of which is beyond the scope of this document. Any given CPL compilation or execution happens in a particular namespace so the names of elements in a group of CPL files that are compiled at once must usually be unique within a Chronocone namespace.

## Language Syntax Structure

Syntactically a CPL file is divided into a number of sections, each of which is introduced by a keyword consisting of `units`, `types`, `concepts`, `measures`, `metrics`, `connectors`, `aggregators`, `spreaders` or `lookups`, each of will be described in detail in subsequent sections. After each keyword the section contains one or more definitions of the corresponding type, which must be indented with a single tab character. Each declaration type will have subsections, each of which will be indented with one more tab than the level before it.

Any section of in a CPL file may occur more than once, reference do not need to be declared before they are used and in fact may be declared in a completely separate file.

CPL supports one style of comments, which are line comments that start with a `#` character and continue to the end of the line. Comments may appear anywhere in a CPL file subject to that constraint.

Note: It's possible that we should accept a number of spaces as being equal to a tab, in which case we'll have to know what the translation from spaces to tabs is. For the moment I'm dodging that, but we can put it in later if we want to.

## Environment Variables

CPL supports the use of environment variables in the form `${<env variable name>}` in any place where the resulting text would result in a valid CPL definition, including inside strings. Environment variable substitution is a macro substitution concept where the substitution conceptually happens before the CPL file is parsed. Environment variable names must start with an alphabetic character and may be any sequence of alphabetic characters that does not include whitespace or the characters `$`, `{` or `}`. The CPL definition doesn't have any opinion on how or where environment variables are defined or how their values are obtained, only that they must be made available to the parser before parsing begins.

## The Metrics Idea As An Organizing Principle

Metrics are the core idea in CPL, the needs of metric representation and calculation drive all of the other elements of the language. A metric is made up of the following elements:

1. **Concepts** capture the dimensions of a metric.
2. **Time Series** capture the data in a metric. Concepts may also capture the measurement units of the data in a metric.
3. Metrics have inputs, which are metrics (including, potentially, the metric itself) and lookups which are used to represent "side data", such as currency conversion tables, weightings and so on.
4. Metrics have a calculation that specifies how any given cell of a time series in the metric is generated from the cells of the input metrics and lookups.

In order to make it possible for metric calculations to be possible and internally coherent Chronocone ensures that the effective inputs to any metric calculation match the characteristics of the effective outputs so that users don't inadvertently produce data that doesn't make sense through misalignment of dimensions/concepts (e.g., a calculation adds the inventory for a product at the region level when it should be added at the warehouse level), misalignment of time (e.g., weekly data for an input is added to daily data for a different input metric) or in terms of units (e.g., a revenue calculation combines sales in US dollars and Euros without doing a conversion first).

While programming flexibility is a wonderful thing in the right circumstances, with great expressive power comes the possibility of introducing subtle errors. Chronocone calculations supply the user with all that is required to express business logic, but intentionally prevents the user from doing things that are almost certainly mistakes. For example, a metric support "a calculation" rather than a general programming capability, which means that there is always a well define meaning for every metric. This also aids in the Chronocone philosophy of strong protection of customer data, since it is not even possible to describe mechanisms that would allow data to escape from the system.

## Concepts

A concept has three properties: A set of zero or more values, and a relationship with other concepts. The order in which values are given or shown is significant, as some operations will depend on the order for calculations, display of data or both. As will be seen in more detailed examples business planning at various levels often adopts a coarser or more granular viewpoint, but in the end is intended to describe the same systems. Thus, a Retail forecast might be done at the level of a country, but then need to be divided into regional or warehouse level forecasts. Likewise, low level data or forecasts might need to be aggregated up to a higher level. One of the goals of Chronocone is to make it easy to describe how these levels related to each other and to make it easy to create metrics without having to precisely predict at what level they'll be used, since that is likely to change based on the needs of the consumers of the metrics.

Every metric has zero or more concepts associated with it. Conceptually speaking, for every combination of the values of those concepts there is a time series. In practice Chronocone will only materialize the time series that are actually needed to generate a particular result, but never in a way that changes the value(s) of outputs in a way that would be visible to the user given their stated purpose (e.g., viewing some set of outputs at some time grain).

Every concept has associated with it either zero or one parent concepts, interpreted in the following way: When a parent/child relationship between concepts exists it means that for every value of the parent concept there are zero or more values of a child concept. The child values for a given parent value don't need to be the same for every parent value, and child values may be shared between parent values. If individual concept values are considered to be tagged by their corresponding concept, the combination of a set of parent and child concepts will always be unique. For example, if we have a concept of "country" and a concept of "region" where region is a child of country and country has the values of Mexico and Canada and the regions for Mexico are North and South and the regions for Canada are North, Central SouthWest and SouthEast (i.e., there is an overlapping regional value of North for each country) the value (country:Mexico, region:North) is distinct from the value (country:Canada, region:North) even though the values of the various regions are not globally unique.

## Time Series

A time series is a sequence of values that are indexed by time, with the following properties/elements:

1. A time series is broken into two **aspects**, `actuals` and `forecast`. Actuals represent observed or defined data, forecast represents predictions about the future. It is allowable for either or both of the actuals or forecast aspect of a time series to be of zero length.
2. A time series has a time grain, such as by-week or by-day. The time grain of a time series applies to both aspects of a time series. At this point it is worth pointing out the difference between document types and documents. A `document type` is what is being specified in a particular combination of CPL files. It tells the system a set of metrics, inputs, calculations and so are related to each other, but not exactly what the outputs are. In particular, the only place in a CPL description where a reference to a time grain exists is for inputs, since that is likely to be a property of a customers systems that is beyond Chronocone's control. It will not be possible to determine the time grains of most metrics until a user tells the system what they want to see or which metrics to push to another system. A `document` is a computed, realized reflection of a `document type`, in which every time series will have a defined grain.
3. A time series has a cell type, such as Currency or the count of an inventory item. Each cell has a value of some kind and an associated unit (e.g., US dollars, number of widgets in an inventory as opposed to the number of gadgets). Keeping track of the units in calculations and ensure that operations create results with the expect units is a key mechanism for detecting mistakes, and provided guidance for AI in creating and updating CPL descriptions. All cells in a time series will always have the same cell type for both the actuals and forecast aspect.

## Units

Units represent the 'units of measure', which is what distinguishes `20 meters` from `20 inches` or `100 dollars` from `100 euros`.

Chronocone defines two units. All other units are derived from concept values.

1. **Unitless** - There is no such thing as a type without a unit in Chronocone, but the type may be `Unitless`. The semantics of Unitless follow the general typing rules with the following caveat: The result of the `*` or `/` operation on Unitless and any other measure is the other measure (including Unitless itself). The rules the prohibit the combination of a unit with other units via operations other than \* or / apply to Unitless just as they do to any other unit.

In practice there will also be some concepts and units that are defined as standard definitions, such as a `Duration` concept with values like `Second`, `Minute`, `Hour`, and `Day`, that will appear to the user to be essentially built-in. These are defined through normal concept, unit, measure, and conversion rules; `Duration` is **not a type**. See [talking-about-time.md](talking-about-time.md) for the corrected time model.

The rules for operations on units are as follows:

1. The `*` and `/` operations on units are always allowed. They are to all intents and purposes preserved in their expression form. For example, if Ft and Lb are units, then `(12 Ft)/(6 Lb) = (2 Ft/Lb)`.
2. Combinations of units for all other operations are illegal unless the units for two types match exactly. For example, USD is the defined unit for us dollars and EURO is a defined unit for Euro, the expression `(10 USD) + (20 EURO)` would be illegal. The expression `(10 USD)*(20 EURO)` would be legal, but would result in a value of (200 USD*EURO), which is unlikely to be a useful type and will almost certainly wind up being incompatible with the specified cell type of a metric at some point during a calculation. On the other hand, if we have a source of data that has units of `USD/EURO` (i.e., as a conversion ratio) the expression `(10 USD) + (20 EURO)*(0.8 USD/EURO)`=`(10 USD) + (16 USD) = (26 USD)`would result in a`+` operation working on identical units and thus be legal.

See [concepts-types-units-measures-syntax.md](concepts-types-units-measures-syntax.md) for details of units declaration syntax and interpretation.

## Types

In the Chronocone system types define the basic type and type constraints for operations on the value part of a measure:

For example, in the expression `20 meter/second` the data is `20` and the units are `meters/second`.

## Data Type

Chronocone defines four ground data types (i.e., type definitions that are inherently defined by the system):

1. **Number** - Number is an unrestricted real number with units of Unitless that can accomodate values that are infinitely large with infinitely fine precision. Note that this is not the same as int or float as defined in many programming languages because those types have limitations on the magnitude and precision of the stored data. The possible constraints on Number are maximum value, minimum value, and precision (i.e., the number of places to the right of the decimal point).
2. **String** - Represents an unrestricted sequence of unicode code points of zero or more length. There is no upper bound on the length of a string. The possible constraints on a String are maximum length, minimum length, and a regular expression that the string of that type must match.
3. **Instant** - Represents one or more UTC moments in time. Every parameter supplied to an `Instant` declaration is a **constraint** on matching moments. Any UTC moment that satisfies all supplied constraints matches that `Instant` declaration. A single `Instant` declaration may match exactly one moment, many moments, or an infinite number of moments. `Instant` has maximum resolution of one second and always uses UTC. Constraints include year, month, day, hour, minute, second, day of week, and week number. See [talking-about-time.md](talking-about-time.md) for details.
4. **Interval** - Represents a span, range, or extent of time, usually described using `Instant` boundaries or constraints together with a unit reference. The canonical parameter is `starts_at`, which accepts the starting `Instant` for an interval declaration.
5. **Boolean** - Represents a logical value of either true or false.

New types may be derived from othe types (whether ground types or other derived types) by naming the source type, constraints on the source type and units. For example, a Currency type might be derived from number by placing a `precision=2` constraint on it and supply appropriate units. A derived type may place stricter constraints on it's parent type, but may never relax constraints. For example, if a Currency type were derived from Number by specifying `precision=2`, it would not be possible to derive a type from Currency that had a precision of 3, but would be possible to derive a type from Currency that had a precision of 1. It is always allowable to derive a more restricted data type from a less restricted data type. Thus, if one input to a calculation at some stage of the calculation was of data type Currency and another was of data type Number with compatible units, then it is permissible to convert the Number to a Currency.

See [concepts-types-units-measures-syntax.md](concepts-types-units-measures-syntax.md) for details of types declaration syntax and interpretation.

Time is a particularly difficult concept to work with that includes a number of nuances that often create confusion. Because of that errors in business logic involving time are not uncommon (e.g., around day light saving time changes or leap years). Chronocone has chosen a set of concepts, types and units that make it easy for users to express intent but at the same time force them to think about what a particular statement means exactly to avoid time related errors. Because time is so tricky it has its own discussion in [talking-about-time.md](talking-about-time.md) that the reader should consult if there is confusion about time related elements in these requirements.

## Measures

Measures are the combination of a type and a unit. The following measures are built-in to Chronocone:

1. **Instant** - Of type `Instant` with unit `Unitless`.
2. **Interval** - Of type `Interval` with units that are derived from a `Duration` concept through normal unit declarations. The type `Interval` defines operations that work with `Instant` constraints and `Interval` measures. Grain references for metrics and connectors are measure names whose resolved type is `Interval` or derived from `Interval`.
3. **Nil**, which is a built-in pseudo-measure. There is a Nil measure for each built-in type, that represents a non-existent value of that type although they are all called Nil from a user point of view. Note that the "non-existent" value is not the same as the value 0, which is a valid value for all numeric types but it is equivalent to the empty string for string types and the value False for boolean types.

`Duration` is not a type. It is a concept or value set whose values define units used when measuring an `Interval`. For example, a particular forecast might be referred to as a 'weekly forecast' (specifying only the duration), but what it really is is a 'week-long forecast that starts at midnight on Monday and ends just short of midnight on the following Monday' (specifying an interval with `starts_at` and a unit). Given our usage rules and standard definitions we should always be able to tell the difference while also allowing CPL specifications to be given in a natural way.

## Metrics, Revisited

Metrics are the main computational declaration in CPL (others are unit conversions and dimension or time transformations). Metrics declarations follow the same general pattern as concept, unit, and type declarations and includes:

1. A metric name
2. The logic that is used to determine it's value.

It should be noted what the definition does not specify:

1. The time grains/intervals of the inputs or outputs
2. The type of inputs or outputs
3. The dimensions/concepts associated with the metric.
4. The type of the cells of the time series captured by the metric.

All of these things are determined by the Chronocone calculation framework, and may involve applying transformations between types, time grains and/or dimensions in order to make the cell(s) of the input metrics align with the corresponding cell(s) of the output metrics across time grain, dimensions, types and units. Other definitions in .cpl files (such a unit conversions or aggregators/spreaders) may be used to make that possible. If it is not possible for the framework to determine how to match cells (e.g., thought the operation of aggregators/spreaders and unit and type conversions) the Chronocone analyzer will indicate an error. As a result, the CPL description isn't enough to do a metric calculation by itself, it's also necessary to specify enough restrictions on enough metrics in a namespace to determine which time grain(s) for which output(s) are of potential interest in the form of an output specification.

See [metrics-syntax.md](metrics-syntax.md) for details of metrics declaration syntax and interpretation.

## Connectors

Planning requires inputs.

The inputs are obtained by the action of connectors, which will produce inputs in a form that the calculation framework understands. From the point of view of the calculation framework the things produced by connectors are identical to metrics, except that they cannot be transformed by aggregators, spreaders or unit conversions although the data in input metrics can be transformed as part of the activity of shaping inputs to other metrics.

See [connectors.md](connectors.md) for details of connectors declaration syntax and interpretation, and [io_specs.md](io_specs.md) for details of IO specifications.

## Dimensional Transformation

One of the aspects of Chronocone modeling is the idea that users can create specifications that are natural for each individual metric, without having to worry about the mechanical details of how they're calculated. In some cases this means that the dimensions of an metric won't match the dimensions of it's inputs. In that case, before the metric is 'called', the dimensions of the input(s) must be changed to match the dimensions of the metric being defined.

Since concepts are organized as a tree, there are two possible ways to transform dimensions. One is to 'roll up' the values of a dimension to a higher level in the tree. This function is implemented by aggregators. The other is to 'roll down' the values of a dimension to a lower level in the tree. This function is implemented by spreaders.

## Aggregators

Aggregators are used to transform more granular data to less granular data. An aggregator specifies what operation(s) should be used to translate data that is distributed across multiple time series in a metric to a single time series that has dimensions that correspond to a higher level concept in the taxonomy. Users are not responsible for determining when to use aggregators, the Chronocone calculations framework will determine which aggregators to use and in what sequence.

See [aggregators-and-spreaders-syntax.md](aggregators-and-spreaders-syntax.md) for details of aggregators declaration syntax and interpretation.

## Spreaders

Spreaders are used to distribute data from a less granular dimension to a more granular dimension, such as dividing inventory orders that are forecast on a Country level down to the Warehouse level. The name comes from the idea that something is "spread" across something else like peanut butter being spread across bread.

See [aggregators-and-spreaders-syntax.md](aggregators-and-spreaders-syntax.md) for details of metrics declaration syntax and interpretation.

## Lookups

A lookup is conceptually similar to a hash table in the sense
that there are zero or more input dimensions, that may come from
either concepts or grain dimensions, that map to a value. A lookup is quite similar to a connector in terms of the data
that it feeds into the system. The main difference is that
connectors, in a fully implemented Chronocone system, will tend
to be read from an external service or database and will generally be expected to be different each time they are read while a lookup is expected to change more slowly and will be read from a file or database. Lookups will tend to be used to hold items like conversion factors or mapping tables. While the tensors produced by connectors will always have a time dimension, lookups will often not have a time dimension (but may have a time dimension). The intent is that data in lookups can be represented at run time as tensors in exactly the way that connector or metric data is represented.

See [lookups-syntax.md](lookups-syntax.md) for details of lookups declaration syntax and interpretation.
