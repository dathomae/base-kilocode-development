# Data Stores Requirements

This document specifies requirements for object storage and relational database usage within ChronoCone.

## Relational Database Requirements

### REL-001: Database Engine Selection

**The system shall use PostgreSQL as the relational database for all structured metadata storage.**

This includes, but is not limited to:
- Namespace families, namespaces, and their policies
- Taxonomy (categories, elements)
- Metric definitions, transformers, shifters, and their DAG relationships
- Timeseries document metadata and lineage
- Storage configuration (e.g., S3 integration)
- Authentication gateway (subject mapping)
- Localization data (display names, tooltips, descriptions per locale)
- Configuration data
- Audit metadata

**Decision rationale:** Graph databases (Neo4j, ArangoDB, Amazon Neptune, TerminusDB, TigerGraph) were evaluated as alternatives for the graph-like aspects of this data (DAG relationships, taxonomy hierarchies, lineage). PostgreSQL was confirmed as the better fit because:

1. **Bitemporality**: No graph database offers native bitemporal support. The **Flush/Purge/Assert** append-only pipeline with `validity_start`/`validity_end` would need to be built from scratch in every graph DB, and graph traversal performance degrades when every node and edge requires temporal filtering.
2. **Localization**: The localization pattern (separate `*_localization` tables, each also bitemporal, per language/country) maps naturally to relational constraints and loses enforcement rigor in property-graph models.
3. **Shallow, predictable traversals**: ChronoCone's graph structures (taxonomy, metric DAGs, lineage) are shallow and bounded, not deep/unpredictable. PostgreSQL recursive CTEs handle these efficiently.
4. **Strong consistency**: ACID transactions with serializable isolation are trivially available in PostgreSQL; graph databases offer weaker guarantees, especially for multi-entity writes.
5. **JSONB flexibility**: Attribute tables use JSONB for flexible values with relational constraints around them, a pattern PostgreSQL supports natively with GIN indexing.
6. **Operational maturity**: PostgreSQL's ecosystem for backup, PITR, replication, monitoring, and extensions (citext, pg_partman, etc.) is substantially deeper than any graph database's.

