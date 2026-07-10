Directives for agentic or AI writers and editors:

1. You MUST leave this list of directives in place as the first item in the document and unaltered as changes are made to this document.
2. When editing this document you MUST use either the Technical Writer or Technical Editor roles. If you are not in one of these modes you MUST NOT edit this document and inform the user of the requirement.
3. This document contains annotations that are labelled as "Notes:" or "Questions:" and you MUST NOT edit these annotations. You should move them to keep them with the content they relate to.

--- Begin Document Content ---

# Part 1: Introduction

## What is ChronoCone?

ChronoCone is a generic planning system designed to solve a wide variety of planning problems. It provides a flexible framework that supports automated planning, human-in-the-loop planning, and collaborative planning across different organizational levels (see Note 1).

The name "ChronoCone" represents widening "cones" of possible future states. Plans can expand either through refinement, where general plans develop into specific ones (e.g., country-level plans break into regional plans), or through temporal extension, where short-term plans extend into longer horizons. Longer-term plans typically have less specificity, more uncertainty, or both.

The system serves diverse planning domains including logistics, manufacturing, and healthcare. These domains share common planning challenges such as coordinating across organizational boundaries, managing uncertainty, and balancing resource constraints against operational goals.

# Part 2: Core Concepts

## The Conceptual Structure of ChronoCone Plans

ChronoCone plans exist at two levels from an end user perspective. High-level plans typically have organizational associations—for example, the Retail plan belongs to one department while different parts of the Logistical plan belong to different departments within the Logistics organization. This two-level structure reflects how real organizations divide planning responsibilities across teams and functions.

Information in ChronoCone flows acyclically in one direction. For instance, the Retail plan produces outputs that feed into the Logistics plans, which then determine warehouse operations and transportation schedules. This unidirectional flow simplifies planning by establishing clear hierarchies and dependencies between different organizational units. When exceptional circumstances require adjustments to higher-level plans based on lower-level outcomes (e.g., a snowstorm limiting warehouse capacity), these adjustments are handled by external processes outside the planning system rather than through feedback loops within the plans themselves.

## Document Types and Documents

ChronoCone distinguishes between document types and documents. A document type defines the structure and logic of a plan including the rules, calculations, and dependencies that govern how inputs transform into outputs. Planning organizations create and maintain these document types, which specify what data the plan requires, what operations to perform on that data, and what results to produce. For example, a warehouse labor document type might specify how to calculate the number of workers needed based on expected order volume and processing rates.

A document represents a specific document type at a particular point in time, derived from a document type plus input data. When a document type runs with concrete input data it produces a document with specific numerical or categorical outputs. For instance, a warehouse labor document might specify that 15 people should work the picking station from 2 PM to 6 PM on Tuesday, March 15th. These outputs guide actual operational decisions and actions.

ChronoCone versions and timestamps both document types and documents, recording metadata such as who made changes and what data sources contributed to the plan (see Note 2). Once created, a document type or document is immutable. Changes require creating a new version rather than modifying the existing one. This versioning approach maintains a complete audit trail of planning decisions and enables analysis of how plans evolved over time.

### Versioning and Immutability

All artifacts in ChronoCone are versioned and immutable. This principle applies to:

- **Document type definitions**: The structure and logic that defines how a plan operates
- **Documents**: The concrete outputs produced when a document type executes with specific input data
- **Transformers**: Functions that produce metric values from inputs
- **Shifters**: Functions that align metrics with different dimensional structures
- **Filters**: Functions that split metrics into constituent parts

When any artifact is modified—whether recalculating a plan, updating a transformer, or adjusting a filter—the system creates a new version rather than updating the existing one. This immutability is fundamental to the system's design and enables several critical capabilities:

1. **Historical reproducibility**: Any historical calculation can be reproduced exactly by retrieving the specific versions of all artifacts that were active at that time
2. **Audit trails**: Complete records of what changed, when, and why support compliance and debugging
3. **Safe evolution**: New versions can be tested and validated before being used in production, with the ability to roll back if problems emerge
4. **Lineage tracking**: The system can trace how any value was derived, which is essential for understanding and explaining planning results

### Lineage Tracking

ChronoCone maintains lineage information at two distinct levels, enabling users to understand both how specific values were computed and how the system's components relate to each other.

**Document-Level Lineage**: For any value in a time series—whether an actual or a forecast—the system can trace the complete flow of inputs and calculations that produced it. This includes:

- The source data that contributed to the value
- The sequence of transformers, shifters, and filters applied
- The specific versions of each function used
- The document type and document that contained the calculation

This lineage enables users to answer questions like "Why does this forecast show 500 units?" by walking backward through the calculation chain to see exactly which inputs and operations produced that result.

**Definition-Level Lineage**: The system tracks relationships between the definitions of metrics, transformers, shifters, and filters as versions change over time. This includes:

