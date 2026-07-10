# Command-Line Tools

Entry points and CLI utilities for ChronoCone operations.

## Responsibilities

- Database tooling and schema management (`dbtool`)
- Code generation and schema generation (`generator`)
- Deployment and operational utilities for FDEs

## Package Boundaries

_(none documented — commands depend on `domain` and the standard library; specific dependencies should be documented when material.)_

## Current Commands

- `dbtool` — Database schema management, migrations, introspection
- `generator` — Code generation for schema artifacts and models

## When you work here

- Commands should be thin wrappers that delegate to domain services
- All business logic lives in the domain layer; cmd/ orchestrates and presents it
- Document command usage with clear examples for FDEs
- Error messages should diagnose issues, not require internal knowledge
- Keep command interfaces stable for customers relying on these tools
