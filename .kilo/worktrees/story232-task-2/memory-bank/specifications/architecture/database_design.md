# Database Design: Bitemporal Model

This document describes the bitemporal data model used throughout the ChronoCone system. This model is critical for maintaining a complete audit trail and historical record of all entities in the "Institutional Brain."

## Infinity Sentinel Value (Year 3000)

Rather than using PostgreSQL's native `'infinity'` value, this system uses a concrete timestamp sentinel: **3000-01-01 00:00:00 UTC**. This design choice is deliberate.

**Why not use SQL 'infinity'?** Comparisons between finite and infinite values require special branching logic (e.g., `transaction_end > some_date` must account for unbounded cases). By using a concrete sentinel value, we force this logic to be explicit and localized in code, preventing subtle bugs where comparisons accidentally treat the unbounded case as a finite boundary.

**Why year 3000?** This value was chosen because:
- It fits safely within int64 in Go (int64 nanoseconds) and all major databases (PostgreSQL, MySQL, SQL Server, Oracle, SQLite all support timestamps well beyond year 3000)
- It is far enough in the future that no practical business logic will encounter it as a real date
- It is deterministic and portable across all systems (unlike SQL syntax variations)

**Implication for code:** Any code that compares with `transaction_end` must explicitly handle the sentinel value if boundary conditions matter. This explicitness prevents bugs where code assumes a finite upper bound.

## Core Principles

### 1. Append-Only Bitemporalism
The system uses an append-only model for all major entities. Data is never destructively updated in place. Instead, we use a **"Flush, Purge, and Assert"** pipeline (see §6).

The model is **fully bitemporal** with two explicit dimensions, each expressed as a half-open interval `[start, end)`:
- **Validity dimension** (`validity_start`, `validity_end`): When the fact is true in the real world.
- **Transaction dimension** (`transaction_start`, `transaction_end`): The window during which the system believed this version of the fact to be current. A record with `transaction_end = 3000-01-01 00:00:00 UTC` (the sentinel value for "infinity") represents the system's current belief.

### 2. Identity Continuity
To maintain object continuity across multiple versions, each entity has a stable ID (e.g., `ns_id`, `category_id`). 
- When inserting a new version of an existing object, the application MUST override the identity column value (using `OVERRIDING SYSTEM VALUE` in PostgreSQL) to reuse the ID from the previous version.
- This allows foreign key-like relationships (application constraints) to remain stable even as the data changes.

### 3. Temporal Columns
All bitemporal tables include four temporal columns:
- `validity_start` (`timestamptz NOT NULL`): When this version became valid in the real world (inclusive).
- `validity_end` (`timestamptz NOT NULL`): When this version ceased to be valid in the real world (exclusive). Enforced: `validity_start < validity_end`.
- `transaction_start` (`timestamptz NOT NULL DEFAULT transaction_timestamp()`): When the system first asserted this row.
- `transaction_end` (`timestamptz NOT NULL DEFAULT 3000-01-01 00:00:00 UTC`): When the system retired this row (i.e., replaced it with a newer belief). The sentinel value `3000-01-01 00:00:00 UTC` means this row represents the system's current belief.

A version represents the system's **current belief** if `transaction_end = 3000-01-01 00:00:00 UTC`.

A version is considered **currently active** if it is the current belief AND covers the present moment:
```sql
WHERE transaction_end = '3000-01-01 00:00:00 UTC' AND validity_start <= now() AND validity_end > now()
```

To reconstruct "what did the system believe at time U?", query:
```sql
WHERE transaction_start <= U AND U < transaction_end
```

> **Deprecated Pattern (Temporal Shadowing):** The previous pattern `ORDER BY transaction_at DESC LIMIT 1` is explicitly prohibited. It is a heuristic that fails under concurrent writes and future-dated assertions. The `transaction_start`/`transaction_end` interval pair makes belief reconstruction a simple, deterministic range predicate.

