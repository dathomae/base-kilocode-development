# Terms

**Timecone.ai**
: The overarching visionary planning system designed to manage multiple possible futures simultaneously. It acts as an Institutional Brain, powered by the deterministic ChronoCone calculation engine.

**Institutional Brain**
: An ontology mapping of business metrics that prevents siloed logic. Rather than individual "hero" analysts writing customized Excel logic, logic is stored centrally and decoupled from execution.

**Consensus Trap**
: The organizational phenomena where discordance is suppressed to force a singular agreed-upon future plan, leading to massive median error and inventory distortion. Timecone.ai breaks the trap by encouraging the generation of multiple, conflicting future scenarios.

**Permissionless Assertion Layers**
: The capacity for agents or domains to submit differing, contradictory future projections without requiring override approval. In the ChronoCone engine, this is expressed by creating entirely new, concurrent DAG plan documents representing the differing assertions.

**Systemic Idea Darwinism**
: The process by which the system evaluates and aggregates Permissionless Assertion Layers. By evaluating actuals against the asserted forecast documents, the system prunes inaccurate models and establishes probability distributions of accurate futures.

**subject_gateway**
: The subject_gateway table represents authenticated users or system actors. It maps external OIDC (OpenID Connect) identities to internal subject identifiers. A special subject with subject_id = -1 represents a system default actor ("nobody").

**namespace_family**
: A namespace_family groups related namespaces under a common owner and type. Families can be application-level (e.g., production, testing), personal (e.g., sandboxes), or team-based (collaborative environments).

**namespace**
: Namespaces provide isolation boundaries for taxonomies, metrics, and time series data. Each namespace has a type (e.g., Production, Alpha, Experiment) that determines its capabilities through policy rules.

**namespace_family_namespace_type**
: This junction table defines the policy capabilities for combinations of family types and namespace types. It controls permissions such as whether an owner can delete a family, whether a namespace can map resources outside its family, and whether resources can be shared.

**category**
: A category defines a classification dimension or grouping. Categories can form hierarchies where a category can expand another element (via taxonomy_parent_id), allowing rich taxonomic structures. Categories are dimensions associated with metric. **Time is not a category**—time is a fundamental property of the time series itself (see Time Grain).

**element**
: An element represents a specific member within a category. Elements are the leaf nodes for classification—the actual values that can be used to tag data or define dimensions.

**category_attr**
: These tables store flexible JSONB attributes on categories and elements. The attr_owner_ns_id allows namespaces to extend categories and elements with custom attributes, even if those objects were defined in different namespaces.

**element_attr**
: These tables store flexible JSONB attributes on categories and elements. The attr_owner_ns_id allows namespaces to extend categories and elements with custom attributes, even if those objects were defined in different namespaces.

**metric**
: A metric represents a measurable quantity with defined dimensions (categories). It references a concept (an element) that provides semantic meaning. Each metric has a time grain that defines the temporal spacing of its time series data points.

**Time Grain**
: The temporal interval between data points in a time series (e.g., hour, day, week, month). Time grain is a property of the metric itself, not a category. All time series within a metric share the same time grain.

**Grain Dimension**
: A dimension associated with time, declared via the `grain <measure>` clause in a CPL metric or connector declaration. Every metric must have a grain. The grain's underlying measure must be an Interval or derived from Interval. At the verification/examples level, grain dimensions are structurally identical to concept dimensions — both are named columns with concept-derived values. The distinction between grain and non-grain dimensions is conceptual (enforced during CPL analysis), not structural.

**Concept Dimension**
: A dimension that categorizes data by a non-time concept (e.g., Region, Product). Unlike grain dimensions, concept dimensions are not constrained to Interval-derived types. At the verification/examples level, concept dimensions are structurally identical to grain dimensions — both are named columns with concept-derived values.

**metric_dim**
: This table defines the dimensional structure of a metric. Each dimension is an ordered reference to a category, establishing what taxonomic values can index the metric. The terms "metric dimension" and "dimension" refer to the categories that define a metric's structure, not to the category table itself.

