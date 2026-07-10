# Story032: Metrics and Calculation Requirements Exploration

## Goal

Analyze and define the major requirements for how metrics are defined, calculated, and related to one another within ChronoCone. This story captures the specification needed for the metrics and calculation system, building on concepts defined in chronocone-concepts.md.

At the discretion of the person working on this story, architecture or design decisions may be made and recorded in the architecture or design documents. Additionally, one or more implementation stories may be defined to capture work that extends beyond the requirements definition itself. The exact result for this story will depend on the complexity of the requirements discovered and the amount of discussion needed to reach clarity.

## References

- [ChronoCone Concepts](../specifications/chronocone-concepts.md) - Key concepts including:
  - Metric definition (timeseries representation with dimensions and grain)
  - Transformers, shifters, filters, and lookups
  - Plan DAG structure
  - Namespace and namespace family organization
- [Introduction](../specifications/requirements/introduction.md) - System overview and deployment contexts
- [Metrics and Calculation Requirements](../specifications/requirements/metrics-and-calculation.md) - Target document to populate

## Dependencies

- [Story027_relational-store-requirements.md](Story027_relational-store-requirements.md) - Not Started (metric metadata requires relational storage)
- [Story029_object-store-requirements.md](Story029_object-store-requirements.md) - Not Started (metric timeseries data requires object storage)

## Dependent Stories

- [Story031_metadata-requirements.md](Story031_metadata-requirements.md) - Not Started (metrics have associated metadata)
- [Story039_external-api-simulator-and-sample-data-requirements.md](Story039_external-api-simulator-and-sample-data-requirements.md) - Not Started (sample data requires understanding of metric structures)
- [Story033_connectors-requirements.md](Story033_connectors-requirements.md) - Not Started (connectors feed transformers)
- Implementation stories will be created after requirements are defined.

## Context

From [chronocone-concepts.md](../specifications/chronocone-concepts.md), the core concepts to address include:

- **Metric**: A timeseries representation of data broken across one or more dimensions with a grain (daily, weekly, etc.)
- **Forecast vs Actuals**: Metrics have both forecast (future) and actuals (past) portions
- **Transformers**: Python functions that calculate metrics from other metrics using connectors
- **Shifters**: Functions for dimension alignment (collapsing or expanding dimensions)
- **Filters**: Functions that split metrics along dimensions
- **Lookups**: Functions providing non-metric data to transformers/shifters/filters
- **Plan DAG**: Metrics connected as a Directed Acyclic Graph

## Purpose of This Section

The metrics-and-calculation.md requirements document must specify:

1. How metrics are defined, including their structure (dimensions, grain, data types)
2. How metric calculations are specified and executed
3. How the plan DAG is constructed and validated
4. How data flows through transformers, shifters, filters, and lookups
5. How the calculation engine manages dependencies and execution order
6. How metric versioning and lineage interact with calculations

## Major Architectural Elements to Elaborate

### 1. Metric Definition Structure
- Dimension system and grain specifications
- Data type constraints for metric values
- Metadata associated with metrics (descriptions, ownership, etc.)
- Relationship between metrics and plans

### 2. Calculation Engine
- DAG construction from transformer function signatures
- Dependency resolution and topological sorting
- Execution strategies (batch vs incremental)
- Performance considerations for large-scale calculations

### 3. Transformer Specifications
- Python function constraints (disallowed features)
- Input/output alignment (time and dimension matching)
- Error handling and validation
- Testing approaches for transformers

### 4. Shifter System
- Dimension collapsing (aggregation) mechanisms
- Dimension expansion (replication) mechanisms
- Automatic shifter selection by the calculation framework
- Shifter naming conventions and uniqueness

### 5. Filter and Lookup Integration
- Filter function signatures and return types
- Lookup table management and caching
- Integration with the transformer execution context

## Exploration Tasks

1. **Define metric structure requirements** - Not Started
   - Specify dimension system requirements
   - Define grain options and constraints
   - Determine metadata requirements for metrics

2. **Specify calculation engine requirements** - Not Started
   - Define DAG construction rules
   - Specify dependency resolution requirements
   - Address cycle detection and prevention
   - Document execution order guarantees

3. **Elaborate transformer requirements** - Not Started
   - Define Python subset allowed in transformers
   - Specify input/output validation requirements
   - Document testing and verification approaches

4. **Define shifter system requirements** - Not Started
   - Specify dimension mapping requirements
   - Define aggregation function requirements
   - Document automatic shifter selection logic

5. **Specify filter and lookup requirements** - Not Started
   - Define filter function interface
   - Specify lookup table structure and management
   - Document caching and performance requirements

6. **Address cross-cutting concerns** - Not Started
   - Error handling and propagation
   - Performance monitoring and optimization
   - Integration with data lineage tracking

## Next Steps (To Be Determined)

After completing this exploration, determine:
- Can the requirements be written as a single implementation story?
- Or are multiple implementation stories needed for different subsystems?
- Which areas require deeper architectural exploration before implementation?

## Notes

- The "Logic travels; data stays" principle from chronocone-concepts.md should guide design decisions
- Transformer Python functions are specifications, not necessarily executed as written
- The system should avoid iterative calculations or searches in transformer execution