### 4. Logical Identity (Keys)
Every versioned entity should have a `key` column:
- **Type**: Must be `CITEXT` for case-insensitive lookup.
- **Constraint**: Should follow the 6-64 character length standard.
- **Role**: Serves as the stable, human-friendly "Slug" or technical key for the entity.
- **Immutability**: Keys are **permanent** after insertion. They never change across versions of an entity. For any `entity_id`, all rows across all versions carry the same `key` value. To change the user-facing displayed name, update the `display_name` in the associated localization table.
- **Scope**: Most entity keys are scoped to a namespace. In the **active partition**, key uniqueness is enforced naturally because only rows with `transaction_end = 3000-01-01 00:00:00 UTC` (the infinity sentinel) exist (one row per entity per validity segment). In the **historical partition**, the GiST exclusion constraint (see §6) prevents overlapping intervals for the same entity, which transitively prevents duplicate keys within any given point in time. `storage_config` is an exception with global uniqueness.

### 5. Localization Pattern
To support multi-lingual display labels without cluttering base tables:
- **Separation**: Human-facing text (`display_name`, `tooltip`, `description`) resides in a separate `*_localization` table.
- **Linking**: The localization table links to the base table's stable numeric ID and includes its own bitemporal validity range.
- **Locale Granularity**: Localization is indexed by `language` and (optionally) `country`.
- **Independence**: This allows display text to be corrected/translated without "closing" the technical configuration or identity of the base entity.

### 6. Active/Historical Partitioning
All bitemporal tables are physically partitioned into two tiers using declarative `PARTITION BY LIST (mode)`:

- **Active Partition** (`mode = 'active'`): Contains only the system's current belief (`transaction_end = 3000-01-01 00:00:00 UTC`). Indexed with standard B-trees on `(entity_id, validity_start DESC)` for sub-millisecond point-in-time lookups. No GiST indexes.
- **Historical Partition** (`mode = 'historical'`): Contains all retired beliefs. Sub-partitioned by `HASH (entity_id)` into 32 physical sub-partitions to parallelize GiST write amplification during batch imports. Each sub-partition carries a bitemporal exclusion constraint:
  ```sql
  EXCLUDE USING gist (
      entity_id WITH =,
      tstzrange(validity_start, validity_end, '[)') WITH &&,
      tstzrange(transaction_start, transaction_end, '[)') WITH &&
  )
  ```

**Rationale:** GiST exclusion constraints are essential for preventing overlapping 2D intervals in the historical record, but they cause significant write amplification (page splits on every insert). By quarantining GiST to the historical partition and using B-trees on the active partition, we optimize the read-hot path (calculations, UI rendering) without sacrificing integrity on the write-cold path (audit, compliance).

### 7. Mutation Pipeline: Flush, Purge, and Assert
The mutation pattern for bitemporal entities is a strict 3-step pipeline executed atomically within a single database transaction. All interval algebra is computed in the Go application layer before any SQL is emitted.

