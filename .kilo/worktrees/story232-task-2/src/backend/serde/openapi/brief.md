# OpenAPI Specification

OpenAPI spec generation, contract definition, and API documentation.

## Responsibilities

- OpenAPI/Swagger spec generation from domain models and handlers
- API contract definition and versioning
- Request/response schema documentation
- Code generation from OpenAPI specs (if applicable)

## Package Boundaries

| Package                | Status         | Purpose                                           |
| ---------------------- | -------------- | ------------------------------------------------- |
| `stdlib encoding/json` | Required       | JSON marshaling/unmarshaling                      |
| `../serde/json_schema` | Available      | Schema generation for request/response validation |
| `../serde/rdbms`       | **Prohibited** | No direct database access                         |
| `../serde/arrow`       | **Prohibited** |                                                   |
| `../serde/mcp`         | **Prohibited** |                                                   |

## Caller-Facing Operations

Handlers use these functions for response formatting:

- **Response Formatting** (`response.go`)
  - `FormatResponse`
  - `MarshalToOpenAPI`
  - `ApplySchema`

## When you work here

- OpenAPI specs are the source of truth for the API contract
- Keep specs synchronized with handler implementations
- Document API versioning strategy and deprecation policies
- Specs should be customer-facing; errors and examples should guide FDEs
- Reference `memory-bank/specifications/patterns/openapi-usage.md` for patterns
