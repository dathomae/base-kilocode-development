# Patterns Table of Contents

This file lists the architectural and design patterns used in the ChronoCone system.

**[backend-testing-design.md](backend-testing-design.md)**
: Defines the testing strategy across backend layers (persistence, domain, handlers). Explains why each layer tests comprehensively against embedded PostgreSQL and how handler tests serve as user documentation. Use this to understand and design tests for the backend.

**[layered-architecture.md](layered-architecture.md)**
: Defines the Presentation, Domain, and Repository layers, their responsibilities, and interactions. Use this pattern to structure the application logic and ensure separation of concerns.

**[service-organization.md](service-organization.md)**
: Defines the directory structure and file organization for backend services (`src/services`, `src/layers`). Use this pattern when creating new services or organizing code within existing services.

**[openapi-usage.md](openapi-usage.md)**
: Defines the patterns and guidelines for using OpenAPI to ensure comprehensive API documentation. Use this pattern when defining API endpoints and data models.

**[symbol-table-driven-parsing.md](symbol-table-driven-parsing.md)**
: Defines the pattern for using symbol table stacks to track parsing context instead of type switches. Use this pattern when extending the CPL compiler with new declaration types.

**[postfix-expression-parsing.md](postfix-expression-parsing.md)**
: Defines the pattern for converting infix expressions to postfix notation via ANTLR listener call order and symbol table definitions. Use this pattern when parsing expressions with operator precedence.

**[fixed-point-arithmetic.md](fixed-point-arithmetic.md)**
: Defines the pattern for fixed-point integer representation of values with specified precision, overflow detection in Go, and static range analysis for type selection. Use this pattern for financial calculations where floating-point rounding errors are unacceptable.

**[brief-template.md](brief-template.md)**
: Defines the required and optional headings for sub-package `brief.md` files. Use this when creating a brief for a new sub-package.
