Directives for agentic or AI writers and editors:

1. You MUST leave this list of directives in place as the first item in the document and unaltered as changes are made to this document.
2. When editing this document you MUST use either the Technical Writer or Technical Editor roles. If you are not in one of these modes you MUST NOT edit this document and inform the user of the requirement.
3. This document contains annotations that are labelled as "Notes:" or "Questions:" and you MUST NOT edit these annotations. You should move them to keep them with the content they relate to.

--- Begin Document Content ---

# MCP Server - Planning

## Purpose

The purpose of the planning MCP server is to help human planners leverage AI to construct and adjust plans. The ChronoCone engine provides a deterministic, rule-based calculation layer, and the MCP server acts as the interface where four primary personas (Planners, Power-planners, Analysts, and System Integrators) use agent-chat interfaces to govern these models safely.
The Planning MCP Server helps planners create or update models, using a PDD (Prompt Driven Development) style of input that will in turn be translated into a programmatic form. This server bridges the gap between the domain expertise of planners and the technical requirements of the calculation engine, enabling planners to express their intent in natural language while producing rigorous, testable code.

A PDD script is a pseudo-executable script used by LLMs and agentic tools to perform structured tasks. Unlike traditional programming languages, PDD scripts are written in natural language with enough structure to guide AI assistants through complex workflows. The Planning MCP Server interprets these scripts and translates them into the Python code that implements transformers, shifters, and filters.

In addition to code generation, the server aids in the generation and updating of test data, executes tests, and provides feedback to the planner to help in fixing errors. This integrated development workflow supports planners through the entire lifecycle of creating and refining calculation logic, from initial concept through tested, production-ready code.

## Target Users

Planners creating or adjusting planning models represent the primary users of this server. These users are typically domain experts who understand the business logic behind the calculations but may not have extensive programming experience. They know what they want to calculate and why, but they need assistance translating that knowledge into correct, efficient code.

The server supports these users by providing an AI-assisted development environment that understands both the ChronoCone domain model and the principles of good software engineering. Planners can describe their intent in natural language, and the server helps them produce code that is correct, well-structured, and thoroughly tested.

## Responsibilities

1. Act as the primary interface for governed autonomy, whereby humans initiate agent interactions, and agents use the MCP to read/write the ontology which governs the planning models.
2. Generate programmatic code from PDD (Prompt Driven Development) style input.
3. Generate test data for planning models.
4. Execute tests and provide feedback to the planner and help in fixing errors.
The Planning MCP Server has three primary responsibilities that together provide comprehensive support for the model development workflow.

**Generating programmatic code from PDD style input** is the core responsibility of the server. Planners describe what they want to calculate using natural language, and the server translates that description into Python code that implements the appropriate transformer, shifter, or filter. This translation requires the server to understand both the domain concepts (metrics, dimensions, time series) and the technical requirements (correct function signatures, proper error handling, efficient algorithms).

**Generating test data for planning models** ensures that the generated code can be validated before it is used in production. The server helps planners create realistic test data that exercises the important cases in their logic. This test data is essential for building confidence that the code works correctly and for catching regressions when the code is modified.

**Executing tests and providing feedback** completes the development workflow. The server runs the generated code against the test data and reports the results to the planner. When tests fail, the server helps diagnose the problem and suggests fixes. This iterative feedback loop is essential for producing high-quality code, especially when the original requirements are ambiguous or incomplete.

## Interfaces

The Planning MCP Server implements three MCP interfaces: Tools, Resources, and Prompts. These interfaces follow the Model Context Protocol specification and are designed to support the code generation, testing, and publication workflows described above.

### Tools Interface

The Tools interface provides callable functions that AI assistants can invoke to create and manage planning models. These tools support the complete lifecycle of transformers, shifters, and filters from initial creation through publication.

#### metric_names

**Purpose**: Lists all metrics in a namespace, including both published and unpublished.

**Parameters**:

- `namespace_family` (string, required): The namespace family to query
- `namespace` (string, required): The namespace to query

**Behavior**: Returns the names of both published and unpublished metrics in the specified namespace family/namespace. Each metric name includes a status tag indicating whether it is published or unpublished. This visibility into unpublished metrics helps planners understand what changes are pending and track their work in progress.

**Returns**: A list of metric names with their publication status (published/unpublished).

#### publish

**Purpose**: Transitions unpublished metrics to the published state.

**Parameters**:

- `namespace_family` (string, required): The namespace family containing the metrics to publish
- `namespace` (string, required): The namespace containing the metrics to publish

**Behavior**: Publishes any unpublished metrics in the specified namespace family/namespace. This action:

- Transitions metrics from the unpublished state to the published state
- Makes the metrics active for production calculations
- Creates new versions in the object store while preserving previous versions
- Records the publication event in the audit trail

**Important Considerations**: Publication is a significant action that affects business decisions. Users should have appropriate permissions, and the action should be performed with care. All unpublished metrics in the namespace will be published; there is no selective publication at the individual metric level.

