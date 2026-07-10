# Story045: Policy Services (Service Pattern)

## Goal

Implement the `PolicyService` for managing the policy domain: namespace families, namespaces, subject gateway (access control), and intra-family cross-namespace promotion rules.

The service provides a high-level API that enforces technical constraints (like key immutability and namespace scoping) while using the Story 043 engine for bitemporal integrity.

## References

- [Data Stores Requirements](../specifications/requirements/data-stores.md) - REL-001 through REL-009
- [schema.sql](../../src/backend/serde/rdbms/schema.sql) - Database schema
- [ChronoCone Concepts](../specifications/chronocone-concepts.md) - Namespace families and access control
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Mutation Orchestrator

## Dependencies

- [Story041_relational-schema-implementation.md](finished-stories/Story041_relational-schema-implementation.md) - Completed
- [go-jet](https://github.com/go-jet/jet) - Type-safe query construction
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Not Started
- [Story044_localization-services.md](Story044_localization-services.md) - Not Started

## Dependent Stories

None currently identified.

## Tasks

1. **Structural Design & Refinement** - Not Started
   a. Create a **Class Diagram** (Mermaid) defining the `PolicyService`, its relationship to the Subject Gateway, and the Namespace/Family hierarchy.
   b. Create a **Sequence Diagram** (Mermaid) visualizing a **Cross-Namespace Promotion** operation, showing the bitemporal lock sequence.
   c. Define the internal structure of the service and any domain-specific interfaces for access control.

2. **Implement Namespace Family Operations** - Not Started
   a. Create `src/backend/domain/policy.go`.
   b. Implement `UpsertNamespaceFamily`, `GetNamespaceFamily`, and `ListNamespaceFamilies`.
   c. Enforce `customer_subject_id` scoping (REL-008).
   d. Use the Story 043 orchestrator for all mutations.
   e. Enforce **Key Immutability** (technical keys cannot be changed across versions).

3. **Implement Namespace Operations** - Not Started
   a. Implement `UpsertNamespace`, `GetNamespace`, and `ListNamespaces`.
   b. Ensure namespaces are correctly scoped to their parent `ns_fam_id`.
   c. Use the Story 043 orchestrator for all mutations.

4. **Implement Subject Gateway & Promotion Rules** - Not Started
   a. Implement subject-to-namespace-family access mapping and validation.
   b. Implement intra-family namespace promotion policies.
   c. Use Go-Jet for all access control queries.

5. **Migrate Store Logic to Service** - Not Started
   a. Identify all policy-related methods in `src/backend/store/rdbms.go`.
   b. Refactor the API layer to call `PolicyService` directly.
   c. Purge the policy "dumb" logic from the old Store implementation.

## Constraints

- All writes must use **Advisory Locks** via the Story 043 engine.
- Key immutability and suspension/reactivation rules (REL-007) must be enforced.
- The service must be the **sole entry point** for policy data.
- No ORM usage; all queries must use **Go-Jet**.
- **Domain Mapping**: Use custom domain structs with `sql` and `json` tags for scanning query results. Do not use Go-Jet generated models in the service or API layers.

## Acceptance Criteria

1. `PolicyService` handles full CRUD for namespace families and namespaces with correct bitemporal behavior.
2. Subject gateway access mapping works correctly.
3. Key immutability and suspension/reactivation are strictly enforced.
4. All operations use the mutation orchestrator with advisory locks.
5. All tests in `domain/policy_test.go` pass.
