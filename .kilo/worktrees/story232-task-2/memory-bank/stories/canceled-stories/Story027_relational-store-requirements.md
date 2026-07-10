# Story027: Relational Store Requirements Exploration

## Goal

Analyze and define the major requirements for relational database usage within ChronoCone. This story captures the specification needed for metadata persistence and other structured data storage with appropriate consistency, performance, and reliability characteristics.

At the discretion of the person working on this story, architecture or design decisions may be made and recorded in the architecture or design documents. Additionally, one or more implementation stories may be defined to capture work that extends beyond the requirements definition itself. The exact result for this story will depend on the complexity of the requirements discovered and the amount of discussion needed to reach clarity.

## References

- [ChronoCone Concepts](../specifications/chronocone-concepts.md) - Key concepts including:
  - Data management and visibility
  - Bitemporal data model
  - Encryption requirements
- [Introduction](../specifications/requirements/introduction.md) - System overview
- [Data Stores Requirements](../specifications/requirements/data-stores.md) - Target document to populate
- [Metrics and Calculation Requirements](../specifications/requirements/metrics-and-calculation.md) - Related (metadata for metrics)
- [Story029](Story029_object-store-requirements.md) - Object store requirements (complementary data storage)
- Existing Stories: Story013 (storage interface), Story020 (ontology services)
- [Database Design: Bitemporal Model](../specifications/architecture/database_design.md) - Bitemporal "Close and Insert" pattern documentation
- [schema.sql](../../src/backend/serde/rdbms/schema.sql) - Consolidated v0 schema file

## Dependencies

None - This story has no precedents. Relational store requirements can be defined independently.

## Dependent Stories

- [Story032_metrics-and-calculation-requirements.md](Story032_metrics-and-calculation-requirements.md) - Not Started (metric calculations depend on relational storage)
- [Story030_versioning-requirements.md](Story030_versioning-requirements.md) - Not Started (versioning storage requires relational storage design)
- [Story031_metadata-requirements.md](Story031_metadata-requirements.md) - Not Started (metadata storage requires relational storage design)
- [Story029_object-store-requirements.md](Story029_object-store-requirements.md) - Not Started (object storage complements relational storage)
- [Story041_relational-schema-implementation.md](finished-stories/Story041_relational-schema-implementation.md) - Completed (schema consolidation and implementation changes)
- [go-jet](https://github.com/go-jet/jet) - Type-safe query construction 
- [Story043_temporal-enforcement.md](Story043_temporal-enforcement.md) - Not Started (cross-cutting temporal enforcement and advisory locks)
- [Story044_localization-services.md](Story044_localization-services.md) - Not Started (cross-cutting localization services)
- [Story045_policy-services.md](Story045_policy-services.md) - Not Started (policy domain: namespace families, namespaces, access control)
- [Story046_taxonomy-services.md](Story046_taxonomy-services.md) - Not Started (taxonomy domain: categories, elements, cascade analysis)
- [Story047_taxonomy-extension-services.md](Story047_taxonomy-extension-services.md) - Not Started (taxonomy extensions: *_attr tables)
- [Story048_plan-type-services.md](Story048_plan-type-services.md) - Not Started (plan types: metrics, transformers, shifters, DAG)
- [Story049_plan-storage-services.md](Story049_plan-storage-services.md) - Not Started (plan storage: ts_doc, blob coordination)
- [Story050_connector-services.md](Story050_connector-services.md) - Not Started (connectors: language, storage_config)

## Context

From [chronocone-concepts.md](../specifications/chronocone-concepts.md):

**Data Types for Relational Storage:**
- **Metadata**: Namespace, plan, metric definitions (relational, query-heavy)
- **Lineage Information**: Execution and definition lineage (graph-like, queryable)
- **Configuration**: System configuration, user preferences
- **Audit Data**: Change history, access logs

**Security Requirements:**
- Data encrypted in transit and at rest
- Keys stored securely
- Strong consistency for metadata operations

**Existing Storage Context:**
- PostgreSQL used for metadata (via Ontology Service)
- Bitemporal data model requires careful schema design
- Story020 established the Ontology Service layer

## Purpose of This Section

The data-stores.md requirements document's **relational database sections** must specify:

1. Relational database architecture (PostgreSQL focus)
2. Schema design for bitemporal data
3. Data access layer patterns
4. Performance and scalability requirements
5. Backup and recovery procedures
6. Encryption and security measures

## Major Architectural Elements to Elaborate

### 1. Relational Database Architecture
- PostgreSQL as primary metadata store
- Schema organization (namespaces, schemas, tables)
- Connection pooling and management
- Read replicas and scaling strategies
- Multi-tenancy considerations

### 2. Schema Design
- Bitemporal data representation
- Natural key vs surrogate key tradeoffs
- Indexing strategies for query performance
- Foreign key relationships and constraints
- JSON/JSONB usage for flexible attributes

### 3. Data Access Layer
- Repository pattern implementation
- ORM vs query builder vs raw SQL decisions
- Transaction management
- Connection lifecycle management
- Query optimization and caching

### 4. Performance and Scalability
- Query performance requirements
- Write throughput for metadata changes
- Read scaling with replicas
- Connection pool sizing
- Query plan optimization

### 5. Backup and Recovery
- Backup strategies for relational data
- Point-in-time recovery
- Disaster recovery procedures
- Backup verification and testing
- Cross-region replication

### 6. Security and Encryption
- Encryption at rest (TDE)
- Encryption in transit (TLS)
- Key management
- Access control at database level
- Row-level security options

## Exploration Tasks

1. **Define relational database architecture** - Done
   - PostgreSQL confirmed as primary store (REL-001)
   - Single schema organization (REL-002)
   - Connection pooling: asyncpg built-in for now (REL-010)
   - Tenancy model: single DB per company (REL-011)

2. **Specify schema design** - Done
   - Bitemporal model with transaction_at column (REL-003)
   - Covering index pattern for temporal queries (REL-006)
   - Immutable names and suspension semantics (REL-007)
   - Namespace-scoped logical keys, v0 bugs identified (REL-008)
   - Application-level referential integrity (REL-005)

3. **Elaborate data access layer** - Done
   - Type-safe query construction with go-jet (REL-004)
   - Advisory lock concurrency with deterministic ordering (REL-009)

4. **Define performance requirements** - Deferred
   - Deferred to operational/deployment stories; architectural decisions captured in REL-006, REL-009

5. **Specify backup and recovery** - Deferred
   - Deferred to Kubernetes/operator deployment story

6. **Define security requirements** - Deferred
   - Deferred to Kubernetes/operator deployment story

## Next Steps (To Be Determined)

After completing this exploration, determine:
- Can the relational store requirements be written as a single implementation story?
- Should the data access layer be a separate implementation story from schema design?
- How should the relational store and object store implementations coordinate?

## Notes

- Relational store requires strong consistency for metadata
- Bitemporal queries can be complex - consider materialized views or query optimization
- Schema migrations must be carefully planned for zero-downtime deployments
- Consider read replicas for query-heavy operations like plan browsing
- Natural keys align with API design (no internal ID exposure)
