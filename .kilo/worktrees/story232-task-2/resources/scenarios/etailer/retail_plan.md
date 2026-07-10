# Retail Planning Example

## Overview

The retail group is responsible for maximizing **net revenue** across the product network. Net revenue is calculated as:

```
Net Revenue = (Units Sold × Price per Unit) - (Units Bought × Cost per Unit)
```

Unsold inventory represents a cost that detracts from net revenue. When inventory accumulates beyond anticipated needs, the retail group can initiate sales or promotions to clear stock.

## Planning Scope

The retail group operates at the **network level** only—forecasts and plans are not broken down by individual warehouses. Therefore, warehouse is never a category in any retail planning metric.

## Time Granularity

All metrics in this plan use **week** as the time grain. Time is a property of the metric itself, defining the spacing and range of the time series data points. It is not a category that varies within the metric.

## Input Metrics

### 1. Demand Forecast

The demand forecast predicts how many units will sell at various price points. **Price is a category** of this metric, with one time series for each price point.

| Category | Description                                              |
| -------- | -------------------------------------------------------- |
| Product  | The specific product being forecasted                    |
| Price    | The price point at which demand is forecasted (Currency) |

**Time Grain**: Week

**Value**: Number of units forecasted to be sold

### 2. Vendor Availability

Vendor availability describes what products are available from vendors, including pricing that varies by quantity ordered. Larger quantities typically have lower per-unit costs but longer lead times. Each time series has both a minimum quantity threshold and an associated price.

| Category | Description                                                     |
| -------- | --------------------------------------------------------------- |
| Product  | The product available for purchase                              |
| Vendor   | The supplier offering the product                               |
| MinQty   | The minimum order quantity for this price/lead time combination |

**Time Grain**: Week

**Values**:

- Unit Price (Currency)
- Lead Time (weeks)
- Quantity Available

Some products may be available from multiple vendors, creating sourcing alternatives.

## Constraints and Business Rules

### Inventory Buffer Constraint

To ensure we never run out of units to sell, the retail group maintains an inventory buffer. However, to minimize carrying costs, this buffer must not exceed **20% of anticipated sales** for the selected price point in any given week.

This constraint creates a tension: sometimes the retail group must order smaller quantities with shorter lead times (even at higher per-unit costs) to avoid stockouts, rather than waiting for larger, cheaper orders that might arrive too late.

### Vendor Selection Trade-offs

For each product and week, the planning process must evaluate:

- Which vendor(s) to order from
- What quantity to order (affecting both price and lead time)
- When to place the order to ensure arrival before stockout

## Output Metrics and Transformations

The retail planning process produces the following bottom-line metrics through a series of transformations:

### 1. Orders Submitted

| Category | Description                    |
| -------- | ------------------------------ |
| Vendor   | The vendor receiving the order |
| Product  | The product being ordered      |

**Time Grain**: Week (the week in which the order is submitted)

**Values**:

- Order Quantity
- Expected Arrival Week

### 2. Item Cost

| Category | Description |
| -------- | ----------- |
| Product  | The product |

**Time Grain**: Week (the week for which the cost applies)

**Value**: Weighted average cost per unit based on orders placed

### 3. Anticipated Vendor Cost

| Category | Description |
| -------- | ----------- |
| Vendor   | The vendor  |

**Time Grain**: Week (the week in which payment is expected)

**Value**: Total cost expected to be paid to the vendor

### 4. Unit Sales

| Category | Description |
| -------- | ----------- |
| Product  | The product |

**Time Grain**: Week (the week of anticipated sales)

**Value**: Number of units expected to sell at the selected price point

### 5. Top Line Revenue

| Category | Description |
| -------- | ----------- |
| Product  | The product |

**Time Grain**: Week (the week of anticipated revenue)

**Value**: Unit Sales × Selected Price

### 6. Net Revenue

| Category | Description |
| -------- | ----------- |
| Product  | The product |

**Time Grain**: Week (the week of anticipated net revenue)

**Value**: Top Line Revenue - (Units Bought × Cost per Unit) - Carrying Cost of Unsold Inventory

## Transformation Flow

```
Demand Forecast (by Product, Price)
         ↓
    [Select Price Point for each Product/Week]
         ↓
Selected Price + Anticipated Sales
         ↓
Vendor Availability (by Product, Vendor, MinQty)
         ↓
    [Select Vendor and Quantity considering:]
    [- Price per unit]
    [- Lead time vs. inventory buffer needs]
    [- 20% max buffer constraint]
         ↓
Orders Submitted → Item Cost → Anticipated Vendor Cost
         ↓
Unit Sales → Top Line Revenue → Net Revenue
```

## Downstream Consumers

The outputs of the Retail plan are consumed by:

1. **Vendor Ordering System**: Receives the Orders Submitted metric to place actual orders with vendors
2. **Customer Service Group**: Reads unit sales and pricing to set delivery expectations
3. **Logistics Group**: Reads anticipated unit sales to plan warehouse operations and transportation capacity

## Detailed Metrics Specification

This section lists each metric used in or produced by the Retail plan with complete structural details.

### Metric 1: Demand Forecast

Input metric sourced from an external demand forecasting system.

