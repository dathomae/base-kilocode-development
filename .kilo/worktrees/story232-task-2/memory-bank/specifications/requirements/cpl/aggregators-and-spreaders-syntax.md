## Aggregators and Spreaders Definition

Chronocone metrics break out the time series in the metric by the combination of concept values that are inherent in input metric or lookup definitions. The concepts involved in this representation are referred to as a dimension of a metric or lookup. For example, if some input was the inventory broken out by Country, Product and Warehouse those are each one of the dimensions. At some point in the metric calculation process it might be necessary to collapse a representation broken out by [Country, Product, Warehouse] to [Country, Product] by combining the data across the Warehouse dimension in some way. The definitions in Chronocone that accomplish that are called aggregators. Likewise, it might be necessary to create a representation by [Country, Product, Warehouse] from a representation broken out by [Country, Product], usually by applying some type of weighted distribution. The definitions in Chronocone that accomplish that are called spreaders.

Note that in general the action of aggregators and spreaders involve information loss. Thus, in the previous example if the dimension sequence went from [Country, Product, Warehouse] -> [Country, Product] -> [Country, Product, Warehouse] we would not expect that the final result was the same as the original metric. The CPL analysis process should always choose the sequence of aggregators and spreaders that results in the least information loss.

## Aggregators Definition

One or more aggregators may be defined in a aggregators section which must begin with the `aggregators` keyword. Individual aggregators definitions must be indented one tab deeper than the aggregators keyword. For example, given the concept, unit and type definitions (see [concepts-types-units-measures-syntax.md](concepts-types-units-measures-syntax.md)):

```
concepts
    Country = ['USA', 'Canada']
    Product = ['widget', 'gadget']
    Package = [EACH, PALLET]
```

```
measures
    Count = Count(unit=Package)
```

```
types
    Count = Number(precision=0, unit=Package)
```

```
aggregators
    from [Country, Product] to [Product] = sum(Country)
```

Declares the existence of an aggregator that will transform a metric that has Country and Product among its dimensions to a metric that replaces the set of time series in a metric for every given value of the combination of the values for the Country and Product concepts with a single time series that has a time series for every value of the Product concept where the data values of each cell are the sum of the values of the corresponding cells in the time series for each value of the Country concept. Other dimensions in affected metrics would not be affected.

There's an important aspect of aggregator and spreader expressions that's important to understand - the concepts take on the effective role of units in the calculations because the values being aggregated are spread across different time series in the metric. These "dimensional units" are distinct from the units of the cell type of the metric. "Dimensional units" need never appear in a CPL description or be directly visible to the user, they are implicitly defined by the aggregator and spreader expressions when they name the input and output dimension lists. The dimensional units of a metric with dimensions of [Country, Product] are (Product/Country) or product per country (i.e. for each product value there are many country values). Our goal is to produce a set of time series with dimensional units of (Product).

In order to make sense, the result of the calculation must produce results that have units that are compatible with the output dimensional units of the aggregator. Summing the Country values spread across a Country metric yields a result with dimensional units of Country. Doing the dimensional unit calculation yields a result of `(Product/Country*Country = Product`, which is correct. This will be a check on the correctness of aggregators. It is important to note that these so-called "dimensional unit" relationships are established by the aggregator definition itself. In the raw data sense there is a many to many relationship between Country and Product. However, because we're collapsing a specific dimension in the context of the aggregator we're implicitly implying that there is a set of Country for each Product, and therefore there is a one to many relationship for any given product and a set of countries that have values for that product value.

In addition to the "dimensional unit" math, there may be units associated with the cell types. Taking the sum over a set of values that must by definition have identical units will always be valid, but if there was a more complex calculation involving multiple metrics it is possible that the cell unit validation might fail even if the dimensional unit calculation succeeds.

Since aggregators are never explicitly referenced in a CPL description they don't have names, unlike concepts, unit, metrics and so on.

The full form of an aggregator definitions is:

```
from <list of concepts> to [<one concept name>] = <expression involving concepts>
```

Where `to` may also be the empty list `[]`, in which case the `from` list must contain exactly one concept name. When `to` is `[]`, the aggregator reduces the single dimension to a scalar (no concept dimensions). It is an error for `to` to be `[]` when `from` contains more than one concept.

The concept name that is the target of the `to` clause must be included in the list of concepts that is the target of the `from` clause, otherwise it is an error
with the understanding that the target of the `to` clause may be `[]` (the empty list) only when `from` lists exactly one concept. It is not necessary to list concepts that aren't involved in the expression. Thus, an aggregator that collapses [Country, Product] to [Product] would work equally well if the dimensional grain of some metric was [Country, Product], [Country, Product, Warehouse] or [Country, Region, Product, Warehouse].

## Spreaders Definition

One or more spreaders may be defined in a spreaders section which must begin with the `spreaders` keyword. Individual spreaders definitions must be indented one tab deeper than the spreaders keyword. For example, given the concept, unit and type definitions (see [concepts-types-units-measures-syntax.md](concepts-types-units-measures-syntax.md)):

```
concepts
    Country = ['USA', 'Canada']
    Product = ['widget', 'gadget']
    Currency = ['USD', 'CAD']
```

```
units
    Count
    Currency
```

```
types
    Sales = Number(precision=0, unit=Currency)
    Customer = Number(precision=0, unit=Count)
```

```
spreaders
    from [Product] to [Country, Product] = Product * 1/size(Country)*Country
```