1. **Flush (Insert to History):** Read the superseded active record(s) that overlap with the new assertion's effective range. Cap their `transaction_end` to `transaction_timestamp()` and `INSERT` them into the `historical` partition. This preserves the complete belief history.
2. **Purge (Delete from Active):** Execute a physical `DELETE` of the original record(s) from the `active` partition. This is the only context in which `DELETE` is permitted on bitemporal tables.
3. **Assert (Insert to Active):** **EBM** computes the disjoint interval splinters in Go from the read-back active state. EBM emits whichever SQL statements (`INSERT`, `UPDATE`, `DELETE`) most efficiently express the result, subject to the invariant that every superseded belief is preserved in the historical partition and no active row carries a `transaction_start` older than the moment that row's current belief was formed. Callers never compute splinters and may emit only `[]postgres.InsertStatement` for the new assertion they want recorded.

   The splinters EBM materializes are conceptually:
   - The **preserved prefix** (the portion of an old record before the new assertion's effective date, if any) — typically expressed as an in-place `UPDATE` of the existing active row's `validity_end` together with a refresh of its `transaction_start`.
   - The **new assertion** (the caller's intended change), inserted with `transaction_start = transaction_timestamp()` and `transaction_end = 3000-01-01 00:00:00 UTC` (the sentinel value for infinity).
   - The **preserved suffix** (the portion of an old record after the new assertion's effective end, if any), inserted similarly.

   The choice of `INSERT` vs `UPDATE` for the prefix is an EBM optimization, not a contract; what is contractual is that (a) every active row whose belief is being shortened, replaced, or removed is first preserved in `historical` with `transaction_end = transaction_timestamp()`, and (b) the active partition's view at any `(T, U)` resolves to exactly one belief.

   This is executed under advisory lock after the staleness check; see [`layered-architecture.md` §2](../patterns/layered-architecture.md).

**Rationale:** Bitemporal honesty rests on two invariants. **Invariant 1 — preservation:** every superseded belief is archived in `historical` with `transaction_end` capped at the moment the belief stopped being held; without this, the audit trail and the §3 unicity predicate are silently corrupted. **Invariant 2 — honest active state:** every active row's `transaction_start` reflects the moment the row's *current* belief (its current `validity_end`, in particular) was formed, not the moment some earlier belief about that entity was first asserted; without this, the truncated active row and its historical predecessor both satisfy the §3 predicate at any `U` in the gap, and the bitemporal model's most basic guarantee — exactly one belief per `(T, U)` point — fails. EBM is the single point in the codebase that holds both invariants, which is why it is the only code permitted to emit `UPDATE` or `DELETE` against bitemporal tables.

### 8. Read Query Dimensionality Discipline
To prevent application code from accidentally entangling the two temporal dimensions into 2D polygon queries, the data access layer enforces strict dimensionality boundaries:

**0D Point Queries (single fact retrieval):**
- **Active path:** `GetActiveFactAt(entityID, effectiveT)` — Queries only the active partition. Resolves the current belief at a specific point in valid time. Hyper-optimized via B-tree and partition pruning.
- **Historical path:** `GetHistoricalFactAt(entityID, effectiveT, knowledgeU)` — Queries the parent table. Resolves what the system believed about `effectiveT` at knowledge time `U`.

**1D Horizontal Query (As-Was Timeline):**
- `GetTimelineAsOf(entityID, knowledgeU)` — Locks the transaction axis at `U`, returns the complete validity timeline as believed at that moment. Guaranteed to return a flat, chronological, non-overlapping array.

**1D Vertical Query (Audit Trail):**
- `GetAuditTrailFor(entityID, effectiveT)` — Locks the validity axis at `T`, returns how the system's belief about that specific date evolved over time. Guaranteed to return a chronological list ordered by `transaction_start`.

**Rationale:** Bitemporal data is inherently a 2D space (validity × transaction). Exposing raw 2D queries to application logic leads to subtle bugs where overlapping intervals in both dimensions produce unexpected result sets. By constraining the DAL to 0D and 1D queries, we guarantee that every result set is a simple, flat list that application code can consume without polygon intersection logic.

## Implementation Guidance for LLMs

### Mutation Boundary: `mutation.go` Is the Sole Owner of `UPDATE` and `DELETE`

All application-level mutations to bitemporal tables **must** pass through `ExecuteBitemporalMutation` (EBM) in `serde/rdbms/mutation.go`. This is not a convention — it is **structurally enforced by the type system**:

- EBM's `builder` callback signature is `func(freshState []T, effectiveAt postgres.Expression) []postgres.InsertStatement`.
- Callers physically **cannot** return an `UPDATE` or `DELETE` statement through this interface. The Go compiler rejects it.
- The only code in the entire codebase that emits `UPDATE` (Flush) or `DELETE` (Purge) against bitemporal tables is the internal machinery of `mutation.go` itself.

**Consequences for code generation:**
- **Application code** (services, handlers) must never import or construct `postgres.UpdateStatement` or `postgres.DeleteStatement` for bitemporal tables. If you find yourself writing one, you are bypassing the mutation pipeline.
- **The builder function** receives the fresh state (post-staleness-check) and the resolved effective timestamp expression. It returns only the `INSERT` statements for the new assertion splinters. The Flush and Purge phases are handled automatically by EBM before the builder's inserts are executed.
- **`ResolveEntityID`** is the only other public function in `mutation.go`. It performs a read-only lookup and emits no mutations.
- **Schema setup code** (DDL, seed data, test fixtures) is exempt from this rule as it operates outside the bitemporal lifecycle.

### General Rules

When generating SQL or application logic for this system:
1. **`UPDATE` and `DELETE` are exclusive to `serde/rdbms/mutation.go`.** No other file in the codebase may construct these statement types for bitemporal tables. This is the single most important architectural constraint.
2. **Indexing follows the partition strategy.** Active partitions use B-tree on `(entity_id, validity_start DESC)`. Historical partitions use GiST exclusion constraints (see §6). Do not add `validity_end` to unique constraints as a workaround for multi-version coexistence — that is an obsolete pattern from the pre-partitioning era.
3. **Always use `CITEXT`** for keys and identifiers in base tables for stable keys.
4. **Display Strings are Case-Sensitive**: Use `TEXT` or `VARCHAR` for `display_name`, `tooltip`, and `description` in localization tables to preserve casing. Only `language` and `country` should use `CITEXT` in these tables.
5. **Always include `transaction_start` and `transaction_end`** on all bitemporal tables. Never use a single `transaction_at` scalar.
6. **Always include `insertion_metadata`** on all bitemporal tables for unstructured audit context (who, why, and how). Do **not** store temporal information in `insertion_metadata`; use the dedicated temporal columns instead.
7. **Type-Safe Query Construction**: Use `go-jet` for the Go backend to construct type-safe SQL queries against generated models. Complex query logic should reside in an adjacent `service_sql.go` file for any given `service.go`, using standard Go control flow for conditional SQL construction.
8. **Follow the Language Constraint**: Localization tables should include a comment referencing the global `language` table.
9. **Never use `ORDER BY transaction_at DESC LIMIT 1`** for belief reconstruction. Use the `transaction_start`/`transaction_end` interval predicate instead.
10. **Interval algebra is computed in Go, inside EBM.** EBM is free to express the result via any combination of `INSERT`, `UPDATE`, or `DELETE`, provided the bitemporal-honesty invariants are preserved. Code outside `serde/rdbms/mutation.go` may not emit `UPDATE` or `DELETE` against bitemporal tables.

## Future Considerations: Graph Query Extensions

PostgreSQL's graph query capabilities are evolving. If ChronoCone's graph traversal needs grow beyond what recursive CTEs handle efficiently (e.g., deep transitive dependency analysis across namespaces at arbitrary historical points in time), two PostgreSQL extensions should be evaluated:

### Apache AGE
- Adds openCypher query language support as a PostgreSQL extension
- Cypher pattern matching (e.g., `(a)-[:FEEDS*1..5]->(b)`) runs directly against existing relational tables
- Full ACID transaction support inherited from PostgreSQL
- Bitemporal constraints and localization patterns remain intact
- Available today as a mature, Apache-licensed extension

### SQL/PGQ (pg_graph, PostgreSQL 18+)
- Implements the SQL/PGQ (Property Graph Queries) standard from SQL:2023
- Defines vertices and edges declaratively within standard SQL
- Standards-based approach with long-term vendor support
- Enables graph pattern matching without leaving SQL
- Emerging in PostgreSQL 18; expected to mature over subsequent releases

### When to Revisit
These extensions should be evaluated if any of the following arise:
1. Recursive CTEs become a measurable performance bottleneck for DAG or lineage traversals
2. Query expressiveness for multi-hop path analysis becomes a developer productivity issue
3. New use cases require deep, variable-depth traversals not anticipated in the current schema