| Attribute            | Value           | Description                                                             |
| -------------------- | --------------- | ----------------------------------------------------------------------- |
| **Namespace Family** | Retail          | Domain of control for the retail organization                           |
| **Namespaces**       | Prod, Dev       | Prod = officially published plan; Dev = plan under development/approval |
| **Metric Name**      | demand_forecast | Predicted unit sales by product and price point                         |
| **Time Grain**       | Weekly          | Each data point represents one week                                     |
| **Forecast Length**  | 6 months        | 26 weeks of forecast data                                               |

**Categories:**

| Category | Meaning                                      | Example Values        |
| -------- | -------------------------------------------- | --------------------- |
| Product  | The specific product being forecasted        | SKU-12345, Widget-Pro |
| Price    | The price point at which demand is predicted | $19.99, $24.99        |

**Value Type:** Number (units forecasted to be sold)

**Data Source:** External demand forecasting system via adapter

---

### Metric 2: Vendor Availability

Input metric sourced from vendor systems describing product availability, pricing, and lead times.

| Attribute            | Value               | Description                                                             |
| -------------------- | ------------------- | ----------------------------------------------------------------------- |
| **Namespace Family** | Retail              | Domain of control for the retail organization                           |
| **Namespaces**       | Prod, Dev           | Prod = officially published plan; Dev = plan under development/approval |
| **Metric Name**      | vendor_availability | Product availability with price/quantity/lead time terms                |
| **Time Grain**       | Weekly              | Each data point represents one week                                     |
| **Forecast Length**  | 6 months            | 26 weeks of availability data                                           |

**Categories:**

| Category | Meaning                                         | Example Values     |
| -------- | ----------------------------------------------- | ------------------ |
| Product  | The product available for purchase              | SKU-12345          |
| Vendor   | The supplier offering the product               | Vendor-A, Vendor-B |
| MinQty   | Minimum order quantity for this price/lead time | 100, 500, 1000     |

**Values (per time series):**

- Unit Price (Currency)
- Lead Time (weeks)
- Quantity Available (number of units)

**Data Source:** External vendor availability systems via adapter

---

_Additional metrics (Orders Submitted, Item Cost, Anticipated Vendor Cost, Unit Sales, Top Line Revenue, Net Revenue) to be documented in subsequent iterations._

## Dynamic Adjustments

The retail planning process must adapt to short-term changes:

- **Demand spikes**: A celebrity's death driving music sales, a product going viral
- **Supply disruptions**: Vendor stockouts, extended lead times
- **Product recalls**: Immediate need to halt sales and adjust inventory positions

When forecasts miss and unwanted inventory accumulates, the retail group initiates clearance sales or promotions to restore the inventory buffer to acceptable levels while maximizing recovery value.

---

## Calculation: min_order

### Inputs

1. **demand_forecast** - Time series indexed by product and price
2. **current_inventory** - Integer number of units indexed by product

### Output

**min_order** - Metric indexed by product and price

### Description

For each product and price in the demand_forecast, min_order represents the number of units that would need to be ordered to satisfy the demand_forecast for that product and price given the current inventory.

### Calculation Logic

**First week:**

```
min_order = max(0, demand_forecast - current_inventory)
```

**Subsequent weeks:**

```
min_order = max(0, demand_forecast - current_inventory - sum of demand_forecast in prior weeks)
```

### Notes

- The calculation accounts for inventory depletion over time
- In week 1, we check if current inventory can satisfy the initial demand
- In later weeks, we accumulate prior weeks' demand to determine how much inventory remains
- The max(0, ...) ensures we never return negative order quantities

### Python Metric Definitions

#### Forecast Version

```python
"""
Transformer: inventory_shortfall_forecast
Namespace Family: Retail
Type: Forecast

This transformer is registered in the database with:
- Input metrics: demand_forecast (Product, Price dimensions), inventory_actuals (Product dimension)
- Output metric: inventory_shortfall (Product, Price dimensions)
- Shifter: inventory_actuals is shifted by replicating across Price dimension

The calculation engine automatically invokes the shifter to align inventory_actuals
(Product only) with demand_forecast (Product, Price) before calling this transformer.
"""

def inventory_shortfall_forecast(
    demand_forecast: int,  # Dimensions: Product, Price
    inventory_actuals: int  # Dimensions: Product, Price (after shifter replication)
) -> int:
    """
    Calculate minimum order quantity needed to satisfy demand.

    First week: min(0, inventory_actuals - demand_forecast)
    Subsequent weeks: min(0, inventory_actuals - sum of prior demand)
    """
    return min(0, inventory_actuals - sum(demand_forecast))
```

### Notes on Implementation

- **Forecast Calculation**: The forecast version uses `inventory_actuals` that are shifted by replicating across the Price dimension (which inventory doesn't have) and across the forecast time periods to match the grain of `demand_forecast`. The shifter is automatically invoked by the calculation engine.

- **Aggregation**: The unadorned `sum(demand_forecast)` in the function body operates within the framework context, summing only the forecast elements that precede the current entry being calculated.

- **Actuals Source**: The actuals version delegates to an `InventoryAdapter` which lazily batches calls for efficient retrieval from external inventory systems.

- **Dimensions**: The `demand_forecast` input has Product and Price dimensions. The `inventory_actuals` has only Product, requiring a shifter to replicate values across the Price dimension to align the grains.