There's an important aspect of aggregator and spreader expressions that's important to understand - the concepts take on the effective role of units in the calculations because the values being aggregated are spread across different time series in the metric. These "dimensional units" are distinct from the units of the cell type of the metric. "Dimensional units" need never appear in a CPL description or be directly visible to the user, they are implicitly defined by the aggregator and spreader expressions when they name the input and output dimension lists. The dimensional units of a metric with dimensions of [Country, Product] are (Product/Country) or product per country (i.e. for each product value there are many country values). Our goal is to produce a set of time series with dimensional units of (Product).

Imagine that some metric is used to compute the sales targets by product world-wide and a second metric divides those sales targets out by country. The world-wide sales_target metric has dimensions of [Product] and the country-specific sales_target metric has dimensions of [Country, Product]. One strategy would be to evenly divide the between the countries. As with aggregators, the concepts involved in the spreader play the role of units in the calculation. When spreading from Product to [Country, Product] the starting metric has dimensional units of `(Product)`, since in the context of the spreader there are many Country values for each product value the result has dimensional units of `(Country*Product)`. In this example the expression `1/size(Country)` where `size(Country)` is the number of countries in the Country concept gives a dimensionless number. When a dimensionless number is multiplied by `(Product*Country)` the result is `(Product*Country)` which is the correct dimensional units for the result.

Conceptually in order to divide one number into another, the system must multiply the number to be divided by a vector to produce a vector. That is, the spread operation can be thought of as `A*[b0, b1, b2]` to produce result `[A*b0, A*b1, A*b2]`. In this case, for every value of the metric with the more aggregated dimension (Product) we must somehow create a vector of the same size as the less aggregated dimension (Country). The spreader definition above does this by creating a vector of the same size as the Country dimension and then multiplying it by each element of the Product dimension. In this example, the number associated with each country is the inverse of the number of countries and the multiplier is (like all concepts) interpreted as the operation it's involved in as being applied to each value associated with the concept.

In the real world it's unlikely that the sales targets would be evenly divided between countries. Instead, the sales targets would be divided according to the number of customers in each country. Suppose that we have another metric which holds actuals and forecasts for the number of customers in each country named customers with a type of Count and dimensions of [Country]. If we somehow knew the fraction of the number of world wide customers belonged to each country we could then multiply that fraction by Sales by Product to do the spread. Unfortunately, there is no data source that will give us that information.

Not all is lost, however. Finding the total number of world-wide customer from the customers metric is trivial, since it's just collapsing the dimensions of Country down to no dimensions, thus:

```
aggregators
    from [Country] to [] = 1/sum(Country)
```

and then in the spreader we can directly name the customers metric in place of the fixed sized constant:

```
spreaders
    from [Product] of Sales to [Country, Product] = Product / customers
```

The expression `Product / customers` seems vaguely to express the idea of spread the product sales over the sets of customers which are broken out by country, but how would Chronocone know what to do? Dimensional unit analysis will be the guide.

The full form of a spreader definition is:

```
from [<list of concepts>] to [<list of concepts plus one more>] = <expression involving concepts>
```

The `<list of concepts plus one more>` that is the target of the `to` clause must include all of the concepts in the list of concepts that is the target of the `from` clause, otherwise it is an error.

The raw customers metric has dimensional units of (Country), since Country is it's only dimension. Product has dimensional units of (Product) by definition. The result of the spreader must have dimensional units of (Country, Product), again by definition. When the spreader expression refers to customers, it could be referring to the basic customer metric with dimensional units of (Country), but if that was the correct choice the dimensional units of the result would be (Product/Country), which doesn't match the require result dimensional unit requirement. In fact, we can tell that in order to get from dimensional units of (Product) to dimensional units of (Product, Country) we would require something that has dimensional units of (1/Country), which happens to be the dimensional units of the aggregator so that will work. The transformation sequence then is as follows:

1. For each cell that represents the same time interval in each time series spread across the customers dimension a vector of values is formed (which has dimensional units of (Country)).
2. That vector is multiplied by the output of the aggregator that collapses [Country] to [] (thus having only one time series in it), which has dimensional units of (1/Country), thus each element of the vector formed in #1 (which is the number of customers in that country) is then divided by the total number of customers in all countries, producing a unitless fraction for each country that represents the proportion of customers in that country.
3. The vector of unitless fractions for each country is then multiplied by the Sales data for each Product value in the source metric, thus dividing out the sales by product world-wide to sales by product and country weighted by the number of customers in each country.

In addition to the dimensional unit correspondence in order to pass the unit checks for the time series cell unit data the spreader must have units of Sales. Since the sequence of #1 and #2 produce a unitless fraction, multiplying that by the cell data unit type of Sales produces a cell unit type of Sales, which is correct.

It's not necessarily obvious at first glance that in order to make this calculation work that we need an aggregator that says

```
aggregators
    from [Country] to [] = 1/sum(Country)
```

and not

```
aggregators
    from [Country] to [] = sum(Country)
```

which might seem more natural in some way, even though we know that in the end finding a multiplier that is the fraction of customers by each country would involve dividing by the total. One of the benefits of dimensional unit analysis is that it makes it obvious that the second choice is wrong because it has dimension units of (Country), when a correct formulation must have dimensional units of (1/Country). We can devise ways to point this out to the user and give hints to AI.

It's also worth noting that the unit analysis for this sequence of operations also gives us confidence in the result. Since the unit type for cell values for the customers metric is Count it might be thought at first glance that we couldn't combine a metric with type Count with a metric with type Sales. However, in this case we're dividing two datums with type Count (customers per country divided by customers per world) to get a unitless fraction, which we can then properly multiply by a type with unit Sales to get a result with type Sales.