**transform**
: A transform represents executable code that transforms input metrics into output metrics. The code is stored in a storage_config location as a **CPL specification** and can be versioned through the bitemporal validity model.

**transformer**
: A **CPL specification** that produces metric values. Each metric has at least a forecast transformer that produces predicted values; metrics receiving external data may also have an actuals transformer that produces historical ground truth values. Transformers are compiled and executed by the calculation engine with inputs and outputs that share the same dimensions, time grain, and forecast length.

**forecast transformer**
: A transformer that produces predicted values for future time periods. Every metric must have a forecast transformer.

**actuals transformer**
: A transformer that produces historical values representing ground truth measurements. Only metrics that receive data from external systems typically have actuals transformers.

**shifter**
: A **CPL specification** that transforms a metric with one set of dimensions into a metric with a different set of dimensions. Shifters enable transformers to work with inputs that have incompatible dimensional structures by performing operations such as aggregation (summing across dimensions), selection (choosing specific elements), or expansion (replicating values across new dimensions).

**filter**
: A **CPL specification** that splits a single input metric into multiple output metrics where the values in the input metric equal the sum of the corresponding values in the output metrics. Used for segmentation, allocation, and scenario analysis.

**transform_input**
: These tables define the metrics consumed and produced by a transform. Input metrics have an input_order to maintain parameter position.

**transform_output**
: These tables define the metrics consumed and produced by a transform. Input metrics have an input_order to maintain parameter position.

**ts_doc_type**
: A ts_doc_type defines the structure of a time series document. It acts as a schema or template for actual document instances.

**ts_doc_row_type**
: A ts_doc_row_type defines a single row type within a document. Each row type specifies which metric structure the rows will follow.

**ts_doc**
: A ts_doc represents an actual time series document instance—a collection of data stored at a specific point in time. The document itself is the output of a deterministic Directed Acyclic Graph (DAG) calculation. The as_of timestamp provides version control.

**ts_doc_tags**
: This table associates documents with taxonomy elements for classification and filtering. Tags reference element instances.

**ts_doc_input_lineage**
: This table tracks data lineage by recording which documents were used as inputs to create other documents. It enables dependency analysis and impact assessment.

**storage_config**
: The storage_config table defines external storage locations for data and code. It is key to decoupling execution logic from data storage. It supports versioning through the bitemporal validity model.

**connector**
: A connector represents a connection to a data source or sink. Each connector has a driver, which is a named plugin that is used to access the data source or sink. Connectors serve two purposes in ChronoCone: accessing external services for pulling/pushing time series data, and accessing object stores for storing time series data files.

**driver**
: A named plugin that implements the logic to access a specific type of data source or sink. Examples include "local" for local filesystem access, "s3" for AWS S3, and service-specific drivers for external APIs.

**Bitemporal Model**
: All major tables implement a bitemporal model using four temporal columns: two for the validity dimension (`validity_start`, `validity_end`) and two for the transaction dimension (`transaction_start`, `transaction_end`). Tables are physically partitioned into an `active` partition (current beliefs, B-tree indexed) and a `historical` partition (retired beliefs, GiST exclusion constrained). Mutations follow the Flush/Purge/Assert pipeline via `ExecuteBitemporalMutation`. See [Database Design](specifications/architecture/database_design.md) for full details.

**Multi-tenant System**
: A system for managing hierarchical taxonomies, time series data, data transformations, and access policies across multiple tenants.

**document type**
: A synonym for [`ts_doc_type`](memory-bank/terms.md:19).

**document**
: A synonym for [`ts_doc`](memory-bank/terms.md:21).

**Actuals**
: The actual (historical) values of a metric. These values always represent past measurements and are the ground truth against which forecasts are compared.

**Forecast**
: Predicted values of a metric. Forecast data may represent either past or future time periods. Forecasts in the past are primarily used for determining forecast accuracy by comparing against actuals.