- Which transformer versions were associated with which metric versions
- How input-output relationships between metrics evolved
- When shifters or filters were added, modified, or deprecated
- The dependency graph connecting all calculation components

This lineage enables users to understand how the planning system itself has evolved, answering questions like "What changed in the labor calculation between March and April?" by showing the differences in the relevant function versions and their relationships.

### Diff Operations

An important operation supported by lineage tracking is the ability to compare two versions of an artifact. The system provides diff capabilities that are primarily surfaced as:

- **Compare two document versions**: Shows what changed between two runs of the same document type, including differences in input data, intermediate calculations, and final outputs. This helps users understand why results differ from one planning cycle to the next.

- **Compare two function versions**: Shows the differences between two versions of a transformer, shifter, or filter, including changes to the code logic, input/output specifications, and associated metadata. This helps users understand how calculation logic has evolved.

These diff operations are essential for debugging, auditing, and understanding the impact of changes to the planning system.

## DAGs and DCFGs

Within a given plan (e.g., a labor plan for a warehouse) a document type or document is conceptually a Directed Acyclic Graph (DAG) that applies some operations to some outputs, which in turn are fed to other operations and so on. Since bare DAGs are difficult for people to comprehend, the intent is to allow plan creators to specify various plan elements in more convenient forms, such as annotate Python code fragments or Domain Specific Languages (see Note 3).

## Working With External Systems

Plans don't exist in isolation, and ChronoCone isn't a "do everything" system. It will need to be able to connect to external systems to both pull in data (e.g., demand forecasts for products, inventory levels in warehouses, productivity targets or historical data for warehouse associates) and export data (e.g., send labor capacity forecasts to a warehouse scheduling system). ChronoCone uses connectors to interface with external systems. Within the boundaries of the ChronoCone system, only ChronoCone-defined types exist, with the transformation to or from those types happening in the connectors.

### Built-in Types

There are a number of built-in types (i.e., those defined natively by ChronoCone as opposed to the entities defined by users):

1. Strings
2. Numbers. Assume that ChronoCone doesn't distinguish between integer and floating point numbers.
3. Currency. Currency will be fixed point and carry a currency type.
4. Dates and times. Assume for the moment that is a common DateTime type.
5. Enumerations to represent categorical values.
6. A simple time series type. The simple time series has a time grain (e.g., hour, day, week), a length and a Number or Enumeration value for each time period.
7. A parameterized time series, which can be thought of as a hash table where the key is some non-time series type and the values are a simple or parameterize time series. For example, a demand forecasting system may be dependent upon price (the lower the price, the higher the demand). Therefore, it could produce a time series demand forecast for each product at different price points. In this case the keys would be a Currency value and the time series data points would be the number of units forecasted to be sold at that price.
8. A Metric is a simple or parameterized time series with a unit of measure and a unique name. I'm going to ignore the idea of namespaces in this document, but that's a straight forward extension of the naming concept.
9. User defined types. For example, a Product that's associated with a demand forecast will have something that distinguishes it from other products or different models of the same product. Likewise, a product may be available from multiple vendors, who need to be distinguished from each other. At the least each user defined type and each instance of a user defined type will need a unique id.
10. Connector defined types. Connectors will need to have knowledge of some types and if those change the connector code may need to change as well. Connector defined types might be a subset of user defined types (or connector defined types and user defined types might be a subset of something else).

# Part 3: Illustrative Example

## Supply Chain Planning Scenario

This example illustrates how ChronoCone handles multi-level planning in a retail supply chain context. The company operates only in the United States and includes:

1. A set of vendors that supply products
2. Multiple warehouses for inventory storage and distribution
3. Delivery vehicles for transporting products from vendors to warehouses and to customers
4. A customer base receiving product deliveries

The overall planning process involves coordination between three organizational groups: Retail, Customer Service, and Logistics. Each group creates and maintains document types that define their planning logic, and these document types execute regularly to produce documents that guide operational decisions. The following sections examine each group's planning responsibilities and show how document types capture their specific planning needs.

### Customer Service Planning

The customer service group handles orders and payments, manages cancellations and returns, and provides customer support. While most functions operate automatically, exceptional situations require human intervention that may alter existing plans. When taking an order, the group commits to specific delivery timeframes, creating constraints for the logistics organization. Some actions face constraints from logistics (product shortages, fulfillment location limitations) or retail (price changes, sale endings), and customer characteristics (payment processing delays) may interact with these constraints. The primary goal is maximizing customer satisfaction while minimizing costs.

[PLACEHOLDER: Example showing customer service document type structure, including:

- Input data sources: retail demand forecasts, logistics capacity commitments, current order volumes, customer service level targets
- Planning logic: order acceptance rules, delivery promise calculations, return/cancellation policies
- Output data: order commitments, delivery promises, expected service levels
- Concrete example showing how the document type balances customer commitments against operational constraints]

