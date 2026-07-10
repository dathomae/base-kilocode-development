## Metrics Definition

One or more metrics may be defined in a `metrics` section which must begin with the `metrics` keyword. Individual metric definitions must be indented one tab deeper than the `metrics` keyword. For example:

```
concepts
    Country = [USA, Canada]
    Product = [widget, gadget]
    Package = [Each, Pallet]
```

```
types
    Count = Number(precision=0)
```

```
measures
    Count = Count(unit=Package)
```

```
metrics
    inventory_actuals
        <...>
```

Declares the existence of a metric named `inventory_actuals`. A metric has no inherent cell type, dimensions, forecast or actuals length, or time grain associated with the metric declaration itself. Those properties are determined by the IO specification supplied to the CPL compiler at synthesis time (see [io_specs.md](io_specs.md)).

The general form of a metric declaration is:

```
<new metric name>
    <metric definition body>
```

A metric declaration has two sub-declarations, each of which must be indented by one tab from the main declaration: `inputs` and `calculate`.

### Metric Inputs

The first sub-declaration lists the metric inputs, introduced by the `inputs` keyword. For example:

```
inventory_actuals
    inputs
        metric demand_forecast

        # inventory_actuals has a daily grain in its connector, but the output
        # has a weekly grain, so a time-oriented aggregator will be needed to
        # do the conversion.
        metric inventory_actuals as inv_actuals where aspect=ACTUALS
```

The general form of a metric input declaration is:

```
<metric|lookup|dimension> <name> [as <alias>] [where <boolean expression>] [backward(<constraint>=<expression>) | forward(<constraint>=<expression>)]
```

The input keyword determines the kind of input:

