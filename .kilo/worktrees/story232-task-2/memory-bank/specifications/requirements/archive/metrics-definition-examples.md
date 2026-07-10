# Metric Definition Examples

This document contains definitions for some plans that illustrate the concepts and syntax of the
metric definition DSL.

For these example the pound sign line comments are statements about data that comes from somewhere
other than the DSL, such as the relational database.

Note: At the moment the example is rather stream-of-conciousness rather than completely thought out,
so there are probably holes in this. Some conventions that I created on the fly:

1. ':=' is the 'create a definition' operator, which can be used in either the formula form of a metric
   or the functional form.
2. In the functional using the ':=' to assign to the value of a metric is
   defining the return value. If we allow the user of ':=' in a metric definition for anything other than
   assignment to the metric name, then we'd be creating sub-metrics. I'm not sure if that's a good idea
   or whether it should be allowed or not, but it's what the natural interpretation implies.

3) Every dimension name defines a type that can be passed to a definition of any type.

## Example 1: Retail Forecast

Note: The definitions in this example are a bit schizophrenic in that some of the earlier connector definitions
include some typing information in their definition, but the later definitions mostly don't. From an implementation
point of view I think that we can store typing information in the db or have it explicit in the definition and make it
work either way. The question
is what's going to work best for the typical user. People tend to like languages where typing data isn't required, but
I think that making typing explicit helps people recognize and understand dimension and/or typing misalignment. Given the
nature of the system and the fact that we're going to include AI to help users figure things out I suspect that either going
without explicit typing or making it optional is the way to do, but we should talk about it.

The goal of the retail forecast is to forecast how much of each product at a country level needs to
be ordered each week, when it will arrive at a warehouse and how much top line revenue will result.

### Plan Vocabulary

This plan defines two data types that can be held in time series:
Unit[int] - This represents the number of units associated with a metric.
LeadTime[int] - This represents the number of days lead time that comes between an order and the delivery of the order.

The plan also uses the following dimensions:
Product[Product] - This represents a product.
Country[Country] - This represents a country.
Price[currency] - This represents the price of a product.

### Inputs

The these are the outside inputs to the Retail plan:

1. The demand forecast by Product and Price, which is the forecasted number of
   units that would be expected to be sold for a particular product at a particular price point. This data comes from
   an external demand forecasting service.
2. The actual number of units sold in the past, which is obtained from a sales tracking service.
3. The current and past inventory held in each warehouse broken out by Product and Warehouse.
4. The lead time that will elapsed before ordering a product and having it arrive at a warehouse.

Each of these is obtained by a Connector and produces a document of some kind.

The demand forecast service connector is defined as follows:

```
# It uses a driver that connects to the demand forecasting service (from database)
# The output of the connector can be a metric, a metric aspect or a lookup.
# The dimensions of this Aspect are [Product, Price] and the datatype is Unit
connector demand_forecast_forecast(demand_forecast_driver: Driver[Aspect[Unit]],
                                   number_of_weeks:int)->Aspect:
    # Connectors will sometimes need to pass parameters to drivers, those
    # parameters will need to know in the connector definition or configuration.
    # This fetches forecasts for the next number_of_weeks weeks.
    demand_forecast_forecast := demand_forecast_driver.get(now(), number_of_weeks)
```

The demand forecast actuals connector is defined as follows:

```
# It uses a driver that connects to the sales tracking service (from database)
# The output of the connector can be a metric, a metric aspect or a lookup.
# The dimensions of this Aspect are [Product, Price] and the datatype is Unit
connector demand_forecast_actuals(demand_forecast_actuals_driver: Driver[Aspect[Unit]],
                                 number_of_weeks:int)->Aspect:
    # This fetches data back from the current date to the given number of weeks.
    demand_forecast_actuals := demand_forecast_actuals_driver.get(now(), number_of_weeks)
```

The inventory actuals connector is defined as follows:

```
# It uses a driver that connects to the inventory tracking service (from database)
# The output of the connector can be a metric, a metric aspect or a lookup.
# The dimensions of this Aspect are [Product] and the datatype is Unit
connector inventory_actuals(inventory_driver: Driver[Metric[Unit]],
                            number_of_weeks:int)->Metric:
    # This fetches data back from the current date to the given number of weeks.
    inventory_actuals := inventory_driver.get(now(), number_of_weeks)
```

