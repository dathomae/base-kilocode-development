# Concepts

**Institutional Brain**
: An ontology mapping of business metrics that prevents siloed logic. Rather than individual analysts writing customized spreadsheet logic, reasoning logic is stored centrally across the organization.
: Related: Ontology, Namespace Mapping, Document Type, Metric

**Name-Based Dependency Resolution**
: Dependencies between entities should always be specified using human-readable names rather than internal database IDs. The system resolves names to IDs internally. This ensures APIs and CLIs remain user-friendly and decouples users from internal database implementation details.
: Related: Internal IDs, Database IDs, Foreign Keys

**Document Type**
: Master blueprint defining activities, dependencies, timelines, rules, calculations, and dependencies that govern how inputs transform into outputs. Planning organizations create and maintain these document types. In the database schema, document types are stored in the [`ts_doc_type`](memory-bank/terms.md:19) table.
: Related: Document

**Document**
: Concrete execution of a document type at a particular point in time, derived from a document type plus input data. Produces specific numerical or categorical outputs that guide actual operational decisions. To construct a document, the system uses document type plus available document inputs, transformers, shifters, and filters to construct a deterministic, discrete Directed Acyclic Graph (DAG) for ground-truth calculation. In the database schema, documents are stored in the [`ts_doc`](memory-bank/terms.md:21) table.
: Related: Document Type

**Probabilistic Aggregation**
: To prevent the Consensus Trap, Timecone.ai models multiple possible futures. Because each base Document is deterministic, probability distributions are created by constructing an aggregating Document that depends on an array of conflicting base Documents (Permissionless Assertion Layers).
: Related: Document

**Information Flow**
: Acyclic flow of information through the planning system in one direction via DAGs. High-level plans typically have organizational associations, and outputs flow downward to simplify dependencies. Branching, overriding, and merging logical assertions happens through distinct document iterations.
: Related: Document Type, Document

**Activity**
: Unit of work within a plan. The smallest actionable component of a document type that defines a specific task or operation to be performed.
: Related: Document Type

**Dependency**
: Relationship constraining activity execution order. Defines prerequisites and sequencing between activities within a plan.
: Related: Activity

**Metric**
: Time-series data collected for analysis and optimization. Includes both actual historical data and forecasted values. Stored in PostgreSQL database with associated metadata.
: Related: Document

**Endpoint**
: External system integration point. Used for pulling data into ChronoCone from external systems or exporting data to external systems. Types of endpoints are defined and discovered via the Core Server.
: Related: Connector

**Actuals**
: The actual (historical) values of a metric, always representing past measurements.
: Related: Metric, Forecast

**Forecast**
: Predicted values of a metric, may be in past or future; past forecasts used for accuracy assessment.
: Related: Metric, Actuals

**Transformer**
: A **CPL specification** that produces metric values. Each metric has at least a forecast transformer; metrics receiving external data may also have an actuals transformer. Transformers are tracked in the database and compiled by the calculation engine.
: Related: Metric, Forecast, Actuals

**Shifter**
: A **CPL specification** that aligns metrics with different dimensional structures. When a transformer requires input metrics with different dimensions than its output, shifters transform the inputs to match. Operations include aggregation, selection, and expansion across dimensions.
: Related: Transformer, Metric Dimension

**Filter**
: A **CPL specification** that splits a single input metric into multiple output metrics where input cardinality equals the sum of output cardinality and all values are the same. Used for segmentation, allocation, and scenario analysis.
: Related: Metric, Transformer

**Metric Dimension**
: The categories that define a metric's dimensional structure. Each dimension is an ordered reference to a category, establishing what taxonomic values can index the metric. The calculation engine requires transformers' input and output metrics to have matching dimensions. The system finds a series of shifters to align input dimensions; therefore, all transfers are written as if the dimensions they require already exist.
: Related: Category, Transformer, Metric, Shifter

**Ontology Service Layer**
: An architectural boundary between application logic and the database. It enforces bitemporal safety, prevents internal ID leakage, and provides high-level control over the project's ontology (namespaces, categories, metrics). It uses standardized return shapes (ServiceResult) and is the only layer permitted to perform mutations.
: Related: DTO, Natural Key, Bitemporal Model, ErrorCode

**Epic Story**
: A collection of related user stories grouped into a larger initiative. Epic stories break major objectives into manageable chunks that can be distributed across multiple development cycles. Each epic has an overarching goal that is achieved when all its constituent stories are completed. Epics provide a higher-level planning abstraction above individual stories. Uses the template at [`resources/templates/epics.md`](resources/templates/epics.md).
: Related: Story, story-implementor, work-planner

**Logical TDD Lifecycle**
: A specialized development process for ChronoCone that combines BDD-style structural planning with strict TDD. It prioritizes defining the **Structural Shape** (signatures/docstrings) and **Logical Contract** (failing tests) before any business logic is permitted. This ensures that the system's "Knowledge" is verified before implementation begins. See `scripts/story-implementor.pdd.script.md` for the mandatory execution steps.
: Related: Structural Definition, Logical Contract, TDD

