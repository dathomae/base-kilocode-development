# MCP Protocol Descriptors

Model Context Protocol descriptors and plumbing for APIs.

## Responsibilities

- MCP tool/resource definitions for domain operations
- Integration with main OpenAPI specification
- Protocol support for conversational APIs and agentic clients
- MCP-specific serialization and marshaling

## Package Boundaries

_(none documented — depends on MCP protocol libraries and the OpenAPI surface.)_

## Architecture

MCP descriptors wrap the main API surface. Conversational APIs and agent interactions use
the same underlying API, enhanced with MCP metadata for machine readability.

## When you work here

- MCP descriptors should enhance, not duplicate, the OpenAPI specification
- Ensure agent clients and UI clients can both consume the API surface
- Document any MCP-specific constraints or protocols
