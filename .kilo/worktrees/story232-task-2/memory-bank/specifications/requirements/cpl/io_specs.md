# IO Specifications

Chronocone CPL specifications are only partial specifications in that
the same specification may consume a variety of inputs and produce a variety of outputs that have different unit types, different dimensions and different time grains. These things are specified not in the CPL or other specification (e.g., definitions taken from the database), but in an _IO Specification_ which is supplied to the CPL compiler when an executable is generated. Like other definitions in Chronocone it is valid to have IO specifications defined in different forms, whether that be file formats or taken from a service API or database provided that they conform to the rules for IO specifications.

## IO Specification Contents

IO specifications shall include the following:

- One or more inputs that name a connector or connectors in the plan specification (e.g., read from one or more CPL files) and the corresponding data source that include the following:
  - The name of the connector in the plan specification.
  - The format of the data source.
  - The driver used to access the data source.
  - Zero or more driver parameters.
- One or more outputs that include the following:
  - The name of the metric in the plan specification.
  - The dimensions of the metric.
  - The celltype of the metric.
  - The grain of the metric.
- A specification for the compute and memory configuration that the synthesized go code will run with. While it is likely that an executable will still run on a machine that has different specifications than in present in the IO Specification, it is likely that it will make less efficient use of the resources available and thus run more slowly. The specification assumes a simplified computer architecture which is made of a number of identical CPU cores, each of which has a cache of a given size. The specification includes the following:
  - The number of CPU cores.
  - The size of the cache in bytes.
  - The size of the memory in bytes.

### JSON IO Specification

This section defines the format of an IO Specification when captured as JSON.

```
{
    "inputs": [
        {
            "connector": "<connector_name>",
            "data_format": "<data source format>",
            "driver": "<driver name>",
            "driver_parameters": [<list of parameter objects>]
        }
    ],
    "outputs": [
        {
            "metric": "<metric name>",
            "dimensions": [<list of category names>],
            "celltype": "<measure name>",
            "grain": ["<interval measure name>"]
        },
        {
            "metric": "yearly_quarterly_sales",
            "dimensions": ["Region", "Product"],
            "celltype": "Revenue",
            "grain": ["Yearly", "Quarterly"]
        }
    ],
    "architecture":{
        "cpu_cores": <number of cores>,
        "cache_size": <cache size in bytes>,
        "main_memory_size": <memory size in bytes>
    }
}
```

#### Inputs

- The connector_name must be the same as the name of a connector in the plan/CPL specification.
- The data format name is the form of the data, such as 'parquet' or 'excel'.
- The driver is the mechanism used to access the data, such as 'file', 'chronocone database', or 'demand service'.
- driver_parameters is a list of zero or more JSON objects that are passed to the driver. The number and type of parameters depends on the driver.

#### Outputs

- The metric must reference a metric name defined in the plan specification.
- The dimensions are a comma separated list of strings that must be the same as concept definitions in the plan specification.
- The celltype is a measure name that must be defined in the plan specification.
- The grain is an ordered array of one or more measure names that must be defined in the plan specification, each with a type that is `Interval` or derived from `Interval`. Single-grain outputs use a single-element array. Multiple grain measures appear in the array in the same order as declared in the connector. IO specifications do not make `Duration` into a type; they bind a particular run to the relevant measure and unit choices already defined through normal concept, unit, and measure declarations. For connector input data, connector metadata and IO-spec time-dimension selection determine which `Duration`-derived unit applies.
