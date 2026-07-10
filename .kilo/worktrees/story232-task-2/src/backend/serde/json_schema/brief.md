# JSON Schema Validation

JSON schema utilities for validating domain data and JSONB columns.

## Responsibilities

- JSON schema generation and compilation from struct definitions
- Validation of client requests and responses against schemas
- JSONB column schema validation
- Schema utilities for domain business logic to enforce data constraints

## Package Boundaries

| Package                   | Status         | Purpose                                                                                                          |
| ------------------------- | -------------- | ---------------------------------------------------------------------------------------------------------------- |
| `stdlib encoding/json`    | Required       | JSON marshaling/unmarshaling                                                                                     |
| External schema libraries | Required       | JSON Schema validation engines - replace this entry with specific libraries once those design decisions get made |
| `../serde/rdbms`          | **Prohibited** |                                                                                                                  |
| `../serde/arrow`          | **Prohibited** |                                                                                                                  |
| `../serde/mcp`            | **Prohibited** |                                                                                                                  |

## Utility Functions

- **Validation** (`validate.go`)
  - `ValidateJSON`
  - `ValidateJSONB`
  - `CompileSchema`
- **Generation** (`generate.go`)
  - `SchemaFromStruct`
  - `SchemaFromTags`

## Usage Pattern

Domain services and handlers use json_schema utilities to validate data.
Example: Domain logic processes foreign namespace assertions using json_schema to
validate that incoming statements match the asserted schema.

## When you work here

- Schemas are utilities for domain business logic, not controllers of it
- Provide clear APIs for domain to check constraints and report validation errors
- Document any schema versioning or compatibility concerns
