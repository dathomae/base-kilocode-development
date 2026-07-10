> **Note to AI Agents:** This document is capturing a set of concepts that has been changed from the initial version captured in the current v0 requirements and v0 architecture. Until this note is removed when there is a conflict between this document and other documents or stories the statements in this document will override any conflicting ideas or statements from other documents.

-- Begin Document Content ---

# Introduction

## What is ChronoCone?

ChronoCone is a generic planning system designed to solve a wide variety of planning problems. It provides a flexible framework that supports automated planning, human-in-the-loop planning, and collaborative planning across different organizational levels (see Note 1).

The name "ChronoCone" represents widening "cones" of possible future states. Plans can expand either through refinement, where general plans develop into specific ones (e.g., country-level plans break into regional plans), or through temporal extension, where short-term plans extend into longer horizons. Longer-term plans typically have less specificity, more uncertainty, or both.

The system serves diverse planning domains including logistics, manufacturing, and healthcare. These domains share common planning challenges such as coordinating across organizational boundaries, managing uncertainty, and balancing resource constraints against operational goals.

## Key Terms

**forecast**  
Predictions of future data such as predictions of buyer purchases in future weeks or the need for warehouse labor in future weeks or months.

**actuals**  
Past observations of some metric, such as the number of units that were actually purchased.

**metric**  
A particular type of thing that is of interest to a business, such as actuals and forecast for customer demand, orders that are placed by vendors, trucking capacity that must be contracted for, warehouse labor for which people must be hired and scheduled. Most metrics will have both actuals (for past data) and forecasts associated with them.

**scheduling vs planning**  
Plans, in the ChronoCone sense, are past observation or forecasts of "how much and when" which may be informed by constraints on people, equipment or other resources but are not dealing with the individual resources themselves. Schedules deal with the allocation and tracking of individual resources. ChronoCone is a planning system, not a scheduling system.

For example, ChronoCone might be used to determine how many labor hours are required in a warehouse on a particular day or how many trucks are needed to ship items from a vendor to a warehouse in a particular week, but would not be concerned with determining when particular people worked in a warehouse or when particular trucks arrived at a vendor or a warehouse—those things are the domain of a scheduling system. While it would be possible for ChronoCone to integrate with a scheduling system (e.g. to gather actuals for hours worked or supply plan outputs that drive a scheduling system), ChronoCone would only send or receive aggregate (as opposed to per-resource) data from such a system.

## Open Questions
These are questions sparked by writing this document. I'm capturing them here for the time being,
although all but the first one don't really belong in the concepts doc.

1. This document is supposed to be using the terminology exposed to customers and users, but I'm not
sure that 'namespace family' and 'namespace' are the right terms to be using for them. Are they?
If not, what are the right terms?

2. How should shifters be defined? Should they be functions or declarations, given that they primarily perform aggregation (to collapse dimensions) or replication (to expand dimensions)? Does the current description adequately support these use cases?

3. How are filters invoked? Is it done by the calculation engine based on metadata associated with transformers?

# Users And Customers
ChronoCone is targeted at a particular set of users (i.e., the people who directly work with the product) and customers (i.e., those people who pay for the product and are ultimately responsible for the success of their business). While some individuals may operate in both roles from time to time, these groups are conceptually distinct.

The customer consists of people in various executive suite roles such as CFO, CIO, CTO, and COO. Their desire to purchase ChronoCone is ultimately driven by their need to run their business efficiently, reduce uncertainty in forcasting, manage and/or reduce risk, and communicate plans to people in their organization and their partners. One of their key concerns is that the proprietary data captured by ChronoCone does not escape their control or become visible to those who are not authorized to see it, either accidentally or due to malicious activity. The customers will generally make policy decisions regarding the uses of ChronoCone and have a special interest in reporting and KPI management. They may or may not be interested in the details of individual plans, but will be interested in the ability to quickly answer questions or concerns that are raised for questions about forecast and data accuracy.

