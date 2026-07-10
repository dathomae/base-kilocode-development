## Concepts, Units and Types

Note: Time is a particularly difficult concept to work with that includes a number of nuances that often create confusion. Because of that errors in business logic involving time are not uncommon (e.g., around day light saving time changes or leap years). Chronocone has chosen a set of concepts, types and units that make it easy for users to express intent but at the same time force them to think about what a particular statement means exactly to avoid time related errors. Because time is so tricky it has its own discussion in [talking-about-time.md](talking-about-time.md) that the reader should consult if there is confusion about time related elements in these requirements.

## Concepts Declaration

Concepts have two properties:

1. They may be related to zero or one parent concepts and zero or more child concepts.
2. They are associated with zero or more ordered values. The meaning of values is context dependent, and will usually depend on the measures associated with the concept.

There is a one to many association between each parent value and a set of child values. The child value associated with each parent value may be different for different parent values.

Concepts are declared in a `concepts` section, with a one tab indentation before each declaration. For example, this is a valid concepts declaration:

```
concepts
    Country=[Values[USA, Canada, Mexico]]
      Region = [Values[North, South]]
```

The final values for concepts may be thought of as strings, although strictly speaking they have no type, only rules for how concept values are represented. Concept values must start with an alphabetic character and may contain alphanumeric characters, the `-` character, and the `_` character and may optionally be enclosed in single quotes. If a concept value contains whitespace it must be enclosed in single quotes.

Where an indentation of a concept definition below another concept definition indicates that there may be many Region associated with a Country, or in concept terms Region is a child concept of Country, Country is a parent concept of Region. If this was the entire set of concepts then Country has no parent concept. Because there is no restiction on the values of Country in the declaration of Region, every value of Country is associated with the values 'North' and 'South' for Region.

The contents of the square brackets are a generator that defines the values of the concept. These are the currently defined generators:

- `Values[<comma seperated list of values>]`, which accepts a list of comma seperated values enclosed in square brackets. For example `Values['New England', 'Mid Atlantic', 'South Atlantic', South]` would generate those exact values as part of the concept definition.
- `Int[<start>, <end>, <step>]`, which accepts a start, end, and step value. For example `Int[0, 10, 1]` would generate the values 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10. The step value is optional and defaults to 1. The end value is inclusive. The start and end values must be integers. The step value must be a positive integer. The start value must be less than the end value.
- `<concept name>[<start concept value>..<end concept value>]`. In this type of generator the initial `concept name` refers to a concept. The concept values must refer to concept values in that concept. The values created by the generator are all concept values included in the range defined by the order of the concept values in the named concept, where the starting concept values must be earlier in the order than the ending concept value. The range is inclusive.
- `Regex[...]`, need to work out the details of this one.

The definition of a concept may include multiple generators enclosed in square brackets. The relative order of the values in the concept will reflect the order of the generators in the definition, regardless of their values. For example, the following definition:

```
concepts
  DayNumbers = [Int[4, 7, 1], Int[1, 3, 1]]
```

will result in a set of concept values that has the following order: 4, 5, 6, 7, 1, 2, 3.

Child concept values need not be the same for all parent concept values. For example:

```
concepts
    Country=[Values[USA, Canada, Mexico]]
      Region = Country[USA] => [Values['New England', 'Mid Atlantic', 'South Atlantic', South]]
      Region = Country[Canada] => [Values[Ontario, Quebec, 'British Columbia']]
```

Where the optional clause `<parent concept name>[<parent concept value>]` asserts a conditional relationship for a set of child concept values
when the parent concept value is equal to the specified value.
With this declaration the Region values associated with the 'USA' value of Country are 'New England', 'Mid Atlantic', 'South Atlantic', and 'South' and the Region values associated with the 'Canada' value of Country are 'Ontario', 'Quebec', and 'British Columbia'. The 'Mexico' value of Country maps to zero values for Region. However, if we had the following declaration:

```
concepts
    Country=Values[[USA, Canada, Mexico]]
      Region = Country[USA] => [Values['New England', 'Mid Atlantic', 'South Atlantic', South]]
      Region = Country[Canada] => [Values[Ontario, Quebec, 'British Columbia']]
      Region = Values[[North, South]]
```

Then the Region values for any Region that doesn't have a more specific definition will be 'North' and 'South', so those would be the values associated with the 'Mexico' Country. The general rule is that the more specific concept definition will be used in preference to the broader or more general definition.

By themselves concepts have no intrinsic meaning, but may be used in the definition of units and the dimensions of metrics and other entities in the system.

The definition syntax for concepts is:

```
<concept name> = [<concept generator>+]
```

where a generator is some string enclosed in square brackets that follows a generator pattern as described above.

## Types Declaration

