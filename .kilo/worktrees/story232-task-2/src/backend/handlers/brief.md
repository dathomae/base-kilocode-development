# API Handlers

HTTP/REST translation layer. Converts incoming requests into domain service calls and serializes responses.

## Responsibilities

- HTTP request parsing and routing
- Basic syntax and presence validation (required fields, type checking; delegate schema validation to utilities)
- Delegate authentication to the security package
- Call exactly one domain service function per operation
- Format responses using OpenAPI-compatible types

## Package Boundaries

Handlers operate as a thin HTTP translation layer with strict package constraints:

| Package                                | Status         | Purpose                                                                |
| -------------------------------------- | -------------- | ---------------------------------------------------------------------- |
| `github.com/danielgtaylor/huma`        | Required       | HTTP framework and request/response handling                           |
| `../serde/json_schema`                 | Available      | Schema validation and JSON unmarshaling                                |
| `../serde/openapi`                     | Available      | OpenAPI documentation and response formatting                          |
| `../security`                          | Required       | Authentication and authorization checks                                |
| `../domain`                            | Required       | All business logic and state orchestration                             |
| `../serde/rdbms`                       | **Prohibited** | No direct database access; all queries must go through domain services |
| `../serde/rdbms/gen/postgres/public/*` | **Prohibited** | No direct table access; no manual query construction                   |

Each handler operation maps 1:1 to exactly one domain function. The domain function owns all business decisions: state discovery, mutation sequencing, EBM orchestration, and authorization logic.

## Standard Handler Pattern

Every handler operation follows this sequence:

```go
func(ctx context.Context, input *RequestType) (*ResponseType, error) {
	// 1. Validate HTTP request syntax (huma + serde/json_schema)
	// 2. Authenticate request (security package)
	// 3. Call exactly one domain function with validated input
	// 4. Format domain response for OpenAPI (serde/openapi)
	// 5. Return HTTP response
}
```

This pattern is nearly universal. Deviations require explicit justification and user approval.

## When you work here

- Handlers are thin delegators. All business logic lives in domain services.
- Parse HTTP input using huma; validate syntax/presence using utilities.
- Delegate authentication and authorization to the security package.
- Call exactly one domain function per handler operation.
- Never import `serde/rdbms`; never make database calls.
- Use `c.InfinityTime` for the unbounded-validity sentinel — never define your own `time.Date(3000, ...)`. Import the constants package as `c` (`import c "chronocone/backend/constants"`). The authoritative definition and design rationale are in `constants/constants.go`.
- Format the domain service response for HTTP and return.

## Lessons Learned

**Never expose internal database IDs in API request or response types**

API request payloads and response bodies must not require or surface internal primary keys or foreign key IDs. Use human-readable natural keys instead (e.g., `namespace_name` rather than `ns_id`). Resolve natural keys to internal IDs inside the domain layer before any database operation, and return appropriate errors if a name lookup fails. Internal IDs are an implementation detail that can change across migrations; exposing them couples consumers to the schema.

**Write handler tests as user-focused examples with clear before/after states**

Handlers are the user-facing layer of the backend. Tests here should read like documentation showing how users interact with the API. Each test should:

1. **Start with a user scenario** — "A planner creates a family" or "The user updates a budget category"
2. **Show BEFORE state clearly** — What exists in the system before the action? (empty, or pre-existing data?)
3. **Demonstrate the action** — The HTTP request the user makes
4. **Show AFTER state** — What changed? What does the response contain? What's now in the database?
5. **Use realistic names** — "acmecorp" not "testfam1"; "engineering" not "getelem1"

This style serves dual purposes: the tests verify the API works correctly (catching type mismatches, missing temporal filters, serialization errors), and they document by example how the API is meant to be used. New FDEs and customers can read these tests to understand the intended workflows.

Test structure: Each test (1) describes a user scenario in the test name, (2) sets up BEFORE state (or relies on empty database), (3) makes HTTP requests to handlers, (4) asserts both the response AND the resulting database state using embedded PostgreSQL.

See `memory-bank/specifications/patterns/backend-testing-design.md` for the complete testing strategy across all backend layers and how handler tests fit into the overall architecture.