**DTO (Data Transfer Object)**
: Pydantic models in the API service layer used to enforce strict input/output shapes. They specifically hide internal database identifiers (e.g., auto-incrementing integers) from external consumers, exposing only safe, human-readable Natural Keys to maintain strict encapsulation and predictable API behavior.

**Natural Key**
: A human-readable business identifier (like `key` or combinations of parent keys, e.g., `ns_fam_key` + `ns_key`) used exclusively by the API and frontend to reference entities. The service layer assumes full responsibility for resolving Natural Keys into their corresponding internal database primary keys prior to execution.

**ErrorCode**
: An enumeration of standardized failure modes (e.g., `NOT_FOUND`, `VALIDATION_FAILED`, `CONFLICT`) used by services to communicate domain errors without leaking HTTP-specific exceptions into the business logic layer.

**AuditContext**
: A mandatory data structure for all database mutations containing `reason`, `agent`, and `auth_chain`. It ensures every change to the Institutional Brain is accompanied by a bitemporal audit trail for accountability and lineage.

**ServiceResult**
: A Golang-style return pattern used across service APIs: a tuple of `(Optional[T], Optional[ErrorCode])`. This forces explicit error handling and prevents uncontrolled exception bubbling.

**serde**
: Short for Serialization/Deserialization. In the ChronoCone architecture, the `serde` layer is the technical boundary responsible for translating between the system's high-level logical models (hierarchical CPL concepts and scoped namespaces) and various external formats including the flat RDBMS schema, columnar data (Arrow), API specifications (OpenAPI/JSON Schema), and communication protocols (MCP).

**Structural Definition**
: The initial phase of the development lifecycle where class definitions, function headers, signatures, and English documentation in docstrings are created. Implementations are restricted to `pass` or empty returns to define the system's "shape" before any logic is added.

**Logical Contract**
: The phase where tests are implemented to verify the structural definition and behavioral intent. These tests must be confirmed to fail before the implementation logic is added, ensuring the contract is valid and enforceable.

## Agent Personas

**work-planner**
: An AI agent persona responsible for creating and managing plans, stories, and task breakdowns. The work-planner creates complex execution plans but does not execute them. Rules are defined in `.kilocode/rules/rules-work-planner/work-planner.md`. Not to be confused with the Planner user persona.

**story-implementor**
: An AI agent persona responsible for executing story files by processing tasks sequentially and routing work to appropriate agents. Rules are defined in `.kilocode/rules/rules-story-implementor/story-implementor.md`. Not to be confused with the System Integrator (SI) user persona.

## Planning and Documentation

**Epic**
: A collection of related user stories grouped into a larger initiative. Epics provide a higher-level planning abstraction that organizes multiple stories toward achieving a common overarching goal. Stored in `memory-bank/epics/`. Uses the template at [`resources/templates/epics.md`](resources/templates/epics.md).

**ERO**
: Epic Restraining Order. A temporary document that captures design decisions specifically related to managing confusion from incremental implementation. EROs bridge the gap between the current implementation state and the eventual/final implementation state by documenting: (1) the current state vs final state gaps, (2) technical decisions made to handle incomplete implementations, and (3) clarifications needed due to mismatch between current and final system state. While called "temporary", this state could potentially last months or years. EROs exist because primary requirements, architecture, and design docs describe the system as it ultimately will be, but there can be significant differences between that final state and the current state due to incomplete implementation. EROs make this gap explicit and documented. Uses the template at [`resources/templates/ero.md`](resources/templates/ero.md).

**technical-writer**
: An AI agent persona responsible for writing technical documentation such as requirements, design documents, and specifications. Rules are defined in `.kilocode/rules/rules-technical-writer/technical-writer.md`.

**technical-editor**
: An AI agent persona responsible for reviewing technical documents for correctness, clarity, and form. Rules are defined in `.kilocode/rules/rules-technical-editor/technical-editor.md`.

**Code**
: An AI agent persona responsible for writing, modifying, and refactoring code. Rules are defined in `.kilocode/rules/rules-code/code.md`.