**Returns**: Confirmation of the publication operation, including the list of metrics that were published and their new version identifiers.

#### push

**Purpose**: Saves transformer, shifter, or filter code as an unpublished artifact.

**Parameters**:

- `namespace_family` (string, required): The namespace family to push to
- `namespace` (string, required): The namespace to push to
- `content` (string, required): The Python code for the transformer, shifter, or filter
- `artifact_type` (string, required): The type of artifact being pushed: "transformer", "shifter", or "filter"
- `name` (string, required): A unique identifier for the artifact within the namespace family
- `function_type` (string, optional): For transformers only: "forecast" or "actuals"

**Behavior**: Pushes the provided code to the object store as an unpublished artifact. The artifact:

- Is stored in the object store but not used for production calculations
- Can be tested and validated before publication
- Remains isolated from published artifacts until explicitly published
- Is associated with the specified namespace family and namespace

**Returns**: Confirmation of the push operation, including the artifact identifier and storage location.

### Resources Interface

The Resources interface provides contextual information that helps planners understand the planning domain and existing models. Resources are read-only and provide background context rather than actionable functions.

#### Metric Form and Purpose

For each metric in the system, the server provides a resource that describes:

- **What the metric measures**: A plain-language description of the business quantity
- **How it is calculated**: A conceptual explanation of the calculation logic
- **What its dimensions represent**: The meaning of each dimensional category
- **Relationships to other metrics**: How this metric fits into the broader planning hierarchy

This information helps planners understand the existing planning models and design new metrics that integrate properly with the existing structure. When creating a new transformer or modifying an existing one, planners can reference this resource to ensure their code aligns with the intended semantics.

#### Domain Definition and Procedures

When defined through the administration endpoint of the Metrics Service, the server provides descriptions and terms used in a namespace family. These definitions are stored in the object store and presented as a resource that explains:

- Business terminology specific to the planning domain
- Procedures and conventions used in the planning process
- Organizational context for interpreting the data
- Standards and best practices for the domain

This context is essential for planners who need to ensure their models align with business requirements. The resource helps bridge the gap between technical implementation and business intent.

### Prompts Interface

The Prompts interface provides PDD (Prompt Driven Development) scripts that guide planners through common development workflows. These scripts offer structured approaches to creating, testing, and publishing calculation logic.

#### Available Prompts

**Creating a New Transformer for a Metric**: Guides planners through the process of defining a new transformer, including:

- Understanding the input metrics and their dimensions
- Defining the output metric structure
- Writing the transformation logic
- Creating test data for validation

**Creating a Shifter to Align Dimensional Structures**: Helps planners create shifters when input and output metrics have different dimensions, including:

- Identifying the dimensional mismatch
- Choosing the appropriate shifter operation (aggregation, selection, expansion)
- Implementing the dimensional transformation
- Testing with representative data

**Creating a Filter to Split a Metric**: Guides planners through creating filters that split a single metric into multiple components, including:

- Defining the split criteria
- Ensuring the output metrics sum to the input
- Implementing the filter logic
- Validating the mathematical properties

**Generating Test Data for Validation**: Helps planners create comprehensive test data that exercises the important cases in their logic, including:

- Identifying edge cases and boundary conditions
- Creating representative input data
- Defining expected outputs for validation
- Documenting the test scenarios

**Debugging and Fixing Errors**: Provides a systematic approach to diagnosing and fixing problems in existing functions, including:

- Interpreting error messages and test failures
- Tracing data flow through the function
- Identifying common error patterns
- Making targeted fixes without introducing regressions

**Publishing Changes to a Namespace**: Guides planners through the publication process, including:

- Reviewing unpublished changes
- Validating that tests pass
- Understanding the impact of publication
- Executing the publish operation

## Design Considerations

The design of the Planning MCP Server must balance several competing concerns.

**Correctness and safety** are paramount. The code generated by the server will be used to make business decisions, so it must be correct. The server must validate generated code against the expected function signatures and data types, and it must ensure that the code does not have unintended side effects or security vulnerabilities.

**Transparency and explainability** ensure that planners understand what the generated code does. The server should not be a black box; instead, it should explain its reasoning and allow planners to review and modify the generated code. This transparency builds trust and enables planners to learn from the AI assistance.

**Iterative refinement** recognizes that the first version of generated code is rarely perfect. The server must support an iterative workflow where planners can provide feedback, request changes, and gradually refine the code until it meets their needs. This iterative approach is more natural for most users than trying to specify everything perfectly upfront.

**Integration with the function lifecycle** ensures that the code produced by the server follows the proper lifecycle for transformers, shifters, and filters. Generated code starts in the local workspace state, can be pushed to unpublished, and eventually can be published for production use. The server must understand and support this lifecycle.

## Related Documentation

- [v0 Architecture Overview](v0_architecture.md)
- [v0 Requirements - MCP Server Planning Section](../requirements/v0_requirements/v0_requirements.md#34-mcp-server---planning)
- [Function Lifecycle](v0_architecture.md#function-lifecycle) in the architecture overview