**Epic Restraining Order (ERO)**
: A temporary document that captures design decisions specifically related to managing confusion from incremental implementation. EROs bridge the gap between the current implementation state and the eventual/final implementation state by documenting: (1) the current state vs final state gaps, (2) technical decisions made to handle incomplete implementations, and (3) clarifications needed due to mismatch between current and final system state. While called "temporary", this state could potentially last months or years. EROs exist because primary requirements, architecture, and design docs describe the system as it ultimately will be, but there can be significant differences between that final state and the current state due to incomplete implementation. EROs make this gap explicit and documented. Uses the template at [`resources/templates/ero.md`](resources/templates/ero.md).
: Related: Epic Story, Story, Architect, technical-writer

**Agent Persona**
: A role that an AI agent takes on when performing work on the ChronoCone project. Each agent persona has specific rules and responsibilities defined in the `.kilocode/rules` directory. Agent personas include work-planner, story-implementor, technical-writer, technical-editor, Code, Architect, Debug, and Ask. Agent personas are distinct from user personas.
: Related: work-planner, story-implementor, technical-writer

**User Persona**
: A type of human user who interacts with the ChronoCone system. User personas are defined in the specifications and represent the people who will directly use the product. User personas include System Integrator (SI), Analyst, Planner, and Power Planner. The Customer is the customer persona representing executive decision-makers. User personas are distinct from agent personas.
: Related: System Integrator, Analyst, Planner, Power Planner, Customer

**Shared-Memory Computation Model**
: The execution model where multiple computation units (CPU cores) share access to a common main memory. In this model, low-level operations take pointers and indices referencing data in shared memory, allowing parallel goroutines to operate on distinct data subsets while reading from and writing to the same address space.
: Related: CPU Core, CPU Cache, Main Memory, Data Locality

**Cache-Optimized Computation**
: A computation strategy that structures data and operations to maximize cache hit rates. By breaking data into appropriately-sized chunks that fit within CPU cache, and accessing data sequentially within chunks, subsequent operations can reuse cached data rather than fetching from slower main memory.
: Related: Chunking Strategy, Data Locality, Contiguous Memory, CPU Cache

**Chunk-Based Parallelization**
: A parallelization approach where tensor data is divided into chunks (contiguous memory blocks) that can be processed independently by different CPU cores. Each core operates on its assigned chunks, with chunk sizes selected to ensure the working set fits in cache for efficient execution.
: Related: Chunking Strategy, CPU Core, Shared-Memory Computation Model

**CPL Processing Lifecycle**
: The four-phase lifecycle that transforms declarative CPL files into executable planning logic: Parsing (syntax analysis producing AST), Analysis (semantic validation producing Data/Control Flow Graph), Synthesis (code generation), and Evaluation (execution producing results). Each phase is independently testable and errors are detected as early as possible.
: Related: Parsing Phase, Analysis Phase, Synthesis Phase, Evaluation Phase, AST, Data and Control Flow Graph

**Declarative over Imperative**
: A core principle of CPL where users specify what they want computed, not how to compute it. This separates the intent from the implementation, allowing the system to optimize execution.
: Related: CPL, Metric

**Dimensional Awareness**
: A core principle of CPL where every value carries dimensional context (e.g., "revenue by country and product"). This ensures calculations are semantically meaningful and dimensionally consistent.
: Related: Dimension, Unit, Metric

**Grain Specification**
: A core principle of CPL where time-based and dimensional granularity is explicit. Every metric has a defined time grain (e.g., hour, day, week) that determines the temporal spacing of data points.
: Related: Time Grain, Metric, Dimension

**Composability**
: A core principle of CPL where metrics can reference other metrics, enabling complex calculation chains. This allows building sophisticated plans from simple, reusable components.
: Related: Metric, Transformer, Document

**External Integration**
: A core principle of CPL providing native support for connecting to HTTP APIs and other data sources through connectors. This enables seamless data import from external systems.
: Related: Connector, Driver, Metric

**Structured Logging**
: An approach to logging where log entries are organized as structured data (key-value pairs) rather than plain text. Go's slog package implements this pattern, making logs easier to search, filter, and analyze programmatically. Each log entry contains a message, severity level, and collection of typed attributes.
: Related: slog, Logger, Attribute

**Severity-based Log Filtering**
: The practice of filtering log entries based on their importance level before they are formatted or output. slog defines levels (Debug, Info, Warn, Error) and discards entries below a configured threshold, improving performance by avoiding unnecessary formatting work.
: Related: slog, Level

**Handler-based Output Formatting**
: A pattern where log output formatting is separated from the logging API through a handler interface. This allows the same application code to write to different output formats (text, JSON) or destinations (files, databases, network services) without code changes.
: Related: slog, Handler, TextHandler, JSONHandler

**Request-scoped Logging**
: A pattern where loggers are associated with request contexts, enabling automatic inclusion of request-specific attributes in all log entries for that request. slog supports this through context integration, allowing consistent tracing across distributed operations.
: Related: slog, Logger, Context

