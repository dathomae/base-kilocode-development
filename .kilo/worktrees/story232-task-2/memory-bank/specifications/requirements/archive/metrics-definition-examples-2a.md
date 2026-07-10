# Slightly extended from metric-definition-examples-2.md

## Taxonomic Declarations

```
concepts
    retail/kpis
        demand
    fulfillment/kpis
        inventory
        inbound_shipment
        outbound_shipment
    retail/products
        category
            product
              price
              revenue
    global/geo
        country
            region
                state
                    gov_postal_code
            metro
                metro_postal_code
    fulfillment/service_zones
        warehouse
            service_zone
                service_postal_code
                    gov_postal_code -> I added this, my asumption is that when the same thing is named twice it refers to the same concept in a different context. From this declaration we can say that there's a one to many relationship between service_postal_code and gov_postal_code. From the declaration that there's a one to many relationship between region and gov_postal_code above, we can deduce that there's a many to many relationship between service_zone and region since there's a many to one relationship between region and state and a many to one relationship between state and gov_postal_code.
```

I'm also going to assume that the '#' sign is a line comment character in the DSL/language.

---- Everything after this point is an actual part of the DSL ---

```
# Unit declarations

units
    # Unit and lead time are the possible data types in a time series in this example.
    unit = int
    lead-time = Days

dimensions # Dimensions have type, but are different from units, so added this declaration section
    # price is a concept that uses currency as it's underlying type. Because it's associated with
    # price in the taxonomy we know there there is a many to one relationship between price and product.
    price =  price[currency]
    revenue = revenue[currency]

    # country is a concept that uses string as it's underlying type. Because it's a super-concept
    # for region, we know that there is a many to one relationship between region and country.
    # because region is a super-concept for state, we know that there is a many to one relationship
    # between state and region, and therefore between state and country.
    country = country[string]

    product = product[string]
    region = region[string]
    warehouse = warehouse[string]


# Importers

# The combination of the demand_forecast_importer and the sales_actuals_importer creates the
# demand_forecast metric that includes both forecast and actuals aspects, although each importer
# supplies only one aspect.
# I don't think that connectors need names because it seems more natural to refer to the metrics or
# lookups they produce in other connectors as opposed to the connectors/importers themselves, since
# the metric/lookup they name are a sufficient handle to figure out the reference.
connector
    metric
        # This connector creates or adds to a metric called 'demand_forecast' with units of 'unit'
        # and dimensions of 'country', 'product' and 'price'
        demand_forecast as unit by [country, product, price]


connector
    metric
        # This connector creates or adds to a metric called 'demand_forecast' with units of 'unit'
        # and dimensions of 'country' and 'product'
        demand_forecast as unit by [country, product]
    driver=https_api
    secret_path=/i/am/chronocone
    server=sales.internal.foo.com
    path=/v1/actuals_record
    params
        start=$fetch_time
        lookback=$number_of_weeks

        # There's no product parameter to the salesactuals endpoint, it returns a paged list of
        # all of the unit sales for each product in the given country.
        country=country
    trigger=weekly
    grain=weekly

    # This connector supplies only the actuals aspect of the demand_forecast metric.
    horizon=actuals,
    on missing country
        error "unrecognized country ${country}"
    on missing product
        fetch from product_table[${country}]

connector
    metric
        # This connector creates a metric called 'inventory_actuals' with units of 'unit' and dimensions
        # of country, region, product and warehouse
        inventory_actuals as unit by [country, region, product, warehouse]
    driver=https_api
    secret_path=/i/am/chronocone
    server=inventory_service.internal.foo.com
    path=/v1/actuals_record
    params
        start=$fetch_time
        lookback=$number_of_weeks
    trigger=weekly
    grain=weekly

    # This connector supplies only the actuals aspect of the demand_forecast metric.
    horizon=actuals,
    on missing region
        error "unrecognized region ${region}"
    on missing country
        error "unrecognized country ${country}"
    on missing warehouse
        fetch from warehouse_lookup_importer
    on missing product
        fetch from product_table[${country}]

connector
    metric
        inventory_actuals as unit by [country, region, product, warehouse]
    driver=https_api
    secret_path=/i/am/chronocone
    server=inventory_service.internal.foo.com
    path=/v1/actuals_record
    params
        start=$fetch_time
        lookback=$number_of_weeks
    trigger=daily
    grain=daily

    # This connector supplies only the actuals aspect of the demand_forecast metric.
    horizon=actuals,
    on missing region
        error "unrecognized region ${region}"
    on missing country
        error "unrecognized country ${country}"
    on missing warehouse
        fetch from warehouses[${region}, ${country}]
    on missing product
        fetch from product_table[${country}]

connector
    lookup
        # This connector creates a lookup called 'order_lead_time' with units of 'lead-time'
        # and dimensions of 'country' and 'product'. In other words, if the caller supplies a country
        # and product the lookup will return the lead time for that product in that country.

        order_lead_time as unit by [country, product]
    driver=https_api
    secret_path=/i/am/chronocone
    server=vendorservice.internal.foo.com
    path=/v1/leadtime
    params
        # There's no product parameter to the vendorservice endpoint, it returns a paged list of
        # the lead times for each product in the given country.
        country=country
    trigger=monthly

    # Note, there is no 'grain' for a lookup, they're just values that are updated from time to time.

    on missing region
        error "unrecognized region ${country}"
    on missing country
        error "unrecognized country ${country}"
    on missing product
        fetch from product_lookup_importer

product_lookup_importer = connector
    lookup
        # This connector creates a lookup called 'product_table' with units of 'product' and dimensions of 'country'
        product_table as product by [country]
    driver=https_api
    secret_path=/i/am/chronocone
    server=sales.internal.foo.com
    path=/v1/productlist
    params
        # There's no product parameter to the vendorservice endpoint, it returns a paged list of
        # the lead times for each product in the given country.
        country=country
    trigger=monthly

    on missing region
        error "unrecognized region ${country}"
    on missing country
        error "unrecognized country ${country}"

warehouse_lookup_importer = connector
    lookup
        # This connector creates a lookup called 'warehouses' with units of 'warehouse' and dimensions of 'region', 'country'
        warehouses as warehouse by [region, country]
    driver=https_api
    secret_path=/i/am/chronocone
    server=logistics_service.internal.foo.com
    path=/v1/warehouse_list
    params
        region=region
        country=country
    trigger=monthly

    on missing region
        error "unrecognized country ${region}"
    on missing country
        error "unrecognized country ${country}"

# The metrics refer to warehouses by country, but the connector return warehouses by region
# and country.
aggregator from [country, region] to [country] =
    inputs
        warehouses as warehouse by [country, region]
    aggregate
        # Use the contatenation operation across the warehouses in a region to bring
        # them up to the country level.
        concat(warehouses[region])

# Metrics

# Revenue is the product of the demand forecast and the price for each product.
revenue = metric by [country, product]
    inputs
        demand_forecast as revenue by [country, product, price]
    grain=weekly
    calculate
        demand_forecast * price


# The dimensions for the demand_forecast metric are [country, product, price], but the
# dimensions for revenue are [country, product]. The calculation engine will look for an aggregator
# that can align the dimensions of demand_forecast to the dimensions of revenue. In this case,
# we'll just choose the maximum price for each product in each country. This would be a case
# where the more sophisticated algorithms would specify that we wanted to maximum the revenue
# given this definition of revenue, so Chronocone would automatically choose the correct price
# rather than deriving it from explicit logic.
# I don't think that the aggregators or spreaders need explicit names, since the calculator
# is going to figure out which one(s) need to be used.
aggregator from [country, product, price] to [country, product] =
    inputs
        # Here we're naming the metrics that the aggregator can work on. It would be possible to
        # come up syntax that would apply to any metric that has a certain unit type to the same
        # metric with the with reduced dimensions.
        demand_forecast as unit by [country, product, price]
        revenue as revenue by [country, product]
    aggregate
        max(price)

# Possible optimization related version, although by calling it a chooser the
# designer could also include an explicit calcuation if they wanted to without
# changing the concept.
#choose from [country, product, price] to [country, product] =
#    inputs
#        demand_forecast as unit by [country, product, price]
#        revenue as revenue by [country, product]
#    choose
#        price
#    goal
#        maximize revenue

# inventory_arrivals is the actual and forecast arrival unit volume for each product in the country
inventory_arrivals = metric by [country, product]
    inputs
        # This is an implicitly calculated version of demand_forecast that needs to be created using
        # an aggregator.
        demand_forecast as unit by [country, product]

        # inventory_actuals has a daily grain, but the output has a weekly grain, so the
        # time oriented aggregator will be needed to do the conversion.
        inventory_actuals as unit by [country, product] where ACTUAL
    grain=weekly
    calculate
        if inventory_actuals
            inventory_actuals
        else
            demand_forecast * 1.2 - lag(inventory,1)

# inventory_actuals comes from the connector with dimensions of [country, region, product, warehouse],
# but is used by the inventory_arrivals metric on the [country, product]. The following two
# aggregators are used successively to align the dimensions.
# This example uses aggregators in cascade. It should be possible to use a single aggregator with
# sum cascaded directly, as in sum(sum(inventory_actuals[warehouse])[region])
aggregator from [country, region, product, warehouse] to [country, region, product] =
    inputs
        inventory_actuals as unit by [country, region, product, warehouse]
    aggregate
        sum(inventory_actuals[warehouse])

aggregator from [country, region, product] to [country, product] =
    inputs
        inventory_actuals as unit by [country, region, product]
    aggregate
        sum(inventory_actuals[region])

# inventory_arrivals actuals are on a daily grain, but they need to be converted to a weekly grain
# to work with the other metrics.
aggregator from daily [country, product] to weekly [country, product] =
    inputs
        inventory_arrivals as unit by [country, product] where ACTUAL
    aggregate
        # The assumption here is that a 'week' for the planner starts on Monday and ends on Sunday.
        # If the day is a sunday than the daily grain to weekly grain conversion is the sum of the
        # previous 7 days, otherwise it's nothing.  There is no change in the dimensions, so we don't
        # need any dimensional logic.
        if day == Sunday
            sum(lag(inventory_arrivals,6), lag(inventory_arrivals,0))

# inventory is the actual and forecast inventory level for each product in the country
inventory = metric by [country, product]
    inputs
        demand_forecast as unit by [country, product]
        inventory_arrivals as unit by [country, product] where ACTUAL
    grain=weekly
    calculate
        # A self-referential calcuation at the metric level, still an easy calc as long as it's a DAG at the
        # time series element level. If we want to allow cycles, we'll need to fall back on
        # solving simultaneous equations and/or optimization techniques to solve for valid
        # values.
        lag(inventory,1) + inventory_arrivals - demand_forecast

# inventory_orders is the actual and forecast orders by week for each product in each country
inventory_orders = metric by [country, product]
    inputs
        inventory_arrivals as unit by [country, product]
        order_lead_time as unit by [country, product]
    grain=weekly
    calculate
        # This example is assuming that the order lead time is specified in days, but the
        # metric is on a weekly grain so we need to convert.
        lead(inventory_arrivals, ceiling(order_lead_time[product],7))

```
