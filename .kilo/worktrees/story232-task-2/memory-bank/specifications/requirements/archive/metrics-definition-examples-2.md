
Taxonomic Declarations
---
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
    global/geo
        region
            country
                state
                    gov_postal_code
            metro
                metro_postal_code
    fulfillment/service_zones
        warehouse
            service_zone
                service_postal_code

units
    each = product[int]
    lot = product[int]
    load = weight[float]
    package = volume[float]
```

Importers
---
```
demand_forecast_importer = connector
    metrics
        demand as lot by [region, category]
    driver=https_api
    secret_path=/i/am/chronocone
    server=demand.internal.foo.com
    path=/v1/demand_forecast
    params
        start=$fetch_time
        duration=$number_of_weeks
    trigger=weekly
    on missing region
        error "unrecognized region ${region}"
    on missing category
        fetch from category_lookup_importer

sales_actuals_importer = connector
    metrics
        inbound_shipment as lot by [region, country, warehouse, product]
        outbound_shipment as each by [region, country, warehouse, product]
        inventory as each by [region, country, warehouse, product]
    driver=https_api
    secret_path=/i/am/chronocone
    server=actuals.internal.foo.com
    path=/v1/actuals_record
    params
        start=$fetch_time
        lookback=$number_of_weeks
    trigger=daily
    on missing region
        error "unrecognized region ${region}"
    on missing country
        error "unrecognized country ${country}"
    on missing warehouse
        fetch from warehouse_lookup_importer
    on missing product
        fetch from product_lookup_importer

warehouse_lookup_importer = connector
    taxons
        fulfillment/service_zones
            warehouse
                service_zone
                    service_postal_code
    ? what's the right way to define this?
```

Aligners
---
```

```

Calculations
---
```
demand as each by [region, country, warehouse, category] =
    inputs
        inventory where ACTUAL
        inbound_shipment where ACTUAL
        demand where FORECAST
    calculate
        if inbound_shipment is present
            inventory + inbound_shipments[-1] - inventory[-1]
        else
            demand
```
Note: input units (each) and dimensions ([region, country, warehouse, category]) is required to be identical to the calculation.