We need to have a place to look up the lead time for a product. Here we're assuming it comes from an
external service.

```
# Assume that the lead time for a product is something we read from the Vendor Orders Service. Also assume for this
# example that lead time doesn't vary by vendor and that there are no limits on the amount ordered or price
# advantage for larger orders.
connector order_lead_time(vendor_orders_service_driver: Driver[Lookup[LeadTime]],
                       product: Product)->Lookup[Product, LeadTime]:
    order_lead_time := vendor_orders_service_driver.get_order_lead_time(product)
```

## Plan Definitions

```
# The demand_forcast metric is the combination of the forecasts aspect from the forecast service
# and the actuals aspect from the sales tracking service.
metric demand_forecast(demand_forecast_forecast: Aspect[Unit],
                       demand_forecast_actuals: Aspect[Unit]) :=
    demand_forecast := Metric(forecast=demand_forecast_forecast, actuals=demand_forecast_actuals)
```

```
# We talked about a different name for the thing that transforms dimensions, but I don't remember the term
# right now. Going with 'alinger' for the moment.
# The aligner collapses the [Product, Price] dimension of the demand_forecast metric to [Product] by selecting the
# time series with the highest price for each product.
aligner demand_forecast_aligner(demand_forecast: Metric[Unit]) := demand_forecast[max(demand_forecast.Price)]
```

```
# The revenus metric is the product of the values in the various time series in the demand_forecast metric with the
# value of the price dimension for each of the time series
# demand_forecast has dimensions of [Product, Price] and since it's the only metric named
# the new metric 'revenue' also has dimension of [Product, Price]
metric revenue := demand_forecast * Price
```

```
# The inventory forecast for each week is the inventory from the previous week plus the inventory arrivals for the week
# minus the demand forecast for the week
# inventory and inventory arrivals has dimensions of [Product] (the assumption with these examples is that output dimensions
# of a metric are captured in the database, which is why there's no mention of them in the metric definition, we could potentially
# to the other way if we wanted to), while demand_forecast has dimensions of [Product, Price]
# The calcuator will look for a sequence of aligners that remove the Price dimension
# Note that inventory self-reference means that this is not a DAG at the metric level, but is' still
# a DAG at the time series element level.
metric inventory:
    # We need to pull the actuals into the inventory metric, since the lag operation will refer to the
    # actuals for the first cell in the time series.
    result = Metric(actuals=inventory_actuals)
    inventory := lag(result,1) + inventory_arrivals - demand_forecast
```

```
# inventory_arrivals is the amount of product that we want to arrive in any given week in the future.
# It has dimesions of [Product]. The aligner will reduce the demand_forecast to just [Product] by
# choosing the maximum price in the demand forecast. The company policy is that we try to keep a
# 20% 'inventory pad' against errors in the demand forecast at the high side.
# If this metric had been defined as
# metric inventory_arrivals := demand_forecast * 1.2 - inventory
# Then the computation graph would no longer be a DAG at the timeseries element level since
# for any given time period inventory would depend on inventory_arrivals but inventory_arrivals
# would depend on inventory. To solve that kind of problem we'd have to reach for optimization tools rather
# than just creating a data flow graph, but it could be done.
metric inventory_arrivals := demand_forecast * 1.2 - lag(inventory,1)
```

```
# Given that we need inventory to arrive at the time designated by inventory_arrivals,
# then the orders need to be placed the number of days before they need to arrive.
# In this formulation inventory_orders has dimensions of [Product] and type of Unit[int].
# Since the lookup result has type LeadTime[int] but is in days, the time grain needs to be
# converted.  This may not be the right way to do it, should this be done by a time grain aligner?
# This is also a case where a forecast can't be allowed to update actuals. We're handing
# that in this case by using lead(inventory_arrivals) here rather than using lag(inventory_orders)
# in the inventory_arrivals definition, but I could see where that distinction could be a bit
# subtle. Do we need to do anything like that?
metric inventory_orders(product: Product) := lead(inventory_arrivals, lookup(product)/7)
```

## Outputs

What the 'output' of this plan is depends on who's looking at it. If the group that places orders
is looking at it the output is inventory_orders. If the retail/sales group is looking at it
the output is revenue. If you're building a labor plan for a warehouse the output is inventory_arrivals.