The users consist of people who will directly use the product to construct plans, integrate ChronoCone with external services or other data sources or use the forecast or actual outputs as part of operating the business.  There are several types of users, each with their own typical functions and concerns:
1) System Integrators (or SIs). These are typically consultants that have deep 
expertise in the general problem domain that the customer business is in (e.g. logistics, budget planning) but will be less knowledgeable about the details of the individual businesses of the customers they work with and for. They will typically engage with a customer by gathering requirements, asking questions and so on to understand the problem(s) the customer has and translate those into ChronoCone terms. SIs will typically have deep expertise in development/programming and will be the prime initial implementors of ChronoCone plans and integration with the customer infrasture, such as implementing connectors to pull data from customer services or other data sources. The ChronoCone organization will primarily work with SIs, rather than the ultimate customers.
SIs will tend to work with their customer companies on an episodic basis, with a strong initial involvement to set up a system and develop plans and integrations which are then handed off to customer employees. SIs may be brought back in to help with major upgrades or to create new types of plans but otherwise will generally be called up to help with problems that customer employees can't solve. SIs are particularly interested in features that make it easy to set up new plans, verify that the logic and structure of plans is correct, track down bugs in planning logic and be able to easily demonstrate to customers that plan outputs are correct for their business.
2) Analysts are customer employees who have good expertise in development/programming, although not necessarily as deep as those of SIs. Analysts will be called upon to improve and adjust existing plans and sometimes create new ones, although they may be less able to create new integrations from scratch. Analysts will have deep understanding of forecasting and the details of their business. The primary desire of analysts is having the ability to quickly understand and modify plan structure and verify it's correctness.
3) Planners are the primary users of plan outputs to run the day to day operations of their business and have the strongest direct experience with the accuracy and appropriateness of plan outputs and the deepest instinctive "feel" for the correctness of outputs. Planners are the users who are most likely to ask questions about plan outputs that may be incorrect and want to distinguish quickly between legitimate reasons for unexpected outputs (e.g. a holiday in a foreign country that the planner didn't know about) and quality or correctness issues (e.g., a system that ChronoCone gets data from is having a transient data quality issue).
4) "Power planners" are users who share some of the characteristics of analysts and planners, and have at least some basic development/programming expertise. They share the concerns of both types of users, but may require more guidance then analysts in understanding the structure of plans and the reason(s) for structuring plans in a certain way.


# Data Management And Visibility
The core philosophy is that "Logic travels; data stays". That is, the emphasis in ChronoCone is tightly controlling data visibility so that it can only be seen and used by the people and systems that the customer deems appropriate. Different customers will make different decisions about their physical infrastructure. Some may insist on only holding data on their own internal networks, computers and storage devices, other may be comfortable with deploy them in a secure off-site data center or cloud provider. Regardless of the exact deployment mechanism, the following constraints will be preserved by the design, implementation and supported deployment options of/for the ChronoCone system:
1) Data will be encrypted in transit and at rest.
2) Keys needed to support the encryption will be created, communicated and stored securely, so they cannot be easily or quietly obtained by unauthorized parties.
3) All users will be adequately authenticated using open standards for authentication, including multi-factor authentication if the customer so desires.
4) All users will be authorized to see and/or modify only the data or definitions that are required for their function. For example, user may be able to see or modify some plans but not others, may be able to see or modify some sets of metadata but not others and so on.
5) ChronoCone distinguishes between "core" data, such as the actuals and forecast numbers associated with metrics and "metadata" which is concerned with the structure of plans and plan deployments. Core data is strongly protected, metadata may be more freely distributed, with appropriate authorization.
6) Features that might potentially cause forecast or actuals core data to exit the control of ChronoCone will not be supported. For example, ChronoCone will not provide the ability to download data to .csv files in the web interface, provide core data to a development environment or send core data to an LLM.
7) Metadata may be more freely distributed. Thus, sending metadata to and allowing it to be manipulated in a development environment would be acceptable, provided dummy data was used for development, debugging and testing. Using AI agents to help users understand the structure of plans or giving advice about how to track down problems would be allowable as long as only metadata (but not core data) is sent to the third party LLMs that power agentic capabilities.

# ChronoCone Key Concepts
ChronoCone is organized around a set of key concepts concerning plan deployment, visibility and structure.
1) A "namespace family" represents a major organizational function of a business such as "Retail Planning" (e.g. forecast customer demand and orders from vendors), "Labor Planning" (e.g., determine how many person-hours of labor is needed in each warehouse in a fulfillment network based on the output of a country or region level retail plan) or "Transportation Planning" (e.g., determing how much of what type of air or ground transport is needed to get items from vendors to warehouses or from warehouses to buyers). The name of a namespace family must be globally unique.
2) A "namespace" represents a particular visibility for plans, such as Production (which would be used to actually run a business), Development (which might be used to surface new or modified metrics for review), or "Dougs Work" which might be used to capture prototype plans for discussion before commiting an organization to development them completely. Plans may be "promoted" by pushing or copying them from one namespace within a namespace family to another namespace in the same namespace family. The name of a namespace must be unique within a namespace family.
3) A plan represents an interconnected set of metrics, having some inputs which are taken from some external system, human input, file upload or other data source, some outputs and optionally come intermediate metrics. A plan is composed of a set of metrics which are connected by a Directed Acyclic Graph (DAG). Multiple plans may exist within a namespace, but a plan name must be unique within a namespace. From the user point of view plans are always updated as a unit (i.e., plans that are internally inconsistent due to partial evaluation are not allowed).
4) A metric is a timeseries representation of some data broken out across one or more dimensions. The time series has a "grain" such as daily or weekly. The dimension represents type of element that the individual data elements are indexed by. For example, a metric for a demand forecast might be broken out by Product and Price (typically the lower the price the higher the demand),a warehouse labor plan might have a labor needed metric broken out by process area and unit size. Every metric has a forecast portion for future dates and an actuals portion for past data. Forecasts for past time periods are still available and may be used to determine things like forecast accuracy.