**Grammar-Driven Parsing**
: An approach to language recognition where a formal grammar defines both the lexical structure (tokens) and syntactic structure (phrases) of a language. The grammar serves as a single source of truth that generates the lexer, parser, and tree walker code automatically.
: Related: ANTLR 4, Grammar, Lexer Rule, Parser Rule

**Lexical Analysis**
: The first phase of language processing where input characters are grouped into tokens (words, symbols, numbers). Lexers apply pattern matching rules to identify token boundaries and types, discarding whitespace and comments as needed.
: Related: ANTLR 4, Token, Lexer Rule

**Syntactic Analysis**
: The second phase of language processing where tokens are organized into a hierarchical structure according to grammar rules. The parser verifies that the token sequence conforms to the language's grammatical rules and builds a parse tree representation.
: Related: ANTLR 4, Parser Rule, Parse Tree

**Parse Tree Walking**
: A pattern for traversing parse trees to perform operations like code generation, interpretation, or analysis. ANTLR supports both listener interfaces (event-driven) and visitor interfaces (explicit traversal), keeping application logic separate from grammars.
: Related: ANTLR 4, Parse Tree, Listener, Visitor

**Rule Context Objects**
: Objects generated by ANTLR that represent the parsing context for each rule invocation. These objects contain references to matched tokens, child rule contexts, and user-defined attributes, providing access to all information about how a rule matched the input.
: Related: ANTLR 4, Parser Rule, Parse Tree

**Semantic Analysis**
: The phase of language processing where meaning is derived from syntactic structure. In ANTLR, semantic predicates and actions enable context-sensitive decisions during parsing, while separate passes can perform type checking and validation.
: Related: ANTLR 4, Semantic Predicate, Action

**Adaptive Parsing**
: ANTLR's ALL(\_) parsing strategy that uses runtime profiling to make parsing decisions. Unlike static LL(k) parsers, adaptive parsers use arbitrary lookahead only when necessary and cache decisions for improved performance on subsequent parses.
: Related: ANTLR 4, ATN, ALL(\_) Parsing

**Left Recursion Elimination**
: ANTLR's automatic transformation of left-recursive grammar rules into equivalent non-left-recursive forms using precedence climbing. This allows natural expression of binary operators while maintaining efficient top-down parsing.
: Related: ANTLR 4, Left Recursion, Parser Rule

**Target-Agnostic Grammars**
: The practice of writing ANTLR grammars without target language-specific code, maximizing reusability across different programming languages. Actions and predicates can be isolated or avoided to create pure declarative grammars.
: Related: ANTLR 4, Grammar, Target Language

**Lexical Modes**
: A technique for context-sensitive tokenization where the lexer switches between different sets of rules based on context (e.g., inside vs. outside string literals). Modes enable recognition of complex languages with different tokenization needs in different regions.
: Related: ANTLR 4, Lexical Mode, Lexer Rule

**Grammar Composition**
: The ability to import and combine grammars, enabling code reuse and grammar inheritance. ANTLR treats imported grammars like object-oriented superclasses, allowing rule overriding and extension while producing a single combined parser.
: Related: ANTLR 4, Grammar, Parser Rule

**No Manufactured Data**
: A core data integrity principle stating that the system must never invent, fabricate, or fill in missing data with default values. Input data must match the CPL specification exactly: the number of rows must equal the Cartesian product of all dimension and grain concept values, with no missing combinations and no duplicate or extraneous rows. If an input data source does not exactly match the structure implied by CPL definitions, the system must terminate with a clear error rather than silently substituting zeros, discarding duplicates, or otherwise altering the data.
: Related: Connector, Transformer, Parquet, Data Integrity

**Predicate-Guided Parsing**
: The use of semantic predicates (boolean conditions evaluated at parse time) to guide parsing decisions. Predicates can disambiguate alternatives, enable context-sensitive parsing, and implement semantic analysis during the parse phase.
: Related: ANTLR 4, Semantic Predicate, Parser Rule

**Dimension Uniformity at Verification Level**
: At the verification and example level (parquet files, expected outputs, ODS spreadsheets), all dimensions — whether declared as grain dimensions (time-related, via the `grain` keyword) or concept dimensions (arbitrary categorical) — are structurally identical. Both derive from concepts, both appear as named columns in data files, and both carry string values from their respective concept element sets. The grain-vs-concept distinction is meaningful only during CPL parsing/analysis where the `grain` keyword enforces time-related constraints. Tools operating at the verification level (generators, verifiers, comparison utilities) should treat the full ordered list of dimensions uniformly rather than splitting them into separate "grain" and "dimensions" categories.
: Related: Dimension, Metric Dimension, Grain Specification, Time Grain, Concept

**LibreOffice Calc (Dev Dependency)**
: A required development environment tool for creating ODS template files and visually inspecting generated spreadsheets during verification. CLI usage: `libreoffice --calc` opens Calc with a new empty spreadsheet; `libreoffice <path>` opens a specific `.ods` file in Calc (the application auto-detects file type by extension). Calc is used during story execution to create the embedded `template.ods` (Task 1 of the generator story) and to visually verify branch formulas and final values during user inspection stops in validation tasks.
: Related: ODS Generator, ODS Verifier, go-ods, ODF (OpenDocument Format)
