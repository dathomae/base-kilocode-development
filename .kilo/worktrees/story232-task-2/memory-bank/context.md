# Context

## Active Tasks

- [ ] Process Correction: Align past work with .cursorrules and Kilocode standards
- [x] [Story240: Delete ODS-Related Code and Resources](stories/finished-stories/Story240_delete-ods-code-and-resources.md) - Completed (all ODS CLI code, tests, utilities, and API documentation deleted; 2544+51+480+796=3871 lines removed)
- [ ] [Story043B: Active/Historical Partitioning](stories/Story043B_active_historical_partitioning.md) - Re-opened pending 043F (043E completed; see 043I task 8)
- [ ] [Story043D: Re-implement Handlers via Service Pattern](stories/Story043D_reimplement_handlers_services_pattern.md) - Planned
- [x] [Story043E: EBM Splinter Correctness](stories/finished-stories/Story043E_ebm_splinter_correctness.md) - Completed (data-loss bug in future-record purge fixed; bitemporal-honesty invariants enforced; specs reconciled)
- [x] [Story043F: Read-Path Helper Correctness](stories/Story043F_read_path_correctness.md) - Implemented (code+vet clean; runtime test pass requires non-root user)
- [ ] [Story043G: Test-Suite Integrity](stories/Story043G_test_integrity.md) - Planned (043E done; depends on 043F)
- [x] [Story043H: Registry and Mutation Hardening](stories/Story043H_registry_and_mutation_hardening.md) - Completed (registry IDColumns/NonMetaColumns explicit; EBM composite-key FLUSH/PURGE correct; ResolveEntityID two-pass deterministic; overridingInsert regex pivot; all 7 tasks done)
- [ ] [Story043I: Documentation and Scaffold Cleanup](stories/Story043I_docs_and_scaffold_cleanup.md) - Planned (043E done; depends on 043F, 043H)
- [ ] [Story274: Fix Conversion Issues & Unify Unit/Measure Implementations](stories/Story274_fix-conversion-issues.md) - Not Started

## Current Context

- Requirements defined in: [`memory-bank/specifications/requirements/`](specifications/requirements/)
- Architecture defined in: [`memory-bank/specifications/architecture/`](specifications/architecture/)
- Design defined in: [`memory-bank/specifications/design/`](specifications/design/)
- Stories defined in: [`memory-bank/stories/`](memory-bank/stories/)

## Completed Work

- Step 1: VSCode multi-root workspace configuration
- Step 2: Go-native backend environment setup and Go-Jet integration
- Story 041: Implemented RDBMS store layer with Go-Jet, strict attribute validation, and updated architectural specifications to reflect the transition from file-based SQL templates to schema-driven persistence.
- **Modernization**: Migrated from Python-based Metrics/Ontology services to the Go-native Service Pattern (Stories 043-052).
- **Stories 020-024**: Ontology and Metrics service refactoring, TDD robustness, and persona selection.
- **Story 043**: Implemented the Bitemporal Mutation Orchestrator and Advisory Lock protocol.
- **Story 043C**: Purged the legacy `store/` repository-pattern abstraction. Migrated bitemporal infrastructure to `serde/rdbms/`, stubbed handlers with 501 status, and updated architectural documentation to reflect the Service-Logic pattern and the Flush/Purge/Assert pipeline.
- **Story 171**: Removed parser/analyzer coupling for connector grain — eliminated `ParserConnectorDefinition`, introduced `GrainDefinition` symbol table type, broke `analysis -> parser` import cycle.

## Next Steps

- [x] Refactor magic values from `services.py` (Tracked in Bug Tracker)
- [ ] Step 3: Set up React/Node frontend environment
- [ ] Step 4: Update CI/CD configuration for multi-root workspace