**Architect**
: An AI agent persona responsible for system design, creating technical specifications, and making architectural decisions. Rules are defined in `.kilocode/rules/rules-architect/architect.md`.

**Debug**
: An AI agent persona responsible for troubleshooting and diagnosing issues in software. Rules are defined in `.kilocode/rules/rules-debug/debug.md`.

**Ask**
: An AI agent persona focused on providing explanations and answering technical questions. Rules are defined in `.kilocode/rules/rules-ask/ask.md`.

## User and Customer Personas

**Customer**
: The customer persona representing executive decision-makers (CFO, CIO, CTO, COO) who purchase and oversee ChronoCone for their organization. Concerned with business efficiency, data security, and high-level reporting. Defined in `memory-bank/specifications/concepts.md`.

**System Integrator (SI)**
: A user persona representing consultants with deep domain expertise who implement ChronoCone plans and integrations. SIs set up systems, create plans, and build connectors. Not to be confused with the story-implementor agent persona. Defined in `memory-bank/specifications/concepts.md`.

**Analyst**
: A user persona representing customer employees with development/programming expertise who improve and adjust existing plans. They have deep understanding of forecasting and business details. Defined in `memory-bank/specifications/concepts.md`.

**Planner**
: A user persona representing the primary users of plan outputs who run day-to-day operations. They have the deepest experience with plan accuracy and correctness. Not to be confused with the work-planner agent persona. Defined in `memory-bank/specifications/concepts.md`.

**Power Planner**
: A user persona representing users who share characteristics of both Analysts and Planners, with some development/programming expertise. They require more guidance than Analysts but share concerns of both user types. Defined in `memory-bank/specifications/concepts.md`.

## CPL (Chronocone Planning Language)

**CPL (Chronocone Planning Language)**
: A domain-specific language (DSL) designed for defining metrics, data sources, calculations, and dimensional relationships in planning and forecasting systems. CPL files use the `.cpl` extension.

**Unit**
: Defines how a value is measured (the scale and representation). Units map names to built-in types that determine how values are stored and interpreted, such as `Count`, `Currency`, or `Day`.

**Dimension**
: Defines the conceptual category that a value represents. A dimension associates a conceptual category with a unit type, allowing values to carry both semantic meaning and measurement context.

**Built-in Type**
: Predefined types available in CPL for unit declarations. Includes `String`, `Number`, `Count`, `Currency`, and time-oriented types (`Second`, `Minute`, `Hour`, `Day`, `Week`, `Month`, `Year`).

**Type Parameter**
: Optional parameters enclosed in parentheses for specifying presentation and formatting attributes on built-in types. For example, `Currency(symbol: '$')` or `Number(decimal_separator: '.')`.

**Aggregator**
: A transformation that aligns data across different dimensional granularities by rolling up or summarizing values from one set of dimensions to another. Aggregators enable metrics with finer granularity to feed into metrics with coarser granularity.

**Spreader**
: A transformation that distributes values from a coarser granularity to a finer granularity. Spreaders are the inverse of aggregators.

**Lookup**
: A reference table that maps between different concept spaces. Lookups provide static or slowly-changing reference data used by transformers, shifters, or filters.

**AST (Abstract Syntax Tree)**
: A tree representation of the structure of CPL declarations produced during the parsing phase. The AST captures the syntactic structure of the source code for further processing.

**Symbol Table**
: A data structure built during parsing and analysis that captures names and their scopes across a CPL file. Maps unit names to type information and dimension names to unit references.

**Data and Control Flow Graph**
: An intermediate representation produced during the analysis phase, oriented around internal operators. This graph represents the plan in an abstract operations language used by the synthesis phase to generate executable code.

**Parsing Phase**
: The first phase of CPL processing where `.cpl` files are read and validated for syntactic correctness. Produces an AST and symbol tables.

**Analysis Phase**
: The second phase of CPL processing where relationships are validated against the source and taxonomy database. Produces a Data and Control Flow Graph.

