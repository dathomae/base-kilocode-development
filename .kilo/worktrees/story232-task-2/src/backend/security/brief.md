# Security Layer

All gateways to external systems and security-sensitive operations.

## Responsibilities

The security directory contains:

- **External connectors provisioning** — Provisioning and obtaining connections for databases, blob storage, and any integration with systems outside
  ChronoCone control
- **Authentication** — User identity verification and session management
- **Authorization** — Access control and permission enforcement (ReBAR/Zanzibar model, TBD)
- **Workflow** — Security-sensitive workflows and gate-keeping logic

**Core principle**: Anything that touches external systems or enforces access control lives here.
This includes provisioning database connections and blob storage access—not just traditional auth code. (Note: Security obtains and configures the connections, while other packages are responsible for importing and actually using them).

## Package Boundaries

| Package                    | Status         | Purpose                                    |
| -------------------------- | -------------- | ------------------------------------------ |
| `stdlib crypto`            | Required       | Cryptographic operations                   |
| `stdlib context`           | Required       | Request context and cancellation           |
| External auth providers    | Required       | OAuth, SAML, MFA providers                 |
| Any package requiring `..` | **Prohibited** | No dependencies on other internal packages |

## Public Operations

- **Authentication** (`auth.go`)
  - `AuthenticateRequest`
  - `ExtractUserFromContext`
  - `ValidateSession`
- **Authorization** (`authz.go`)
  - `AuthorizeAction`
  - `CheckPermission`
  - `GetUserPermissions`

## Design Principles

From `memory-bank/specifications/chronocone-concepts.md`:

- Data is encrypted in transit and at rest
- Users authenticated via open standards with MFA
- Users authorized to see/modify only data required for their function
- Core data strongly protected; metadata may be more freely distributed (with authorization)
- No features that cause core data to exit ChronoCone control

## Authorization Model (TBD)

ReBAR/Zanzibar-based authorization model. Details and final design pending
`memory-bank/stories/Story025_identity-authentication-authorization-requirements.md`.

## When you work here

- Code here is security-sensitive. Changes require careful review.
- Document assumptions about encryption, key management, and data access
- External integrations (blob, database, third-party services) must be security-aware
- Error messages should not leak sensitive information
- Assume FDEs may deploy this in regulated customer environments

## References

- **Critical**: `memory-bank/specifications/chronocone-concepts.md` — Security principles and user types
- **Critical**: `memory-bank/stories/Story025_identity-authentication-authorization-requirements.md` — Requirements exploration for identity, authentication, and authorization
- **Critical**: `memory-bank/specifications/requirements/identity-authentication-authorization.md` — Target requirements doc (not yet written)