For future graph query needs, see [Future Considerations in Database Design](../architecture/database_design.md#future-considerations-graph-query-extensions).

### REL-002: Schema Organization

**All relational tables shall reside in a single PostgreSQL schema** (e.g., `public` or a dedicated `chronocone` schema). Domain-level grouping (policy, taxonomy, morphisms, timeseries, connectors) is an organizational convenience in documentation and source files, not a database-level separation.

The schema definition shall be maintained as a single SQL file (`schema.sql`) with dependency-correct table ordering and clear section comment blocks for navigability.

### REL-003: Full Bitemporal Model

**Every bitemporal table shall implement both the validity dimension and the transaction dimension as dedicated, indexed columns.**

The bitemporal columns on every versioned table are:

| Column | Type | Purpose |
|---|---|---|
| `validity_start` | `timestamptz NOT NULL` | When the fact became true in the real world (validity dimension, inclusive) |
| `validity_end` | `timestamptz NOT NULL` | When the fact ceased being true in the real world (validity dimension, exclusive) |
| `transaction_at` | `timestamptz NOT NULL DEFAULT transaction_timestamp()` | When the system recorded this row (transaction dimension) |
| `insertion_metadata` | `JSONB NOT NULL` | Unstructured audit context: who, why, how (actor, reason, operation type) |

**Separation of concerns:**
- The three `timestamptz` columns enforce the *when* of both temporal dimensions at the schema level, with proper indexing for range queries.
- The `insertion_metadata` JSONB column carries the *why/who/how* audit context (actor identity, operation reason, change description) but no longer carries temporal information.

**Decision rationale:** The v0 design stored the transaction timestamp inside `insertion_metadata` JSONB as `created_at`. This was inadequate because:
1. JSONB-extracted timestamps cannot use B-tree range indexes for efficient temporal queries
2. No schema enforcement ensures the timestamp exists in every row
3. Dual-dimension queries (validity × transaction) require mixing native columns with JSON extraction, degrading both clarity and performance

**Dual-dimension query pattern:**
```sql
-- "What did we believe entity X looked like at system-time S, for validity-time V?"
WHERE validity_start <= V AND validity_end > V
  AND transaction_at <= S
ORDER BY transaction_at DESC LIMIT 1
```

### REL-004: RDBMS Infrastructure (Bitemporal Toolkit)

**The RDBMS infrastructure shall prioritize explicit SQL control and type-safe execution.**

The implementation of the RDBMS store layer utilizes Go and the `go-jet` library for type-safe query construction. This ensures that all table and column references are verified at compile time.

The data access pattern consists of:

1. **Type-safe Query Construction**: All queries are constructed using the `go-jet` library, ensuring that table and column references are verified at compile time against the generated models.
2. **Adjacent SQL Modules**: For any service `service.go`, complex query logic resides in an adjacent `service_sql.go` file. This file contains the `go-jet` statements and any conditional logic.
3. **Go-Native Conditionals**: Instead of a custom preprocessor, standard Go `if` statements and logic are used to dynamically build `go-jet` statements (e.g., conditionally adding `.WHERE(...)` or `.INNER_JOIN(...)` clauses).
4. **Natural Parameter Handling**: `go-jet` handles parameter substitution and sanitization natively, providing protection against SQL injection without requiring a separate preprocessor layer.

**Decision rationale:** ORMs were rejected for this system because:

1. **Flush/Purge/Assert pipeline fights UPDATE semantics**: The append-only bitemporal model fundamentally does not use UPDATE for data changes. ORMs are built around UPDATE; adapting them to "flush superseded records, purge from active, and assert new splinters" requires extensive customization that obscures rather than simplifies.
2. **Complex cascades require visible SQL**: Business operations like tombstoning a taxonomy element require conditional, multi-hop cascades. These are application logic, not FK cascades. `go-jet` allows these operations to be expressed as explicit, tunable SQL that remains type-safe.
3. **Bitemporality contaminates every query**: Every relationship load, filter, and join requires temporal predicates on both dimensions. In an ORM, this is a fragile global filter. In `go-jet`, these predicates are explicit and auditable parts of the query construction.
4. **SQL quality is the competitive advantage**: The bitemporal, localizable metadata store is core to ChronoCone. Direct control via `go-jet` ensures query quality across both performance and readability dimensions while maintaining Go's type safety.


### REL-005: Application-Level Referential Integrity

**Referential integrity between bitemporal entities shall be enforced at the service API layer, not via database foreign key constraints.**

The schema shall not use `REFERENCES` / `FOREIGN KEY` constraints between bitemporal tables. Relationships (e.g., `category.ns_id` → `namespace.ns_id`, `metric_dim.dimension_id` → taxonomy elements) are logical application constraints enforced by the service APIs that mediate all database access.

**Decision rationale:**

1. **FK constraints are incompatible with bitemporal versioning**: Multiple rows share the same logical ID across versions. A traditional FK like `REFERENCES namespace(ns_id)` would require referencing a specific version, which requires GiST exclusion indices — these are **10-100x slower** than B-tree indices and do not scale.
2. **B-tree FK index overhead scales poorly**: Even without bitemporality, FK index maintenance adds write-path latency that grows with table size. Since all access is API-mediated, this overhead provides no safety beyond what the APIs already enforce.

**Exception: Strict Relational Integrity for Attribute Modeling**:
To support the "strict" attribute model, some tables use `FOREIGN KEY` constraints referencing the **physical primary keys** (`serial` / `int`) of other versioned rows. This ensures that a specific version of a dependent entity (e.g., an `element_attribute`) is strictly bound to a specific version of its definition (e.g., `category_shape`). While this introduces write-time overhead, it is justified by the mission-critical requirement for schema-driven attribute validation.

**Validation pattern — batch in-memory verification:**

Service APIs enforce referential integrity using in-memory structural validation rather than row-by-row database logic. For example:

- A batch taxonomy update pulls the entire taxonomy into a tree-oriented memory structure, validates structural integrity before and after the update in-situ, then writes the changes. This is dramatically faster than any row-by-row constraint logic the database could perform.
- Cascade impact analysis (e.g., "what metric dimensions are affected by this taxonomy change?") uses `WHERE IN ({affected_ids})` queries to fetch all dependent rows in a single round-trip, validates in memory, and writes the resulting closures/insertions in a single transaction.

**Exception**: Non-bitemporal reference tables (e.g., `namespace_type`, `namespace_family_type`) may use traditional FK constraints where appropriate, since these tables do not have versioned rows.

### REL-006: Indexing Strategy

**Every bitemporal table shall use a covering index pattern:**

```sql
UNIQUE (entity_id, validity_end) INCLUDE (validity_start, transaction_at)
```

**Query execution rationale:**

A B-tree can efficiently use only one inequality per scan. For a bitemporal query:

```sql
WHERE entity_id = {X}
  AND validity_end > {target_time}
  AND validity_start <= {target_time}
  AND transaction_at <= {system_time}    -- "what did we believe at system_time?"
  -- OR transaction_at >= {system_time}  -- "what was recorded after system_time?"
```

Both `transaction_at` directions are important query patterns:
- `transaction_at <= S` — reconstruct historical system belief ("what was the forecast 3 days ago based on what we knew then?")
- `transaction_at >= S` — find what was recorded after a point ("what actuals arrived since last review?")

The index works as follows:

1. **Equality seek** on `entity_id` — the B-tree navigates directly to the cluster of rows for this entity (typically ~20 versions over the entity's lifetime)
2. **Range scan** on `validity_end > target_time` — the one efficient inequality; within the entity's cluster, rows are ordered by `validity_end`, so this positions the scan past all historically-closed versions. The candidate set includes the current active row (`validity_end = 3000-01-01`) plus any future-dated rows (scheduled changes are a core feature of validity ranges), typically a small set
3. **Leaf-page filtering** on `validity_start` and `transaction_at` — because these are `INCLUDE` columns, the remaining predicates are evaluated directly from the index leaf pages **without heap access**

**Key insight**: The `INCLUDE` columns are for WHERE clause performance (avoiding heap fetches during predicate filtering), not primarily for SELECT clause projection. The covering index ensures the filtering step can be resolved entirely within the index structure.

Additional properties:
- **Uniqueness enforcement** — one active version per entity at the B-tree level
- **Minimal index size** — only two key columns in the B-tree internal nodes; `INCLUDE` columns appear in leaf pages only

Logical key indices follow the same `INCLUDE` pattern:

```sql
-- Scoped entity example
UNIQUE (ns_id, name, validity_end) INCLUDE (validity_start, transaction_at)

-- Relationship table example
UNIQUE (transformer_id, input_order, validity_end) INCLUDE (validity_start, transaction_at)

-- Localization table example
UNIQUE (entity_id, language, country, validity_end) INCLUDE (validity_start, transaction_at)
```

**Partitioning — future consideration:**

Time-based partitioning (e.g., via TimescaleDB hypertables) is not required initially. If the append-only growth of bitemporal tables causes performance degradation, partitioning should be evaluated at that time. TimescaleDB's `set_chunk_time_interval()` allows forward-looking chunk interval changes without downtime, though existing data retains its original chunk boundaries.

### REL-007: Immutable Names and Name Suspension

**Entity names are immutable after insertion.** Names (`CITEXT`) serve as permanent logical business keys with the same longevity as synthetic `entity_id` columns. Once an entity is created with a name, that name never changes across any version of that entity.

If the user-facing displayed name needs to change, the change is made in the associated localization table (`display_name`), not in the base table's `name` column.

This invariant means:
- For any entity_id, all rows across all versions carry the same `name` value
- Natural key lookups by name always identify the same logical entity regardless of the time queried
- The natural key and synthetic key are equivalent for entity identification — the natural key is simply more human-friendly and suitable for use outside the database

**Name suspension and reactivation:**

Tombstoning an entity (closing its `validity_end`) suspends the name but **does not free it**. The name remains permanently bound to its original `entity_id`. If a user later attempts to create an entity with a name that was previously used in the same scope:

- The system treats this as a **reactivation** of the suspended concept — a new version is inserted using the original `entity_id` (via `OVERRIDING SYSTEM VALUE`), not a new entity
- If the intent is to create a genuinely different concept, a different name must be chosen (e.g., `'demand_v2'`)

**Design rationale:** Supporting concept suspension and reactivation is more important and semantically clearer than allowing a new concept to attach to an old name. Name recycling would mean the same natural key identifies different logical entities at different points in time, breaking the equivalence between names and entity IDs.

**Enforcement**: Application-level. The service API layer must:
1. Ensure Flush/Purge/Assert pipeline always carry the original name into the new version
2. On entity creation, check whether the name was ever used in the same scope (including tombstoned entities) and treat matches as reactivations, not new creations

### REL-008: Namespace-Scoped Logical Keys

**All major entity logical keys shall be scoped to a namespace (`ns_id`).** The v0 schema contains bugs where `metric`, `transformer`, and `shifter` use globally-unique logical keys (`UNIQUE (name, validity_end)`). These must be corrected to namespace-scoped keys:

| Table | v0 (bug) | Corrected |
|---|---|---|
| `metric` | `UNIQUE (key, validity_end)` | `UNIQUE (ns_id, key, validity_end)` |
| `transformer` | `UNIQUE (key, validity_end)` | `UNIQUE (ns_id, key, validity_end)` |
| `shifter` | `UNIQUE (key, validity_end)` | `UNIQUE (ns_id, key, validity_end)` |

**`storage_config` remains globally unique** — it is not namespace-scoped.

**Design rationale:** While one might construct a cross-namespace DAG, ChronoCone intentionally will not. This is a deliberate simplification of both the logical design and computational intensity. DAGs are contained within a namespace.

**The system shall use PostgreSQL transaction-level advisory locks for concurrency control during Flush/Purge/Assert pipeline operations.** The default transaction isolation level is `READ COMMITTED`.

#### Rule 1: Always Use Transaction-Level Locks

**Use `pg_advisory_xact_lock(key bigint)` exclusively.** Never use `pg_advisory_lock()` (session-level).

- `pg_advisory_xact_lock` is automatically released at `COMMIT` or `ROLLBACK`
- `pg_advisory_lock` (session-level) persists until explicitly released or session disconnect — in a connection pool, this leaks locks into unrelated requests

#### Rule 2: Lock Key Derivation

Advisory lock keys are a single `bigint` (64-bit signed integer) composed of two parts:

```
┌──────────────┬──────────────────────────────────┐
│ 10 bits      │ 54 bits                          │
│ table type   │ entity_id                        │
└──────────────┴──────────────────────────────────┘
```

```go
func AdvisoryLockKey(tableType int16, entityID int64) int64 {
    // tableType fits in 10 bits (0-1023)
    // entityID fits in 54 bits (0 to 2^54 - 1)
    return (int64(tableType) << 54) | entityID
}
```

The `table_type` constants must be documented and stable. They prevent collisions between entity IDs in different tables (e.g., `category_id=7` vs. `metric_id=7`). Assign constants in a single registry:

```go
// Single source of truth for advisory lock types
const (
    TableTypeNamespaceFamily = 1
    TableTypeNamespace       = 2
    TableTypeCategory        = 3
    TableTypeElement         = 4
    TableTypeMetric          = 5
    TableTypeTransformer     = 6
    TableTypeShifter         = 7
    TableTypeStorageConfig   = 8
)
```

#### Rule 3: Deterministic Lock Ordering and Re-Read

**All lock acquisitions must sort lock keys in ascending numeric order before acquiring them.** This applies to both single-entity and batch operations.

**Acquire all locks in a single database round-trip** using `unnest` over a sorted array:

```sql
SELECT pg_advisory_xact_lock(lock_key)
  FROM unnest($1::bigint[]) AS lock_key
```

**After acquiring locks, always re-read the affected data.** The database state may have changed while waiting for locks — another transaction may have committed changes to the same entities between the initial read (calculate phase) and lock acquisition. Operating on stale data would produce incorrect results.

For batch operations, follow a three-phase pattern:

1. **Calculate phase** — read database state, determine all entities that will be modified. No locks are held during this phase.
2. **Lock phase** — sort the computed `advisory_lock_key` values in ascending order, open a transaction, acquire all locks in a single round-trip.
3. **Re-read and write phase** — re-read the current state of all affected entities (now guaranteed stable because locks are held), recompute changes if needed, then execute the updates and inserts.

```go
// Batch mutation example using Go-native Service Pattern
func (s *TaxonomyService) BatchUpdate(ctx context.Context, changes []Change) error {
    // 1. Calculate phase: determine affected entities
    affected := s.computeChanges(changes)
    
    // 2. Lock phase: acquire locks in deterministic order
    keys := s.deriveSortedLockKeys(affected)
    
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    // Acquire all locks in a single round-trip
    _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(k) FROM unnest($1::bigint[]) AS k", pq.Array(keys))
    
    // 3. Re-read and write phase
    current := s.reReadEntities(ctx, tx, affected)
    finalChanges := s.reCompute(current, changes)
    
    for _, entity := range finalChanges {
        s.mutateBitemporal(ctx, tx, entity)
    }
    
    return tx.Commit()
}
```

This deterministic ordering prevents deadlocks. If two concurrent operations affect overlapping entities, both attempt locks in the same order — one waits cleanly for the other.

#### Rule 4: All Write Paths Must Participate

Advisory locks are **cooperative** — they have no connection to PostgreSQL's internal row or table locks. If any write path modifies a bitemporal entity without first acquiring its advisory lock, the concurrency protocol is broken.

- Every service API that performs the Flush/Purge/Assert pipeline must acquire the advisory lock
- Direct SQL modifications (ops scripts, migrations, manual fixes) bypass the protocol — this is a known trade-off accepted for operational access
- Code reviews must verify that new write paths include advisory lock acquisition

#### Rule 5: Reads Never Lock

Read operations do not acquire advisory locks. PostgreSQL's MVCC (Multi-Version Concurrency Control) at `READ COMMITTED` isolation ensures reads see a consistent snapshot regardless of concurrent writes. Advisory locks are exclusively for serializing writes.

#### Rule 6: Use Blocking Locks

Use `pg_advisory_xact_lock` (blocking) in all normal code paths. Do not use `pg_try_advisory_xact_lock` (non-blocking / fail-fast) unless a specific use case requires it. Concurrent batch operations should wait for each other, not fail.

**Design rationale:** `SERIALIZABLE` isolation was rejected because it introduces serialization failures (`40001` errors) that require application-level retry logic and adds overhead to every transaction, including reads. Advisory locks are more targeted — they serialize only the specific entities being modified, while reads and unrelated writes proceed without contention.

### REL-010: Connection Pooling

**The application shall use a standard Go database connection pool (e.g., via `database/sql` or `pgxpool`).** External connection pooling (e.g., PgBouncer as a sidecar or shared service) shall be evaluated when the deployment target moves to Kubernetes.

For the current iteration (docker-compose with a single application container and PostgreSQL container), the built-in pool is sufficient. No external pooler is required.

### REL-011: Tenancy Model

**Each company deployment uses a single shared database.** Namespace families provide organizational isolation within that database — they are not separate databases or schemas.

A single company has multiple namespace families representing different organizational concerns (e.g., demand planning, economics, personal). A user may have access to multiple namespace families simultaneously. Even single-tenant deployments are architecturally multi-tenant at the namespace family level; single tenancy simply means lower scale expectations.

Multi-company deployments (true SaaS multi-tenancy) are not a current requirement. If eventually needed, the isolation boundary would be at the database level (separate database per company), not at the schema or row level within a shared database.

## Deferred Topics

The following relational database topics are deferred to operational/deployment stories (Kubernetes/Helm operator):

- **Performance & Scalability** — quantitative query performance targets, write throughput requirements, read replica strategy, connection pool sizing. The architectural decisions (covering indices, advisory locks, batch in-memory validation) are captured; quantitative targets require real workload data.
- **Backup & Recovery** — backup strategies, point-in-time recovery (PITR), disaster recovery, cross-region replication. These depend on the deployment infrastructure (docker-compose vs. Kubernetes vs. managed RDS).
- **Security & Encryption** — encryption at rest (TDE), encryption in transit (TLS), key management, row-level security. These are deployment-level concerns; the single-company-per-database tenancy model (REL-011) reduces the urgency of row-level security.

## Object Store Requirements

*Not Yet Written*