**Synthesis Phase**
: The third phase of CPL processing where the validated intermediate representation is transformed into an executable form (e.g., Go code).

**Evaluation Phase**
: The fourth phase of CPL processing where the synthesized form is executed to produce planning outputs, including calculated metric values.

## slog (Structured Logging)

**slog**
: Go's standard library structured logging package (`log/slog`). It organizes log data as key-value pairs, enabling automated processing and consistent formatting across different output destinations.

**Logger (slog)**
: The primary interface for creating structured log entries in the slog package. Provides methods for each severity level (Debug, Info, Warn, Error) and accepts message strings with alternating key-value attribute pairs.

**Handler (slog)**
: An interface in the slog package that determines how log entries are formatted and where they are written. Built-in handlers include `TextHandler` for human-readable output and `JSONHandler` for machine-readable output.

**TextHandler**
: A built-in slog handler that formats log entries as key-value pairs in a human-readable text format. It is the default handler for the package-level logging functions.

**JSONHandler**
: A built-in slog handler that formats log entries as JSON objects. This format is preferred for production environments where logs are processed by aggregation systems.

**Level (slog)**
: A severity level in slog indicating the importance of a log entry. The levels in increasing order are: Debug (-4), Info (0), Warn (4), and Error (8).

**Attribute (slog)**
: The fundamental building block of structured logging in slog, represented by the `slog.Attr` type. Each attribute consists of a string key and a typed `slog.Value`, forming key-value pairs that provide context for log messages.

**Value (slog)**
: The `slog.Value` type holds attribute data with one of several kinds: Bool, Duration, Float64, Int64, String, Time, Uint64, Any, Group, or LogValuer.

**LogValuer**
: An interface in the slog package allowing types to customize their log representation. Types implementing `LogValue() Value` can control how they appear in logs and prevent accidental logging of sensitive fields.

**Group (slog)**
: A named collection of attributes in slog that creates hierarchical structures. Groups organize related attributes under a common namespace, useful for HTTP request details, database parameters, or any logically related data.

## Computation Model

**CPU Core**
: An independent processing unit within a CPU that can execute instructions. Each core has its own computation unit and cache, allowing parallel execution of multiple threads or goroutines.

**CPU Cache**
: A small, fast memory hierarchy (typically L1, L2, and sometimes L3) within a CPU core or shared among cores. Cache stores frequently accessed data from main memory to reduce access latency. L1 cache is fastest and smallest, while higher levels are larger but slower.

**Main Memory**
: The primary volatile memory (RAM) shared by all CPU cores in a system. While significantly larger than cache, it has higher access latency. All data for Chronocone computations resides in main memory during execution.

**Chunking Strategy**
: A function that divides tensor data into contiguous blocks (chunks) sized to optimize cache utilization. The strategy controls the granularity of parallel processing and ensures that the working set of an operation fits within CPU cache for efficient execution.

**Data Locality**
: The principle of organizing data access patterns so that related data is stored contiguously in memory and accessed sequentially. Good data locality maximizes cache hits and minimizes memory bandwidth usage.

**Contiguous Memory**
: A block of memory where data elements are stored sequentially without gaps. Chronocone tensors use contiguous arrays to enable efficient cache utilization and vectorized operations.

## ANTLR 4

**ANTLR 4**
: ANother Tool for Language Recognition - A powerful parser generator for reading, processing, executing, or translating structured text or binary files. ANTLR generates lexers, parsers, and tree walkers from grammars.

**Grammar (ANTLR)**
: A declarative specification of language syntax consisting of lexer rules (token definitions) and parser rules (syntactic structure). ANTLR grammars are defined in `.g4` files and support multiple target languages.

**Lexer Rule (ANTLR)**
: A rule in an ANTLR grammar that defines how to recognize tokens from input characters. Lexer rule names must start with a capital letter. Lexer rules can include character sets, string literals, fragments, and lexer commands like `skip` or `channel`.

**Parser Rule (ANTLR)**
: A rule in an ANTLR grammar that defines syntactic structure by combining tokens and other parser rules. Parser rule names must start with a lowercase letter. Rules can have alternatives, labels, parameters, return values, and local variables.

