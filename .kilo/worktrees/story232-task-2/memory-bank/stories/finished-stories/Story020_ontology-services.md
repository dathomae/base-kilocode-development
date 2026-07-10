# Story 020: Ontology Services
**Status**: Done
**Epic**: API Layer Foundation
**Owner**: @horatiar

## Overview
Our transition toward an ontology-driven, multi-future calculation engine requires a service layer that enforces strict bitemporal constraints and architectural boundaries around our domain model (Namespaces, Metrics, etc.). 

This story establishes the **Ontology Services** API. We are building a strict abstraction layer between the database and the application logic, ensuring that database internal identities (`_ids`) are never exposed, all mutations correctly handle bitemporal validities, and API shapes are strictly governed by Data Transfer Objects (DTOs).

## References
* `memory-bank/lessons-learned.md` (Rule: "Never expose internal IDs in user-facing interfaces")
* `memory-bank/terms.md` (Definitions for DTOs and Natural Keys)
* `Story019_timeseries-parser-typescript.md` (Formatting Reference)

## Dependencies
* `src/metrics/models.py` (Pydantic schemas)
* Database schema definitions (`namespace`, `metric`, `transform`, `taxonomy`)

## Tasks
- [x] Defines strictly bounded API verbs: `fetch`, `upsert`, `expire`. Any generic mutations must fall under an `upsert`.
- [x] Remove all database IDs from service signatures; enforce the use of String "Natural Keys" (e.g., `ns_fam_name`).
- [x] Return strictly typed Golang-style tuples `(Optional[DTO], Optional[ErrorCode])` for explicit failure handling.
- [x] Abstract all execution code (Transforms and Shifters) storage. The API caller provides raw Python strings (`code_content`), and the internal service dynamically resolves the backend `storage_config` independently.
- [x] Implement DTO models (`models.py`) to enforce safe API output structures, permanently eradicating `Dict[str, Any]` returns.
- [x] Require a mandatory `AuditContext` on all mutations for robust lineage.

## Decisions
1. **Verbs**: We limit ourselves to `fetch`, `upsert`, and `expire`. There are and will be no "updates" or "creates" or "patches". 
    * `upsert` closes the past validity and opens a new row.
    * `expire` performs soft-deletes by closing the current bitemporal row.
2. **DTO Outputs**: Internal database structures like `Metric` and `NamespaceFamily` require internal foreign keys. These are **strictly prohibited** from exiting the API layer. We explicitly wrap all return payloads in DTO classes (e.g., `MetricDTO`) that resolve these IDs back into their human-readable Natural Keys for the API consumer.
3. **Ontology Scoping**: All mutations accept an `effective_timestamp`. If omitted, it defaults to the transaction's current execute time. Past timestamps return `ErrorCode.VALIDATION_FAILED` to enforce correct flow of time.
4. **Implementation Abstraction**: The `upsert_transformer` and `upsert_shifter` endpoints no longer require the caller to possess logic dictating S3 paths or Git references. The caller supplies the raw `code_content`. The internal service handles the underlying storage strategy.
5. **DRY Pipeline Testing**: To test the Bitemporal mechanics, the tests are abstracted into heavily parametrized fixtures applying the same core temporal conditions iteratively against every fetch, upsert, and expire function globally.

## Notes
* `AuditContext` enforces a minimum 5-word, 15-character `reason` to prevent lazy auditing entries.
* Taxonomy operations (`upsert_taxonomy`) process completely recursive element/category trees via Pydantic `CategoryNode`.
