# Serialization Layer

Multi-backend serialization abstraction for domain models. Provides utilities, validation,
and orchestration for converting between internal and external representations.

## Responsibilities

- Serialization utilities for various formats and protocols
- Multi-step workflow orchestration with callback hooks where needed
- Format-agnostic interfaces for Unified Models
- Backend-specific implementation details (Arrow, JSON schema, OpenAPI, MCP, PostgreSQL)
- Data validation and schema enforcement

## Package Boundaries

_(none at this level — see child briefs (`arrow`, `json_schema`, `mcp`, `openapi`, `rdbms`) for backend-specific boundaries.)_

## Interaction Patterns

**Utility Functions**: Many serde operations are straightforward utilities. Domain calls them
with parameters and gets results back (e.g., validate JSON against schema, serialize to Arrow).

**Workflow Orchestration**: Some operations are multi-step workflows where domain provides
callbacks/lambdas at key decision points. Example: mutation orchestration accepts domain-provided
INSERT statements as part of the Flush/Purge/Assert pipeline.

## Current Backends

- `arrow/` — Apache Arrow columnar serialization for inter-program communication
- `json_schema/` — JSON schema validation utilities for domain logic
- `mcp/` — MCP protocol descriptors for APIs
- `openapi/` — OpenAPI specification and contract definition
- `rdbms/` — PostgreSQL persistence with bitemporal support

## Type Mapping and Registration

**Architectural Constraint**: The `BitemporalTableRegistry` and type mapping system
maintain a clean separation between generated models (from go-jet) and hand-authored domain models.
This allows `ExecuteBitemporalMutation` to work transparently with domain models.

**The Public Bridge**: `RegisterDomainModel(domainType, generatedType)` acts as the public API bridge.
Domain model packages must call this at package initialization to explicitly declare their persistence mapping.
While domain packages perform this registration, the internal registries (`BitemporalTableRegistry`) and mappers themselves
remain hidden behind the `serde/rdbms` public API and should not be exposed to or queried by other packages.

**Rationale**: Exposing the registry mapping data would create tight coupling across the system. Instead, domain packages declare their mappings one way during initialization, and the mapping data stays contained within `serde/rdbms`.

## Public API: serde/rdbms

### Mutation Orchestration

- **ExecuteBitemporalMutation[T LockableEntity]** — Multi-phase mutation pipeline (lock, validate, insert, fetch)
  - Requires knownState []T implementing LockableEntity (EntityID, NaturalKey)
  - Optionally accepts a custom BitemporalMutationBuilder for INSERT generation
  - Returns mutated/inserted entities with IDs populated and latest database state
  - Internally handles validation, locking, FLUSH/PURGE/ASSERT, and result fetching

- **DefaultBitemporalInserter[T LockableEntity]** — Generic insert statement builder
  - Used when custom builder not provided to ExecuteBitemporalMutation
  - Handles OVERRIDING SYSTEM VALUE for ID preservation across new/existing entities
  - Batches large inserts to avoid query size limits

- **RegisterDomainModel(domainType, generatedType)** — Register domain→generated mapping
  - Called at package init time from domain model packages
  - Required for ExecuteBitemporalMutation to accept domain models
  - Must be called before any mutations using that domain type

### Temporal Predicates (Query Helpers)

- **IsValidAt(start, end, validAt)** — Predicate: row is valid at a given instant
- **AsOf(start, end, v)** — Predicate: row represents belief as of instant v
- **BelievedAt(start, end, s)** — Predicate: row was believed-true at instant s
- **RecordedAfter(start, s)** — Predicate: row recorded after instant s
- **IsActive()** — Predicate: row is in active partition

### Utilities

- **IsNoRows(err)** — Check if error signals empty result set from go-jet's qrm

### Internal API (Private)

The following are implementation details and should not be imported or called from other packages:

- `getActiveFactAt[T]`, `getHistoricalFactAt[T]`, `getTimelineAsOf[T]`, `getAuditTrailFor[T]` — State query helpers (private)
- `resolveEntityID()` — Entity ID resolution (private)
- `acquireLocks()`, `advisoryLockKey()` — Locking utilities (private)
- `allCols()` — Reflection column projection helper (private)

## When you work here

- Design utilities as simple function calls when possible
- Use callback-driven orchestration only where the workflow truly requires it
- Keep backends independent and testable in isolation
- Document the interface and expected behavior clearly
- **Do not expose type registries or mappers** — use public API functions instead

## References

- `memory-bank/specifications/patterns/service-organization.md` — Unified model pattern (reference as needed)