**Parse Tree (ANTLR)**
: A tree data structure representing how the parser recognized the input. Interior nodes are phrase names (rule names) and leaves are input tokens. Parse trees preserve all input information and can be walked using listeners or visitors.

**Listener (ANTLR)**
: An interface mechanism for processing parse trees where the parser triggers `enter` and `exit` methods as it walks the tree. Listeners keep application code separate from grammars. ANTLR generates listener interfaces and base classes from grammars.

**Visitor (ANTLR)**
: An alternative mechanism to listeners for processing parse trees where the application explicitly calls `visit` methods on child nodes. Visitors provide more control over tree traversal but require manual walking of children.

**Token (ANTLR)**
: The smallest unit of input recognized by the lexer. Tokens have attributes including text, type, line number, position, index, and channel. Tokens are passed to the parser for syntactic analysis.

**Lexical Mode (ANTLR)**
: A mechanism in ANTLR lexer grammars that groups rules by context, allowing different tokenization behavior in different contexts (e.g., inside vs. outside XML tags). Modes enable the creation of sublexers. Only available in lexer grammars.

**Semantic Predicate (ANTLR)**
: A boolean expression in the target language (wrapped in `{...}?`) that determines whether an alternative should be considered during parsing. Predicates enable context-sensitive parsing decisions and disambiguation.

**Action (ANTLR)**
: A block of target language code embedded in a grammar (enclosed in `{...}`) that executes during parsing. Actions can access token and rule attributes, and are used to perform side effects or compute values during recognition.

**Fragment (ANTLR)**
: A lexer rule modifier that defines a helper rule for recognizing token components without producing a token itself. Fragments are reusable patterns used by other lexer rules but are invisible to the parser.

**Left Recursion (ANTLR)**
: A grammar pattern where a rule references itself as the first element of an alternative. ANTLR 4 automatically rewrites direct left-recursive rules into non-left-recursive equivalents using precedence climbing.

**Channel (ANTLR)**
: A mechanism for classifying tokens so the parser can ignore certain types (like whitespace or comments). The default channel is 0; hidden tokens go to channel `Token.HIDDEN_CHANNEL`. Custom channels can be defined.

**Alternative Label (ANTLR)**
: A label (using the `#` operator) attached to alternatives in parser rules that generates specific listener/visitor methods for each alternative. All alternatives in a rule must be labeled or none of them.

**Element Label (ANTLR)**
: A label (using the `=` operator) attached to rule elements that creates fields in the rule context object, allowing access to specific tokens or rule matches within actions. List labels using `+=` collect multiple matches into a list.

**ATN (Augmented Transition Network)**
: The internal graph data structure used by ANTLR to represent grammars for prediction and parsing decisions. The ALL(\*) parsing strategy uses the ATN to make adaptive lookahead decisions.

**ALL(\*) Parsing**
: ANTLR 4's adaptive LL(\*) parsing strategy that uses the ATN to make parsing decisions with arbitrary lookahead when necessary. It caches parsing decisions to improve performance on subsequent parses.

**XPath (ANTLR)**
: A query language for selecting nodes in parse trees using path expressions like `//method` or `/class/field`. ANTLR provides XPath support for extracting specific subtrees from parse results.

**Tree Pattern (ANTLR)**
: A pattern matching mechanism for testing if parse tree subtrees have a particular structure and extracting components. Patterns use angle bracket tags like `<ID>` to match rule or token references.

**Target Language (ANTLR)**
: The programming language for which ANTLR generates parser code. Available targets include Java, C#, Python, JavaScript, TypeScript, Go, C++, Swift, PHP, and Dart. Each target has its own runtime library.

**RecognitionException (ANTLR)**
: The base exception class thrown by ANTLR-generated recognizers when parsing fails. Subclasses include `NoViableAltException`, `InputMismatchException`, `LexerNoViableAltException`, and `FailedPredicateException`.