The `types` declaration defines new data types that are derived from other types, either the Chronocone ground types or other derived types.

For example, each of these is a valid type declaration:

```
types
    Currency = Number(precision=2)
```

it is not an error for the same name to be used for a concept, unit, type and measure. The general rule is that the name is used in the context of the declaration, so when a name that is a concept is used as a type, the type declaration will be used, when used as a unit the unit declaration will be used and when used as a concept the concept declaration will be used.

In this case the type Currency is derived from the base type Number. Number has unlimited magnitude, unlimited precision and has units of `Unitless`, the derived Currency has unlimited size but always has exactly two decimal places of precision and units of the value of the Currency unit (in this example one of 'USD', 'Yen' or 'Euro').

Unit declarations may use other derived types in addition to the base type. Type derivations must be at least as restrictive as the type they are being derived from, or it's an error. Imagine that a brokerage used four decimal places of precision for certain currency calculations. This declaration would produce an error:

```
types
    Currency = Number(precision=2)
    BrokerageCurrency = Currency(precision=4)

```

because BrokerageCurrency is attempting to increase the precision of Currency, which is not allowed. However, this declaration would be valid:

```
types
    BrokerageCurrency = Number(precision=4)
    Currency = BrokerageCurrency(precision=2)
```

because Currency is more restrictive than BrokerageCurrency. Also, because Currency doesn't specify a unit, the unit is inherited from the derived type. It is permissible to define a unit expression when declaring a derived type provided that the unit expression is compatible with the base type. That means that given the unit declarations, including equivalency expressions, it must be possible to perform some series of unit transformations that will convert the type associated with the base type to the type associated with the declared unit.

Each type in the system defines a linear path from the type being used to one of the Chronocone defined base types. Each type is unique and cannot be converted to a more restricted type since that may result in loss of information. For example, if the type of a metric being defined was `BrokerageCurrency` and the type of input A was `BrokerageCurrency` and the type of input B was `Currency` then a calculate expression of `A + B` would be an error because it would require the type of B to be converted to BrokerageCurrency, which is not allowed by default because conversions to a more restricted in general are lossy transformations. Thus a BrokerageCurrency value that was converted from a BrokerageCurrency type to a Currency type and then back to a BrokerageCurrency type would not necessarily be the same as the original value. However, conversion from a more restricted type to a less restricted type is always allowed, since it cannot result in loss of information. Thus if the type of the metric was `Currency` instead of `BrokerageCurrency` then the expression `A + B` would be valid.

Note: In the last proposal I got the direction of the disallowed conversion backwards, so I'm just disallowing information losing type conversions in general. However, if we wanted to we could allow potentially lossy conversions by adding a `lossy` constraint or something, but it might be a feature that we disallow flows that would potentially result in information loss, since that introduces subtle errors, so I didn't do that.

## Measures

Measures are the combination of a value and a unit. Units represent the 'units of measure', which is what distinguishes `20 meters` from `20 inches` or `100 dollars` from `100 euros`. In Chronocone parlance, the number in each of those examples is the 'measure value' while the 'unit of measure' is the 'unit'.

In addition to the definition of what measures may be reference in actuals and forecasts, measure declarations capture the logic or formulas that are required to convert between one measure and another. The general form
of a measure declaration is:

```
<measure group>=<type name>(unit=<concept name>)
   conversion
     pattern
       <expression>
     calculation
       <expression>
```

If a concept is used as the unit for a measure group and hasn't been referenced in a unit declaration then it's use as a unit causes it to be defined as a unit. Thus, if
there is some CPL description that doesn't include any pure units conversions the units clause might not appear at all.

A 'measure group' is a set of measures that have the same type and concept values. For example, if the Currency concept has the values USD, EURO and YEN, then the Currency measure group would be the set of measures that have the type Currency and the unit values USD, EURO and YEN.

Just because the individual measures are in the same group doens't mean that they can be combined, a unit conversion clause must be defined to allow the measures to be combined. The first part of a unit conversion clause is a pattern that defines the conversion that is allowed, specified as an boolean/arthimetic expression that references the same or other measure group names or measure names. The terms that may be referenced in the pattern are measure group names or measure names which optionally take an identifier that is a concept value in the measure group for any given conversion. These aliases may be used in the calculation expression.

For example, in the following declaration:

```
measures
  Currency = Currency(unit=Currency)
    conversion
      pattern
        Currency(c1)/Currency(c2)
      calculation
        currency_conversion_lookup_ratio[c1, c2]
    conversion
      pattern
        Currency(c1)=Currency(c2)
      calculation
        currency_conversion_lookup_absolute_value[c1, c2]
```

The first reference to Currency is a measure group name, which is made up of the type Currency, with units that are the values of the concept Currency.

