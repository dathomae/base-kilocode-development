# Service Organization Pattern (Go-Native)

This pattern defines the directory structure and implementation strategy for the modernized, Go-native backend domain services. It prioritizes **Unified Data Representation** and **Protocol-Driven Mutations** to ensure bitemporal integrity with minimal mapping overhead.

## Core Philosophy

1.  **Unified Representation**: Use a single domain struct for database mapping, business logic, and API serialization. Eliminate the "mapping tax" of separate Repository/Domain/DTO structs.
2.  **Service-Logic Pattern**: Behavior resides in **Domain Services** (e.g., `PolicyService`). Structs remain lightweight data containers with `sql` and `json` tags.
3.  **Reflective Persistence**: Use **Go-Jet** for type-safe query building, but leverage its reflection-based mapping to hydrate custom domain structs.
4.  **Bitemporal Orchestration**: All mutations follow the **Story 043 Protocol** (Lock → Read → Compute → Write) via a centralized orchestrator.

## Directory Structure

The Go backend is organized to maximize visibility and simplify imports. The `domain` package is typically flat, but may be subdivided into sub-packages when grouping improves cohesion.

```text
src/backend/
├── cmd/                # Entry points (main.go)
├── domain/             # Domain Services and Unified Models
│   ├── policy.go       # PolicyService and Policy structs
│   ├── taxonomy.go     # TaxonomyService and Category/Element structs
│   ├── entities/       # Entity types (optional subdirectory for grouping)
│   │   ├── category.go # Category entity
│   │   └── element.go  # Element entity
│   └── ...
├── serde/
│   └── rdbms/          # RDBMS Serialization and Mutation Orchestrator
│       ├── mutation.go # ExecuteBitemporalMutation
│       └── gen/        # Go-Jet Generated Schema
```

`domain/` MAY be subdivided into sub-packages by entity grouping. Flat is preferred when the entity count is small; group when a flat list would exceed roughly a dozen files or when cohesive entity families benefit from a shared internal vocabulary.

## Unified Model Pattern

Domain objects are hand-authored structs that serve as the sole source of truth for a domain entity.

```go
// src/backend/domain/taxonomy.go

type Element struct {
    // Field names map to RDBMS columns by Go-Jet normalization; json tags map
    // to API contracts. Use alias tags only for rare qrm lookup escapes.
    ID          int64     `json:"id"`
    Key         string    `json:"key"`
    NamespaceID int64     `json:"ns_id"`
    IsSuspended bool      `json:"is_suspended"`
    
    // Bitemporal metadata (Required for Story 043 Protocol)
    TransactionStart time.Time `json:"transaction_start"`
    TransactionEnd   time.Time `json:"transaction_end"`
    ValidityStart    time.Time `json:"validity_start"`
    ValidityEnd      time.Time `json:"validity_end"`
}
```

## Service Implementation Pattern

Services are responsible for coordinating persistence, validation, and temporal logic.

```go
type TaxonomyService struct {
    db *sql.DB
}

func (s *TaxonomyService) GetElement(ctx context.Context, nsID int64, key string) (*Element, error) {
    // Use Go-Jet for type-safe query construction
    stmt := SELECT(ElementTable.AllColumns).
            FROM(ElementTable).
            WHERE(ElementTable.NsID.EQ(Int64(nsID)).
              AND(ElementTable.ElementKey.EQ(String(key))))

    var dest Element
    // Use reflection to scan directly into the Unified Model
    err := stmt.QueryContext(ctx, s.db, &dest)
    return &dest, err
}
```

## Bitemporal Mutation Protocol (Story 043 / 043B)

All state changes must use **ExecuteBitemporalMutation (EBM)** in `serde/rdbms/mutation.go` to ensure atomic consistency across validity ranges. EBM enforces the **Flush, Purge, and Assert** pipeline:

1.  **Acquire Advisory Lock**: Lock the logical entity ID to serialize mutations.
2.  **Staleness Check**: Re-read the current state under the lock. Compare the caller's provided `TransactionStart` with the database's latest `TransactionStart`. If the DB's max `transaction_start` is **greater than** the caller's known value, the record was modified by another process. Abort with a concurrency error.
3.  **Compute New Splinters**: **EBM** calculates the prefix, new assertion, and suffix versions in Go and emits whichever SQL statements most efficiently express them. Callers provide only the new-assertion `INSERT` statements via the builder callback — the type system structurally prevents callers from emitting `UPDATE` or `DELETE`. See [`database_design.md` §7](../architecture/database_design.md) for the bitemporal-honesty invariants this step preserves.
4.  **Flush**: Archive the superseded active record(s) into the `historical` partition (cap `transaction_end` to now, INSERT to history).
5.  **Purge**: DELETE the original record(s) from the `active` partition.
6.  **Assert**: INSERT the computed splinters (preserved prefix, new assertion, preserved suffix) into the `active` partition with `transaction_start = now`, `transaction_end = 3000-01-01 00:00:00 UTC` (the infinity sentinel).
7.  **Commit Transaction**: Release lock on completion.

## JSONB and Attribute Mapping

For tables utilizing `JSONB` columns (e.g., `element_attr`, `storage_config`), the Unified Model should use nested structs rather than raw maps. 

1.  **Type Safety**: Nested structs provide compile-time safety for attribute access.
2.  **Scanner/Valuer**: These nested structs must implement `sql.Scanner` and `driver.Value` (typically via a generic `JSONB` wrapper) to allow Go-Jet to hydrate them directly from the database.
3.  **Tag Consistency**: The `json` tags on nested struct fields should match both the keys stored in the database and the keys expected by the API.

## API Serialization and JSON Tags

The Unified Model serves as the API contract. This ensures that the frontend and backend share a common vocabulary.

1.  **Snake Case**: Use `snake_case` for all JSON tags. A 1:1 correspondence between JSON tag and SQL column name is the **default bias** because it minimizes mental mapping, but Go-Jet persistence mapping is driven by type/field naming discipline plus rare `alias` tags, not by `sql` tags. The API contract may still deviate from storage names when there is a legitimate reason.
2.  **Read-Only Fields**: Use `json:"id,omitempty"` for primary keys that are generated by the database.
3.  **Internal Metadata**: Bitemporal metadata (e.g., `transaction_start`) should be included in the JSON representation to allow the **Staleness Detection** check on the "round trip."
4.  **Excluded Fields**: Use `json:"-"` for fields that are strictly internal and must never be exposed via the API (e.g., internal sequence numbers or sensitive technical flags).

## Constraints for Agents

1.  **NO DTOs**: Do not create separate "Request/Response" structs unless the API contract significantly deviates from the Domain Model.
2.  **NO GENERATED STRUCTS**: Do not use `serde/rdbms/model` structs in the service layer. Use them only for Go-Jet table/column references.
3.  **SOLE ENTRY POINT**: Every domain must have a dedicated service. Other services must call the domain service, not the database directly.
4.  **REFLECTIVE MAPPING**: Do not use `sql` tags in domain structs; Go-Jet ignores them on the read and insert paths. Use Go type and field names that normalize to the target table and columns, with `alias:"<table>[.<column>]"` only as the rare escape hatch for qrm lookup.
5.  **JSONB HYDRATION**: Map JSONB columns to nested structs. Avoid `map[string]any` to ensure type safety across the service and API layers.
6.  **API CONTRACT**: The Unified Model's `json` tags **are** the API contract. Any changes to these tags must be treated as a breaking API change.
