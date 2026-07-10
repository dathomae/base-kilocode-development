// Package constants holds cross-cutting sentinel values shared across all backend layers.
// It imports only from the Go standard library so every layer — serde/rdbms, domain, handlers,
// security, compute — can import it without creating a circular dependency.
//
// See brief.md for the full rationale behind this package's design.
package constants

import "time"

// InfinityTime is the Go representation of the bitemporal "unbounded future" sentinel.
// Used for validity_end and transaction_end columns to represent an open-ended validity range.
//
// DESIGN CHOICE: We use a concrete timestamp (3000-01-01 00:00:00 UTC) instead of
// PostgreSQL's native 'infinity' value because comparisons between finite and infinite
// values require special branching logic. By using a concrete sentinel, we force this
// logic to be explicit and localized, preventing subtle bugs where code forgets to handle
// the unbounded case.
//
// PORTABILITY: Year 3000 was chosen because it:
//   - Fits safely within int64 in Go and all major databases (PostgreSQL, MySQL, SQL Server, Oracle, SQLite)
//   - Is far enough in the future that no practical business logic will encounter it
//   - Remains deterministic and portable across systems (unlike SQL 'infinity' syntax)
//
// AGENT NOTE: This is the single authoritative definition. Do not define your own
// time.Date(3000, ...) anywhere in the backend. Import this package with the alias c
// (import c "chronocone/backend/constants") and use c.InfinityTime at call sites.
// When parseTimestamp receives "infinity", it returns this value.
//
// Do NOT confuse with PostgreSQL's DEFAULT 'infinity' string; this is the concrete timestamp.
var InfinityTime = time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)