## Elements Of A Metric
There are four specifications associated with metrics, in addition to the data and metadata mentioned above, that specify how data flows into or out of a plan, how data elements are derived from other data elements and how metrics are connected. In all cases the specification is captured in **CPL (Chronocone Planning Language)**. These CPL blocks are processed/analyzed by the synthesis phase to produce an efficient executable form for the Go-native calculation engine. These specifications come in four types, transformers, shifters, filters and lookups.

### Connectors
Connectors are software elements that pull or push data somewhere. It's 
anticipated that there will be two general types of connectors:
1) Connectors that are the integration points with outside systems, usually customer systems. 
2) Connectors that work with internal ChronoCone stores such as S3 or local file stores.

### Transformers
Transformers are responsible for three types of functions:
1) Using connectors of some type to pull or receive data from some outside system or some human input (e.g., file upload).
2) Using connectors to push data to some outside system (generally for leaf metrics in a plan DAG).
3) Specifying the logic that is used to calcuate an element of a metric using input from some other metric(s).

For metrics that are derived from other metrics, the ChronoCone calculation engine is responsible for managing the calculation so that the right input data elements are selected to align with the data output elements both in time and with matching dimensions, possibly requiring the set of shifters. The CPL code for the transformer serves as the logical contract for this derivation.

The intent is that the environment in which transformers are run will be
such that there is no need for iterative calculations or searches. If 
some scenario is found that would require those kinds of capabilities the
preference is that the shifter/lookup/filter approach would be extended to
eliminate the need rather than that iterative operations be supported.

Transformer function names are the same as metric names as are it's input metrics. The transformer function and input parameter names are used to construct the plan DAG. Metric (and thus transformer) names must be unique within a namespace, but may be used in zero or more plans in that namespace (obviously with the same definitions and inputs).

### Shifters
To make it possible for the respective data cells of input and output time
series to be aligned for processing by transformers it is necessary to make sure the dimensionality of the inputs matches with that of the outputs, either by collapsing or expanding dimensions. Shifters are functions that have the same name as a dimension and inputs that have the same name(s) as some dimensions. The shifter will return a result that represents the equivalent value(s) of the output dimension(s).

The names of shifters must be unique within the namespace, but that's mostly to facilitate development since the calculation framework will decide which shifters to use when.

### Filters
Filters are functions that split metrics along some dimension, such as dividing a national plan into regions or a retail plan among warehouses. The filter will take as input a cell in a metric and output a map that specifies how the input should be divided along the values of one or more dimensions.

Filters have names unique within a namespace.

### Lookups
A lookup provides non-metric data to a transformer, shifter or filter oriented along some dimension. For example, a transformer that will calcuating an inventory allocation for a metric that has dimensions of Product and Warehouse might need to take into account limitations due to the physical capacity of the warehouse. Those limitations aren't forecasts/metrics since they're unchanging for long periods of time. In order to make that functionality available without having users use ad-hoc mechanisms like hard coded dictionaries, lookups will provide that ability.

A lookup is used by a transformer, shifter or filter by including it's name in the input parameters for a function.

Lookup names are unique within a namespace.

# AI Enablement
One of the key features of ChronoCone is the use of AI agents to facilitate the creation and update of metrics and to help planners to answer questions about plans.  A key constraint however, is that ChronoCone will never send core data to LLMs or agents (which would in turn send them to LLMs). Sending metadata to agents/LLMs is acceptable. 

## SI/Analyst/Power Planner Development
One of the two key uses for AI is to help users who are creating or updating metrics to help with the **CPL specifications** for transformers, shifters, filters and lookups and to help with the creation of input data that is used to ensure that those functions are working correctly. Since no core data will ever be surfaced in a development environment there are no special restrictions that need to be placed on the use of LLMs for this set of users.

## AI For Planners
Since planners deal with core data, it will not be possible to create an AI experience that depends on the core data. Thus, it will not be possible for an agent to answer the question 'what is the forecast accuracy for the last 6 months?' or 'what is the forecasted inventory for some product for week 37?'. However it is possible to build AI that can answer the question 'how do I find out what the forecast accuracy has been for the last 6 months?', constructing a link with appropriate parameters and provide it to the user. Likewise, a planner might ask how they can determine the source of an unexpected change in a forecast or why actuals seem to be out of line and get advice and coaching about the procedure to do so.

It will be necessary to put limits on the actions that agents can take. For example, while merely constructing the link using metadata to find forecast accuracy numbers can't leak core data to an agent, the ability to render/scrape web page could potentially allow it to see core data.