- **`metric`** — a reference to another metric defined in the CPL specification.
- **`lookup`** — a reference to a lookup defined in the CPL specification.
- **`dimension`** — a reference to one of the dimensions of the metric being calculated. Dimension inputs make that dimension compulsory for the metric (it cannot be collapsed by aggregators prior to this metric's calculation).

The `as` clause is optional. If absent, the name of the input is its own alias. No two aliases may be the same within a single metric's input block.

The `where` clause is optional. When present it filters or constrains the input. The only select-style operator currently defined is `aspect=<aspect>`, where `<aspect>` is one of:

- `ACTUALS` — filter the input to actual (historical) data points only.
- `FORECAST` — filter the input to forecast data points only.

Multiple `where` conditions may be joined using the boolean operators `and`, `or`, and `not`.

The optional `backward(<constraint>=<expression>)` or `forward(<constraint>=<expression>)` clause shifts the input metric in time. The `<constraint>` identifies the time measure (e.g., `weeks`, `days`) and the expression provides the shift amount. The expression may be any valid expression (e.g., a numeric constant, a lookup reference, or an arithmetic expression) and is evaluated per cell at calculation time.

> **Reserved keywords:** `backward` and `forward` are reserved keywords in CPL and cannot be used as identifiers.

For example:

```
inventory
    inputs
        metric order_arrivals # The arrival week for previous orders
        metric inventory as prior_inventory backward(weeks=1)
    calculate
        ordered_arrivals+prior_inventory
```

Without a `backward` or `forward` clause, each input refers to the "current" cell (the cell at the same time coordinate as the cell being calculated) from the perspective of the metric author.

### Connector Input Registration

When a metric references a connector in its `inputs` block, the connector-provided data must be available to the postfix evaluator as a `MetricComputation`. The lifecycle proceeds as follows:

1. Connector data is loaded from Parquet files into the `inputs` map, keyed by connector name.
2. During synthesis, `generateConnectorRegistrationCode()` emits a loop that creates a `MetricComputation` object for each entry in the `inputs` map.
3. Each `MetricComputation` is registered on the controller, and its tensor is set to the loaded connector tensor.
4. The postfix evaluator resolves connector inputs via `controller.GetMetricComputation(connectorName)`, treating them identically to computed metrics.

### Self-References

Metrics may reference themselves in calculations, but only through an alias that uses a `backward` or `forward` clause. Direct self-reference in a `calculate` clause (naming the metric itself without an alias) is prohibited because an unaliased input always refers to the current cell, which would create a circular dependency at the cell level. The only valid self-reference is to an earlier or later cell, obtained via an alias.

For example, a metric that computes ending inventory from prior inventory plus arrivals minus demand:

```
inventory
    inputs
        metric inventory as prior_inventory backward(weeks=1)
        metric inventory_arrivals
        metric demand_forecast
    calculate
        prior_inventory + inventory_arrivals - demand_forecast
```

### Metric Calculation

The business logic associated with a metric is specified in the `calculate` clause. For example:

```
inventory_actuals
    inputs
        metric demand_forecast
        metric inventory_actuals as inv_actuals where aspect=ACTUALS
    calculate
        if inv_actuals
            inv_actuals
        else if
            demand_forecast * 1.2 - prior_inventory
        end if
```

The body of the `calculate` clause consists of a single statement. The result of that statement is the value of the metric for the cell being calculated. This is a purely functional notation: there are no variables or side effects. The value of a conditional expression is the value of the expression in whichever branch is selected as the true branch.

The `calculate` clause supports:

- `if` / `else if` / `else` / `end if` conditional expressions
- `when` / `end when` alternative conditional syntax (see below)
- Common arithmetic operators (`+`, `-`, `*`, `/`) for their respective types
- Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`)
- Logical operators (`and`, `or`, `not`)
- Signed numeric constants and string constants
- `Nil` literal
- Parenthesized expressions
- Dimension indexing on metrics and lookups (e.g., `order_lead_time[product]`)

The `calculate` clause must produce exactly one result under all possible conditions; otherwise it is a compile-time error. For example:

```
calculate
    if inv_actuals
        inv_actuals
    end if
```

would produce an error because behavior is undefined when `inv_actuals` is falsy. However:

```
calculate
    if True
        inv_actuals
    end if
```

is valid because the condition is tautologically true.

#### Truthiness Rules

In boolean contexts (conditions of `if`, `else if`, `when`, and operands of logical operators):

- `Nil` is equivalent to `False`.
- The boolean value `False` is falsy.
- For numeric values, zero is falsy; any non-zero value is truthy.
- For strings, the empty string is falsy; any non-empty string is truthy.
- All other non-Nil values are truthy.

#### Nil in Arithmetic

`Nil` cannot participate in arithmetic expressions. Any attempt to use `Nil` as an operand of `+`, `-`, `*`, or `/` is a compile-time or runtime error.

### `when` Syntax

To reduce visual clutter, an alternative form of conditional syntax is allowed:

```
when
    inv_actuals != Nil => inv_actuals
    default => demand_forecast * 1.2 - prior_inventory
end when
```

The general form is:

```
when
    <boolean condition> => <expression>
    <boolean condition> => <expression>
    ...
    default => <expression>
end when
```

`when` is interpreted identically to a chain of `if` / `else if` / `else`: the first line is an `if`, subsequent lines are `else if`, and the `default` line is `else`. Fall-through is not allowed, and the result must be fully specified. `default` is a reserved keyword in CPL and cannot be used as an identifier.

The conceptual model when writing business logic is that, for any given calculation, there is one time series cell value for each input, and the boolean or arithmetic expressions operate on those values. It is the responsibility of the Chronocone framework to ensure that the correct cell values are used, given the inputs, including any restrictions.

## Short Form Metrics

Metric definitions may be supplied in an abbreviated short form:

```
<metric name> = <calculation>
```

where the calculation is a valid calculation expression and the references are to unaliased metric names, lookup names, or dimension names. No `where` or `backward`/`forward` clauses are permitted in short form. Thus, the following metric definition is legal:

```
revenue = demand_forecast * Price
```

This is equivalent to:

```
revenue
    inputs
        metric demand_forecast
        dimension Price
    calculate
        demand_forecast * Price
```

In short form, each name in the calculation expression is resolved as follows:

1. If the name matches a metric declared in the CPL specification, it is treated as a `metric` input.
2. If the name matches a lookup declared in the CPL specification, it is treated as a `lookup` input.
3. If the name matches one of the dimensions of the metric (as specified in the IO specification), it is treated as a `dimension` input.
4. A name collision between a lookup, a dimension or a metric in a short form calculation is a compile-time error.

## Compilation Scope

The CPL compiler generates code only for the calculations required to produce the outputs specified in the IO specification. Metrics that are not named in the IO specification and are not direct or indirect inputs to any metric named in the IO specification are excluded from compilation (dead-code elimination).