When an order is taken, a commitment is made to deliver the product to the customer within a certain time. These promises put constraints on the actions of the logistics group. Some actions may be constrained by logistical constraints, such as a product running out, a product being available but not in a place that makes it possible to fulfill an original customer promise, or no longer being available at the original price (e.g., because a sale has ended). In some cases customer characteristics (e.g., difficulties with payment that take time to resolve) may interact with these constraints.

### Logistics Planning

The logistics group allocates inventory to warehouses in a manner that respects customer commitments while managing labor and transportation capacity constraints across multiple regions. Physical transportation constraints come from vehicle capabilities, distances, traffic patterns, and route characteristics. For this example, assume the company uses only trucks of various sizes. Trucks must depart by a "critical pull time" to reach destinations within committed timeframes.

Warehouse operations follow main process paths:

1. Receive: unloading trucks, removing packaging, storing products in inventory
2. Pick: retrieving items from shelves for customer orders
3. Sort: aggregating items for multi-item orders in designated locations
4. Pack: boxing items, sealing, labeling, and conveying to loading areas
5. Load: placing packed orders onto outbound trucks

Processing times vary by item type (small items process faster than large items). Other process paths like inventory cleaning exist but have looser time constraints and can partially compensate for forecast errors. The main processing rate constraint is the number of workers assigned to each process path at specific times. Too many workers create idle time and increase costs; too few workers require overtime at 1.5 times regular pay rates, also increasing costs. The primary goal is cost minimization while satisfying customer promises and meeting processing rate requirements.

[PLACEHOLDER: Example showing logistics document type structure, including:

- Input data sources: customer service order commitments, current inventory positions, warehouse capacities, transportation constraints, labor availability
- Planning logic: inventory distribution optimization, warehouse capacity scheduling, transportation routing, labor scheduling algorithms
- Output data: warehouse inventory allocations, labor schedules by process path, transportation routes and schedules
- Concrete example showing how the document type coordinates multiple logistics activities to meet customer commitments cost-effectively]

Organizations may maintain multiple related plans. For example, a high-level warehouse distribution plan might feed into separate labor and transportation plans for individual warehouses, with regional transportation plans mediating between the distribution plan and warehouse-specific plans.

# Part 4: Advanced Topics

## AI Enablement And User Experience

While agentic AI is still in it's infancy, it seems likely that it will be a key part of the user experience for ChronoCone, but at the same time agentic AI and LLMs have limitations that make it unlikely that the planning problem can simply be implemented as a system of agents. First, consistent, multi-level planning requires a high degree of consistency in the results of the planning process which is not a property of current agentic architectures which require frequent checks and revisions by humans to be effect with current technology. Business need to be certain that the planning process is consistent and reliable. Similar concerns apply if we choose to make plan optimization a part of the product in a way that strictly respects business rules.

# Implementation Considerations

This section contains notes that provide additional context, details, and considerations for various aspects of the ChronoCone system. These notes preserve important information about design decisions, scope considerations, and technical alternatives.

**Note 1: On scope and complexity**

The capabilities and the examples intentionally includes some features/discussions that are suggested by some of the more complex elements of real supply chains. However, it is possible that we don't want to handle these problems in their full complexity (e.g., because ChronoCone isn't specialized for supply chains), that some of these elements should be handled by non-ChronoCone systems or that some features might be desirable to include someday, but not in early releases.

**Note 2: Document storage and versioning optimization**

It's likely that in practice most changes in document types or documents will be small. It may be useful to use a current document plus delta scheme to represent older documents to reduce storage requirements, but that is a discussion that is beyond the scope of this document. We wouldn't want to do it the other way around (i.e., a base document with deltas to derive the more recent documents) since all of the deltas would need to applied to the base document to get the more recent documents, and the most recent document would be the one that would be the one that is used most. The result would be a system that gets slower and slower as time goes on from the end user perspective.

**Note 3: DCFG as an alternative representation**

An attractive alternative for the underlying representation of low level plans would be a Data and Control Flow Graph (DCFG). DCFGs are used commonly in compilers and represent the structure of a program/computation on two levels. The first level is a Control Flow Graph (CFG), which represents the decision making part of a program (e.g., loops, conditionals). Within the body of control flow elements (e.g., the body of a loop, the body of an if statement) calculations are represented as a Data Flow Graph (DFG). Data flow graphs are acyclic, with inputs flowing through a series of calculations to produce one or more outputs (in compiler theory these are known as "basic blocks"). The reasons to choose a DCFG formulation include the following:

1. Since we're already talking about using bits of Python code to capture some plan elements, we could use the well known conversion algorithms used in every compiler to do the conversion.
2. There are well known DCFG optimization techniques that we could apply to plans to reduce their computation requirements.
3. DCFGs work in both directions, so we can synthesize efficient, executable version of a plan to do the actual evaluation. One of the options would be to synthesize code that runs on GPUs. As GPUs become a more common part of execution environments that use LLMs, this option may become attractive as a means of minimizing LLM inference cost (i.e., by running local rather than cloud models).