There are two conversion patterns defined. The first produces a ratio between two currencies and thus the number supplied by the lookup referenced in that calculation is
a unitless ratio. The measure that corresponds to each Currency in the ratio is captured by an alias, which then can be used to pass to a lookup. For example, if Currency had values of USD and EURO, then the measures in the measure group 'Currency' are USD and EURO. The measure ratios defined in the pattern for the first unit conversion patter are USD/USD, USD/EURO (i.e., 'how many dollars are in an euro'), EURO/USD and EURO/EURO.

An alterate version of the conversion is shown in the second unit conversion pattern, which would work with a lookup table that returns the number of the second type of currency for a single unit of the first type of currency (e.g., number of Euros per Dollar). In this case there's a simple equivalence, although number that would be returned for the equivalence (e.g., 1.2 Euros equals a dollar) and the ratio version which would return the ratio between currencies (e.g., 0.8 Euros per dollar) are different because they're stated in different terms. The Chronocone analysis will attempt to find conversions that match the units required to compute metrics, rearranging the terms if necessary. Thus, in practice, have both of those conversions would be redundant. They're both shown here to illustrate some possible conversion scenarios.

Imagine another scenario where there is no lookup defined from which the data can be retrieved. CPL also supports the ability to name a concept value that is part of measure group, provided that the value exists in the concept. For example, suppose that in the previous example suppose that multiple currency units might appear in the inputs but
all calculations are done in US dollars (USD) and that USD is one of the values for the Currency concept. Also imagine that there is no system input for the Euro or Yen conversion and we want to hard code the conversion value. Then the following declaration could capture that relationship:

```
measures
  Currency = Currency(unit=Currency)
    conversion
      pattern
        USD=Currency(other_currency)
      calculation
        when
          other_currency = EURO => 1.2
          other_currency = YEN => 160
        end when
```

If a concept value for Currency is not specified in the clauses for the when clause and there is no default it would be an error at compile time. Note that this conversion
depends only on the units of the measure and not on its value, so it could have been specified in the units section. In this example we only have one measure group that has
units of Currency. If there were multiple measure groups defined that had units of currency either the conversion would need to be defined for each one or it would be defined in the units clause. The second method is preferable to avoid copy/pasting identical code and opening the possibility of different conversion rules for the same unit for different measures. That said, it would be preferable to capture unit conversions in the units clause when possible in the units clause rather than the measures clause to improve maintainability of code if and when additional measures are added to a description.

The previous examples only referenced the unit element of a measure. It is also allowable to reference the measure part and use it in patterns and calculations. For example,
if we wanted to convert from hours to days including with adjustments for daylight savings time we could do the following:

```
concepts
  Duration = [Values[Second, Minute, Hour, Day]]

units
  Duration = Duration

types
  Midnight = Instant(hour = 0, minute = 0, second = 0)
  DayInterval = Interval(starts_at = Midnight, unit = Duration)

measures
  Day = DayInterval(unit = Duration)
    conversion
      pattern
        Hour = Day(other_day)
      calculation
        when
          other_day = 03-09-2025 => 23
          other_day = 11-02-2025 => 25
          default => 24
        end when
```

This combination of pattern and calculation says two things:

1. This is a rule to convert the value part of a measure that has units of Day into one that has units of Hour. The inverse is not automatically implied, thus this does not imply that it's possible to calculate Days from Hour; there would need to be an explicit conversion defined for that. The Chronocone analyzer will, however, feel free to rearrange the pattern using unit algebra according to the usual rules, so if some metric were to calculate a ratio with units of (Hour/Day) starting with Day that would be possible but calculating the same ratio starting with Hour would not.
2. Certain Day values are associated with certain Hour values.

Note: `Duration` is a concept that supplies values for units; it is **not a type**. The measure `Day` is declared from the derived type `DayInterval`, which is based on `Interval`.

Discussion: So far all of the pattern declarations only involve references to entire concepts in the role of units. It would be possible to extend that
to refer to concept values as we do in the calculation clause, but I'm having trouble thinking of an example where that would be more useful than having the
equivalent thing in the calculation clause so I'm leaving that out for now. If we can think of a scenario where it's useful we could include it.

## Units Declarations

The units declaration section is identical to the measures declaration section except that it start with the keyword `units` rather than `measures` and patterns and calculations in units declarations may
reference only units, not the value part of a measure. Thus, if there was a Duration concept that include Days, Hours and Seconds values then assuming that we ignore leap seconds the conversion from Seconds to Hours depends only on the units and so could be included in the units section. Taking daylight savings time into account implies
that the conversion from Hours to Days depends on the value (i.e., which day of the year it is), some days being 23 hours to the day, some 24 hours to the day and some 25 hours to the day and therefore can't be included in the units declaration section.
