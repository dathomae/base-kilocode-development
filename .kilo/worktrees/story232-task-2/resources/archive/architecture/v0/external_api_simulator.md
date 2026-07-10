Directives for agentic or AI writers and editors:

1. You MUST leave this list of directives in place as the first item in the document and unaltered as changes are made to this document.
2. When editing this document you MUST use either the Technical Writer or Technical Editor roles. If you are not in one of these modes you MUST NOT edit this document and inform the user of the requirement.
3. This document contains annotations that are labelled as "Notes:" or "Questions:" and you MUST NOT edit these annotations. You should move them to keep them with the content they relate to.

--- Begin Document Content ---

# External API Simulator

## Purpose

The External API Simulator provides simulated external data for demonstrations and testing. It serves canned data for use in demos when the system is not yet connected to real external systems. This component is essential for demonstrating the full system workflow without requiring access to production data sources.

The simulator addresses a practical reality of planning system development: real external integrations are complex, time-consuming, and often require access to systems that may not be available during development. Rather than waiting for these integrations to be complete, the simulator allows the ChronoCone system to be demonstrated and tested with realistic data that mimics what the real integrations would provide.

This approach supports the v0 architecture's emphasis on demonstrability. Stakeholders can see the complete system in action, from data ingestion through calculation to visualization, without needing to understand or access the external systems that would eventually provide the data. The simulator also provides a stable, predictable data environment for testing, which is valuable for validating system behavior and reproducing issues.

## Responsibilities

The External API Simulator has the following responsibilities:

1. **Metric Data Supply**: Provide data for individual metrics through the `/metric_data` endpoint, returning values that match the types defined for each metric in the database.

2. **Type Lookup**: Query the Metrics Service to determine the correct data type for a requested metric before generating or returning values.

3. **Configuration Management**: Load and interpret optional configuration files that specify predetermined data values for specific metrics and time ranges.

4. **Random Value Generation**: Generate random values when no configuration exists for a requested metric, ensuring values fall within appropriate ranges and formats.

5. **Data Type Support**: Support integer, float, and string data types as defined in the metric metadata.

## API Endpoint

### `/metric_data` Endpoint

The simulator exposes a single primary endpoint for retrieving simulated metric data.

**HTTP Method**: GET

**Query Parameters**:

| Parameter     | Type   | Required | Description                                                                                                                                                       |
| ------------- | ------ | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `metric_name` | string | Yes      | The name of the metric for which data is being requested                                                                                                          |
| `time_grain`  | string | Yes      | The temporal interval for the data (e.g., "hour", "day", "week", "month")                                                                                         |
| `time_range`  | string | Yes      | The time range for the data, specified as ISO 8601 formatted start and end times separated by a forward slash (e.g., "2024-01-01T00:00:00Z/2024-01-31T23:59:59Z") |
| `data_type`   | string | Yes      | Either "actuals" or "forecast" to specify which type of data to return                                                                                            |

**Response**: JSON array of values, where each value corresponds to a time point within the requested range. The data type of each value matches the metric's defined type.

**Example Request**:

```
GET /metric_data?metric_name=units_sold&time_grain=day&time_range=2024-01-01T00:00:00Z/2024-01-07T23:59:59Z&data_type=actuals
```

**Example Response** (for an integer metric):

```json
[342, 456, 289, 512, 678, 423, 389]
```

## Configuration File Format

The configuration file uses YAML format to specify predetermined values for metrics. The file structure allows specifying both actuals and forecast values for specific time ranges.

```yaml
metrics:
  - name: units_sold
    actuals:
      - time_range: "2024-01-01T00:00:00Z/2024-01-07T23:59:59Z"
        values: [350, 400, 300, 500, 650, 450, 400]
      - time_range: "2024-01-08T00:00:00Z/2024-01-14T23:59:59Z"
        values: [420, 380, 510, 490, 560, 430, 410]
    forecast:
      - time_range: "2024-02-01T00:00:00Z/2024-02-07T23:59:59Z"
        values: [380, 420, 350, 480, 600, 440, 390]
  - name: customer_count
    actuals:
      - time_range: "2024-01-01T00:00:00Z/2024-01-31T23:59:59Z"
        values:
          [
            1200,
            1350,
            1100,
            1450,
            1600,
            1280,
            1150,
            1300,
            1400,
            1250,
            1500,
            1700,
            1320,
            1200,
            1280,
            1380,
            1180,
            1480,
            1650,
            1300,
            1180,
            1250,
            1350,
            1150,
            1420,
            1580,
            1260,
            1120,
            1220,
            1320,
            1120,
            1400,
          ]
```

When a request matches a configured metric and time range, the simulator returns the configured values. If the request falls outside the configured ranges, the simulator generates random values.

## Random Value Generation

When no configuration exists for a requested metric and time range, the simulator generates random values according to the metric's data type:

| Data Type | Generation Rule                                           |
| --------- | --------------------------------------------------------- |
| Integer   | Random integer in the range [100, 1000]                   |
| Float     | Random floating-point number in the range [100.0, 1000.0] |
| String    | Random alphanumeric string of exactly 10 characters       |

The number of values generated corresponds to the number of time points in the requested range, based on the specified time grain.

## Design Considerations

The design of the External API Simulator must balance several competing concerns.

**Realism** ensures that the simulated data is representative of what real external systems would provide. This includes realistic value ranges, appropriate data types, and plausible patterns over time. Unrealistic data can lead to false conclusions about system behavior and may not reveal issues that would occur with real data.

**Reproducibility** ensures that the same simulation scenario produces the same results each time it is run. This reproducibility is essential for testing, where the same test case should produce the same outcome consistently. It also supports debugging, where being able to reproduce a specific scenario is often the key to identifying and fixing problems.

**Configurability** allows different simulation scenarios to be defined and selected. Different demonstrations or tests may require different data patterns, and the simulator should support this variety without requiring code changes. Configuration might include the time range of data, the values returned, the timing of responses, and the introduction of specific conditions like errors or delays.

**Integration transparency** means that the simulator should be interchangeable with real external systems from the perspective of the ChronoCone components that use it. This transparency ensures that the system can transition from simulated to real data sources without requiring changes to the core logic.

## Related Documentation

- [v0 Architecture Overview](v0_architecture.md)
- [v0 Requirements - External API Simulator Section](../requirements/v0_requirements/v0_requirements.md#36-external-api-simulator)
- [Connectors](v0_architecture.md#connectors) in the architecture overview
